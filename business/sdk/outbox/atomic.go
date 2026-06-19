package outbox

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// WriteAtomic runs a cascade bus's write (entity write + Emit) atomically. With no
// transaction on the context it begins one on the Writer's base pool, rebinds the bus
// to it via newWithTx, runs fn, and commits — rolling back on any error. When a caller
// transaction is already in flight it JOINS it: rebinds the bus onto it, runs fn, and
// does NOT commit (the caller owns the commit). A nil *Writer (a bus built without an
// outbox / pre-cutover) runs fn on the unmodified bus with no transaction management.
//
// newWithTx is the bus's own (*Business).NewWithTx method expression; it rebinds the
// storer to tx. The same tx is placed on ctx (by BeginOrJoin) so the bus's
// outbox.Emit rides it too — both writes share one transaction.
func WriteAtomic[B any, T any](
	ctx context.Context,
	w *Writer,
	self B,
	newWithTx func(B, sqldb.CommitRollbacker) (B, error),
	fn func(ctx context.Context, bus B) (T, error),
) (T, error) {
	var zero T

	if w == nil {
		return fn(ctx, self)
	}

	ctx, tx, owned, err := sqldb.BeginOrJoin(ctx, sqldb.NewBeginner(w.db))
	if err != nil {
		return zero, fmt.Errorf("outbox: begin-or-join: %w", err)
	}

	bus, err := newWithTx(self, tx)
	if err != nil {
		if owned {
			_ = tx.Rollback()
		}
		return zero, err
	}

	out, err := fn(ctx, bus)
	if err != nil {
		if owned {
			_ = tx.Rollback()
		}
		return zero, err
	}

	if owned {
		if err := tx.Commit(); err != nil {
			return zero, fmt.Errorf("outbox: commit: %w", err)
		}
	}
	return out, nil
}

// WriteAtomicVoid is WriteAtomic for bus methods that return only an error (Delete).
func WriteAtomicVoid[B any](
	ctx context.Context,
	w *Writer,
	self B,
	newWithTx func(B, sqldb.CommitRollbacker) (B, error),
	fn func(ctx context.Context, bus B) error,
) error {
	_, err := WriteAtomic(ctx, w, self, newWithTx,
		func(ctx context.Context, bus B) (struct{}, error) {
			return struct{}{}, fn(ctx, bus)
		})
	return err
}
