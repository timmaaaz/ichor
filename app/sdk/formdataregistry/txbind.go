package formdataregistry

import (
	"context"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// NewWithTxer is satisfied by any app (or bus) that can return a transaction-bound copy of
// itself. The type parameter A is the concrete receiver type, so TxBind preserves it without
// interface erasure: each form-capable app implements NewWithTx(tx) (*App, error).
type NewWithTxer[A any] interface {
	NewWithTx(sqldb.CommitRollbacker) (A, error)
}

// TxBind returns a transaction-bound copy of app when the context carries a transaction
// (set via sqldb.WithTx), so registry-driven writes ride the caller's tx. When no tx is present
// it returns app unchanged — pool-bound and back-compatible for non-transactional callers.
//
// formdataapp.UpsertFormData enrolls its tx on the context, so every registry Create/Update
// closure self-binds to that tx through this helper. The shared registry is never mutated: the
// tx flows as request-scoped data and a fresh tx-bound app is constructed per call, so this is
// safe under concurrent submits.
//
// The NewWithTx error is propagated, never swallowed — a failure must surface rather than
// silently fall back to the base pool, which would reintroduce the non-atomicity this fixes.
func TxBind[A NewWithTxer[A]](ctx context.Context, app A) (A, error) {
	tx, ok := sqldb.GetTx(ctx)
	if !ok {
		return app, nil
	}
	return app.NewWithTx(tx)
}
