package sqldb

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

// WithTx adds a transaction to the context
func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetTx retrieves a transaction from the context
func GetTx(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx, ok
}

// WithCommitRollbacker stores a unit-of-work transaction on the context when it is
// concretely a *sqlx.Tx (the type DBBeginner.Begin returns), so ctx-tx readers such
// as outbox.Emit run on the request's transaction. A non-*sqlx.Tx (e.g. a test fake)
// is ignored. This is the bridge the app/sdk/mid transaction middleware uses so it
// can populate the tx-on-ctx carrier without importing sqlx (layer purity).
func WithCommitRollbacker(ctx context.Context, tx CommitRollbacker) context.Context {
	if sqlxTx, ok := tx.(*sqlx.Tx); ok {
		return WithTx(ctx, sqlxTx)
	}
	return ctx
}

// GetTxExecutor retrieves the originating unit-of-work transaction from the
// context as an sqlx.ExtContext, suitable for running a query on the same tx
// as the entity write (a *sqlx.Tx satisfies sqlx.ExtContext). It reports false
// when no transaction is present on the context, in which case the caller must
// decide whether to fall back to the base connection pool.
func GetTxExecutor(ctx context.Context) (sqlx.ExtContext, bool) {
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, false
	}
	return tx, true
}

// Execer interface for both DB and Tx
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
