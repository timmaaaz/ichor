// Package all binds all the routes into the specified app.
package all

import (
	"context"
	"time"

	"github.com/timmaaaz/ichor/api/domain/http/assets/approvalstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assetapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assetconditionapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assettagapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assettypeapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/tagapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/userassetapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/validassetapi"
	"github.com/timmaaaz/ichor/api/domain/http/config/formapi"
	"github.com/timmaaaz/ichor/api/domain/http/config/formfieldapi"
	"github.com/timmaaaz/ichor/api/domain/http/config/pageactionapi"
	"github.com/timmaaaz/ichor/api/domain/http/config/pageconfigapi"
	"github.com/timmaaaz/ichor/api/domain/http/config/pagecontentapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/contactinfosapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/pageapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/roleapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/rolepageapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/tableaccessapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/userroleapi"
	"github.com/timmaaaz/ichor/api/domain/http/dataapi"
	"github.com/timmaaaz/ichor/api/domain/http/formdata/formdataapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/approvalapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/commentapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/officeapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/reportstoapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/titleapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/inspectionapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/inventoryadjustmentapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/inventoryitemapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/inventorylocationapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/inventorytransactionapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/lottrackingsapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/serialnumberapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/transferorderapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/warehouseapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/zoneapi"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/purchaseorderapi"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/purchaseorderlineitemapi"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/purchaseorderlineitemstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/purchaseorderstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/supplierapi"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/supplierproductapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/brandapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/costhistoryapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/metricsapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/physicalattributeapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/productcategoryapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/productcostapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/customersapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/lineitemfulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/orderfulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/orderlineitemsapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/ordersapi"

	"github.com/timmaaaz/ichor/api/domain/http/assets/fulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/checkapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/cityapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/countryapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/regionapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/streetapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/timezoneapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/homeapi"
	"github.com/timmaaaz/ichor/api/domain/http/introspectionapi"

	"github.com/timmaaaz/ichor/api/domain/http/rawapi"

	"github.com/timmaaaz/ichor/api/domain/http/core/userapi"

	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/app/domain/assets/approvalstatusapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assetapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assetconditionapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assettagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
	"github.com/timmaaaz/ichor/app/domain/assets/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/assets/tagapp"
	"github.com/timmaaaz/ichor/app/domain/assets/userassetapp"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/domain/config/formapp"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfosapp"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/domain/core/roleapp"
	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
	"github.com/timmaaaz/ichor/app/domain/core/tableaccessapp"
	"github.com/timmaaaz/ichor/app/domain/core/userapp"
	userroleappimport "github.com/timmaaaz/ichor/app/domain/core/userroleapp"
	"github.com/timmaaaz/ichor/app/domain/formdata/formdataapp"
	"github.com/timmaaaz/ichor/app/domain/geography/cityapp"
	"github.com/timmaaaz/ichor/app/domain/geography/streetapp"
	"github.com/timmaaaz/ichor/app/domain/geography/timezoneapp"
	"github.com/timmaaaz/ichor/app/domain/hr/approvalapp"
	"github.com/timmaaaz/ichor/app/domain/hr/commentapp"
	"github.com/timmaaaz/ichor/app/domain/hr/homeapp"
	"github.com/timmaaaz/ichor/app/domain/hr/officeapp"
	"github.com/timmaaaz/ichor/app/domain/hr/reportstoapp"
	"github.com/timmaaaz/ichor/app/domain/hr/titleapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inspectionapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/serialnumberapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/transferorderapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/warehouseapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/zoneapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemstatusapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderstatusapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierapp"
	"github.com/timmaaaz/ichor/app/domain/procurement/supplierproductapp"
	"github.com/timmaaaz/ichor/app/domain/products/brandapp"
	"github.com/timmaaaz/ichor/app/domain/products/costhistoryapp"
	"github.com/timmaaaz/ichor/app/domain/products/metricsapp"
	"github.com/timmaaaz/ichor/app/domain/products/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/domain/products/productapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcategoryapp"
	"github.com/timmaaaz/ichor/app/domain/products/productcostapp"
	"github.com/timmaaaz/ichor/app/domain/sales/customersapp"
	"github.com/timmaaaz/ichor/app/domain/sales/lineitemfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/sales/orderfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus/stores/approvalstatusdb"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	validassetdb "github.com/timmaaaz/ichor/business/domain/assets/validassetbus/stores/assetdb"
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
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus/stores/contactinfosdb"
	"github.com/timmaaaz/ichor/business/domain/introspectionbus"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus/stores/pagedb"
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

	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus/stores/assetconditiondb"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus/store/assettagdb"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus/stores/tagdb"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus/stores/userassetdb"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus"
	"github.com/timmaaaz/ichor/business/domain/hr/officebus/stores/officedb"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus/store/reportstodb"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus/stores/titledb"

	productapi "github.com/timmaaaz/ichor/api/domain/http/products/productapi"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus/stores/assettypedb"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	fulfillmentstatusdb "github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus/stores"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	citydb "github.com/timmaaaz/ichor/business/domain/geography/citybus/stores/citydb"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus/stores/countrydb"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus/stores/regiondb"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus/stores/streetdb"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus/stores/timezonedb"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/hr/approvalbus/stores/approvaldb"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus/stores/commentdb"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus/stores/homedb"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	inventoryproductdb "github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() *add {
	return &add{}
}

type add struct {
	UserBus *userbus.Business
}

// InitializeDependencies sets up and returns shared dependencies
// This is called BEFORE Add() to set up shared instances
func (a *add) InitializeDependencies(cfg mux.Config) {
	delegate := delegate.New(cfg.Log)
	userApprovalStatusBus := approvalbus.NewBusiness(cfg.Log, delegate, approvaldb.NewStore(cfg.Log, cfg.DB))
	userBus := userbus.NewBusiness(cfg.Log, delegate, userApprovalStatusBus,
		usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB), time.Minute))

	// Store it in the struct
	a.UserBus = userBus

	// Store other buses if needed later
}

// Add implements the RouterAdder interface.
func (a add) Add(app *web.App, cfg mux.Config) {

	// Construct the business domain packages we need here so we are using the
	// sames instances for the different set of domain apis.
	delegate := delegate.New(cfg.Log)
	userApprovalStatusBus := approvalbus.NewBusiness(cfg.Log, delegate, approvaldb.NewStore(cfg.Log, cfg.DB))

	if a.UserBus == nil {
		a.InitializeDependencies(cfg)
	}

	// userBus := userbus.NewBusiness(cfg.Log, delegate, userApprovalStatusBus, usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB), time.Minute))
	userApprovalCommentBus := commentbus.NewBusiness(cfg.Log, delegate, a.UserBus, commentdb.NewStore(cfg.Log, cfg.DB))

	homeBus := homebus.NewBusiness(cfg.Log, a.UserBus, delegate, homedb.NewStore(cfg.Log, cfg.DB))
	countryBus := countrybus.NewBusiness(cfg.Log, delegate, countrydb.NewStore(cfg.Log, cfg.DB))
	regionBus := regionbus.NewBusiness(cfg.Log, delegate, regiondb.NewStore(cfg.Log, cfg.DB))
	cityBus := citybus.NewBusiness(cfg.Log, delegate, citydb.NewStore(cfg.Log, cfg.DB))
	streetBus := streetbus.NewBusiness(cfg.Log, delegate, streetdb.NewStore(cfg.Log, cfg.DB))
	timezoneBus := timezonebus.NewBusiness(cfg.Log, delegate, timezonedb.NewStore(cfg.Log, cfg.DB))
	approvalStatusBus := approvalstatusbus.NewBusiness(cfg.Log, delegate, approvalstatusdb.NewStore(cfg.Log, cfg.DB))
	fulfillmentStatusBus := fulfillmentstatusbus.NewBusiness(cfg.Log, delegate, fulfillmentstatusdb.NewStore(cfg.Log, cfg.DB))
	assetConditionBus := assetconditionbus.NewBusiness(cfg.Log, delegate, assetconditiondb.NewStore(cfg.Log, cfg.DB))
	assetTypeBus := assettypebus.NewBusiness(cfg.Log, delegate, assettypedb.NewStore(cfg.Log, cfg.DB))
	validAssetBus := validassetbus.NewBusiness(cfg.Log, delegate, validassetdb.NewStore(cfg.Log, cfg.DB))
	tagBus := tagbus.NewBusiness(cfg.Log, delegate, tagdb.NewStore(cfg.Log, cfg.DB))
	assetTagBus := assettagbus.NewBusiness(cfg.Log, delegate, assettagdb.NewStore(cfg.Log, cfg.DB))
	titleBus := titlebus.NewBusiness(cfg.Log, delegate, titledb.NewStore(cfg.Log, cfg.DB))
	reportsToBus := reportstobus.NewBusiness(cfg.Log, delegate, reportstodb.NewStore(cfg.Log, cfg.DB))
	officeBus := officebus.NewBusiness(cfg.Log, delegate, officedb.NewStore(cfg.Log, cfg.DB))
	userAssetBus := userassetbus.NewBusiness(cfg.Log, delegate, userassetdb.NewStore(cfg.Log, cfg.DB))
	assetBus := assetbus.NewBusiness(cfg.Log, delegate, assetdb.NewStore(cfg.Log, cfg.DB))

	contactInfosBus := contactinfosbus.NewBusiness(cfg.Log, delegate, contactinfosdb.NewStore(cfg.Log, cfg.DB))
	customersBus := customersbus.NewBusiness(cfg.Log, delegate, customersdb.NewStore(cfg.Log, cfg.DB))

	brandBus := brandbus.NewBusiness(cfg.Log, delegate, branddb.NewStore(cfg.Log, cfg.DB))
	productCategoryBus := productcategorybus.NewBusiness(cfg.Log, delegate, productcategorydb.NewStore(cfg.Log, cfg.DB))
	productBus := productbus.NewBusiness(cfg.Log, delegate, inventoryproductdb.NewStore(cfg.Log, cfg.DB))
	physicalAttributeBus := physicalattributebus.NewBusiness(cfg.Log, delegate, physicalattributedb.NewStore(cfg.Log, cfg.DB))

	productCostBus := productcostbus.NewBusiness(cfg.Log, delegate, productcostdb.NewStore(cfg.Log, cfg.DB))
	costHistoryBus := costhistorybus.NewBusiness(cfg.Log, delegate, costhistorydb.NewStore(cfg.Log, cfg.DB))

	warehouseBus := warehousebus.NewBusiness(cfg.Log, delegate, warehousedb.NewStore(cfg.Log, cfg.DB))
	zoneBus := zonebus.NewBusiness(cfg.Log, delegate, zonedb.NewStore(cfg.Log, cfg.DB))
	inventoryLocationBus := inventorylocationbus.NewBusiness(cfg.Log, delegate, inventorylocationdb.NewStore(cfg.Log, cfg.DB))
	inventoryItemBus := inventoryitembus.NewBusiness(cfg.Log, delegate, inventoryitemdb.NewStore(cfg.Log, cfg.DB))

	purchaseOrderLineItemStatusBus := purchaseorderlineitemstatusbus.NewBusiness(cfg.Log, delegate, purchaseorderlineitemstatusdb.NewStore(cfg.Log, cfg.DB))
	purchaseOrderStatusBus := purchaseorderstatusbus.NewBusiness(cfg.Log, delegate, purchaseorderstatusdb.NewStore(cfg.Log, cfg.DB))
	supplierBus := supplierbus.NewBusiness(cfg.Log, delegate, supplierdb.NewStore(cfg.Log, cfg.DB))
	supplierProductBus := supplierproductbus.NewBusiness(cfg.Log, delegate, supplierproductdb.NewStore(cfg.Log, cfg.DB))
	purchaseOrderBus := purchaseorderbus.NewBusiness(cfg.Log, delegate, purchaseorderdb.NewStore(cfg.Log, cfg.DB))
	purchaseOrderLineItemBus := purchaseorderlineitembus.NewBusiness(cfg.Log, delegate, purchaseorderlineitemdb.NewStore(cfg.Log, cfg.DB))

	metricsBus := metricsbus.NewBusiness(cfg.Log, delegate, metricsdb.NewStore(cfg.Log, cfg.DB))
	inspectionBus := inspectionbus.NewBusiness(cfg.Log, delegate, inspectiondb.NewStore(cfg.Log, cfg.DB))

	lotTrackingsBus := lottrackingsbus.NewBusiness(cfg.Log, delegate, lottrackingsdb.NewStore(cfg.Log, cfg.DB))
	serialNumberBus := serialnumberbus.NewBusiness(cfg.Log, delegate, serialnumberdb.NewStore(cfg.Log, cfg.DB))

	roleBus := rolebus.NewBusiness(cfg.Log, delegate, rolecache.NewStore(cfg.Log, roledb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))
	pageBus := pagebus.NewBusiness(cfg.Log, delegate, pagedb.NewStore(cfg.Log, cfg.DB))
	rolePageBus := rolepagebus.NewBusiness(cfg.Log, delegate, rolepagedb.NewStore(cfg.Log, cfg.DB))
	userRoleBus := userrolebus.NewBusiness(cfg.Log, delegate, userrolecache.NewStore(cfg.Log, userroledb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))
	tableAccessBus := tableaccessbus.NewBusiness(cfg.Log, delegate, tableaccesscache.NewStore(cfg.Log, tableaccessdb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))

	permissionsBus := permissionsbus.NewBusiness(cfg.Log, delegate, permissionscache.NewStore(cfg.Log, permissionsdb.NewStore(cfg.Log, cfg.DB), 60*time.Minute), userRoleBus, tableAccessBus, roleBus)

	introspectionBus := introspectionbus.NewBusiness(cfg.Log, cfg.DB)

	inventoryTransactionBus := inventorytransactionbus.NewBusiness(cfg.Log, delegate, inventorytransactiondb.NewStore(cfg.Log, cfg.DB))
	inventoryAdjustmentBus := inventoryadjustmentbus.NewBusiness(cfg.Log, delegate, inventoryadjustmentdb.NewStore(cfg.Log, cfg.DB))
	transferOrderBus := transferorderbus.NewBusiness(cfg.Log, delegate, transferorderdb.NewStore(cfg.Log, cfg.DB))

	orderFulfillmentStatusBus := orderfulfillmentstatusbus.NewBusiness(cfg.Log, delegate, orderfulfillmentstatusdb.NewStore(cfg.Log, cfg.DB))
	lineItemFulfillmentStatusBus := lineitemfulfillmentstatusbus.NewBusiness(cfg.Log, delegate, lineitemfulfillmentstatusdb.NewStore(cfg.Log, cfg.DB))
	ordersBus := ordersbus.NewBusiness(cfg.Log, delegate, ordersdb.NewStore(cfg.Log, cfg.DB))
	orderLineItemsBus := orderlineitemsbus.NewBusiness(cfg.Log, delegate, orderlineitemsdb.NewStore(cfg.Log, cfg.DB))

	configStore := tablebuilder.NewConfigStore(cfg.Log, cfg.DB)
	tableStore := tablebuilder.NewStore(cfg.Log, cfg.DB)

	formFieldBus := formfieldbus.NewBusiness(cfg.Log, delegate, formfielddb.NewStore(cfg.Log, cfg.DB))
	formBus := formbus.NewBusiness(cfg.Log, delegate, formdb.NewStore(cfg.Log, cfg.DB), formFieldBus)
	pageContentBus := pagecontentbus.NewBusiness(cfg.Log, delegate, pagecontentdb.NewStore(cfg.Log, cfg.DB))
	pageActionBus := pageactionbus.NewBusiness(cfg.Log, delegate, pageactiondb.NewStore(cfg.Log, cfg.DB))
	pageConfigBus := pageconfigbus.NewBusiness(cfg.Log, delegate, pageconfigdb.NewStore(cfg.Log, cfg.DB), pageContentBus, pageActionBus)

	// =========================================================================
	// Initialize Workflow Infrastructure
	// =========================================================================

	var eventPublisher *workflow.EventPublisher

	if cfg.RabbitClient != nil && cfg.RabbitClient.IsConnected() {
		workflowStore := workflowdb.NewStore(cfg.Log, cfg.DB)
		workflowBus := workflow.NewBusiness(cfg.Log, workflowStore)

		workflowEngine := workflow.NewEngine(cfg.Log, cfg.DB, workflowBus)
		if err := workflowEngine.Initialize(context.Background(), workflowBus); err != nil {
			cfg.Log.Error(context.Background(), "workflow engine init failed", "error", err)
		} else {
			workflowQueue := rabbitmq.NewWorkflowQueue(cfg.RabbitClient, cfg.Log)
			queueManager, err := workflow.NewQueueManager(cfg.Log, cfg.DB, workflowEngine, cfg.RabbitClient, workflowQueue)
			if err != nil {
				cfg.Log.Error(context.Background(), "queue manager creation failed", "error", err)
			} else {
				if err := queueManager.Initialize(context.Background()); err != nil {
					cfg.Log.Error(context.Background(), "queue manager init failed", "error", err)
				} else if err := queueManager.Start(context.Background()); err != nil {
					cfg.Log.Error(context.Background(), "queue manager start failed", "error", err)
				} else {
					eventPublisher = workflow.NewEventPublisher(cfg.Log, queueManager)
					cfg.Log.Info(context.Background(), "workflow event infrastructure initialized")

					// Register delegate handlers for workflow event firing
					delegateHandler := workflow.NewDelegateHandler(cfg.Log, eventPublisher)

					// Register Sales domain -> workflow events
					delegateHandler.RegisterDomain(delegate, ordersbus.DomainName, ordersbus.EntityName)
					delegateHandler.RegisterDomain(delegate, customersbus.DomainName, customersbus.EntityName)
					delegateHandler.RegisterDomain(delegate, orderlineitemsbus.DomainName, orderlineitemsbus.EntityName)
					delegateHandler.RegisterDomain(delegate, orderfulfillmentstatusbus.DomainName, orderfulfillmentstatusbus.EntityName)
					delegateHandler.RegisterDomain(delegate, lineitemfulfillmentstatusbus.DomainName, lineitemfulfillmentstatusbus.EntityName)

					// Register Assets domain -> workflow events
					delegateHandler.RegisterDomain(delegate, assetbus.DomainName, assetbus.EntityName)
					delegateHandler.RegisterDomain(delegate, validassetbus.DomainName, validassetbus.EntityName)
					delegateHandler.RegisterDomain(delegate, userassetbus.DomainName, userassetbus.EntityName)
					delegateHandler.RegisterDomain(delegate, assettypebus.DomainName, assettypebus.EntityName)
					delegateHandler.RegisterDomain(delegate, assetconditionbus.DomainName, assetconditionbus.EntityName)
					delegateHandler.RegisterDomain(delegate, assettagbus.DomainName, assettagbus.EntityName)
					delegateHandler.RegisterDomain(delegate, tagbus.DomainName, tagbus.EntityName)
					delegateHandler.RegisterDomain(delegate, approvalstatusbus.DomainName, approvalstatusbus.EntityName)
					delegateHandler.RegisterDomain(delegate, fulfillmentstatusbus.DomainName, fulfillmentstatusbus.EntityName)

					// Register Core domain -> workflow events
					delegateHandler.RegisterDomain(delegate, userbus.DomainName, userbus.EntityName)
					delegateHandler.RegisterDomain(delegate, rolebus.DomainName, rolebus.EntityName)
					delegateHandler.RegisterDomain(delegate, userrolebus.DomainName, userrolebus.EntityName)
					delegateHandler.RegisterDomain(delegate, tableaccessbus.DomainName, tableaccessbus.EntityName)
					delegateHandler.RegisterDomain(delegate, pagebus.DomainName, pagebus.EntityName)
					delegateHandler.RegisterDomain(delegate, rolepagebus.DomainName, rolepagebus.EntityName)
					delegateHandler.RegisterDomain(delegate, contactinfosbus.DomainName, contactinfosbus.EntityName)

					// Register HR domain -> workflow events
					delegateHandler.RegisterDomain(delegate, approvalbus.DomainName, approvalbus.EntityName)
					delegateHandler.RegisterDomain(delegate, commentbus.DomainName, commentbus.EntityName)
					delegateHandler.RegisterDomain(delegate, homebus.DomainName, homebus.EntityName)
					delegateHandler.RegisterDomain(delegate, officebus.DomainName, officebus.EntityName)
					delegateHandler.RegisterDomain(delegate, reportstobus.DomainName, reportstobus.EntityName)
					delegateHandler.RegisterDomain(delegate, titlebus.DomainName, titlebus.EntityName)

					// Register Geography domain -> workflow events
					// Note: countrybus and regionbus are read-only (no Create/Update/Delete) and don't need event registration
					delegateHandler.RegisterDomain(delegate, citybus.DomainName, citybus.EntityName)
					delegateHandler.RegisterDomain(delegate, streetbus.DomainName, streetbus.EntityName)
					delegateHandler.RegisterDomain(delegate, timezonebus.DomainName, timezonebus.EntityName)

					// Register Products domain -> workflow events
					delegateHandler.RegisterDomain(delegate, productbus.DomainName, productbus.EntityName)
					delegateHandler.RegisterDomain(delegate, productcategorybus.DomainName, productcategorybus.EntityName)
					delegateHandler.RegisterDomain(delegate, brandbus.DomainName, brandbus.EntityName)
					delegateHandler.RegisterDomain(delegate, productcostbus.DomainName, productcostbus.EntityName)
					delegateHandler.RegisterDomain(delegate, costhistorybus.DomainName, costhistorybus.EntityName)
					delegateHandler.RegisterDomain(delegate, physicalattributebus.DomainName, physicalattributebus.EntityName)
					delegateHandler.RegisterDomain(delegate, metricsbus.DomainName, metricsbus.EntityName)

					// Register Procurement domain -> workflow events
					delegateHandler.RegisterDomain(delegate, supplierbus.DomainName, supplierbus.EntityName)
					delegateHandler.RegisterDomain(delegate, supplierproductbus.DomainName, supplierproductbus.EntityName)
					delegateHandler.RegisterDomain(delegate, purchaseorderbus.DomainName, purchaseorderbus.EntityName)
					delegateHandler.RegisterDomain(delegate, purchaseorderlineitembus.DomainName, purchaseorderlineitembus.EntityName)
					delegateHandler.RegisterDomain(delegate, purchaseorderstatusbus.DomainName, purchaseorderstatusbus.EntityName)
					delegateHandler.RegisterDomain(delegate, purchaseorderlineitemstatusbus.DomainName, purchaseorderlineitemstatusbus.EntityName)

					// Register Inventory domain -> workflow events
					delegateHandler.RegisterDomain(delegate, warehousebus.DomainName, warehousebus.EntityName)
					delegateHandler.RegisterDomain(delegate, zonebus.DomainName, zonebus.EntityName)
					delegateHandler.RegisterDomain(delegate, inventorylocationbus.DomainName, inventorylocationbus.EntityName)
					delegateHandler.RegisterDomain(delegate, inventoryitembus.DomainName, inventoryitembus.EntityName)
					delegateHandler.RegisterDomain(delegate, inventorytransactionbus.DomainName, inventorytransactionbus.EntityName)
					delegateHandler.RegisterDomain(delegate, inventoryadjustmentbus.DomainName, inventoryadjustmentbus.EntityName)
					delegateHandler.RegisterDomain(delegate, transferorderbus.DomainName, transferorderbus.EntityName)
					delegateHandler.RegisterDomain(delegate, inspectionbus.DomainName, inspectionbus.EntityName)
					delegateHandler.RegisterDomain(delegate, lottrackingsbus.DomainName, lottrackingsbus.EntityName)
					delegateHandler.RegisterDomain(delegate, serialnumberbus.DomainName, serialnumberbus.EntityName)

					// Config domain
					delegateHandler.RegisterDomain(delegate, formbus.DomainName, formbus.EntityName)
					delegateHandler.RegisterDomain(delegate, formfieldbus.DomainName, formfieldbus.EntityName)
					delegateHandler.RegisterDomain(delegate, pageconfigbus.DomainName, pageconfigbus.EntityName)
					delegateHandler.RegisterDomain(delegate, pagecontentbus.DomainName, pagecontentbus.EntityName)
					delegateHandler.RegisterDomain(delegate, pageactionbus.DomainName, pageactionbus.EntityName)

					// Additional domains can be registered here as they implement event.go files
				}
			}
		}
	}

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homeapi.Routes(app, homeapi.Config{
		Log:            cfg.Log,
		UserBus:        a.UserBus,
		HomeBus:        homeBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	rawapi.Routes(app)

	introspectionapi.Routes(app, introspectionapi.Config{
		Log:              cfg.Log,
		IntrospectionBus: introspectionBus,
		AuthClient:       cfg.AuthClient,
		PermissionsBus:   permissionsBus,
	})

	userapi.Routes(app, userapi.Config{
		Log:            cfg.Log,
		UserBus:        a.UserBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	countryapi.Routes(app, countryapi.Config{
		CountryBus:     countryBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	regionapi.Routes(app, regionapi.Config{
		RegionBus:      regionBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	cityapi.Routes(app, cityapi.Config{
		CityBus:        cityBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	streetapi.Routes(app, streetapi.Config{
		StreetBus:      streetBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	timezoneapi.Routes(app, timezoneapi.Config{
		TimezoneBus:    timezoneBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	approvalstatusapi.Routes(app, approvalstatusapi.Config{
		ApprovalStatusBus: approvalStatusBus,
		AuthClient:        cfg.AuthClient,
		Log:               cfg.Log,
		PermissionsBus:    permissionsBus,
	})

	fulfillmentstatusapi.Routes(app, fulfillmentstatusapi.Config{
		FulfillmentStatusBus: fulfillmentStatusBus,
		AuthClient:           cfg.AuthClient,
		Log:                  cfg.Log,
		PermissionsBus:       permissionsBus,
	})

	assetconditionapi.Routes(app, assetconditionapi.Config{
		AssetConditionBus: assetConditionBus,
		AuthClient:        cfg.AuthClient,
		Log:               cfg.Log,
		PermissionsBus:    permissionsBus,
	})

	assettypeapi.Routes(app, assettypeapi.Config{
		AssetTypeBus:   assetTypeBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	validassetapi.Routes(app, validassetapi.Config{
		ValidAssetBus:  validAssetBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	tagapi.Routes(app, tagapi.Config{
		TagBus:         tagBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	titleapi.Routes(app, titleapi.Config{
		TitleBus:       titleBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	assettagapi.Routes(app, assettagapi.Config{
		AssetTagBus:    assetTagBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	reportstoapi.Routes(app, reportstoapi.Config{
		ReportsToBus:   reportsToBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	officeapi.Routes(app, officeapi.Config{
		OfficeBus:      officeBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	userassetapi.Routes(app, userassetapi.Config{
		UserAssetBus:   userAssetBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	assetapi.Routes(app, assetapi.Config{
		AssetBus:       assetBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	approvalapi.Routes(app, approvalapi.Config{
		UserApprovalStatusBus: userApprovalStatusBus,
		AuthClient:            cfg.AuthClient,
		Log:                   cfg.Log,
		PermissionsBus:        permissionsBus,
	})

	commentapi.Routes(app, commentapi.Config{
		Log:                    cfg.Log,
		UserApprovalCommentBus: userApprovalCommentBus,
		AuthClient:             cfg.AuthClient,
		PermissionsBus:         permissionsBus,
	})

	contactinfosapi.Routes(app, contactinfosapi.Config{
		ContactInfosBus: contactInfosBus,
		AuthClient:      cfg.AuthClient,
		Log:             cfg.Log,
		PermissionsBus:  permissionsBus,
	})

	customersapi.Routes(app, customersapi.Config{
		CustomersBus:   customersBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	brandapi.Routes(app, brandapi.Config{
		BrandBus:       brandBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	productcategoryapi.Routes(app, productcategoryapi.Config{
		ProductCategoryBus: productCategoryBus,
		AuthClient:         cfg.AuthClient,
		Log:                cfg.Log,
		PermissionsBus:     permissionsBus,
	})

	warehouseapi.Routes(app, warehouseapi.Config{
		WarehouseBus:   warehouseBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	// Permissions endpoints
	roleapi.Routes(app, roleapi.Config{
		Log:            cfg.Log,
		RoleBus:        roleBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	pageapi.Routes(app, pageapi.Config{
		Log:            cfg.Log,
		PageBus:        pageBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	rolepageapi.Routes(app, rolepageapi.Config{
		Log:            cfg.Log,
		RolePageBus:    rolePageBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	userroleapi.Routes(app, userroleapi.Config{
		Log:            cfg.Log,
		UserRoleBus:    userRoleBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	tableaccessapi.Routes(app, tableaccessapi.Config{
		Log:            cfg.Log,
		TableAccessBus: tableAccessBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	productapi.Routes(app, productapi.Config{
		ProductBus:     productBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	physicalattributeapi.Routes(app, physicalattributeapi.Config{
		PhysicalAttributeBus: physicalAttributeBus,
		AuthClient:           cfg.AuthClient,
		Log:                  cfg.Log,
		PermissionsBus:       permissionsBus,
	})

	productcostapi.Routes(app, productcostapi.Config{
		ProductCostBus: productCostBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	purchaseorderlineitemstatusapi.Routes(app, purchaseorderlineitemstatusapi.Config{
		PurchaseOrderLineItemStatusBus: purchaseOrderLineItemStatusBus,
		AuthClient:                     cfg.AuthClient,
		Log:                            cfg.Log,
		PermissionsBus:                 permissionsBus,
	})

	purchaseorderstatusapi.Routes(app, purchaseorderstatusapi.Config{
		PurchaseOrderStatusBus: purchaseOrderStatusBus,
		AuthClient:             cfg.AuthClient,
		Log:                    cfg.Log,
		PermissionsBus:         permissionsBus,
	})

	supplierapi.Routes(app, supplierapi.Config{
		SupplierBus:    supplierBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	costhistoryapi.Routes(app, costhistoryapi.Config{
		CostHistoryBus: costHistoryBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	supplierproductapi.Routes(app, supplierproductapi.Config{
		SupplierProductBus: supplierProductBus,
		AuthClient:         cfg.AuthClient,
		Log:                cfg.Log,
		PermissionsBus:     permissionsBus,
	})

	purchaseorderapi.Routes(app, purchaseorderapi.Config{
		PurchaseOrderBus: purchaseOrderBus,
		AuthClient:       cfg.AuthClient,
		Log:              cfg.Log,
		PermissionsBus:   permissionsBus,
	})

	purchaseorderlineitemapi.Routes(app, purchaseorderlineitemapi.Config{
		PurchaseOrderLineItemBus: purchaseOrderLineItemBus,
		AuthClient:               cfg.AuthClient,
		Log:                      cfg.Log,
		PermissionsBus:           permissionsBus,
	})

	metricsapi.Routes(app, metricsapi.Config{
		Log:            cfg.Log,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
		MetricsBus:     metricsBus,
	})

	lottrackingsapi.Routes(app, lottrackingsapi.Config{
		LotTrackingsBus: lotTrackingsBus,
		AuthClient:      cfg.AuthClient,
		Log:             cfg.Log,
		PermissionsBus:  permissionsBus,
	})

	zoneapi.Routes(app, zoneapi.Config{
		ZoneBus:        zoneBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	inventorylocationapi.Routes(app, inventorylocationapi.Config{
		InventoryLocationBus: inventoryLocationBus,
		AuthClient:           cfg.AuthClient,
		Log:                  cfg.Log,
		PermissionsBus:       permissionsBus,
	})

	inventoryitemapi.Routes(app, inventoryitemapi.Config{
		InventoryItemBus: inventoryItemBus,
		AuthClient:       cfg.AuthClient,
		Log:              cfg.Log,
		PermissionsBus:   permissionsBus,
	})

	inspectionapi.Routes(app, inspectionapi.Config{
		InspectionBus:  inspectionBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	serialnumberapi.Routes(app, serialnumberapi.Config{
		SerialNumberBus: serialNumberBus,
		AuthClient:      cfg.AuthClient,
		Log:             cfg.Log,
		PermissionsBus:  permissionsBus,
	})

	inventorytransactionapi.Routes(app, inventorytransactionapi.Config{
		InventoryTransactionBus: inventoryTransactionBus,
		AuthClient:              cfg.AuthClient,
		Log:                     cfg.Log,
		PermissionsBus:          permissionsBus,
	})

	inventoryadjustmentapi.Routes(app, inventoryadjustmentapi.Config{
		InventoryAdjustmentBus: inventoryAdjustmentBus,
		AuthClient:             cfg.AuthClient,
		Log:                    cfg.Log,
		PermissionsBus:         permissionsBus,
	})

	transferorderapi.Routes(app, transferorderapi.Config{
		TransferOrderBus: transferOrderBus,
		AuthClient:       cfg.AuthClient,
		Log:              cfg.Log,
		PermissionsBus:   permissionsBus,
	})

	orderfulfillmentstatusapi.Routes(app, orderfulfillmentstatusapi.Config{
		Log:                       cfg.Log,
		OrderFulfillmentStatusBus: orderFulfillmentStatusBus,
		AuthClient:                cfg.AuthClient,
		PermissionsBus:            permissionsBus,
	})

	lineitemfulfillmentstatusapi.Routes(app, lineitemfulfillmentstatusapi.Config{
		Log:                          cfg.Log,
		LineItemFulfillmentStatusBus: lineItemFulfillmentStatusBus,
		AuthClient:                   cfg.AuthClient,
		PermissionsBus:               permissionsBus,
	})

	ordersapi.Routes(app, ordersapi.Config{
		Log:            cfg.Log,
		OrderBus:       ordersBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	orderlineitemsapi.Routes(app, orderlineitemsapi.Config{
		Log:               cfg.Log,
		OrderLineItemsBus: orderLineItemsBus,
		AuthClient:        cfg.AuthClient,
		PermissionsBus:    permissionsBus,
	})

	// data
	dataapi.Routes(app, dataapi.Config{
		Log:            cfg.Log,
		ConfigStore:    configStore,
		TableStore:     tableStore,
		PageActionApp:  pageactionapp.NewApp(pageActionBus),
		PageConfigApp:  pageconfigapp.NewApp(pageConfigBus),
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	// config
	formapi.Routes(app, formapi.Config{
		Log:            cfg.Log,
		FormBus:        formBus,
		FormFieldBus:   formFieldBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	formfieldapi.Routes(app, formfieldapi.Config{
		Log:            cfg.Log,
		FormFieldBus:   formFieldBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	pageactionapi.Routes(app, pageactionapi.Config{
		Log:            cfg.Log,
		PageActionBus:  pageActionBus,
		DB:             cfg.DB,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	pageconfigapi.Routes(app, pageconfigapi.Config{
		Log:            cfg.Log,
		PageConfigBus:  pageConfigBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	pagecontentapi.Routes(app, pagecontentapi.Config{
		Log:            cfg.Log,
		PageContentBus: pageContentBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	// formdata - dynamic multi-entity operations
	// Build registry with entity registrations
	formDataRegistry, err := buildFormDataRegistry(
		userapp.NewApp(a.UserBus),
		assetapp.NewApp(assetBus),
		roleapp.NewApp(roleBus),
		pageapp.NewApp(pageBus),
		rolepageapp.NewApp(rolePageBus),
		tableaccessapp.NewApp(tableAccessBus),
		userroleappimport.NewApp(userRoleBus),
		contactinfosapp.NewApp(contactInfosBus),
		assetconditionapp.NewApp(assetConditionBus),
		assettypeapp.NewApp(assetTypeBus),
		fulfillmentstatusapp.NewApp(fulfillmentStatusBus),
		tagapp.NewApp(tagBus),
		assettagapp.NewApp(assetTagBus),
		validassetapp.NewApp(validAssetBus),
		userassetapp.NewApp(userAssetBus),
		approvalstatusapp.NewApp(approvalStatusBus),
		cityapp.NewApp(cityBus),
		streetapp.NewApp(streetBus),
		timezoneapp.NewApp(timezoneBus),
		commentapp.NewApp(userApprovalCommentBus),
		approvalapp.NewApp(userApprovalStatusBus),
		reportstoapp.NewApp(reportsToBus),
		officeapp.NewApp(officeBus),
		homeapp.NewApp(homeBus),
		titleapp.NewApp(titleBus),
		inspectionapp.NewApp(inspectionBus),
		inventoryadjustmentapp.NewApp(inventoryAdjustmentBus),
		inventorylocationapp.NewApp(inventoryLocationBus),
		inventorytransactionapp.NewApp(inventoryTransactionBus),
		serialnumberapp.NewApp(serialNumberBus),
		transferorderapp.NewApp(transferOrderBus),
		warehouseapp.NewApp(warehouseBus),
		zoneapp.NewApp(zoneBus),
		inventoryitemapp.NewApp(inventoryItemBus),
		lottrackingsapp.NewApp(lotTrackingsBus),
		purchaseorderlineitemstatusapp.NewApp(purchaseOrderLineItemStatusBus),
		purchaseorderstatusapp.NewApp(purchaseOrderStatusBus),
		purchaseorderapp.NewApp(purchaseOrderBus),
		purchaseorderlineitemapp.NewApp(purchaseOrderLineItemBus),
		supplierapp.NewApp(supplierBus),
		supplierproductapp.NewApp(supplierProductBus),
		brandapp.NewApp(brandBus),
		costhistoryapp.NewApp(costHistoryBus),
		metricsapp.NewApp(metricsBus),
		physicalattributeapp.NewApp(physicalAttributeBus),
		productcategoryapp.NewApp(productCategoryBus),
		productcostapp.NewApp(productCostBus),
		productapp.NewApp(productBus),
		customersapp.NewApp(customersBus),
		orderlineitemsapp.NewApp(orderLineItemsBus),
		ordersapp.NewApp(ordersBus),
		lineitemfulfillmentstatusapp.NewApp(lineItemFulfillmentStatusBus),
		orderfulfillmentstatusapp.NewApp(orderFulfillmentStatusBus),
		formapp.NewApp(formBus),
		formfieldapp.NewApp(formFieldBus),
	)
	if err != nil {
		cfg.Log.Error(context.Background(), "failed to build formdata registry", "error", err)
		// Continue without formdata support rather than failing startup
	} else {
		// Initialize formdata app and routes
		formDataApp := formdataapp.NewApp(formDataRegistry, cfg.DB, formBus, formFieldBus)
		formDataApp.SetEventPublisher(eventPublisher)

		formdataapi.Routes(app, formdataapi.Config{
			FormdataApp:    formDataApp,
			AuthClient:     cfg.AuthClient,
			PermissionsBus: permissionsBus,
		})

		cfg.Log.Info(context.Background(), "formdata routes initialized",
			"entities", len(formDataRegistry.ListEntities()))
	}

}
