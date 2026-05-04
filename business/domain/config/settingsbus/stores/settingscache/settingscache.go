// Package settingscache wraps a settingsbus.Storer with an in-process snapshot
// cache. It exists so the steady-state read path for /v1/config/settings
// answers from memory instead of issuing a database round-trip per prefix on
// every floor scan.
//
// House style: Ichor caches a Storer by wrapping it in a sturdyc-backed
// decorator (see usercache, rolecache, permissionscache, tableaccesscache,
// userrolecache, currencycache). This package follows that house shape but
// uses a "snapshot" sub-variant that is new in this repo:
//
//	per-entity caches (e.g. usercache):  one cache key per UUID; Query and
//	                                     Count are pass-through.
//	snapshot cache (this package):       one cache key for the whole table;
//	                                     Query, Count, and QueryByKey all
//	                                     derive from the snapshot, with
//	                                     filter / order / paginate done in
//	                                     memory.
//
// The snapshot variant is appropriate when (a) the table is small and bounded
// (config.settings is admin-edited lever keys, low hundreds in practice) and
// (b) the read pattern is dominated by prefix queries that per-entity caching
// can't accelerate. Do not adopt this variant for unbounded or write-heavy
// tables — full eviction on any write becomes expensive.
//
// Strategy: cache the full resolved settings slice (settings table LEFT JOINed
// with config.scenario_setting_overrides against the active scenario) under a
// single key. Filter, order, and paginate in memory.
//
// Invalidation: Create/Update/Delete on this Storer evict the snapshot.
// NewWithTx returns a tx-bound storer that ALSO evicts on inner-store success
// (eager — see txStore comment for the rollback caveat). Scenario swaps and
// override changes do not fire delegate events today (see scenariobus.Load
// comment), so the TTL bounds staleness for those paths.
//
// See docs/arch/auth.md "SettingsCache" section for the cross-cache reference.
package settingscache

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

const (
	snapshotKey = "snapshot"
	// snapshotPage fetches the entire settings table in one shot. page.Parse
	// accepts up to 1000 rows; settings is bounded by admin-edited lever keys
	// (low hundreds in practice), so this comfortably holds the full set.
	// snapshot() emits a warning when len(result) >= snapshotCap so the cap
	// becomes observable before silent truncation turns into a correctness
	// bug in prod.
	snapshotPage = "1000"
	snapshotCap  = 1000
)

// Store implements settingsbus.Storer with an in-memory snapshot cache.
type Store struct {
	log    *logger.Logger
	storer settingsbus.Storer
	cache  *sturdyc.Client[[]settingsbus.Setting]
}

// NewStore constructs a cache-wrapping Storer. Concurrent first-callers share
// one underlying fetch via sturdyc's internal singleflight.
func NewStore(log *logger.Logger, storer settingsbus.Storer, ttl time.Duration) *Store {
	const capacity = 64
	const numShards = 2
	const evictionPercentage = 10

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[[]settingsbus.Setting](capacity, numShards, ttl, evictionPercentage),
	}
}

// NewWithTx constructs a Storer bound to the given transaction. The returned
// storer wraps the inner tx storer so Create/Update/Delete invalidate the
// snapshot cache after the inner write succeeds. See txStore for caveats.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (settingsbus.Storer, error) {
	inner, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}
	return &txStore{Storer: inner, parent: s}, nil
}

// txStore wraps a tx-bound Storer and invalidates the parent's snapshot on
// successful Create/Update/Delete. Eviction is eager — it fires when the
// inner store call returns nil error, NOT on tx commit — because
// sqldb.CommitRollbacker has no post-commit hook. If the tx rolls back, we
// have evicted unnecessarily; this is safe (one extra DB fetch on next read)
// but worth knowing. The other Storer methods (Query/Count/QueryByKey/
// NewWithTx) are inherited from the embedded inner storer.
type txStore struct {
	settingsbus.Storer
	parent *Store
}

func (t *txStore) Create(ctx context.Context, setting settingsbus.Setting) error {
	if err := t.Storer.Create(ctx, setting); err != nil {
		return err
	}
	t.parent.invalidate()
	return nil
}

func (t *txStore) Update(ctx context.Context, setting settingsbus.Setting) error {
	if err := t.Storer.Update(ctx, setting); err != nil {
		return err
	}
	t.parent.invalidate()
	return nil
}

func (t *txStore) Delete(ctx context.Context, setting settingsbus.Setting) error {
	if err := t.Storer.Delete(ctx, setting); err != nil {
		return err
	}
	t.parent.invalidate()
	return nil
}

// Create proxies to the underlying storer, then evicts the snapshot.
func (s *Store) Create(ctx context.Context, setting settingsbus.Setting) error {
	if err := s.storer.Create(ctx, setting); err != nil {
		return err
	}
	s.invalidate()
	return nil
}

// Update proxies to the underlying storer, then evicts the snapshot.
func (s *Store) Update(ctx context.Context, setting settingsbus.Setting) error {
	if err := s.storer.Update(ctx, setting); err != nil {
		return err
	}
	s.invalidate()
	return nil
}

// Delete proxies to the underlying storer, then evicts the snapshot.
func (s *Store) Delete(ctx context.Context, setting settingsbus.Setting) error {
	if err := s.storer.Delete(ctx, setting); err != nil {
		return err
	}
	s.invalidate()
	return nil
}

// Query returns settings matching filter/orderBy/page. On cache miss, the
// underlying storer is asked for the entire resolved set; subsequent calls
// derive their slice from the cached snapshot.
func (s *Store) Query(ctx context.Context, filter settingsbus.QueryFilter, orderBy order.By, p page.Page) ([]settingsbus.Setting, error) {
	snap, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	out := applyFilter(snap, filter)
	sortBy(out, orderBy)
	return paginate(out, p), nil
}

// Count returns the number of settings matching filter, served from the
// cached snapshot when available. Avoids the per-row clone that applyFilter
// does — count doesn't need to hand bytes to the caller.
func (s *Store) Count(ctx context.Context, filter settingsbus.QueryFilter) (int, error) {
	snap, err := s.snapshot(ctx)
	if err != nil {
		return 0, err
	}
	return countMatching(snap, filter), nil
}

// QueryByKey returns the setting with the given key. Resolves from the
// cached snapshot when present; falls back to the underlying storer
// otherwise so a key lookup never needlessly fetches the whole table.
// The returned Setting is cloned so callers cannot mutate cached state
// through the json.RawMessage Value field.
func (s *Store) QueryByKey(ctx context.Context, key string) (settingsbus.Setting, error) {
	if cached, exists := s.cache.Get(snapshotKey); exists {
		for _, st := range cached {
			if st.Key == key {
				return cloneSetting(st), nil
			}
		}
		return settingsbus.Setting{}, settingsbus.ErrNotFound
	}
	return s.storer.QueryByKey(ctx, key)
}

func (s *Store) invalidate() {
	s.cache.Delete(snapshotKey)
}

func (s *Store) snapshot(ctx context.Context) ([]settingsbus.Setting, error) {
	return s.cache.GetOrFetch(ctx, snapshotKey, func(ctx context.Context) ([]settingsbus.Setting, error) {
		rows, err := s.storer.Query(ctx,
			settingsbus.QueryFilter{},
			order.NewBy(settingsbus.OrderByKey, order.ASC),
			page.MustParse("1", snapshotPage))
		if err != nil {
			return nil, err
		}
		// Snapshot relies on settings being small (low hundreds). If the
		// row count ever hits the cap, the snapshot is silently truncated
		// and QueryByKey would start reporting ErrNotFound for keys past
		// the cutoff. Warn loudly so we can revisit before that becomes a
		// correctness bug in prod.
		if len(rows) >= snapshotCap {
			s.log.Warn(ctx, "settingscache.snapshot.truncated",
				"cap", snapshotCap,
				"len", len(rows),
				"msg", "snapshot at page cap; raise cap or paginate before settings grow further")
		}
		return rows, nil
	})
}

// =============================================================================
// in-memory equivalents of settingsdb's filter / order / page logic.
// Keep these in sync with stores/settingsdb/{filter,order}.go semantics.
// =============================================================================

// cloneSetting deep-copies a Setting so callers cannot mutate cached state
// through Value (json.RawMessage is []byte — a reference type). Other
// Setting fields are value types (UUID, string, time.Time) and are safe to
// share via shallow copy.
func cloneSetting(s settingsbus.Setting) settingsbus.Setting {
	if len(s.Value) > 0 {
		v := make(json.RawMessage, len(s.Value))
		copy(v, s.Value)
		s.Value = v
	}
	return s
}

// matches applies the QueryFilter predicates to a single Setting. Centralized
// so applyFilter and countMatching agree on filter semantics — keep in sync
// with stores/settingsdb/filter.go.
func matches(s settingsbus.Setting, filter settingsbus.QueryFilter) bool {
	if filter.Key != nil && s.Key != *filter.Key {
		return false
	}
	if filter.Prefix != nil && !strings.HasPrefix(s.Key, *filter.Prefix) {
		return false
	}
	return true
}

func applyFilter(in []settingsbus.Setting, filter settingsbus.QueryFilter) []settingsbus.Setting {
	if filter.Key == nil && filter.Prefix == nil {
		out := make([]settingsbus.Setting, len(in))
		for i := range in {
			out[i] = cloneSetting(in[i])
		}
		return out
	}
	out := make([]settingsbus.Setting, 0, len(in))
	for _, s := range in {
		if !matches(s, filter) {
			continue
		}
		out = append(out, cloneSetting(s))
	}
	return out
}

// countMatching returns the number of Settings that satisfy filter without
// allocating a result slice or cloning Value bytes.
func countMatching(in []settingsbus.Setting, filter settingsbus.QueryFilter) int {
	if filter.Key == nil && filter.Prefix == nil {
		return len(in)
	}
	n := 0
	for _, s := range in {
		if matches(s, filter) {
			n++
		}
	}
	return n
}

func sortBy(in []settingsbus.Setting, by order.By) {
	desc := by.Direction == order.DESC
	cmp := func(a, b settingsbus.Setting) bool {
		switch by.Field {
		case settingsbus.OrderByCreatedDate:
			return a.CreatedDate.Before(b.CreatedDate)
		case settingsbus.OrderByUpdatedDate:
			return a.UpdatedDate.Before(b.UpdatedDate)
		default:
			return a.Key < b.Key
		}
	}
	sort.SliceStable(in, func(i, j int) bool {
		if desc {
			return cmp(in[j], in[i])
		}
		return cmp(in[i], in[j])
	})
}

func paginate(in []settingsbus.Setting, p page.Page) []settingsbus.Setting {
	rows := p.RowsPerPage()
	if rows <= 0 {
		return []settingsbus.Setting{}
	}
	offset := (p.Number() - 1) * rows
	if offset >= len(in) {
		return []settingsbus.Setting{}
	}
	end := offset + rows
	if end > len(in) {
		end = len(in)
	}
	return in[offset:end]
}
