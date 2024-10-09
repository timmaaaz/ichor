// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus/stores/assetconditiondb"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assettypebus/stores/assettypedb"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/homebus/stores/homedb"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	citydb "github.com/timmaaaz/ichor/business/domain/location/citybus/stores/citydb"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus/stores/countrydb"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus/stores/regiondb"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	streetdb "github.com/timmaaaz/ichor/business/domain/location/streetbus/stores/streetdb"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/domain/userbus"
	"github.com/timmaaaz/ichor/business/domain/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/domain/vproductbus"
	"github.com/timmaaaz/ichor/business/domain/vproductbus/stores/vproductdb"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/migrate"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/docker"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// BusDomain represents all the business domain apis needed for testing.
type BusDomain struct {
	Delegate       *delegate.Delegate
	Home           *homebus.Business
	AssetType      *assettypebus.Business
	AssetCondition *assetconditionbus.Business
	Asset          *assetbus.Business
	Product        *productbus.Business
	User           *userbus.Business
	Country        *countrybus.Business
	Region         *regionbus.Business
	City           *citybus.Business
	Street         *streetbus.Business
	VProduct       *vproductbus.Business
}

func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	delegate := delegate.New(log)
	countryBus := countrybus.NewBusiness(log, delegate, countrydb.NewStore(log, db))
	regionBus := regionbus.NewBusiness(log, delegate, regiondb.NewStore(log, db))
	cityBus := citybus.NewBusiness(log, delegate, citydb.NewStore(log, db))
	streetBus := streetbus.NewBusiness(log, delegate, streetdb.NewStore(log, db))

	assetTypeBus := assettypebus.NewBusiness(log, delegate, assettypedb.NewStore(log, db))
	assetConditionBus := assetconditionbus.NewBusiness(log, delegate, assetconditiondb.NewStore(log, db))
	assetBus := assetbus.NewBusiness(log, delegate, assetdb.NewStore(log, db))

	userBus := userbus.NewBusiness(log, delegate, usercache.NewStore(log, userdb.NewStore(log, db), time.Hour))
	productBus := productbus.NewBusiness(log, userBus, delegate, productdb.NewStore(log, db))
	homeBus := homebus.NewBusiness(log, userBus, delegate, homedb.NewStore(log, db))
	vproductBus := vproductbus.NewBusiness(vproductdb.NewStore(log, db))

	return BusDomain{
		Delegate:       delegate,
		Home:           homeBus,
		AssetType:      assetTypeBus,
		AssetCondition: assetConditionBus,
		Asset:          assetBus,
		Product:        productBus,
		User:           userBus,
		Country:        countryBus,
		Region:         regionBus,
		City:           cityBus,
		Street:         streetBus,
		VProduct:       vproductBus,
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
