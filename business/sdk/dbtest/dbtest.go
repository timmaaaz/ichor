// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"

	"bitbucket.org/superiortechnologies/ichor/business/domain/homebus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/homebus/stores/homedb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus/stores/countrydb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus/stores/regiondb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus/stores/productdb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus/stores/usercache"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus/stores/userdb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/vproductbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/vproductbus/stores/vproductdb"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/delegate"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/migrate"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/sqldb"
	"bitbucket.org/superiortechnologies/ichor/foundation/docker"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"bitbucket.org/superiortechnologies/ichor/foundation/otel"
	"github.com/jmoiron/sqlx"
)

// BusDomain represents all the business domain apis needed for testing.
type BusDomain struct {
	Delegate *delegate.Delegate
	Home     *homebus.Business
	Product  *productbus.Business
	User     *userbus.Business
	Country  *countrybus.Business
	Region   *regionbus.Business
	VProduct *vproductbus.Business
}

func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	delegate := delegate.New(log)
	countryBus := countrybus.NewBusiness(log, delegate, countrydb.NewStore(log, db))
	regionBus := regionbus.NewBusiness(log, delegate, regiondb.NewStore(log, db))
	userBus := userbus.NewBusiness(log, delegate, usercache.NewStore(log, userdb.NewStore(log, db), time.Hour))
	productBus := productbus.NewBusiness(log, userBus, delegate, productdb.NewStore(log, db))
	homeBus := homebus.NewBusiness(log, userBus, delegate, homedb.NewStore(log, db))
	vproductBus := vproductbus.NewBusiness(vproductdb.NewStore(log, db))

	return BusDomain{
		Delegate: delegate,
		Country:  countryBus,
		Region:   regionBus,
		Home:     homeBus,
		Product:  productBus,
		User:     userBus,
		VProduct: vproductBus,
	}
}

// =============================================================================

// Database owns state for running and shutting down tests.
type Database struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	BusDomain BusDomain
}

// NewDatabase creates a new test database inside the database that was started
// to handle testing. The database is migrated to the current version and
// a connection pool is provided with business domain packages.
func NewDatabase(t *testing.T, testName string) *Database {
	image := "postgres:16.4"
	name := "servicetest"
	port := "5432"
	dockerArgs := []string{"-e", "POSTGRES_PASSWORD=postgres"}
	appArgs := []string{"-c", "log_statement=all"}

	c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
	if err != nil {
		t.Fatalf("Starting database: %v", err)
	}

	t.Logf("Name    : %s\n", c.Name)
	t.Logf("HostPort: %s\n", c.HostPort)

	dbM, err := sqldb.Open(sqldb.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.HostPort,
		Name:       "postgres",
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqldb.StatusCheck(ctx, dbM); err != nil {
		t.Fatalf("status check database: %v", err)
	}

	// -------------------------------------------------------------------------

	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 4)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	dbName := string(b)

	t.Logf("Create Database: %s\n", dbName)
	if _, err := dbM.ExecContext(context.Background(), "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("creating database %s: %v", dbName, err)
	}

	// -------------------------------------------------------------------------

	db, err := sqldb.Open(sqldb.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.HostPort,
		Name:       dbName,
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Logf("Migrate Database: %s\n", dbName)
	if err := migrate.Migrate(ctx, db); err != nil {
		t.Logf("Logs for %s\n%s:", c.Name, docker.DumpContainerLogs(c.Name))
		t.Fatalf("Migrating error: %s", err)
	}

	t.Logf("Seed Database: %s\n", dbName)
	if err := migrate.Seed(ctx, db); err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(ctx) })

	// -------------------------------------------------------------------------

	t.Cleanup(func() {
		t.Helper()

		t.Logf("Drop Database: %s\n", dbName)
		if _, err := dbM.ExecContext(context.Background(), "DROP DATABASE "+dbName); err != nil {
			t.Fatalf("dropping database %s: %v", dbName, err)
		}

		db.Close()
		dbM.Close()

		t.Logf("******************** LOGS (%s) ********************\n\n", testName)
		t.Log(buf.String())
		t.Logf("******************** LOGS (%s) ********************\n", testName)
	})

	return &Database{
		DB:        db,
		Log:       log,
		BusDomain: newBusDomains(log, db),
	}
}

// =============================================================================

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}

// FloatPointer is a helper to get a *float64 from a float64. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func FloatPointer(f float64) *float64 {
	return &f
}

// BoolPointer is a helper to get a *bool from a bool. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func BoolPointer(b bool) *bool {
	return &b
}

// UserNamePointer is a helper to get a *Name from a string. It's in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func UserNamePointer(value string) *userbus.Name {
	name := userbus.MustParseName(value)
	return &name
}

// ProductNamePointer is a helper to get a *Name from a string. It's in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func ProductNamePointer(value string) *productbus.Name {
	name := productbus.MustParseName(value)
	return &name
}
