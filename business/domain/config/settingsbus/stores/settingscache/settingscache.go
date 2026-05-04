// Package settingscache wraps a settingsbus.Storer with an in-process snapshot
// cache. It exists so the steady-state read path for /v1/config/settings
// answers from memory instead of issuing a database round-trip per prefix on
// every floor scan.
//
// Strategy: cache the full resolved settings slice (settings table LEFT JOINed
// with config.scenario_setting_overrides against the active scenario) under a
// single key. Filter, order, and paginate in memory. Settings is small and
// read-mostly, so a single snapshot covers all prefix queries cheaply.
//
// Invalidation: Create/Update/Delete on this Storer evict the snapshot.
// Scenario swaps and override changes do not fire delegate events today
// (see scenariobus.Load comment), so the TTL bounds staleness for those
// paths.
package settingscache

import (
	"context"
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
	// snapshotPage fetches the entire settings table in one shot. 1000 is the
	// page-package upper bound; settings is bounded by admin-edited lever keys
	// (low hundreds in practice), so this comfortably holds the full set.
	snapshotPage = "1000"
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

// NewWithTx constructs a Storer bound to the given transaction. The wrapped
// storer is delegated to so transactional writes do not touch the cache.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (settingsbus.Storer, error) {
	return s.storer.NewWithTx(tx)
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
// cached snapshot when available.
func (s *Store) Count(ctx context.Context, filter settingsbus.QueryFilter) (int, error) {
	snap, err := s.snapshot(ctx)
	if err != nil {
		return 0, err
	}
	return len(applyFilter(snap, filter)), nil
}

// QueryByKey returns the setting with the given key. Resolves from the
// cached snapshot when present; falls back to the underlying storer
// otherwise so a key lookup never needlessly fetches the whole table.
func (s *Store) QueryByKey(ctx context.Context, key string) (settingsbus.Setting, error) {
	if cached, exists := s.cache.Get(snapshotKey); exists {
		for _, st := range cached {
			if st.Key == key {
				return st, nil
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
		return s.storer.Query(ctx,
			settingsbus.QueryFilter{},
			order.NewBy(settingsbus.OrderByKey, order.ASC),
			page.MustParse("1", snapshotPage))
	})
}

// =============================================================================
// in-memory equivalents of settingsdb's filter / order / page logic.
// Keep these in sync with stores/settingsdb/{filter,order}.go semantics.
// =============================================================================

func applyFilter(in []settingsbus.Setting, filter settingsbus.QueryFilter) []settingsbus.Setting {
	if filter.Key == nil && filter.Prefix == nil {
		out := make([]settingsbus.Setting, len(in))
		copy(out, in)
		return out
	}
	out := make([]settingsbus.Setting, 0, len(in))
	for _, s := range in {
		if filter.Key != nil && s.Key != *filter.Key {
			continue
		}
		if filter.Prefix != nil && !strings.HasPrefix(s.Key, *filter.Prefix) {
			continue
		}
		out = append(out, s)
	}
	return out
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
