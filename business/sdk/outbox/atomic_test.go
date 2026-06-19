package outbox_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// fakeBus stands in for a cascade *Business: it holds a tx-bindable executor and
// writes a probe row, exactly as a real storer.Create would.
type fakeBus struct{ exec sqlx.ExtContext }

func (b *fakeBus) NewWithTx(tx sqldb.CommitRollbacker) (*fakeBus, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}
	return &fakeBus{exec: ec}, nil
}

func (b *fakeBus) writeProbe(ctx context.Context, id string) error {
	_, err := b.exec.ExecContext(ctx, `INSERT INTO atomic_probe (id) VALUES ($1)`, id)
	return err
}

func Test_WriteAtomic(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_WriteAtomic")
	ctx := context.Background()
	_, err := db.DB.Exec(`CREATE TABLE atomic_probe (id TEXT PRIMARY KEY)`)
	require.NoError(t, err)
	w := outbox.NewWriter(db.Log, db.DB, map[string]string{}, nil)
	count := func(id string) int {
		var n int
		require.NoError(t, db.DB.GetContext(ctx, &n, `SELECT count(*) FROM atomic_probe WHERE id=$1`, id))
		return n
	}

	t.Run("begin path commits on success", func(t *testing.T) {
		_, err := outbox.WriteAtomic(ctx, w, &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				return struct{}{}, b.writeProbe(ctx, "ok")
			})
		require.NoError(t, err)
		require.Equal(t, 1, count("ok"), "begin path must commit the in-tx write")
	})

	t.Run("begin path rolls the in-tx write back when fn errors (the emit-failure shape)", func(t *testing.T) {
		boom := errors.New("emit failed")
		_, err := outbox.WriteAtomic(ctx, w, &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				if e := b.writeProbe(ctx, "rollback"); e != nil {
					return struct{}{}, e
				}
				return struct{}{}, boom // emit fails AFTER the entity write
			})
		require.ErrorIs(t, err, boom)
		require.Equal(t, 0, count("rollback"), "begin path must roll the entity write back with the failed emit")
	})

	t.Run("join path uses the caller tx and does NOT commit (owner does)", func(t *testing.T) {
		caller, err := db.DB.Beginx()
		require.NoError(t, err)
		callerCtx := sqldb.WithTx(ctx, caller)

		_, err = outbox.WriteAtomic(callerCtx, w, &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				return struct{}{}, b.writeProbe(ctx, "joined")
			})
		require.NoError(t, err)
		require.NoError(t, caller.Rollback(), "caller still owns the tx")
		require.Equal(t, 0, count("joined"), "join path must NOT commit; the caller's rollback removed the write")
	})

	t.Run("nil writer runs fn on the unmodified bus with no tx", func(t *testing.T) {
		_, err := outbox.WriteAtomic(ctx, (*outbox.Writer)(nil), &fakeBus{exec: db.DB}, (*fakeBus).NewWithTx,
			func(ctx context.Context, b *fakeBus) (struct{}, error) {
				return struct{}{}, b.writeProbe(ctx, "nilw")
			})
		require.NoError(t, err)
		require.Equal(t, 1, count("nilw"), "nil writer ⇒ pool write, no tx management")
	})
}
