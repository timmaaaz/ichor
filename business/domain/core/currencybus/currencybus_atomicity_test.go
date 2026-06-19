package currencybus_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus/stores/currencydb"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// poisonOutbox makes any INSERT into workflow.cascade_outbox fail, so a cascade
// emit raises an error. Safe because dbtest gives each test an isolated database.
func poisonOutbox(t *testing.T, db *dbtest.Database) {
	t.Helper()
	for _, s := range []string{
		`CREATE OR REPLACE FUNCTION workflow.fail_outbox() RETURNS trigger
		   AS $$ BEGIN RAISE EXCEPTION 'poisoned outbox'; END; $$ LANGUAGE plpgsql`,
		`CREATE TRIGGER poison_outbox BEFORE INSERT ON workflow.cascade_outbox
		   FOR EACH ROW EXECUTE FUNCTION workflow.fail_outbox()`,
	} {
		_, err := db.DB.Exec(s)
		require.NoError(t, err)
	}
}

func currencyCount(t *testing.T, db *dbtest.Database, code string) int {
	t.Helper()
	var n int
	require.NoError(t, db.DB.GetContext(context.Background(), &n,
		`SELECT count(*) FROM core.currencies WHERE code = $1`, code))
	return n
}

// seedUserID seeds one real user and returns its ID. core.currencies.created_by /
// updated_by both FK to core.users(id), so the entity INSERT would fail on the FK
// (before the poisoned emit) if we used a random uuid — defeating the test's point.
func seedUserID(t *testing.T, ctx context.Context, db *dbtest.Database) uuid.UUID {
	t.Helper()
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, db.BusDomain.User)
	require.NoError(t, err)
	require.Len(t, usrs, 1)
	return usrs[0].ID
}

// Test_Currencybus_Atomicity_BeginPath proves the FF#2 fix: on a simple write (no
// caller tx), a failed cascade emit rolls the entity write back with it.
func Test_Currencybus_Atomicity_BeginPath(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Currencybus_Atomicity_BeginPath")
	ctx := context.Background()

	// Seed a real user BEFORE poisoning the outbox (user.Create emits too).
	uid := seedUserID(t, ctx, db)

	bus := currencybus.NewBusiness(db.Log, db.BusDomain.Delegate,
		currencydb.NewStore(db.Log, db.DB)).WithOutbox(db.BusDomain.OutboxWriter)

	poisonOutbox(t, db)

	_, err := bus.Create(ctx, currencybus.NewCurrency{
		Code: "XTS", Name: "Atomicity Probe", Symbol: "¤", Locale: "en", DecimalPlaces: 2,
		IsActive: true, SortOrder: 1, CreatedBy: &uid,
	})
	require.Error(t, err, "Create must fail when its cascade emit fails")

	require.Equal(t, 0, currencyCount(t, db, "XTS"),
		"FF#2: the entity write must roll back WITH the failed cascade emit (atomic) — "+
			"on master it is left committed (the lost-cascade gap)")
}

// Test_Currencybus_Atomicity_NoPoolFallbackWarn proves the wrapped Create rides a tx
// (does NOT hit the emit.go pool fallback) and commits exactly one cascade row. The
// committed-row count is the primary signal; the warn assertion is belt-and-suspenders.
func Test_Currencybus_Atomicity_NoPoolFallbackWarn(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Currencybus_Atomicity_NoPoolFallbackWarn")
	ctx := context.Background()
	const warn = "outbox: no transaction on context"

	uid := seedUserID(t, ctx, db)

	// db.Log's captured buffer is not exposed on *Database, so wire a fresh
	// bytes.Buffer-backed logger + Writer (cf. outbox_test.go bufWriter) into the bus
	// to inspect the pool-fallback warn. The Writer needs the currency domain↔entity
	// mapping so Emit resolves the entity name (cosmetic; the row commits either way).
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelWarn, "TEST", func(context.Context) string { return "" })
	w := outbox.NewWriter(log, db.DB,
		map[string]string{currencybus.DomainName: currencybus.EntityName}, nil)

	bus := currencybus.NewBusiness(log, db.BusDomain.Delegate,
		currencydb.NewStore(log, db.DB)).WithOutbox(w)

	cur, err := bus.Create(ctx, currencybus.NewCurrency{
		Code: "XTT", Name: "OnTx", Symbol: "¤", Locale: "en", DecimalPlaces: 2,
		IsActive: true, SortOrder: 1, CreatedBy: &uid,
	})
	require.NoError(t, err)

	var n int
	require.NoError(t, db.DB.GetContext(ctx, &n,
		`SELECT count(*) FROM workflow.cascade_outbox WHERE domain = $1 AND action = $2`,
		currencybus.DomainName, currencybus.ActionCreated))
	require.Equal(t, 1, n, "wrapped Create must commit exactly one cascade_outbox row on its own tx")

	require.NotContains(t, buf.String(), warn,
		"wrapped Create must ride its own tx — Emit must not fall back to the base pool")
	_ = cur
}
