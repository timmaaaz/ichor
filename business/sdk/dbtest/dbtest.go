// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus/stores/approvalstatusdb"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus/stores/assetdb"
	validassetdb "github.com/timmaaaz/ichor/business/domain/assets/validassetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus/stores/contactinfodb"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus/stores/branddb"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus/stores/organizationalunitdb"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus/stores/roledb"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus/stores/userroledb"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus/stores/approvaldb"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus/stores/commentdb"

	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus/stores/assetconditiondb"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus/store/assettagdb"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus/stores/assettypedb"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	fulfillmentstatusdb "github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus/stores"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus/stores/tagdb"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus/stores/userassetdb"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/homebus/stores/homedb"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	citydb "github.com/timmaaaz/ichor/business/domain/location/citybus/stores/citydb"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus/stores/countrydb"
	"github.com/timmaaaz/ichor/business/domain/location/officebus"
	"github.com/timmaaaz/ichor/business/domain/location/officebus/stores/officedb"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus/stores/regiondb"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	streetdb "github.com/timmaaaz/ichor/business/domain/location/streetbus/stores/streetdb"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus/store/reportstodb"
	"github.com/timmaaaz/ichor/business/domain/users/titlebus"
	"github.com/timmaaaz/ichor/business/domain/users/titlebus/stores/titledb"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/userdb"
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
	Delegate *delegate.Delegate

	// Locations
	Home    *homebus.Business
	Country *countrybus.Business
	Region  *regionbus.Business
	City    *citybus.Business
	Street  *streetbus.Business
	Office  *officebus.Business

	// Users
	User                *userbus.Business
	Title               *titlebus.Business
	ReportsTo           *reportstobus.Business
	UserApprovalStatus  *approvalbus.Business
	UserApprovalComment *commentbus.Business

	// Assets
	ApprovalStatus    *approvalstatusbus.Business
	FulfillmentStatus *fulfillmentstatusbus.Business
	Tag               *tagbus.Business
	AssetTag          *assettagbus.Business
	ValidAsset        *validassetbus.Business
	AssetType         *assettypebus.Business
	AssetCondition    *assetconditionbus.Business
	UserAsset         *userassetbus.Business
	Asset             *assetbus.Business

	// Core
	ContactInfo *contactinfobus.Business

	// Inventory
	Brand *brandbus.Business

	// ETC
	Product  *productbus.Business
	VProduct *vproductbus.Business

	// Permissions
	Role               *rolebus.Business
	UserRole           *userrolebus.Business
	OrganizationalUnit *organizationalunitbus.Business
}

func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	delegate := delegate.New(log)

	// Users
	userapprovalstatusbus := approvalbus.NewBusiness(log, delegate, approvaldb.NewStore(log, db))
	userBus := userbus.NewBusiness(log, delegate, userapprovalstatusbus, usercache.NewStore(log, userdb.NewStore(log, db), time.Hour))
	reportsToBus := reportstobus.NewBusiness(log, delegate, reportstodb.NewStore(log, db))
	userApprovalCommentBus := commentbus.NewBusiness(log, delegate, userBus, commentdb.NewStore(log, db))
	titlebus := titlebus.NewBusiness(log, delegate, titledb.NewStore(log, db))

	// Locations
	countryBus := countrybus.NewBusiness(log, delegate, countrydb.NewStore(log, db))
	regionBus := regionbus.NewBusiness(log, delegate, regiondb.NewStore(log, db))
	cityBus := citybus.NewBusiness(log, delegate, citydb.NewStore(log, db))
	streetBus := streetbus.NewBusiness(log, delegate, streetdb.NewStore(log, db))
	homeBus := homebus.NewBusiness(log, userBus, delegate, homedb.NewStore(log, db))
	officeBus := officebus.NewBusiness(log, delegate, officedb.NewStore(log, db))

	// Assets
	assetTypeBus := assettypebus.NewBusiness(log, delegate, assettypedb.NewStore(log, db))
	validAssetBus := validassetbus.NewBusiness(log, delegate, validassetdb.NewStore(log, db))
	assetConditionBus := assetconditionbus.NewBusiness(log, delegate, assetconditiondb.NewStore(log, db))
	approvalstatusBus := approvalstatusbus.NewBusiness(log, delegate, approvalstatusdb.NewStore(log, db))
	fulfillmentstatusBus := fulfillmentstatusbus.NewBusiness(log, delegate, fulfillmentstatusdb.NewStore(log, db))
	tagBus := tagbus.NewBusiness(log, delegate, tagdb.NewStore(log, db))
	assetTagBus := assettagbus.NewBusiness(log, delegate, assettagdb.NewStore(log, db))
	userAssetBus := userassetbus.NewBusiness(log, delegate, userassetdb.NewStore(log, db))
	assetBus := assetbus.NewBusiness(log, delegate, assetdb.NewStore(log, db))

	// Core
	contactInfoBus := contactinfobus.NewBusiness(log, delegate, contactinfodb.NewStore(log, db))

	// Inventory
	brandBus := brandbus.NewBusiness(log, delegate, branddb.NewStore(log, db))

	// Products
	productBus := productbus.NewBusiness(log, userBus, delegate, productdb.NewStore(log, db))
	vproductBus := vproductbus.NewBusiness(vproductdb.NewStore(log, db))

	// Permissions
	roleBus := rolebus.NewBusiness(log, roledb.NewStore(log, db))
	userRoleBus := userrolebus.NewBusiness(log, userroledb.NewStore(log, db))
	organizationalunitBus := organizationalunitbus.NewBusiness(log, organizationalunitdb.NewStore(log, db))

	return BusDomain{
		Delegate:            delegate,
		Home:                homeBus,
		AssetType:           assetTypeBus,
		ValidAsset:          validAssetBus,
		Product:             productBus,
		User:                userBus,
		UserApprovalStatus:  userapprovalstatusbus,
		UserApprovalComment: userApprovalCommentBus,
		Country:             countryBus,
		Region:              regionBus,
		City:                cityBus,
		Street:              streetBus,
		VProduct:            vproductBus,
		ApprovalStatus:      approvalstatusBus,
		FulfillmentStatus:   fulfillmentstatusBus,
		AssetCondition:      assetConditionBus,
		Tag:                 tagBus,
		AssetTag:            assetTagBus,
		Title:               titlebus,
		ReportsTo:           reportsToBus,
		Office:              officeBus,
		UserAsset:           userAssetBus,
		Asset:               assetBus,
		ContactInfo:         contactInfoBus,
		Brand:               brandBus,
		Role:                roleBus,
		UserRole:            userRoleBus,
		OrganizationalUnit:  organizationalunitBus,
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

	_, err = db.Exec("SET TIME ZONE 'America/New_York'")
	if err != nil {
		t.Fatalf("Error setting time zone: %v", err)
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

func UUIDPointer(value uuid.UUID) *uuid.UUID {
	return &value
}
