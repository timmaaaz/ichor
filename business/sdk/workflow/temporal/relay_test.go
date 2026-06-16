package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func relayLog() *logger.Logger {
	return logger.New(io.Discard, logger.LevelError, "TEST", func(context.Context) string { return "" })
}

// fakeDispatcher is a test EventDispatcher: it records every dispatched event and
// returns errFn's verdict, letting tests drive success/failure deterministically.
type fakeDispatcher struct {
	mu     sync.Mutex
	events []workflow.TriggerEvent
	err    error
}

func (f *fakeDispatcher) OnEntityEvent(_ context.Context, e workflow.TriggerEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
	return f.err
}

func (f *fakeDispatcher) snapshot() []workflow.TriggerEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]workflow.TriggerEvent(nil), f.events...)
}

// outboxRow builds a pending outbox row carrying a delegate event for entityName.
func outboxRow(t *testing.T, entityName, action, eventType string, entityID uuid.UUID, entity, before map[string]any) outbox.Outbox {
	t.Helper()
	params := workflow.DelegateEventParams{EntityID: entityID, Entity: entity, BeforeEntity: before}
	raw, err := json.Marshal(params)
	require.NoError(t, err)
	payload, err := json.Marshal(delegate.Data{Domain: entityName, Action: action, RawParams: raw})
	require.NoError(t, err)
	return outbox.Outbox{
		ID: uuid.New(), Domain: entityName, Action: action,
		EventType: eventType, EntityName: entityName, Payload: payload,
	}
}

func countOutbox(t *testing.T, db *dbtest.Database) int {
	t.Helper()
	var n int
	require.NoError(t, db.DB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM workflow.cascade_outbox`).Scan(&n))
	return n
}

func TestRelay_DrainsInSeqOrderThenDeletes(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_RelayDrain")
	ctx := context.Background()
	store := outbox.NewStore(db.Log)

	a := outboxRow(t, "alpha", workflow.ActionCreated, workflow.EventTypeOnCreate, uuid.New(), nil, nil)
	b := outboxRow(t, "bravo", workflow.ActionCreated, workflow.EventTypeOnCreate, uuid.New(), nil, nil)
	c := outboxRow(t, "charlie", workflow.ActionCreated, workflow.EventTypeOnCreate, uuid.New(), nil, nil)
	for _, o := range []outbox.Outbox{a, b, c} {
		require.NoError(t, store.Insert(ctx, db.DB, o))
	}

	fake := &fakeDispatcher{}
	relay := NewRelay(db.Log, db.DB, fake, RelayConfig{})

	n, err := relay.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, n)

	got := fake.snapshot()
	require.Len(t, got, 3)
	require.Equal(t, "alpha", got[0].EntityName, "dispatched in seq (insertion) order")
	require.Equal(t, "bravo", got[1].EntityName)
	require.Equal(t, "charlie", got[2].EntityName)
	// EventID is the row id — the dedup key the trigger derives the workflow id from.
	require.Equal(t, a.ID, got[0].EventID)

	require.Equal(t, 0, countOutbox(t, db), "delete-on-publish removed every drained row")
}

func TestRelay_RetriesThenMarksDeadAfterMaxAttempts(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_RelayDead")
	ctx := context.Background()
	store := outbox.NewStore(db.Log)

	row := outboxRow(t, "alpha", workflow.ActionCreated, workflow.EventTypeOnCreate, uuid.New(), nil, nil)
	require.NoError(t, store.Insert(ctx, db.DB, row))

	fake := &fakeDispatcher{err: errors.New("dispatch boom")}
	relay := NewRelay(db.Log, db.DB, fake, RelayConfig{MaxAttempts: 3})

	// Each failing poll re-claims the still-pending row and bumps attempts.
	for i := 1; i <= 3; i++ {
		n, err := relay.ProcessBatch(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, n, "row still pending and re-dispatched on attempt %d", i)
	}

	var attempts int
	var dead bool
	require.NoError(t, db.DB.QueryRowContext(ctx,
		`SELECT attempts, dead FROM workflow.cascade_outbox WHERE id = $1`, row.ID).Scan(&attempts, &dead))
	require.Equal(t, 3, attempts)
	require.True(t, dead, "row dead after MaxAttempts exhausted")

	// A dead row is skipped by FetchPending — no head-of-line block.
	n, err := relay.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n, "dead row is no longer claimed")
	require.Len(t, fake.snapshot(), 3, "no further dispatch attempts on a dead row")
}

func TestRelay_ReapsAgedDeadRowsOnly(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_RelayReap")
	ctx := context.Background()
	store := outbox.NewStore(db.Log)

	old := outboxRow(t, "alpha", workflow.ActionCreated, workflow.EventTypeOnCreate, uuid.New(), nil, nil)
	young := outboxRow(t, "bravo", workflow.ActionCreated, workflow.EventTypeOnCreate, uuid.New(), nil, nil)
	require.NoError(t, store.Insert(ctx, db.DB, old))
	require.NoError(t, store.Insert(ctx, db.DB, young))
	require.NoError(t, store.MarkAttempt(ctx, db.DB, old.ID, "x", true))
	require.NoError(t, store.MarkAttempt(ctx, db.DB, young.ID, "x", true))
	_, err := db.DB.ExecContext(ctx,
		`UPDATE workflow.cascade_outbox SET created_at = now() - interval '30 days' WHERE id = $1`, old.ID)
	require.NoError(t, err)

	relay := NewRelay(db.Log, db.DB, &fakeDispatcher{}, RelayConfig{}) // default 7d window

	n, err := relay.Reap(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	require.Equal(t, 1, countOutbox(t, db), "young dead row retained within the window")
}

// TestRelay_BuildEventEnrichment is the golden check: the relay reconstructs the
// TriggerEvent from a stored row identically to the old DelegateHandler path
// (it reuses the same extractEntityData / computeFieldChanges).
func TestRelay_BuildEventEnrichment(t *testing.T) {
	t.Parallel()
	r := &Relay{log: relayLog()}
	ctx := context.Background()

	t.Run("on_create extracts id + raw data + event id", func(t *testing.T) {
		entityID := uuid.New()
		row := outboxRow(t, "orders", workflow.ActionCreated, workflow.EventTypeOnCreate, entityID,
			map[string]any{"id": entityID.String(), "status": "pending"}, nil)

		ev, ok := r.buildEvent(ctx, row)
		require.True(t, ok)
		require.Equal(t, workflow.EventTypeOnCreate, ev.EventType)
		require.Equal(t, "orders", ev.EntityName)
		require.Equal(t, entityID, ev.EntityID)
		require.Equal(t, row.ID, ev.EventID, "EventID = row id (dedup key)")
		require.Equal(t, "pending", ev.RawData["status"])
	})

	t.Run("on_update diffs before/after into field changes", func(t *testing.T) {
		entityID := uuid.New()
		row := outboxRow(t, "orders", workflow.ActionUpdated, workflow.EventTypeOnUpdate, entityID,
			map[string]any{"id": entityID.String(), "status": "approved"},
			map[string]any{"id": entityID.String(), "status": "pending"})

		ev, ok := r.buildEvent(ctx, row)
		require.True(t, ok)
		require.Contains(t, ev.FieldChanges, "status")
		require.Equal(t, "pending", ev.FieldChanges["status"].OldValue)
		require.Equal(t, "approved", ev.FieldChanges["status"].NewValue)
	})

	t.Run("corrupt payload is not dispatchable", func(t *testing.T) {
		row := outbox.Outbox{ID: uuid.New(), EventType: workflow.EventTypeOnCreate, Payload: []byte("not json")}
		_, ok := r.buildEvent(ctx, row)
		require.False(t, ok, "undecodable payload returns ok=false so the relay retires it dead")
	})
}

func TestRelay_DecodeLineage(t *testing.T) {
	t.Parallel()

	require.Empty(t, decodeLineage(nil).Visited, "nil lineage starts a fresh chain")
	require.Empty(t, decodeLineage([]byte("garbage")).Visited, "garbage degrades to empty")

	l := decodeLineage([]byte(`{"visited":["rule:entity"]}`))
	require.Equal(t, []string{"rule:entity"}, l.Visited)

	// Round-trips a real lineage built via the carrier's own API.
	src := WorkflowLineage{}.With(uuid.New(), uuid.New())
	b, err := json.Marshal(src)
	require.NoError(t, err)
	require.Equal(t, src.Visited, decodeLineage(b).Visited)
}
