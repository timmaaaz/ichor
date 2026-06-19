package sqldb

import "context"

// BeginOrJoin returns a transaction to run a unit of work on, satisfying the
// begin-or-JOIN invariant. If the context already carries a transaction (a caller
// opened one) it is returned with owned=false — the caller MUST NOT commit it; the
// owner does. Otherwise a fresh transaction is begun on bgn, placed on the returned
// context (so ctx-tx readers such as outbox.Emit ride it), and returned with
// owned=true, in which case the caller owns Commit/Rollback.
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
