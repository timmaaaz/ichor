package actionhandlers_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// eventRecorder records the EventID of every dispatched cascade event, in dispatch order.
type eventRecorder struct {
	mu  sync.Mutex
	ids []uuid.UUID
}

func (r *eventRecorder) OnEntityEvent(_ context.Context, e workflow.TriggerEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ids = append(r.ids, e.EventID)
	return nil
}

func (r *eventRecorder) snapshot() []uuid.UUID {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]uuid.UUID(nil), r.ids...)
}

func truncateOutbox(t *testing.T, db *dbtest.Database) {
	t.Helper()
	_, err := db.DB.ExecContext(context.Background(), `TRUNCATE workflow.cascade_outbox`)
	require.NoError(t, err)
}

func outboxRowCount(t *testing.T, db *dbtest.Database) int {
	t.Helper()
	var n int
	require.NoError(t, db.DB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM workflow.cascade_outbox`).Scan(&n))
	return n
}

func outboxIDsBySeq(t *testing.T, db *dbtest.Database) []uuid.UUID {
	t.Helper()
	rows, err := db.DB.QueryContext(context.Background(),
		`SELECT id FROM workflow.cascade_outbox ORDER BY seq`)
	require.NoError(t, err)
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())
	return ids
}

// Test_Cascade_BusRollback_NoRowNoCascade is the DESIGN §8 I1 decisive upgrade: roll back a
// real bus write and assert BOTH legs in one test — no outbox row persisted AND no cascade
// dispatched. (Previously only composed from separate unit tests.)
func Test_Cascade_BusRollback_NoRowNoCascade(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeI1")
	ctx := context.Background()

	catID := mustSeedCategory(t, ctx, db, "I1-OLD")
	truncateOutbox(t, db)

	rec := &eventRecorder{}
	relay := workflowtemporal.NewRelay(db.Log, db.DB, rec, workflowtemporal.RelayConfig{})

	// A real bus Update on a tx the test rolls back. The Update both writes the entity and
	// emits the cascade row — on the same tx, so the rollback must discard both.
	tx, err := db.DB.Beginx()
	require.NoError(t, err)
	txCtx := sqldb.WithTx(ctx, tx)
	catBus, err := db.BusDomain.ProductCategory.NewWithTx(tx)
	require.NoError(t, err)
	cat, err := catBus.QueryByID(txCtx, catID)
	require.NoError(t, err)
	nd := "I1-NEW"
	_, err = catBus.Update(txCtx, cat, productcategorybus.UpdateProductCategory{Description: &nd})
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())

	// Leg 1: the rolled-back bus write left no outbox row.
	require.Equal(t, 0, outboxRowCount(t, db), "rolled-back bus write must leave no cascade_outbox row")

	// Leg 2: with no row, the relay dispatches no cascade.
	n, err := relay.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n, "no cascade dispatched for a rolled-back write")
	require.Empty(t, rec.snapshot())

	// And the entity itself is unchanged.
	require.Equal(t, "I1-OLD", categoryDescription(t, ctx, db, catID))
}

// Test_Cascade_SameEntityUpdates_DispatchInSeqOrder is the DESIGN §8 I6 decisive upgrade: two
// real bus updates to ONE entity end-to-end must cascade in seq order (previously only the
// relay-unit test with synthetic rows across 3 distinct domains).
func Test_Cascade_SameEntityUpdates_DispatchInSeqOrder(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeI6")
	ctx := context.Background()

	catID := mustSeedCategory(t, ctx, db, "v0")
	truncateOutbox(t, db)

	rec := &eventRecorder{}
	relay := workflowtemporal.NewRelay(db.Log, db.DB, rec, workflowtemporal.RelayConfig{})

	// Two committed updates to the SAME entity -> two outbox rows with increasing seq.
	updateCat := func(desc string) {
		cat, err := db.BusDomain.ProductCategory.QueryByID(ctx, catID)
		require.NoError(t, err)
		d := desc
		_, err = db.BusDomain.ProductCategory.Update(ctx, cat, productcategorybus.UpdateProductCategory{Description: &d})
		require.NoError(t, err)
	}
	updateCat("v1")
	updateCat("v2")

	// Capture the seq-ordered ids before the relay deletes them on publish.
	wantOrder := outboxIDsBySeq(t, db)
	require.Len(t, wantOrder, 2, "two same-entity updates must produce two cascade rows")

	n, err := relay.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, wantOrder, rec.snapshot(),
		"same-entity updates must dispatch in seq order (EventID == outbox row id)")
}
