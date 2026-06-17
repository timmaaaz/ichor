package transferorderapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// Test_TransferOrderExecute_NoPhantomOnRollback is the DESIGN §8 "on-a-tx" trip-wire (T3) and
// the regression guard for the F9 phantom-cascade-on-rollback fix.
//
// transferorderapp.Execute opens its own transaction (BeginTxx) and builds buses via
// NewWithTx. NewWithTx swaps only the storer; before F9 the handler never enrolled that tx on
// the context (sqldb.WithTx). So when transferorderbus.Execute cascaded, outbox.Emit found no
// tx on ctx and committed the workflow.cascade_outbox row on the BASE POOL — immediately and
// independent of the handler tx. When Execute then rolls back on insufficient source stock,
// the entity write vanishes but the cascade row survives, so the relay would reliably dispatch
// a PHANTOM workflow for an operation that never happened.
//
// The test drives a real transfer order to in_transit, then executes it with NO inventory at
// the source location so it rolls back at DecrementQuantity (ErrInsufficientStock) — AFTER
// transferorderbus.Execute has already emitted transfer_orders.updated. It asserts Execute
// added ZERO net cascade_outbox rows.
//
//	RED  (pre-fix): the pool-committed emit(s) survive the rollback  -> +N rows.
//	GREEN (post-fix): Execute enrolls its tx on ctx, the emit rides   -> 0 rows.
//
// StartTest wires the outbox into every bus but starts NO relay (no TemporalClient), so rows
// accumulate and the count is stable. We snapshot AFTER the approve/claim status-walk (whose
// emits land on the pool) and measure only Execute's delta.
func Test_TransferOrderExecute_NoPhantomOnRollback(t *testing.T) {
	test := apitest.StartTest(t, "Test_TransferOrderExecuteAtomicity")

	sd, err := insertSeedData(test.DB, test.Auth)
	require.NoError(t, err, "seeding")

	ctx := context.Background()
	db := test.DB

	// Walk a pending transfer order: pending -> approved -> in_transit, so Execute proceeds
	// past its status guard and reaches the first cascade emit before the rollback trigger.
	// TestSeedTransferOrders sorts its returned slice by random UUID, so the slice order (and
	// thus a given index's status) is non-deterministic — locate a pending order by status
	// rather than assuming a position.
	var toIDStr string
	for i := range sd.TransferOrders {
		if sd.TransferOrders[i].Status == "pending" {
			toIDStr = sd.TransferOrders[i].TransferID
			break
		}
	}
	require.NotEmpty(t, toIDStr, "seed must include at least one pending transfer order")
	toID, err := uuid.Parse(toIDStr)
	require.NoError(t, err, "parsing transfer order id")
	to, err := db.BusDomain.TransferOrder.QueryByID(ctx, toID)
	require.NoError(t, err, "querying seeded transfer order")
	require.Equal(t, "pending", to.Status, "selected transfer order must be pending to walk the status machine")

	approverID := sd.Admins[0].ID
	approved, err := db.BusDomain.TransferOrder.Approve(ctx, to, approverID, "")
	require.NoError(t, err, "approve pending -> approved")
	inTransit, err := db.BusDomain.TransferOrder.Claim(ctx, approved, approverID)
	require.NoError(t, err, "claim approved -> in_transit")
	require.Equal(t, "in_transit", inTransit.Status, "transfer order must be in_transit before execute")

	// Snapshot AFTER the setup emits; we only care about Execute's delta.
	before := countCascadeOutbox(t, db)

	// Execute via HTTP so the real auth middleware supplies the userID mid.GetUserID requires.
	// No inventory exists at the source location -> DecrementQuantity returns ErrInsufficientStock
	// (FailedPrecondition -> 400) and the handler transaction rolls back.
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/v1/inventory/transfer-orders/%s/execute", toID), nil)
	r.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
	test.ServeHTTP(w, r)

	require.Equal(t, http.StatusBadRequest, w.Code,
		"execute must fail its insufficient-stock precondition and roll back; body=%s", w.Body.String())

	after := countCascadeOutbox(t, db)
	require.Equal(t, before, after,
		"phantom cascade: a rolled-back Execute left %d cascade_outbox row(s) committed on the base pool. "+
			"transferorderapp.Execute must enroll its tx on ctx (ctx = sqldb.WithTx(ctx, tx)) so the cascade "+
			"Emit rides — and rolls back with — the same transaction.", after-before)
}

func countCascadeOutbox(t *testing.T, db *dbtest.Database) int {
	t.Helper()
	var n int
	require.NoError(t, db.DB.QueryRowContext(context.Background(),
		`SELECT count(*) FROM workflow.cascade_outbox`).Scan(&n))
	return n
}
