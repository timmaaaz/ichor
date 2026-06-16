package sqldb_test

import (
	"context"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// TestGetTxExecutor_NoTx covers the only branch of GetTxExecutor with non-trivial
// logic: when no transaction has been placed on the context, it must report
// (nil, false) so callers can decide to fall back to the base pool. The positive
// branch (a WithTx-populated *sqlx.Tx surfacing as a usable sqlx.ExtContext) needs
// a real transaction and is exercised by the outbox Emit integration test, which
// writes on the ctx tx — see business/sdk/outbox.
func TestGetTxExecutor_NoTx(t *testing.T) {
	ec, ok := sqldb.GetTxExecutor(context.Background())
	if ok {
		t.Fatalf("expected ok=false when no tx on context, got true")
	}
	if ec != nil {
		t.Fatalf("expected nil executor when no tx on context, got %T", ec)
	}

	// GetTx itself must agree there is no tx (guards against a future refactor
	// where GetTxExecutor and GetTx diverge on the absent-tx contract).
	if tx, ok := sqldb.GetTx(context.Background()); ok || tx != nil {
		t.Fatalf("expected GetTx to report (nil,false) with no tx on context, got (%v,%v)", tx, ok)
	}
}
