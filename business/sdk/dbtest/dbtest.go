// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus/stores/approvalstatusdb"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus/stores/assetdb"
	validassetdb "github.com/timmaaaz/ichor/business/domain/assets/validassetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus/stores/contactinfosdb"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus/stores/permissionscache"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus/stores/permissionsdb"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus/stores/rolecache"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus/stores/roledb"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus/stores/tableaccesscache"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus/stores/tableaccessdb"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus/stores/userrolecache"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus/stores/userroledb"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus/stores/approvaldb"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus/stores/commentdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus/stores/inspectiondb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus/stores/inventoryadjustmentdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus/stores/inventoryitemdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/stores/inventorylocationdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus/stores/lottrackingsdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus/stores/serialnumberdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus/stores/transferorderdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus/stores/warehousedb"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus/stores/zonedb"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/stores/supplierdb"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/stores/supplierproductdb"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus/stores/branddb"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/stores/costhistorydb"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus/stores/metricsdb"
	"github.com/timmaaaz/ichor/business/domain/products/physicalattributebus"
	"github.com/timmaaaz/ichor/business/domain/products/physicalattributebus/stores/physicalattributedb"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus/stores/productcategorydb"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcostbus/stores/productcostdb"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	customersdb "github.com/timmaaaz/ichor/business/domain/sales/customersbus/stores/contactinfosdb"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus/stores/lineitemfulfillmentstatusdb"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus/stores/orderfulfillmentstatusdb"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus/stores/orderlineitemsdb"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/stores/ordersdb"

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
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	citydb "github.com/timmaaaz/ichor/business/domain/geography/citybus/stores/citydb"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus/stores/countrydb"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus/stores/regiondb"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	streetdb "github.com/timmaaaz/ichor/business/domain/geography/streetbus/stores/streetdb"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus/stores/homedb"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus/stores/officedb"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"

	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus/store/reportstodb"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus/stores/titledb"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/migrate"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
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
	ContactInfos *contactinfosbus.Business
	Customers    *customersbus.Business

	// Inventory
	Brand             *brandbus.Business
	ProductCategory   *productcategorybus.Business
	Product           *productbus.Business
	PhysicalAttribute *physicalattributebus.Business
	InventoryItem     *inventoryitembus.Business

	// Warehouse
	Warehouse         *warehousebus.Business
	Zones             *zonebus.Business
	InventoryLocation *inventorylocationbus.Business

	// Permissions
	Role        *rolebus.Business
	UserRole    *userrolebus.Business
	TableAccess *tableaccessbus.Business
	Permissions *permissionsbus.Business

	// Finance
	ProductCost *productcostbus.Business

	// Supplier
	Supplier        *supplierbus.Business
	SupplierProduct *supplierproductbus.Business
	CostHistory     *costhistorybus.Business

	// Quality
	Metrics    *metricsbus.Business
	Inspection *inspectionbus.Business

	// Lots
	LotTrackings *lottrackingsbus.Business
	SerialNumber *serialnumberbus.Business

	// Movement
	InventoryTransaction *inventorytransactionbus.Business
	InventoryAdjustment  *inventoryadjustmentbus.Business
	TransferOrder        *transferorderbus.Business

	// Order
	OrderFulfillmentStatus    *orderfulfillmentstatusbus.Business
	LineItemFulfillmentStatus *lineitemfulfillmentstatusbus.Business
	Order                     *ordersbus.Business
	OrderLineItem             *orderlineitemsbus.Business

	// Workflow
	Workflow *workflow.Business
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
	contactInfosBus := contactinfosbus.NewBusiness(log, delegate, contactinfosdb.NewStore(log, db))
	customersBus := customersbus.NewBusiness(log, delegate, customersdb.NewStore(log, db))

	// Inventory
	brandBus := brandbus.NewBusiness(log, delegate, branddb.NewStore(log, db))
	productCategoryBus := productcategorybus.NewBusiness(log, delegate, productcategorydb.NewStore(log, db))
	productBus := productbus.NewBusiness(log, delegate, productdb.NewStore(log, db))
	physicalAttributeBus := physicalattributebus.NewBusiness(log, delegate, physicalattributedb.NewStore(log, db))
	inventoryItemBus := inventoryitembus.NewBusiness(log, delegate, inventoryitemdb.NewStore(log, db))

	// Warehouses
	warehouseBus := warehousebus.NewBusiness(log, delegate, warehousedb.NewStore(log, db))
	zoneBus := zonebus.NewBusiness(log, delegate, zonedb.NewStore(log, db))
	inventoryLocationBus := inventorylocationbus.NewBusiness(log, delegate, inventorylocationdb.NewStore(log, db))

	// Permissions
	roleBus := rolebus.NewBusiness(log, delegate, rolecache.NewStore(log, roledb.NewStore(log, db), 60*time.Minute))
	userRoleBus := userrolebus.NewBusiness(log, delegate, userrolecache.NewStore(log, userroledb.NewStore(log, db), 60*time.Minute))
	tableAccessBus := tableaccessbus.NewBusiness(log, delegate, tableaccesscache.NewStore(log, tableaccessdb.NewStore(log, db), 60*time.Minute))
	permissionsBus := permissionsbus.NewBusiness(log, delegate, permissionscache.NewStore(log, permissionsdb.NewStore(log, db), 60*time.Minute), userRoleBus, tableAccessBus, roleBus)

	// Finance
	productCostBus := productcostbus.NewBusiness(log, delegate, productcostdb.NewStore(log, db))
	costHistoryBus := costhistorybus.NewBusiness(log, delegate, costhistorydb.NewStore(log, db))

	// Suppliers
	supplierBus := supplierbus.NewBusiness(log, delegate, supplierdb.NewStore(log, db))
	supplierProductBus := supplierproductbus.NewBusiness(log, delegate, supplierproductdb.NewStore(log, db))

	// Quality
	metricsBus := metricsbus.NewBusiness(log, delegate, metricsdb.NewStore(log, db))
	inspectionBus := inspectionbus.NewBusiness(log, delegate, inspectiondb.NewStore(log, db))

	// Lots
	lotTrackingsBus := lottrackingsbus.NewBusiness(log, delegate, lottrackingsdb.NewStore(log, db))
	serialNumberBus := serialnumberbus.NewBusiness(log, delegate, serialnumberdb.NewStore(log, db))

	// Movement
	inventoryTransactionBus := inventorytransactionbus.NewBusiness(log, delegate, inventorytransactiondb.NewStore(log, db))
	inventoryAdjustmentBus := inventoryadjustmentbus.NewBusiness(log, delegate, inventoryadjustmentdb.NewStore(log, db))
	transferOrderBus := transferorderbus.NewBusiness(log, delegate, transferorderdb.NewStore(log, db))

	// Orders
	orderFulfillmentStatusBus := orderfulfillmentstatusbus.NewBusiness(log, delegate, orderfulfillmentstatusdb.NewStore(log, db))
	lineItemFulfillmentStatusBus := lineitemfulfillmentstatusbus.NewBusiness(log, delegate, lineitemfulfillmentstatusdb.NewStore(log, db))
	ordersBus := ordersbus.NewBusiness(log, delegate, ordersdb.NewStore(log, db))
	orderLineItemsBus := orderlineitemsbus.NewBusiness(log, delegate, orderlineitemsdb.NewStore(log, db))

	// Workflow
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	return BusDomain{
		Delegate:                  delegate,
		Home:                      homeBus,
		AssetType:                 assetTypeBus,
		ValidAsset:                validAssetBus,
		User:                      userBus,
		UserApprovalStatus:        userapprovalstatusbus,
		UserApprovalComment:       userApprovalCommentBus,
		Country:                   countryBus,
		Region:                    regionBus,
		City:                      cityBus,
		Street:                    streetBus,
		ApprovalStatus:            approvalstatusBus,
		FulfillmentStatus:         fulfillmentstatusBus,
		AssetCondition:            assetConditionBus,
		Tag:                       tagBus,
		AssetTag:                  assetTagBus,
		Title:                     titlebus,
		ReportsTo:                 reportsToBus,
		Office:                    officeBus,
		UserAsset:                 userAssetBus,
		Asset:                     assetBus,
		ContactInfos:              contactInfosBus,
		Customers:                 customersBus,
		Brand:                     brandBus,
		Warehouse:                 warehouseBus,
		Role:                      roleBus,
		UserRole:                  userRoleBus,
		ProductCategory:           productCategoryBus,
		TableAccess:               tableAccessBus,
		Permissions:               permissionsBus,
		Product:                   productBus,
		PhysicalAttribute:         physicalAttributeBus,
		ProductCost:               productCostBus,
		Supplier:                  supplierBus,
		CostHistory:               costHistoryBus,
		SupplierProduct:           supplierProductBus,
		Metrics:                   metricsBus,
		LotTrackings:              lotTrackingsBus,
		Zones:                     zoneBus,
		InventoryLocation:         inventoryLocationBus,
		InventoryItem:             inventoryItemBus,
		Inspection:                inspectionBus,
		SerialNumber:              serialNumberBus,
		InventoryTransaction:      inventoryTransactionBus,
		InventoryAdjustment:       inventoryAdjustmentBus,
		TransferOrder:             transferOrderBus,
		OrderFulfillmentStatus:    orderFulfillmentStatusBus,
		LineItemFulfillmentStatus: lineItemFulfillmentStatusBus,
		Order:                     ordersBus,
		OrderLineItem:             orderLineItemsBus,
		Workflow:                  workflowBus,
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

	// cmd := exec.Command("docker", "rm", "-f", name)
	// cmd.Run() // Ignore error - container might not exist

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

// Float64Pointer is a helper to get a *float64 from a float64. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func Float64Pointer(f float64) *float64 {
	return &f
}

func Float32Pointer(f float32) *float32 {
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

func UUIDPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// normalizeJSON compares two json.RawMessage values semantically and if they're equal,
// returns the 'got' value to use for both (to avoid formatting differences in comparison).
func normalizeJSON(got, exp json.RawMessage) (json.RawMessage, json.RawMessage) {
	// Handle nil/empty cases
	if len(got) == 0 && len(exp) == 0 {
		return got, exp
	}
	if len(got) == 0 || len(exp) == 0 {
		return got, exp
	}

	// Parse both JSON values
	var gotJSON, expJSON interface{}
	if err := json.Unmarshal(got, &gotJSON); err != nil {
		return got, exp
	}
	if err := json.Unmarshal(exp, &expJSON); err != nil {
		return got, exp
	}

	// If semantically equal, use 'got' for both to avoid formatting diffs
	if reflect.DeepEqual(gotJSON, expJSON) {
		return got, got
	}

	return got, exp
}

// NormalizeJSONFields handles both single structs and slices of structs
func NormalizeJSONFields(got, exp interface{}) {
	gotVal := reflect.ValueOf(got)
	expVal := reflect.ValueOf(exp)

	// Handle pointers - need to get the element we can actually modify
	if expVal.Kind() == reflect.Ptr {
		expVal = expVal.Elem()
	}
	if gotVal.Kind() == reflect.Ptr {
		gotVal = gotVal.Elem()
	}

	switch gotVal.Kind() {
	case reflect.Slice:
		// Handle slices
		if expVal.Kind() != reflect.Slice {
			return
		}

		minLen := gotVal.Len()
		if expVal.Len() < minLen {
			minLen = expVal.Len()
		}

		for i := 0; i < minLen; i++ {
			normalizeJSONInStruct(gotVal.Index(i), expVal.Index(i))
		}

	case reflect.Struct:
		// Handle single struct
		normalizeJSONInStruct(gotVal, expVal)
	}
}

// normalizeJSONInStruct normalizes JSON fields within a single struct
func normalizeJSONInStruct(gotVal, expVal reflect.Value) {
	// Handle pointer types
	if gotVal.Kind() == reflect.Ptr {
		if gotVal.IsNil() || expVal.IsNil() {
			return
		}
		gotVal = gotVal.Elem()
		expVal = expVal.Elem()
	}

	if gotVal.Kind() != reflect.Struct {
		return
	}

	gotType := gotVal.Type()

	for i := 0; i < gotVal.NumField(); i++ {
		field := gotType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check if field is *json.RawMessage (pointer)
		if field.Type == reflect.TypeOf((*json.RawMessage)(nil)) {
			gotFieldPtr := gotVal.Field(i)
			expFieldPtr := expVal.Field(i)

			// Skip if either is nil
			if gotFieldPtr.IsNil() || expFieldPtr.IsNil() {
				continue
			}

			// Dereference the pointers to get the actual json.RawMessage values
			gotField := gotFieldPtr.Elem().Interface().(json.RawMessage)
			expField := expFieldPtr.Elem().Interface().(json.RawMessage)

			// Compare and normalize
			_, normalized := normalizeJSON(gotField, expField)

			// Update the expected field by setting the value through the pointer
			if expFieldPtr.CanSet() {
				normalizedCopy := make(json.RawMessage, len(normalized))
				copy(normalizedCopy, normalized)
				expFieldPtr.Set(reflect.ValueOf(&normalizedCopy))
			}
		} else if field.Type == reflect.TypeOf(json.RawMessage{}) {
			// Handle non-pointer json.RawMessage
			gotField := gotVal.Field(i).Interface().(json.RawMessage)
			expField := expVal.Field(i).Interface().(json.RawMessage)

			_, normalized := normalizeJSON(gotField, expField)
			if expVal.Field(i).CanSet() {
				expVal.Field(i).Set(reflect.ValueOf(normalized))
			}
		}
	}
}
