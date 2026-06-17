package actionhandlers_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// rywRecorder is an EventDispatcher that, on each dispatched cascade event, reads back the
// entity's current committed state via the injected read func. It lets the test prove exactly
// WHAT a cascaded consumer would observe at dispatch time.
type rywRecorder struct {
	mu   sync.Mutex
	read func() string
	seen []string
}

func (r *rywRecorder) OnEntityEvent(_ context.Context, _ workflow.TriggerEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seen = append(r.seen, r.read())
	return nil
}

func (r *rywRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.seen...)
}

// Test_Cascade_ReadYourWrites_Decisive is the DESIGN §8 I3 read-your-writes proof. It replaces
// the timing-luck assertion in TestCascade_M2_LiveCascade, which concedes (cascade_m2_test.go:181)
// it passes only because worker dispatch lag exceeds the pre-commit→commit window.
//
// The outbox closes the race STRUCTURALLY: a cascade row is written inside the entity's
// transaction, and the relay drains pending rows FOR UPDATE SKIP LOCKED — so a row cannot be
// dispatched until the entity write commits. We prove that deterministically by driving
// relay.ProcessBatch synchronously around the commit boundary: no worker, no lag, no timing.
//
//	pre-commit:  the cascade row is invisible (MVCC) -> 0 dispatched. A pre-commit dispatcher
//	             (the pre-F2 detached go func on context.Background) would have fired here and
//	             read the STALE value — that is the leg that would be RED without the outbox.
//	post-commit: exactly 1 dispatched, and the cascaded reader observes the COMMITTED new value.
func Test_Cascade_ReadYourWrites_Decisive(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeRYW")
	ctx := context.Background()

	const oldDesc, newDesc = "RYW-OLD", "RYW-NEW"
	catID := mustSeedCategory(t, ctx, db, oldDesc)

	// Clear the seed's committed cascade rows so the relay only sees the row this test creates.
	_, err := db.DB.ExecContext(ctx, `TRUNCATE workflow.cascade_outbox`)
	require.NoError(t, err)

	rec := &rywRecorder{read: func() string { return categoryDescription(t, ctx, db, catID) }}
	relay := workflowtemporal.NewRelay(db.Log, db.DB, rec, workflowtemporal.RelayConfig{})

	// Update the category to newDesc AND emit its cascade row on ONE uncommitted transaction.
	tx, err := db.DB.Beginx()
	require.NoError(t, err)
	txCtx := sqldb.WithTx(ctx, tx)

	catBus, err := db.BusDomain.ProductCategory.NewWithTx(tx)
	require.NoError(t, err)
	cat, err := catBus.QueryByID(txCtx, catID)
	require.NoError(t, err)
	nd := newDesc
	_, err = catBus.Update(txCtx, cat, productcategorybus.UpdateProductCategory{Description: &nd})
	require.NoError(t, err)

	// Pre-commit: the relay cannot see the uncommitted row.
	n, err := relay.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n, "relay must NOT dispatch a cascade row whose transaction has not committed")
	require.Empty(t, rec.snapshot(), "no cascade dispatched before commit")

	require.NoError(t, tx.Commit())

	// Post-commit: the row is visible; the relay dispatches it and the cascaded reader observes
	// the COMMITTED write — read-your-writes guaranteed by the outbox, not by worker-lag timing.
	n, err = relay.ProcessBatch(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n, "relay dispatches the now-committed cascade row exactly once")
	require.Equal(t, []string{newDesc}, rec.snapshot(),
		"the cascaded consumer read the committed new value (read-your-writes)")
}
