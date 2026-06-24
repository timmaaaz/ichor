// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
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
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus/stores/currencycache"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus/stores/currencydb"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus/stores/pagedb"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus/stores/paymenttermdb"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus/stores/permissionscache"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus/stores/permissionsdb"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus/stores/rolecache"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus/stores/roledb"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus/stores/rolepagedb"
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
	"github.com/timmaaaz/ichor/business/domain/introspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb"
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
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus/stores/lotlocationdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus/stores/lottrackingsdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus/stores/picktaskdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus/stores/putawaytaskdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus/stores/serialnumberdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus/stores/transferorderdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus/stores/warehousedb"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus/stores/zonedb"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/stores/labeldb"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus/stores/purchaseorderdb"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus/stores/purchaseorderlineitemdb"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus/stores/purchaseorderlineitemstatusdb"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus/stores/purchaseorderstatusdb"
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
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus/stores/scenariodb"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus/stores/actionpermissionsdb"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"

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
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus/stores/timezonedb"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus/stores/homedb"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus/stores/officedb"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus/stores/productuomdb"

	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus/stores/userpreferencesdb"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus/store/reportstodb"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus/stores/titledb"

	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formbus/stores/formdb"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus/stores/formfielddb"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus/stores/pageactiondb"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus/stores/pageconfigdb"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus/stores/pagecontentdb"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/stores/settingscache"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/stores/settingsdb"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/migrate"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflowdomains"
	"github.com/timmaaaz/ichor/foundation/docker"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// BusDomain represents all the business domain apis needed for testing.
type BusDomain struct {
	Delegate *delegate.Delegate

	// OutboxWriter is the transactional-outbox Writer injected into every cascade bus
	// above (F8). Exposed — like Delegate — so tests/handlers that write OUTSIDE a bus
	// (e.g. the Path-C generic data handlers, which write raw SQL then synthesize) can
	// share the same Writer instance via data.WithOutbox(db.BusDomain.OutboxWriter).
	OutboxWriter *outbox.Writer

	// Locations
	Home     *homebus.Business
	Country  *countrybus.Business
	Region   *regionbus.Business
	City     *citybus.Business
	Street   *streetbus.Business
	Timezone *timezonebus.Business
	Office   *officebus.Business

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
	ContactInfos    *contactinfosbus.Business
	Customers       *customersbus.Business
	PaymentTerm     *paymenttermbus.Business
	Currency        *currencybus.Business
	UserPreferences *userpreferencesbus.Business

	// Inventory
	Brand             *brandbus.Business
	ProductCategory   *productcategorybus.Business
	Product           *productbus.Business
	ProductUOM        *productuombus.Business
	PhysicalAttribute *physicalattributebus.Business
	InventoryItem     *inventoryitembus.Business

	// Warehouse
	Warehouse         *warehousebus.Business
	Zones             *zonebus.Business
	InventoryLocation *inventorylocationbus.Business

	// Permissions
	Role        *rolebus.Business
	Page        *pagebus.Business
	RolePage    *rolepagebus.Business
	UserRole    *userrolebus.Business
	TableAccess *tableaccessbus.Business
	Permissions *permissionsbus.Business

	// Introspection
	Introspection *introspectionbus.Business

	// Finance
	ProductCost *productcostbus.Business

	// Supplier
	Supplier        *supplierbus.Business
	SupplierProduct *supplierproductbus.Business
	CostHistory     *costhistorybus.Business

	// Purchase Orders
	PurchaseOrderStatus         *purchaseorderstatusbus.Business
	PurchaseOrderLineItemStatus *purchaseorderlineitemstatusbus.Business
	PurchaseOrder               *purchaseorderbus.Business
	PurchaseOrderLineItem       *purchaseorderlineitembus.Business

	// Quality
	Metrics    *metricsbus.Business
	Inspection *inspectionbus.Business

	// Lots
	LotTrackings *lottrackingsbus.Business
	SerialNumber *serialnumberbus.Business
	LotLocation  *lotlocationbus.Business

	// Movement
	InventoryTransaction *inventorytransactionbus.Business
	InventoryAdjustment  *inventoryadjustmentbus.Business
	TransferOrder        *transferorderbus.Business
	PutAwayTask          *putawaytaskbus.Business
	PickTask             *picktaskbus.Business
	CycleCountSession    *cyclecountsessionbus.Business
	CycleCountItem       *cyclecountitembus.Business

	// Labels
	Label *labelbus.Business

	// Scenarios
	Scenario *scenariobus.Business

	// Order
	OrderFulfillmentStatus    *orderfulfillmentstatusbus.Business
	LineItemFulfillmentStatus *lineitemfulfillmentstatusbus.Business
	Order                     *ordersbus.Business
	OrderLineItem             *orderlineitemsbus.Business

	// Workflow
	Workflow          *workflow.Business
	Alert             *alertbus.Business
	ActionPermissions *actionpermissionsbus.Business
	ApprovalRequest   *approvalrequestbus.Business

	// Data
	ConfigStore *tablebuilder.ConfigStore
	TableStore  *tablebuilder.Store

	// Config
	Form        *formbus.Business
	FormField   *formfieldbus.Business
	PageAction  *pageactionbus.Business
	PageConfig  *pageconfigbus.Business
	PageContent *pagecontentbus.Business
	Settings    *settingsbus.Business
}

func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	delegate := delegate.New(log)

	// F8 harness parity: build the transactional-outbox Writer and inject it into the
	// same cascade buses production wires (mirrors all.go), so integration tests exercise
	// the live outbox+relay cascade path rather than the legacy delegate path. The
	// domain→entity map comes from the (now business-layer) workflowdomains registry; the
	// lineage extractor is the exported temporal helper. The Writer makes every cascade
	// bus write persist an outbox row; cascades only DISPATCH where a relay runs (the
	// workflow rigs start one), so non-workflow tests just emit harmless rows dropped at
	// teardown.
	entityForDomain := make(map[string]string)
	for _, r := range workflowdomains.Registrations() {
		entityForDomain[r.Domain] = r.Entity
	}
	outboxWriter := outbox.NewWriter(log, db, entityForDomain, temporal.MarshalLineageFromContext)

	// Users
	userapprovalstatusbus := approvalbus.NewBusiness(log, delegate, approvaldb.NewStore(log, db)).WithOutbox(outboxWriter)
	userBus := userbus.NewBusiness(log, delegate, userapprovalstatusbus, usercache.NewStore(log, userdb.NewStore(log, db), time.Hour)).WithOutbox(outboxWriter)
	reportsToBus := reportstobus.NewBusiness(log, delegate, reportstodb.NewStore(log, db)).WithOutbox(outboxWriter)
	userApprovalCommentBus := commentbus.NewBusiness(log, delegate, userBus, commentdb.NewStore(log, db)).WithOutbox(outboxWriter)
	titlebus := titlebus.NewBusiness(log, delegate, titledb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Locations
	countryBus := countrybus.NewBusiness(log, delegate, countrydb.NewStore(log, db))
	regionBus := regionbus.NewBusiness(log, delegate, regiondb.NewStore(log, db))
	cityBus := citybus.NewBusiness(log, delegate, citydb.NewStore(log, db)).WithOutbox(outboxWriter)
	streetBus := streetbus.NewBusiness(log, delegate, streetdb.NewStore(log, db)).WithOutbox(outboxWriter)
	timezoneBus := timezonebus.NewBusiness(log, delegate, timezonedb.NewStore(log, db)).WithOutbox(outboxWriter)
	homeBus := homebus.NewBusiness(log, userBus, delegate, homedb.NewStore(log, db)).WithOutbox(outboxWriter)
	officeBus := officebus.NewBusiness(log, delegate, officedb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Assets
	assetTypeBus := assettypebus.NewBusiness(log, delegate, assettypedb.NewStore(log, db)).WithOutbox(outboxWriter)
	validAssetBus := validassetbus.NewBusiness(log, delegate, validassetdb.NewStore(log, db)).WithOutbox(outboxWriter)
	assetConditionBus := assetconditionbus.NewBusiness(log, delegate, assetconditiondb.NewStore(log, db)).WithOutbox(outboxWriter)
	approvalstatusBus := approvalstatusbus.NewBusiness(log, delegate, approvalstatusdb.NewStore(log, db)).WithOutbox(outboxWriter)
	fulfillmentstatusBus := fulfillmentstatusbus.NewBusiness(log, delegate, fulfillmentstatusdb.NewStore(log, db)).WithOutbox(outboxWriter)
	tagBus := tagbus.NewBusiness(log, delegate, tagdb.NewStore(log, db)).WithOutbox(outboxWriter)
	assetTagBus := assettagbus.NewBusiness(log, delegate, assettagdb.NewStore(log, db)).WithOutbox(outboxWriter)
	userAssetBus := userassetbus.NewBusiness(log, delegate, userassetdb.NewStore(log, db)).WithOutbox(outboxWriter)
	assetBus := assetbus.NewBusiness(log, delegate, assetdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Core
	contactInfosBus := contactinfosbus.NewBusiness(log, delegate, contactinfosdb.NewStore(log, db)).WithOutbox(outboxWriter)
	customersBus := customersbus.NewBusiness(log, delegate, customersdb.NewStore(log, db)).WithOutbox(outboxWriter)
	paymentTermBus := paymenttermbus.NewBusiness(log, delegate, paymenttermdb.NewStore(log, db)).WithOutbox(outboxWriter)
	currencyBus := currencybus.NewBusiness(log, delegate, currencycache.NewStore(log, currencydb.NewStore(log, db), 60*time.Minute)).WithOutbox(outboxWriter)
	userPreferencesBus := userpreferencesbus.NewBusiness(log, userpreferencesdb.NewStore(log, db))

	// Inventory
	brandBus := brandbus.NewBusiness(log, delegate, branddb.NewStore(log, db)).WithOutbox(outboxWriter)
	productCategoryBus := productcategorybus.NewBusiness(log, delegate, productcategorydb.NewStore(log, db)).WithOutbox(outboxWriter)
	productBus := productbus.NewBusiness(log, delegate, productdb.NewStore(log, db)).WithOutbox(outboxWriter)
	productUOMBus := productuombus.NewBusiness(log, delegate, productuomdb.NewStore(log, db))
	physicalAttributeBus := physicalattributebus.NewBusiness(log, delegate, physicalattributedb.NewStore(log, db)).WithOutbox(outboxWriter)
	inventoryItemBus := inventoryitembus.NewBusiness(log, delegate, inventoryitemdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Warehouses
	warehouseBus := warehousebus.NewBusiness(log, delegate, warehousedb.NewStore(log, db)).WithOutbox(outboxWriter)
	zoneBus := zonebus.NewBusiness(log, delegate, zonedb.NewStore(log, db)).WithOutbox(outboxWriter)
	inventoryLocationBus := inventorylocationbus.NewBusiness(log, delegate, inventorylocationdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Permissions
	roleBus := rolebus.NewBusiness(log, delegate, rolecache.NewStore(log, roledb.NewStore(log, db), 60*time.Minute)).WithOutbox(outboxWriter)
	pageBus := pagebus.NewBusiness(log, delegate, pagedb.NewStore(log, db)).WithOutbox(outboxWriter)
	rolePageBus := rolepagebus.NewBusiness(log, delegate, rolepagedb.NewStore(log, db)).WithOutbox(outboxWriter)
	userRoleBus := userrolebus.NewBusiness(log, delegate, userrolecache.NewStore(log, userroledb.NewStore(log, db), 60*time.Minute)).WithOutbox(outboxWriter)
	tableAccessBus := tableaccessbus.NewBusiness(log, delegate, tableaccesscache.NewStore(log, tableaccessdb.NewStore(log, db), 60*time.Minute)).WithOutbox(outboxWriter)
	permissionsBus := permissionsbus.NewBusiness(log, delegate, permissionscache.NewStore(log, permissionsdb.NewStore(log, db), 60*time.Minute), userRoleBus, tableAccessBus, roleBus)

	// Introspection
	introspectionBus := introspectionbus.NewBusiness(log, db)

	// Finance
	productCostBus := productcostbus.NewBusiness(log, delegate, productcostdb.NewStore(log, db)).WithOutbox(outboxWriter)
	costHistoryBus := costhistorybus.NewBusiness(log, delegate, costhistorydb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Suppliers
	supplierBus := supplierbus.NewBusiness(log, delegate, supplierdb.NewStore(log, db)).WithOutbox(outboxWriter)
	supplierProductBus := supplierproductbus.NewBusiness(log, delegate, supplierproductdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Purchase Orders
	purchaseOrderStatusBus := purchaseorderstatusbus.NewBusiness(log, delegate, purchaseorderstatusdb.NewStore(log, db)).WithOutbox(outboxWriter)
	purchaseOrderLineItemStatusBus := purchaseorderlineitemstatusbus.NewBusiness(log, delegate, purchaseorderlineitemstatusdb.NewStore(log, db)).WithOutbox(outboxWriter)
	purchaseOrderBus := purchaseorderbus.NewBusiness(log, delegate, purchaseorderdb.NewStore(log, db)).WithOutbox(outboxWriter)
	purchaseOrderLineItemBus := purchaseorderlineitembus.NewBusiness(log, delegate, purchaseorderlineitemdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Quality
	metricsBus := metricsbus.NewBusiness(log, delegate, metricsdb.NewStore(log, db)).WithOutbox(outboxWriter)
	inspectionBus := inspectionbus.NewBusiness(log, delegate, inspectiondb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Lots
	lotTrackingsBus := lottrackingsbus.NewBusiness(log, delegate, lottrackingsdb.NewStore(log, db)).WithOutbox(outboxWriter)
	serialNumberBus := serialnumberbus.NewBusiness(log, delegate, serialnumberdb.NewStore(log, db)).WithOutbox(outboxWriter)
	lotLocationBus := lotlocationbus.NewBusiness(log, delegate, lotlocationdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Movement
	inventoryTransactionBus := inventorytransactionbus.NewBusiness(log, delegate, inventorytransactiondb.NewStore(log, db)).WithOutbox(outboxWriter)
	inventoryAdjustmentBus := inventoryadjustmentbus.NewBusiness(log, delegate, inventoryadjustmentdb.NewStore(log, db)).WithOutbox(outboxWriter)
	transferOrderBus := transferorderbus.NewBusiness(log, delegate, transferorderdb.NewStore(log, db)).WithOutbox(outboxWriter)
	putAwayTaskBus := putawaytaskbus.NewBusiness(log, delegate, putawaytaskdb.NewStore(log, db)).WithOutbox(outboxWriter)
	pickTaskBus := picktaskbus.NewBusiness(log, delegate, picktaskdb.NewStore(log, db)).WithOutbox(outboxWriter)
	cycleCountSessionBus := cyclecountsessionbus.NewBusiness(log, delegate, cyclecountsessiondb.NewStore(log, db)).WithOutbox(outboxWriter)
	cycleCountItemBus := cyclecountitembus.NewBusiness(log, delegate, cyclecountitemdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Labels — printer is nil at the BusDomain layer; tests that exercise
	// printing inject a recording printer through the API stack via
	// mux.Config.LabelPrinter. Direct bus-level tests do not print.
	labelBus := labelbus.NewBusiness(log, delegate, labeldb.NewStore(log, db), nil).WithOutbox(outboxWriter)

	// Scenarios — beginner is required for transactional Load/Reset.
	// Pass "" for scenariosRoot: dbtest contexts do not exercise the YAML
	// worker-zone path.
	scenarioBus := scenariobus.NewBusiness(log, delegate, scenariodb.NewStore(log, db), sqldb.NewBeginner(db), "").WithOutbox(outboxWriter)

	// Orders
	orderFulfillmentStatusBus := orderfulfillmentstatusbus.NewBusiness(log, delegate, orderfulfillmentstatusdb.NewStore(log, db)).WithOutbox(outboxWriter)
	lineItemFulfillmentStatusBus := lineitemfulfillmentstatusbus.NewBusiness(log, delegate, lineitemfulfillmentstatusdb.NewStore(log, db)).WithOutbox(outboxWriter)
	ordersBus := ordersbus.NewBusiness(log, delegate, ordersdb.NewStore(log, db)).WithOutbox(outboxWriter)
	orderLineItemsBus := orderlineitemsbus.NewBusiness(log, delegate, orderlineitemsdb.NewStore(log, db)).WithOutbox(outboxWriter)

	// Workflow
	workflowBus := workflow.NewBusiness(log, delegate, workflowdb.NewStore(log, db)).WithOutboxEmitter(outboxWriter.Emit)
	alertBus := alertbus.NewBusiness(log, alertdb.NewStore(log, db))
	actionPermissionsBus := actionpermissionsbus.NewBusiness(log, actionpermissionsdb.NewStore(log, db))

	// Data
	configBus := tablebuilder.NewConfigStore(log, db)
	tableBus := tablebuilder.NewStore(log, db)

	// Config
	formFieldBus := formfieldbus.NewBusiness(log, delegate, formfielddb.NewStore(log, db)).WithOutbox(outboxWriter)
	formBus := formbus.NewBusiness(log, delegate, formdb.NewStore(log, db), formFieldBus).WithOutbox(outboxWriter)
	pageContentBus := pagecontentbus.NewBusiness(log, delegate, pagecontentdb.NewStore(log, db)).WithOutbox(outboxWriter)
	pageActionBus := pageactionbus.NewBusiness(log, delegate, pageactiondb.NewStore(log, db)).WithOutbox(outboxWriter)
	pageConfigBus := pageconfigbus.NewBusiness(log, delegate, pageconfigdb.NewStore(log, db), pageContentBus, pageActionBus).WithOutbox(outboxWriter)
	settingsBus := settingsbus.NewBusiness(log, delegate, settingscache.NewStore(log, settingsdb.NewStore(log, db), 30*time.Second))

	return BusDomain{
		Delegate:                    delegate,
		OutboxWriter:                outboxWriter,
		Home:                        homeBus,
		AssetType:                   assetTypeBus,
		ValidAsset:                  validAssetBus,
		User:                        userBus,
		UserApprovalStatus:          userapprovalstatusbus,
		UserApprovalComment:         userApprovalCommentBus,
		Country:                     countryBus,
		Region:                      regionBus,
		City:                        cityBus,
		Street:                      streetBus,
		Timezone:                    timezoneBus,
		ApprovalStatus:              approvalstatusBus,
		FulfillmentStatus:           fulfillmentstatusBus,
		AssetCondition:              assetConditionBus,
		Tag:                         tagBus,
		AssetTag:                    assetTagBus,
		Title:                       titlebus,
		ReportsTo:                   reportsToBus,
		Office:                      officeBus,
		UserAsset:                   userAssetBus,
		Asset:                       assetBus,
		ContactInfos:                contactInfosBus,
		Customers:                   customersBus,
		PaymentTerm:                 paymentTermBus,
		Currency:                    currencyBus,
		UserPreferences:             userPreferencesBus,
		Brand:                       brandBus,
		Warehouse:                   warehouseBus,
		Role:                        roleBus,
		Page:                        pageBus,
		RolePage:                    rolePageBus,
		UserRole:                    userRoleBus,
		ProductCategory:             productCategoryBus,
		TableAccess:                 tableAccessBus,
		Permissions:                 permissionsBus,
		Introspection:               introspectionBus,
		Product:                     productBus,
		ProductUOM:                  productUOMBus,
		PhysicalAttribute:           physicalAttributeBus,
		ProductCost:                 productCostBus,
		Supplier:                    supplierBus,
		CostHistory:                 costHistoryBus,
		SupplierProduct:             supplierProductBus,
		PurchaseOrderStatus:         purchaseOrderStatusBus,
		PurchaseOrderLineItemStatus: purchaseOrderLineItemStatusBus,
		PurchaseOrder:               purchaseOrderBus,
		PurchaseOrderLineItem:       purchaseOrderLineItemBus,
		Metrics:                     metricsBus,
		LotTrackings:                lotTrackingsBus,
		LotLocation:                 lotLocationBus,
		Zones:                       zoneBus,
		InventoryLocation:           inventoryLocationBus,
		InventoryItem:               inventoryItemBus,
		Inspection:                  inspectionBus,
		SerialNumber:                serialNumberBus,
		InventoryTransaction:        inventoryTransactionBus,
		InventoryAdjustment:         inventoryAdjustmentBus,
		TransferOrder:               transferOrderBus,
		PutAwayTask:                 putAwayTaskBus,
		PickTask:                    pickTaskBus,
		CycleCountSession:           cycleCountSessionBus,
		CycleCountItem:              cycleCountItemBus,
		Label:                       labelBus,
		Scenario:                    scenarioBus,
		OrderFulfillmentStatus:      orderFulfillmentStatusBus,
		LineItemFulfillmentStatus:   lineItemFulfillmentStatusBus,
		Order:                       ordersBus,
		OrderLineItem:               orderLineItemsBus,
		Workflow:                    workflowBus,
		Alert:                       alertBus,
		ActionPermissions:           actionPermissionsBus,
		ApprovalRequest:             approvalrequestbus.NewBusiness(log, delegate, approvalrequestdb.NewStore(log, db)),
		ConfigStore:                 configBus,
		TableStore:                  tableBus,
		Form:                        formBus,
		FormField:                   formFieldBus,
		PageAction:                  pageActionBus,
		PageConfig:                  pageConfigBus,
		PageContent:                 pageContentBus,
		Settings:                    settingsBus,
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
		User:         "postgres",
		Password:     "postgres",
		Host:         c.HostPort,
		Name:         "postgres",
		MaxIdleConns: 1,
		MaxOpenConns: 1,
		DisableTLS:   true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	// Reclaim leftover test databases from prior runs (e.g. a run killed before
	// cleanup, or one that t.Fatalf'd mid-setup) BEFORE the per-test timeout
	// below starts — a one-time sweep must not eat into the migrate/seed budget.
	// Runs once per process and only drops databases whose owning process is gone.
	orphanSweepOnce.Do(func() { reapOrphanedDatabases(t, dbM) })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqldb.StatusCheck(ctx, dbM); err != nil {
		t.Fatalf("status check database: %v", err)
	}

	// -------------------------------------------------------------------------

	// Create a uniquely-named database. The name embeds the creating PID (so the
	// reaper can tell live databases from orphans) plus 12 random chars, making
	// collisions astronomically unlikely; the retry is a deterministic backstop.
	var dbName string
	for attempt := 0; ; attempt++ {
		dbName = newTestDBName()
		t.Logf("Create Database: %s\n", dbName)
		if _, err = dbM.ExecContext(context.Background(), "CREATE DATABASE "+dbName); err == nil {
			break
		}
		if attempt < 5 && strings.Contains(err.Error(), "already exists") {
			continue
		}
		t.Fatalf("creating database %s: %v", dbName, err)
	}

	// -------------------------------------------------------------------------

	// Register teardown immediately after CREATE so a t.Fatalf during the open/
	// migrate/seed steps below cannot orphan the database we just created. db is
	// opened next; the closure nil-guards it for the pre-open failure paths.
	var db *sqlx.DB
	var buf bytes.Buffer
	t.Cleanup(func() {
		t.Helper()

		// Close the test-database pool BEFORE dropping the database. With idle
		// connections now retained (MaxIdleConns > 0, for connection reuse),
		// DROP DATABASE would otherwise fail with "database is being accessed by
		// other users" because the pool still holds open connections to it.
		if db != nil {
			db.Close()
		}

		t.Logf("Drop Database: %s\n", dbName)
		if _, err := dbM.ExecContext(context.Background(), "DROP DATABASE "+dbName); err != nil {
			// Don't fail the test on a drop hiccup — the orphan reaper reclaims
			// anything left behind on the next run.
			t.Logf("dropping database %s: %v", dbName, err)
		}

		dbM.Close()

		t.Logf("******************** LOGS (%s) ********************\n\n", testName)
		t.Log(buf.String())
		t.Logf("******************** LOGS (%s) ********************\n", testName)
	})

	// -------------------------------------------------------------------------

	db, err = sqldb.Open(sqldb.Config{
		User:         "postgres",
		Password:     "postgres",
		Host:         c.HostPort,
		Name:         dbName,
		MaxIdleConns: 4,
		MaxOpenConns: 4,
		DisableTLS:   true,
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

	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(ctx) })

	return &Database{
		DB:        db,
		Log:       log,
		BusDomain: newBusDomains(log, db),
	}
}

// =============================================================================

// orphanSweepOnce guards reapOrphanedDatabases so it runs at most once per test
// process, on the first NewDatabase call.
var orphanSweepOnce sync.Once

// newTestDBName returns a unique Postgres database name of the form
// ichortest_<pid>_<12 random chars>. The PID lets reapOrphanedDatabases tell
// databases owned by a live test process from orphans left by a dead one; the
// 12 random chars (36^12 ≈ 4.7e18) make collisions astronomically unlikely.
func newTestDBName() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("ichortest_%d_%s", os.Getpid(), string(b))
}

// reapOrphanedDatabases drops leftover test databases in the shared servicetest
// container whose owning process is no longer alive. Running it once per process
// keeps a series of killed/interrupted runs from accumulating hundreds of orphan
// databases (which then collide against the name space). DROP DATABASE fails on
// a database with active connections, so a genuinely live database is skipped.
// maxOrphanReap bounds how many orphan databases the reaper drops per run. Each
// DROP DATABASE forces a Postgres checkpoint, so dropping a large backlog at once
// triggers an I/O storm that starves concurrent migrations (observed: ~700 drops
// produced a 260s checkpoint that stalled an entire run). Capping keeps the storm
// small; any remainder is reclaimed over subsequent runs.
const maxOrphanReap = 32

func reapOrphanedDatabases(t *testing.T, dbM *sqlx.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rows, err := dbM.QueryContext(ctx,
		"SELECT datname FROM pg_database WHERE datistemplate = false AND datname <> 'postgres'")
	if err != nil {
		t.Logf("orphan reaper: listing databases: %v", err)
		return
	}

	var candidates []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		candidates = append(candidates, name)
	}
	rows.Close()

	dropped, remaining := 0, 0
	for _, name := range candidates {
		if !isReapableTestDB(name) {
			continue
		}
		if dropped >= maxOrphanReap {
			remaining++
			continue
		}
		if _, err := dbM.ExecContext(ctx, "DROP DATABASE "+name); err != nil {
			continue // likely "being accessed by other users" — a live database
		}
		dropped++
	}
	if dropped > 0 {
		t.Logf("orphan reaper: dropped %d leftover test database(s)", dropped)
	}
	if remaining > 0 {
		t.Logf("orphan reaper: %d orphan(s) remain (per-run cap %d); reclaiming over subsequent runs", remaining, maxOrphanReap)
	}
}

// isReapableTestDB reports whether name is a test database safe to drop because
// no live process owns it. It only ever returns true for the two name shapes
// this harness creates, so it can never target a non-test database:
//   - new format ichortest_<pid>_<rand>: reaped only when <pid> is not alive
//   - legacy 4-lowercase-letter names: always reapable, since the current
//     harness never creates them, so no live process can own one
func isReapableTestDB(name string) bool {
	if strings.HasPrefix(name, "ichortest_") {
		parts := strings.Split(name, "_") // ["ichortest", "<pid>", "<rand>"]
		if len(parts) == 3 && isLowerAlphaNum(parts[2]) {
			if pid, err := strconv.Atoi(parts[1]); err == nil {
				return !processAlive(pid)
			}
		}
		return false // unrecognized ichortest_ shape — leave it alone
	}
	return isLegacyTestDBName(name)
}

// isLegacyTestDBName matches the old 4-random-lowercase-letter naming scheme.
func isLegacyTestDBName(name string) bool {
	if len(name) != 4 {
		return false
	}
	for i := 0; i < len(name); i++ {
		if name[i] < 'a' || name[i] > 'z' {
			return false
		}
	}
	return true
}

func isLowerAlphaNum(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// processAlive reports whether a process with the given PID is currently running.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
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
func normalizeJSONInStruct(gotVal, expVal reflect.Value) {
	// Handle pointer types - dereference to get to the actual struct
	for gotVal.Kind() == reflect.Ptr {
		if gotVal.IsNil() {
			return
		}
		gotVal = gotVal.Elem()
	}

	for expVal.Kind() == reflect.Ptr {
		if expVal.IsNil() {
			return
		}
		expVal = expVal.Elem()
	}

	// Now both should be structs
	if gotVal.Kind() != reflect.Struct || expVal.Kind() != reflect.Struct {
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
			// Handle non-pointer json.RawMessage - check the actual value's type first
			gotFieldVal := gotVal.Field(i)
			expFieldVal := expVal.Field(i)

			// Verify the actual interface value is json.RawMessage before asserting
			if gotFieldVal.CanInterface() && expFieldVal.CanInterface() {
				gotInterface := gotFieldVal.Interface()
				expInterface := expFieldVal.Interface()

				// Check if both are actually json.RawMessage
				gotRaw, gotOk := gotInterface.(json.RawMessage)
				expRaw, expOk := expInterface.(json.RawMessage)

				if gotOk && expOk {
					_, normalized := normalizeJSON(gotRaw, expRaw)
					if expFieldVal.CanSet() {
						expFieldVal.Set(reflect.ValueOf(normalized))
					}
				}
			}
		}
	}
}

// func normalizeJSONInStruct(gotVal, expVal reflect.Value) {
// 	// Handle pointer types - dereference to get to the actual struct
// 	for gotVal.Kind() == reflect.Ptr {
// 		if gotVal.IsNil() {
// 			return
// 		}
// 		gotVal = gotVal.Elem()
// 	}

// 	for expVal.Kind() == reflect.Ptr {
// 		if expVal.IsNil() {
// 			return
// 		}
// 		expVal = expVal.Elem()
// 	}

// 	// Now both should be structs
// 	if gotVal.Kind() != reflect.Struct || expVal.Kind() != reflect.Struct {
// 		return
// 	}

// 	gotType := gotVal.Type()

// 	for i := 0; i < gotVal.NumField(); i++ {
// 		field := gotType.Field(i)

// 		// Skip unexported fields
// 		if !field.IsExported() {
// 			continue
// 		}

// 		// Check if field is *json.RawMessage (pointer)
// 		if field.Type == reflect.TypeOf((*json.RawMessage)(nil)) {
// 			gotFieldPtr := gotVal.Field(i)
// 			expFieldPtr := expVal.Field(i)

// 			// Skip if either is nil
// 			if gotFieldPtr.IsNil() || expFieldPtr.IsNil() {
// 				continue
// 			}

// 			// Dereference the pointers to get the actual json.RawMessage values
// 			gotField := gotFieldPtr.Elem().Interface().(json.RawMessage)
// 			expField := expFieldPtr.Elem().Interface().(json.RawMessage)

// 			// Compare and normalize
// 			_, normalized := normalizeJSON(gotField, expField)

// 			// Update the expected field by setting the value through the pointer
// 			if expFieldPtr.CanSet() {
// 				normalizedCopy := make(json.RawMessage, len(normalized))
// 				copy(normalizedCopy, normalized)
// 				expFieldPtr.Set(reflect.ValueOf(&normalizedCopy))
// 			}
// 		} else if field.Type == reflect.TypeOf(json.RawMessage{}) {
// 			// Handle non-pointer json.RawMessage
// 			gotField := gotVal.Field(i).Interface().(json.RawMessage)
// 			expField := expVal.Field(i).Interface().(json.RawMessage)

// 			_, normalized := normalizeJSON(gotField, expField)
// 			if expVal.Field(i).CanSet() {
// 				expVal.Field(i).Set(reflect.ValueOf(normalized))
// 			}
// 		}
// 	}
// }
