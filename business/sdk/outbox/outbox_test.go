package outbox_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// makeData builds a delegate.Data carrying DelegateEventParams, mirroring what a
// cascade bus passes to delegate.Call / outbox.Emit.
func makeData(t *testing.T, domain, action string, entityID uuid.UUID) delegate.Data {
	t.Helper()
	raw, err := json.Marshal(workflow.DelegateEventParams{
		EntityID: entityID,
		Entity:   map[string]any{"id": entityID.String(), "status": "pending"},
	})
	require.NoError(t, err)
	return delegate.Data{Domain: domain, Action: action, RawParams: raw}
}

// countRows is a tiny scoped helper for asserting table state from the base pool.
func countRows(t *testing.T, db *dbtest.Database) int {
	t.Helper()
	var n int
	require.NoError(t, db.DB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM workflow.cascade_outbox`).Scan(&n))
	return n
}

func truncate(t *testing.T, db *dbtest.Database) {
	t.Helper()
	_, err := db.DB.ExecContext(context.Background(), `TRUNCATE workflow.cascade_outbox`)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Pure tests (no DB)
// ---------------------------------------------------------------------------

// TestNilWriterNoop pins the inert-until-cutover contract: a nil *Writer Emits
// nothing and returns nil, so the buses can carry b.outbox.Emit calls before the
// real Writer is injected (DESIGN §6).
func TestNilWriterNoop(t *testing.T) {
	t.Parallel()
	var w *outbox.Writer
	require.NoError(t, w.Emit(context.Background(), makeData(t, "order", workflow.ActionCreated, uuid.New())))
}

// TestPayloadRoundTrip pins that the delegate.Data persisted as the row payload
// re-decodes to the same event (the relay re-hydrates the TriggerEvent from it).
func TestPayloadRoundTrip(t *testing.T) {
	t.Parallel()
	data := makeData(t, "order", workflow.ActionUpdated, uuid.New())

	payload, err := json.Marshal(data)
	require.NoError(t, err)

	var got delegate.Data
	require.NoError(t, json.Unmarshal(payload, &got))

	require.Equal(t, data.Domain, got.Domain)
	require.Equal(t, data.Action, got.Action)
	require.JSONEq(t, string(data.RawParams), string(got.RawParams))
}

// ---------------------------------------------------------------------------
// Store tests (DB)
// ---------------------------------------------------------------------------

func TestStore(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_OutboxStore")
	ctx := context.Background()
	store := outbox.NewStore(db.Log)

	insert := func(o outbox.Outbox) {
		t.Helper()
		require.NoError(t, store.Insert(ctx, db.DB, o))
	}
	row := func(domain string) outbox.Outbox {
		return outbox.Outbox{
			ID: uuid.New(), Domain: domain, Action: workflow.ActionCreated,
			EventType: workflow.EventTypeOnCreate, EntityName: domain + "s",
			Payload: json.RawMessage(`{"Domain":"` + domain + `"}`),
		}
	}

	// fetchPending runs inside a tx (FOR UPDATE SKIP LOCKED) and rolls back, so the
	// locks release without consuming the rows.
	fetchPending := func(limit int) []outbox.Outbox {
		t.Helper()
		tx, err := db.DB.Beginx()
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		got, err := store.FetchPending(ctx, tx, limit)
		require.NoError(t, err)
		return got
	}

	t.Run("insert+fetch ordered by seq", func(t *testing.T) {
		truncate(t, db)
		a, b, c := row("a"), row("b"), row("c")
		insert(a)
		insert(b)
		insert(c)

		got := fetchPending(10)
		require.Len(t, got, 3)
		// Insertion order == seq order (BIGSERIAL).
		require.Equal(t, a.ID, got[0].ID)
		require.Equal(t, b.ID, got[1].ID)
		require.Equal(t, c.ID, got[2].ID)
		require.True(t, got[0].Seq < got[1].Seq && got[1].Seq < got[2].Seq)
	})

	t.Run("delete-on-publish", func(t *testing.T) {
		truncate(t, db)
		o := row("a")
		insert(o)
		require.NoError(t, store.DeletePublished(ctx, db.DB, o.ID))
		require.Empty(t, fetchPending(10))
		require.Equal(t, 0, countRows(t, db))
	})

	t.Run("mark dead is skipped by fetch", func(t *testing.T) {
		truncate(t, db)
		o := row("a")
		insert(o)
		require.NoError(t, store.MarkAttempt(ctx, db.DB, o.ID, "boom", true))
		require.Empty(t, fetchPending(10), "dead rows must not be fetched")
		require.Equal(t, 1, countRows(t, db), "dead row stays until reaped")
	})

	t.Run("reap deletes aged dead rows only", func(t *testing.T) {
		truncate(t, db)
		old, young, live := row("old"), row("young"), row("live")
		insert(old)
		insert(young)
		insert(live)
		// Make old+young dead; backdate old's created_at past the window.
		require.NoError(t, store.MarkAttempt(ctx, db.DB, old.ID, "x", true))
		require.NoError(t, store.MarkAttempt(ctx, db.DB, young.ID, "x", true))
		_, err := db.DB.ExecContext(ctx,
			`UPDATE workflow.cascade_outbox SET created_at = now() - interval '30 days' WHERE id = $1`, old.ID)
		require.NoError(t, err)

		n, err := store.Reap(ctx, db.DB, time.Now().Add(-7*24*time.Hour))
		require.NoError(t, err)
		require.Equal(t, int64(1), n, "only the aged dead row is reaped")
		require.Equal(t, 2, countRows(t, db), "young-dead + live rows remain")
	})

	t.Run("FOR UPDATE SKIP LOCKED avoids double-processing", func(t *testing.T) {
		truncate(t, db)
		insert(row("a"))

		// tx1 claims the only pending row and holds the lock.
		tx1, err := db.DB.Beginx()
		require.NoError(t, err)
		defer func() { _ = tx1.Rollback() }()
		claimed, err := store.FetchPending(ctx, tx1, 10)
		require.NoError(t, err)
		require.Len(t, claimed, 1)

		// tx2 must skip the locked row rather than block or double-claim it.
		tx2, err := db.DB.Beginx()
		require.NoError(t, err)
		defer func() { _ = tx2.Rollback() }()
		second, err := store.FetchPending(ctx, tx2, 10)
		require.NoError(t, err)
		require.Empty(t, second, "second relay must skip the row locked by the first")
	})
}

// ---------------------------------------------------------------------------
// Emit tests (DB) — F1.3 / F0.2 positive branch
// ---------------------------------------------------------------------------

func TestEmit(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDatabase(t, "Test_OutboxEmit")
	ctx := context.Background()

	const lineageJSON = `{"visited":["rule:entity"]}`
	newWriter := func() *outbox.Writer {
		return outbox.NewWriter(db.Log, db.DB,
			map[string]string{"order": "orders"},
			func(context.Context) []byte { return []byte(lineageJSON) })
	}

	t.Run("writes on the ctx transaction and persists lineage", func(t *testing.T) {
		truncate(t, db)
		w := newWriter()

		tx, err := db.DB.Beginx()
		require.NoError(t, err)
		txCtx := sqldb.WithTx(ctx, tx)

		entityID := uuid.New()
		require.NoError(t, w.Emit(txCtx, makeData(t, "order", workflow.ActionUpdated, entityID)))

		// Before commit, the row is invisible to the base pool (proving it rode the tx).
		require.Equal(t, 0, countRows(t, db), "row must not be visible on the pool before commit")

		require.NoError(t, tx.Commit())
		require.Equal(t, 1, countRows(t, db))

		var domain, eventType, entityName string
		var lineage []byte
		require.NoError(t, db.DB.QueryRowContext(ctx,
			`SELECT domain, event_type, entity_name, lineage FROM workflow.cascade_outbox`).
			Scan(&domain, &eventType, &entityName, &lineage))
		require.Equal(t, "order", domain)
		require.Equal(t, workflow.EventTypeOnUpdate, eventType)
		require.Equal(t, "orders", entityName, "entity resolved from injected domain map")
		require.JSONEq(t, lineageJSON, string(lineage))
	})

	t.Run("rolled-back tx leaves no row (atomicity)", func(t *testing.T) {
		truncate(t, db)
		w := newWriter()

		tx, err := db.DB.Beginx()
		require.NoError(t, err)
		txCtx := sqldb.WithTx(ctx, tx)
		require.NoError(t, w.Emit(txCtx, makeData(t, "order", workflow.ActionCreated, uuid.New())))
		require.NoError(t, tx.Rollback())

		require.Equal(t, 0, countRows(t, db), "rolling back the unit of work discards the outbox row")
	})

	t.Run("pool fallback writes when no tx on ctx", func(t *testing.T) {
		truncate(t, db)
		w := newWriter()
		require.NoError(t, w.Emit(ctx, makeData(t, "order", workflow.ActionCreated, uuid.New())))
		require.Equal(t, 1, countRows(t, db), "fallback still persists (degraded, non-atomic)")
	})

	t.Run("returns error when the executor fails", func(t *testing.T) {
		truncate(t, db)
		w := newWriter()

		// A committed tx is a dead executor: the INSERT fails, and Emit must
		// propagate that error so the bus can roll back its unit of work.
		tx, err := db.DB.Beginx()
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		err = w.Emit(sqldb.WithTx(ctx, tx), makeData(t, "order", workflow.ActionCreated, uuid.New()))
		require.Error(t, err, "Emit must return the executor error, not swallow it")
	})

	// DESIGN §8 poison backstop (I2): the outbox INSERT and the entity write share one tx.
	// If the outbox INSERT fails, PostgreSQL aborts the transaction and downgrades the
	// COMMIT to ROLLBACK, so the co-tx entity write can never commit on its own. Nothing
	// asserted this before F9 ("poison" had 0 hits in the test tree).
	t.Run("poison backstop: a failed outbox INSERT aborts the tx, taking the co-tx write with it", func(t *testing.T) {
		truncate(t, db)
		w := newWriter()
		store := outbox.NewStore(db.Log)

		// Pre-commit a row on the pool; its id is the collision used to fail a later outbox
		// INSERT and poison the shared tx.
		poison := outbox.Outbox{
			ID: uuid.New(), Domain: "order", Action: workflow.ActionCreated,
			EventType: workflow.EventTypeOnCreate, EntityName: "orders",
			Payload: json.RawMessage(`{"Domain":"order"}`),
		}
		require.NoError(t, store.Insert(ctx, db.DB, poison))
		require.Equal(t, 1, countRows(t, db))

		tx, err := db.DB.Beginx()
		require.NoError(t, err)
		txCtx := sqldb.WithTx(ctx, tx)

		// The co-tx write that MUST roll back if the outbox INSERT fails. It stands in for the
		// entity write, which in production shares this exact transaction with the cascade Emit.
		require.NoError(t, w.Emit(txCtx, makeData(t, "order", workflow.ActionUpdated, uuid.New())))

		// Force the outbox INSERT FAILURE on the SAME tx: a duplicate primary key violates the
		// unique constraint and poisons the transaction (every later statement errors until rollback).
		require.Error(t, store.Insert(txCtx, tx, poison), "duplicate id must fail the outbox INSERT")

		// COMMIT on a poisoned tx is downgraded to ROLLBACK by PostgreSQL. lib/pq may surface
		// that as a commit error or silently — either way the unit of work is gone, which the
		// row count below proves.
		commitErr := tx.Commit()
		t.Logf("commit on poisoned tx returned: %v", commitErr)

		// Backstop holds: the co-tx write rolled back with the failed INSERT; only the
		// pre-committed poison row survives. Entity row and (new) outbox row both absent.
		require.Equal(t, 1, countRows(t, db),
			"poison backstop: a failed outbox INSERT must abort the tx and roll back the co-tx write")
	})
}
