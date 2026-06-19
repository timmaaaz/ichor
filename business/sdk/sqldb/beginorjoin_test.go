package sqldb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

func Test_BeginOrJoin(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_BeginOrJoin")
	ctx := context.Background()

	t.Run("begins a fresh tx when ctx has none", func(t *testing.T) {
		newCtx, tx, owned, err := sqldb.BeginOrJoin(ctx, sqldb.NewBeginner(db.DB))
		require.NoError(t, err)
		require.True(t, owned, "no caller tx ⇒ this call owns the new tx")
		require.NotNil(t, tx)

		got, ok := sqldb.GetTx(newCtx)
		require.True(t, ok, "the new tx must be placed on the returned ctx")
		require.NotNil(t, got, "ctx carries the begun *sqlx.Tx so outbox.Emit can ride it")
		require.NoError(t, tx.Rollback())
	})

	t.Run("joins the caller's tx when one is on ctx, and reports not-owned", func(t *testing.T) {
		caller, err := db.DB.Beginx()
		require.NoError(t, err)
		t.Cleanup(func() { _ = caller.Rollback() })
		callerCtx := sqldb.WithTx(ctx, caller)

		newCtx, tx, owned, err := sqldb.BeginOrJoin(callerCtx, sqldb.NewBeginner(db.DB))
		require.NoError(t, err)
		require.False(t, owned, "a caller tx is in flight ⇒ JOIN, do not own/commit it")
		require.Equal(t, caller, tx, "must return the SAME caller tx, not a nested one")
		got, _ := sqldb.GetTx(newCtx)
		require.Equal(t, caller, got, "ctx still carries the caller's tx unchanged")
	})
}
