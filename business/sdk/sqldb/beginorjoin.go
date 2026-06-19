package sqldb

import "context"

// BeginOrJoin returns a transaction to run a unit of work on, satisfying the
// begin-or-JOIN invariant. If the context already carries a transaction (a caller
// opened one) it is returned with owned=false — the caller MUST NOT commit it; the
// owner does. Otherwise a fresh transaction is begun on bgn and returned with
// owned=true, in which case the caller owns Commit/Rollback. The new tx is also
// placed on the returned context via WithCommitRollbacker — but only when it is
// concretely a *sqlx.Tx (what DBBeginner.Begin yields), so ctx-tx readers such as
// outbox.Emit ride it. A Beginner whose Begin returns a non-*sqlx.Tx (e.g. a test
// fake) is not placed on ctx and those readers fall back to the pool.
//
// It never opens a nested transaction when one is already in flight: begin-or-JOIN,
// not begin-always.
func BeginOrJoin(ctx context.Context, bgn Beginner) (context.Context, CommitRollbacker, bool, error) {
	if tx, ok := GetTx(ctx); ok {
		return ctx, tx, false, nil
	}

	tx, err := bgn.Begin()
	if err != nil {
		return ctx, nil, false, err
	}

	return WithCommitRollbacker(ctx, tx), tx, true, nil
}
