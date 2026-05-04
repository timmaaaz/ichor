package settingscache_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/stores/settingscache"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// fakeStorer — in-memory Storer used to exercise the cache without a database.
// Tracks call counts and supports deterministic blocking for race tests.
// =============================================================================

type fakeStorer struct {
	mu              sync.Mutex
	settings        map[string]settingsbus.Setting
	queryCalls      atomic.Int64
	countCalls      atomic.Int64
	queryByKeyCalls atomic.Int64
	gate            chan struct{}
	gated           atomic.Int64
}

func newFakeStorer(initial ...settingsbus.Setting) *fakeStorer {
	m := make(map[string]settingsbus.Setting, len(initial))
	for _, s := range initial {
		m[s.Key] = s
	}
	return &fakeStorer{settings: m}
}

func (f *fakeStorer) NewWithTx(_ sqldb.CommitRollbacker) (settingsbus.Storer, error) {
	return f, nil
}

func (f *fakeStorer) Create(_ context.Context, s settingsbus.Setting) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.settings[s.Key]; exists {
		return settingsbus.ErrUniqueEntry
	}
	f.settings[s.Key] = s
	return nil
}

func (f *fakeStorer) Update(_ context.Context, s settingsbus.Setting) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.settings[s.Key] = s
	return nil
}

func (f *fakeStorer) Delete(_ context.Context, s settingsbus.Setting) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.settings, s.Key)
	return nil
}

func (f *fakeStorer) Query(ctx context.Context, _ settingsbus.QueryFilter, _ order.By, _ page.Page) ([]settingsbus.Setting, error) {
	f.queryCalls.Add(1)
	if f.gate != nil {
		f.gated.Add(1)
		<-f.gate
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]settingsbus.Setting, 0, len(f.settings))
	for _, s := range f.settings {
		out = append(out, s)
	}
	return out, nil
}

func (f *fakeStorer) Count(_ context.Context, _ settingsbus.QueryFilter) (int, error) {
	f.countCalls.Add(1)
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.settings), nil
}

func (f *fakeStorer) QueryByKey(_ context.Context, key string) (settingsbus.Setting, error) {
	f.queryByKeyCalls.Add(1)
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.settings[key]
	if !ok {
		return settingsbus.Setting{}, settingsbus.ErrNotFound
	}
	return s, nil
}

// =============================================================================
// helpers
// =============================================================================

func newSetting(key, value string) settingsbus.Setting {
	return settingsbus.Setting{
		Key:         key,
		Value:       json.RawMessage(value),
		Description: "test",
		CreatedDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
		UpdatedDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
	}
}

func newCache(t *testing.T, ttl time.Duration, initial ...settingsbus.Setting) (*settingscache.Store, *fakeStorer) {
	t.Helper()
	log := logger.New(io.Discard, logger.LevelInfo, "settingscache_test", func(context.Context) string { return "" })
	fake := newFakeStorer(initial...)
	store := settingscache.NewStore(log, fake, ttl)
	return store, fake
}

func mustQuery(t *testing.T, store *settingscache.Store, prefix *string) []settingsbus.Setting {
	t.Helper()
	filter := settingsbus.QueryFilter{Prefix: prefix}
	got, err := store.Query(context.Background(), filter, order.NewBy(settingsbus.OrderByKey, order.ASC), page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	return got
}

func keys(in []settingsbus.Setting) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = s.Key
	}
	return out
}

func ptr[T any](v T) *T { return &v }

// =============================================================================
// tests
// =============================================================================

func TestQuery_FirstCallFetchesSecondCallHitsCache(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute,
		newSetting("pick.scan_required", `true`),
		newSetting("receive.tolerance", `0.05`),
	)

	mustQuery(t, store, ptr("pick."))
	mustQuery(t, store, ptr("pick."))

	if got := fake.queryCalls.Load(); got != 1 {
		t.Fatalf("queryCalls = %d, want 1 (cache should serve second call)", got)
	}
}

func TestQuery_DifferentPrefixesShareSnapshot(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute,
		newSetting("pick.scan_required", `true`),
		newSetting("pick.allow_partial", `false`),
		newSetting("receive.tolerance", `0.05`),
		newSetting("receive.auto_label", `true`),
		newSetting("transfer.confirm", `true`),
	)

	pick := mustQuery(t, store, ptr("pick."))
	recv := mustQuery(t, store, ptr("receive."))
	xfer := mustQuery(t, store, ptr("transfer."))

	if len(pick) != 2 || len(recv) != 2 || len(xfer) != 1 {
		t.Fatalf("unexpected slice lengths: pick=%d receive=%d transfer=%d", len(pick), len(recv), len(xfer))
	}
	if got := fake.queryCalls.Load(); got != 1 {
		t.Fatalf("queryCalls = %d, want 1 (single snapshot serves all prefixes)", got)
	}
}

func TestQuery_PrefixFilterHonored(t *testing.T) {
	t.Parallel()

	store, _ := newCache(t, time.Minute,
		newSetting("pick.a", `1`),
		newSetting("receive.b", `2`),
		newSetting("pick.c", `3`),
	)

	got := mustQuery(t, store, ptr("pick."))
	if want := []string{"pick.a", "pick.c"}; !equalSets(keys(got), want) {
		t.Fatalf("got %v, want %v", keys(got), want)
	}
}

func TestQuery_OrderByKeyAsc(t *testing.T) {
	t.Parallel()

	store, _ := newCache(t, time.Minute,
		newSetting("pick.c", `1`),
		newSetting("pick.a", `2`),
		newSetting("pick.b", `3`),
	)

	got, err := store.Query(context.Background(), settingsbus.QueryFilter{Prefix: ptr("pick.")},
		order.NewBy(settingsbus.OrderByKey, order.ASC), page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	want := []string{"pick.a", "pick.b", "pick.c"}
	if k := keys(got); !equalOrdered(k, want) {
		t.Fatalf("got %v, want %v", k, want)
	}
}

func TestQuery_OrderByKeyDesc(t *testing.T) {
	t.Parallel()

	store, _ := newCache(t, time.Minute,
		newSetting("pick.a", `1`),
		newSetting("pick.b", `2`),
		newSetting("pick.c", `3`),
	)

	got, err := store.Query(context.Background(), settingsbus.QueryFilter{Prefix: ptr("pick.")},
		order.NewBy(settingsbus.OrderByKey, order.DESC), page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	want := []string{"pick.c", "pick.b", "pick.a"}
	if k := keys(got); !equalOrdered(k, want) {
		t.Fatalf("got %v, want %v", k, want)
	}
}

func TestQuery_Pagination(t *testing.T) {
	t.Parallel()

	store, _ := newCache(t, time.Minute,
		newSetting("a", `1`), newSetting("b", `2`), newSetting("c", `3`),
		newSetting("d", `4`), newSetting("e", `5`),
	)

	got, err := store.Query(context.Background(), settingsbus.QueryFilter{},
		order.NewBy(settingsbus.OrderByKey, order.ASC), page.MustParse("2", "2"))
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if k := keys(got); !equalOrdered(k, []string{"c", "d"}) {
		t.Fatalf("page 2 size 2 got %v, want [c d]", k)
	}
}

func TestCreate_InvalidatesCache(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute, newSetting("pick.a", `1`))

	mustQuery(t, store, ptr("pick.")) // populate
	if err := store.Create(context.Background(), newSetting("pick.b", `2`)); err != nil {
		t.Fatalf("create: %v", err)
	}
	got := mustQuery(t, store, ptr("pick.")) // expect refetch

	if fake.queryCalls.Load() != 2 {
		t.Fatalf("queryCalls = %d, want 2 (Create must invalidate)", fake.queryCalls.Load())
	}
	if !equalSets(keys(got), []string{"pick.a", "pick.b"}) {
		t.Fatalf("post-create got %v, want [pick.a pick.b]", keys(got))
	}
}

func TestUpdate_InvalidatesCache(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute, newSetting("pick.a", `1`))

	mustQuery(t, store, ptr("pick."))
	if err := store.Update(context.Background(), newSetting("pick.a", `999`)); err != nil {
		t.Fatalf("update: %v", err)
	}
	got := mustQuery(t, store, ptr("pick."))

	if fake.queryCalls.Load() != 2 {
		t.Fatalf("queryCalls = %d, want 2 (Update must invalidate)", fake.queryCalls.Load())
	}
	if string(got[0].Value) != "999" {
		t.Fatalf("post-update value = %s, want 999", got[0].Value)
	}
}

func TestDelete_InvalidatesCache(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute,
		newSetting("pick.a", `1`),
		newSetting("pick.b", `2`),
	)

	mustQuery(t, store, ptr("pick."))
	if err := store.Delete(context.Background(), newSetting("pick.a", `1`)); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got := mustQuery(t, store, ptr("pick."))

	if fake.queryCalls.Load() != 2 {
		t.Fatalf("queryCalls = %d, want 2 (Delete must invalidate)", fake.queryCalls.Load())
	}
	if !equalSets(keys(got), []string{"pick.b"}) {
		t.Fatalf("post-delete got %v, want [pick.b]", keys(got))
	}
}

func TestQueryByKey_HitsSnapshotAfterPrime(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute, newSetting("pick.a", `1`))

	// Prime the snapshot.
	mustQuery(t, store, ptr("pick."))

	got, err := store.QueryByKey(context.Background(), "pick.a")
	if err != nil {
		t.Fatalf("querybykey: %v", err)
	}
	if got.Key != "pick.a" {
		t.Fatalf("key = %s, want pick.a", got.Key)
	}
	if fake.queryByKeyCalls.Load() != 0 {
		t.Fatalf("queryByKeyCalls = %d, want 0 (must serve from snapshot)", fake.queryByKeyCalls.Load())
	}
}

func TestQueryByKey_FallsBackToStorerOnSnapshotMiss(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute, newSetting("pick.a", `1`))

	// Do not prime the snapshot — direct QueryByKey hits storer.
	if _, err := store.QueryByKey(context.Background(), "pick.a"); err != nil {
		t.Fatalf("querybykey: %v", err)
	}
	if fake.queryByKeyCalls.Load() != 1 {
		t.Fatalf("queryByKeyCalls = %d, want 1 (snapshot empty so storer must be hit)", fake.queryByKeyCalls.Load())
	}
}

func TestCount_UsesSnapshot(t *testing.T) {
	t.Parallel()

	store, fake := newCache(t, time.Minute,
		newSetting("pick.a", `1`),
		newSetting("pick.b", `2`),
		newSetting("receive.c", `3`),
	)

	n, err := store.Count(context.Background(), settingsbus.QueryFilter{Prefix: ptr("pick.")})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
	if fake.countCalls.Load() != 0 {
		t.Fatalf("countCalls = %d, want 0 (Count must serve from snapshot)", fake.countCalls.Load())
	}
}

func TestConcurrentQueries_ShareSingleFetch(t *testing.T) {
	t.Parallel()

	const N = 50
	store, fake := newCache(t, time.Minute,
		newSetting("pick.a", `1`),
		newSetting("pick.b", `2`),
	)
	fake.gate = make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(N)
	results := make([]error, N)

	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			_, results[i] = store.Query(context.Background(),
				settingsbus.QueryFilter{Prefix: ptr("pick.")},
				order.NewBy(settingsbus.OrderByKey, order.ASC),
				page.MustParse("1", "100"))
		}(i)
	}

	// Wait until at least one goroutine entered the gated fetch.
	deadline := time.Now().Add(2 * time.Second)
	for fake.gated.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}
	if fake.gated.Load() == 0 {
		t.Fatal("no goroutine reached the storer.Query gate within 2s")
	}

	close(fake.gate)
	wg.Wait()

	for i, err := range results {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
	// Exactly one fetch should have been performed despite N concurrent callers.
	if got := fake.queryCalls.Load(); got != 1 {
		t.Fatalf("queryCalls = %d, want 1 (singleflight: %d concurrent first-callers must share one fetch)", got, N)
	}
}

func TestStorerInterfaceSatisfied(t *testing.T) {
	t.Parallel()
	var _ settingsbus.Storer = (*settingscache.Store)(nil)
}

func TestNewWithTx_Passthrough(t *testing.T) {
	t.Parallel()

	store, _ := newCache(t, time.Minute)
	got, err := store.NewWithTx(nil)
	if err != nil {
		t.Fatalf("newwithtx: %v", err)
	}
	if got == nil {
		t.Fatal("newwithtx returned nil storer")
	}
}

// Sanity check: ErrUniqueEntry from underlying storer is propagated by Create.
func TestCreate_PropagatesStorerError(t *testing.T) {
	t.Parallel()

	store, _ := newCache(t, time.Minute, newSetting("pick.a", `1`))
	err := store.Create(context.Background(), newSetting("pick.a", `2`))
	if !errors.Is(err, settingsbus.ErrUniqueEntry) {
		t.Fatalf("err = %v, want ErrUniqueEntry", err)
	}
}

// =============================================================================
// tiny set/order helpers
// =============================================================================

func equalSets(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	m := make(map[string]int, len(want))
	for _, k := range want {
		m[k]++
	}
	for _, k := range got {
		m[k]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

func equalOrdered(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

