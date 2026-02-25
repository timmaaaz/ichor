// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"time"

	"github.com/timmaaaz/ichor/api/domain/http/assets/approvalstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assetapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assettagapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/tagapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/userassetapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/validassetapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/contactinfosapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/roleapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/tableaccessapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/userroleapi"
	"github.com/timmaaaz/ichor/api/domain/http/config/settingsapi"
	"github.com/timmaaaz/ichor/api/domain/http/dataapi"
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
	"github.com/timmaaaz/ichor/api/domain/http/procurement/supplierapi"
	"github.com/timmaaaz/ichor/api/domain/http/procurement/supplierproductapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/brandapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/costhistoryapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/metricsapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/physicalattributeapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/productapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/productcategoryapi"
	"github.com/timmaaaz/ichor/api/domain/http/products/productcostapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/customersapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/lineitemfulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/orderfulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/orderlineitemsapi"
	"github.com/timmaaaz/ichor/api/domain/http/sales/ordersapi"

	"github.com/timmaaaz/ichor/api/domain/http/assets/assetconditionapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assettypeapi"

	"github.com/timmaaaz/ichor/api/domain/http/assets/fulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/checkapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/userapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/cityapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/countryapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/regionapi"
	"github.com/timmaaaz/ichor/api/domain/http/geography/streetapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/approvalapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/commentapi"
	"github.com/timmaaaz/ichor/api/domain/http/hr/homeapi"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"

	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus/stores/approvalstatusdb"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
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
	"github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"
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

	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus/stores/assetconditiondb"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus/stores/assettypedb"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	fulfillmentstatusdb "github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus/stores"

	"github.com/timmaaaz/ichor/business/domain/products/productbus"

	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus/stores/citydb"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus/stores/countrydb"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus/stores/regiondb"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	streetdb "github.com/timmaaaz/ichor/business/domain/geography/streetbus/stores/streetdb"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
	"github.com/timmaaaz/ichor/business/domain/hr/homebus/stores/homedb"

	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/stores/settingsdb"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
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
	userApprovalCommentBus := commentbus.NewBusiness(cfg.Log, delegate, a.UserBus, commentdb.NewStore(cfg.Log, cfg.DB))

	homeBus := homebus.NewBusiness(cfg.Log, a.UserBus, delegate, homedb.NewStore(cfg.Log, cfg.DB))
	countryBus := countrybus.NewBusiness(cfg.Log, delegate, countrydb.NewStore(cfg.Log, cfg.DB))
	regionBus := regionbus.NewBusiness(cfg.Log, delegate, regiondb.NewStore(cfg.Log, cfg.DB))
	cityBus := citybus.NewBusiness(cfg.Log, delegate, citydb.NewStore(cfg.Log, cfg.DB))
	streetBus := streetbus.NewBusiness(cfg.Log, delegate, streetdb.NewStore(cfg.Log, cfg.DB))

	approvalStatusBus := approvalstatusbus.NewBusiness(cfg.Log, delegate, approvalstatusdb.NewStore(cfg.Log, cfg.DB))
	fulfillmentStatusBus := fulfillmentstatusbus.NewBusiness(cfg.Log, delegate, fulfillmentstatusdb.NewStore(cfg.Log, cfg.DB))
	assetConditionBus := assetconditionbus.NewBusiness(cfg.Log, delegate, assetconditiondb.NewStore(cfg.Log, cfg.DB))

	assetTypeBus := assettypebus.NewBusiness(cfg.Log, delegate, assettypedb.NewStore(cfg.Log, cfg.DB))
	validAssetBus := validassetbus.NewBusiness(cfg.Log, delegate, validassetdb.NewStore(cfg.Log, cfg.DB))
	tagBus := tagbus.NewBusiness(cfg.Log, delegate, tagdb.NewStore(cfg.Log, cfg.DB))
	assetTagBus := assettagbus.NewBusiness(cfg.Log, delegate, assettagdb.NewStore(cfg.Log, cfg.DB))

	titlebus := titlebus.NewBusiness(cfg.Log, delegate, titledb.NewStore(cfg.Log, cfg.DB))

	reportsToBus := reportstobus.NewBusiness(cfg.Log, delegate, reportstodb.NewStore(cfg.Log, cfg.DB))

	officeBus := officebus.NewBusiness(cfg.Log, delegate, officedb.NewStore(cfg.Log, cfg.DB))
	userAssetBus := userassetbus.NewBusiness(cfg.Log, delegate, userassetdb.NewStore(cfg.Log, cfg.DB))
	assetBus := assetbus.NewBusiness(cfg.Log, delegate, assetdb.NewStore(cfg.Log, cfg.DB))

	contactInfosBus := contactinfosbus.NewBusiness(cfg.Log, delegate, contactinfosdb.NewStore(cfg.Log, cfg.DB))
	customersBus := customersbus.NewBusiness(cfg.Log, delegate, customersdb.NewStore(cfg.Log, cfg.DB))

	brandBus := brandbus.NewBusiness(cfg.Log, delegate, branddb.NewStore(cfg.Log, cfg.DB))
	productCategoryBus := productcategorybus.NewBusiness(cfg.Log, delegate, productcategorydb.NewStore(cfg.Log, cfg.DB))
	productBus := productbus.NewBusiness(cfg.Log, delegate, productdb.NewStore(cfg.Log, cfg.DB))
	physicalAttributeBus := physicalattributebus.NewBusiness(cfg.Log, delegate, physicalattributedb.NewStore(cfg.Log, cfg.DB))

	warehouseBus := warehousebus.NewBusiness(cfg.Log, delegate, warehousedb.NewStore(cfg.Log, cfg.DB))
	zoneBus := zonebus.NewBusiness(cfg.Log, delegate, zonedb.NewStore(cfg.Log, cfg.DB))
	inventoryLocationBus := inventorylocationbus.NewBusiness(cfg.Log, delegate, inventorylocationdb.NewStore(cfg.Log, cfg.DB))
	inventoryItemBus := inventoryitembus.NewBusiness(cfg.Log, delegate, inventoryitemdb.NewStore(cfg.Log, cfg.DB))

	supplierBus := supplierbus.NewBusiness(cfg.Log, delegate, supplierdb.NewStore(cfg.Log, cfg.DB))
	supplierProductBus := supplierproductbus.NewBusiness(cfg.Log, delegate, supplierproductdb.NewStore(cfg.Log, cfg.DB))

	metricsBus := metricsbus.NewBusiness(cfg.Log, delegate, metricsdb.NewStore(cfg.Log, cfg.DB))
	inspectionBus := inspectionbus.NewBusiness(cfg.Log, delegate, inspectiondb.NewStore(cfg.Log, cfg.DB))

	lotTrackingsBus := lottrackingsbus.NewBusiness(cfg.Log, delegate, lottrackingsdb.NewStore(cfg.Log, cfg.DB))
	serialNumberBus := serialnumberbus.NewBusiness(cfg.Log, delegate, serialnumberdb.NewStore(cfg.Log, cfg.DB))
	settingsBus := settingsbus.NewBusiness(cfg.Log, delegate, settingsdb.NewStore(cfg.Log, cfg.DB))

	productCostBus := productcostbus.NewBusiness(cfg.Log, delegate, productcostdb.NewStore(cfg.Log, cfg.DB))
	costHistoryBus := costhistorybus.NewBusiness(cfg.Log, delegate, costhistorydb.NewStore(cfg.Log, cfg.DB))

	roleBus := rolebus.NewBusiness(cfg.Log, delegate, rolecache.NewStore(cfg.Log, roledb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))
	userRoleBus := userrolebus.NewBusiness(cfg.Log, delegate, userrolecache.NewStore(cfg.Log, userroledb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))
	tableAccessBus := tableaccessbus.NewBusiness(cfg.Log, delegate, tableaccesscache.NewStore(cfg.Log, tableaccessdb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))

	permissionsBus := permissionsbus.NewBusiness(cfg.Log, delegate, permissionscache.NewStore(cfg.Log, permissionsdb.NewStore(cfg.Log, cfg.DB), 60*time.Minute), userRoleBus, tableAccessBus, roleBus)

	inventoryTransactionBus := inventorytransactionbus.NewBusiness(cfg.Log, delegate, inventorytransactiondb.NewStore(cfg.Log, cfg.DB))
	inventoryAdjustmentBus := inventoryadjustmentbus.NewBusiness(cfg.Log, delegate, inventoryadjustmentdb.NewStore(cfg.Log, cfg.DB))
	transferOrderBus := transferorderbus.NewBusiness(cfg.Log, delegate, transferorderdb.NewStore(cfg.Log, cfg.DB))

	orderFulfillmentStatusBus := orderfulfillmentstatusbus.NewBusiness(cfg.Log, delegate, orderfulfillmentstatusdb.NewStore(cfg.Log, cfg.DB))
	lineItemFulfillmentStatusBus := lineitemfulfillmentstatusbus.NewBusiness(cfg.Log, delegate, lineitemfulfillmentstatusdb.NewStore(cfg.Log, cfg.DB))
	ordersBus := ordersbus.NewBusiness(cfg.Log, delegate, ordersdb.NewStore(cfg.Log, cfg.DB))
	orderLineItemsBus := orderlineitemsbus.NewBusiness(cfg.Log, delegate, orderlineitemsdb.NewStore(cfg.Log, cfg.DB))

	configStore := tablebuilder.NewConfigStore(cfg.Log, cfg.DB)
	tableStore := tablebuilder.NewStore(cfg.Log, cfg.DB)

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homeapi.Routes(app, homeapi.Config{
		UserBus:    a.UserBus,
		HomeBus:    homeBus,
		AuthClient: cfg.AuthClient,
	})

	userapi.Routes(app, userapi.Config{
		UserBus:    a.UserBus,
		AuthClient: cfg.AuthClient,
	})

	countryapi.Routes(app, countryapi.Config{
		CountryBus: countryBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	regionapi.Routes(app, regionapi.Config{
		RegionBus:  regionBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	cityapi.Routes(app, cityapi.Config{
		CityBus:    cityBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	streetapi.Routes(app, streetapi.Config{
		StreetBus:  streetBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	approvalstatusapi.Routes(app, approvalstatusapi.Config{
		ApprovalStatusBus: approvalStatusBus,
		AuthClient:        cfg.AuthClient,
		Log:               cfg.Log,
	})

	fulfillmentstatusapi.Routes(app, fulfillmentstatusapi.Config{
		FulfillmentStatusBus: fulfillmentStatusBus,
		AuthClient:           cfg.AuthClient,
		Log:                  cfg.Log,
	})

	assetconditionapi.Routes(app, assetconditionapi.Config{
		AssetConditionBus: assetConditionBus,
		AuthClient:        cfg.AuthClient,
		Log:               cfg.Log,
	})

	assettypeapi.Routes(app, assettypeapi.Config{
		AssetTypeBus: assetTypeBus,
		AuthClient:   cfg.AuthClient,
		Log:          cfg.Log,
	})

	validassetapi.Routes(app, validassetapi.Config{
		ValidAssetBus:  validAssetBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	tagapi.Routes(app, tagapi.Config{
		TagBus:     tagBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	assettagapi.Routes(app, assettagapi.Config{
		AssetTagBus: assetTagBus,
		AuthClient:  cfg.AuthClient,
		Log:         cfg.Log,
	})
	titleapi.Routes(app, titleapi.Config{
		TitleBus:   titlebus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	reportstoapi.Routes(app, reportstoapi.Config{
		ReportsToBus: reportsToBus,
		AuthClient:   cfg.AuthClient,
		Log:          cfg.Log,
	})

	officeapi.Routes(app, officeapi.Config{
		OfficeBus:  officeBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})
	userassetapi.Routes(app, userassetapi.Config{
		UserAssetBus: userAssetBus,
		AuthClient:   cfg.AuthClient,
		Log:          cfg.Log,
	})

	assetapi.Routes(app, assetapi.Config{
		AssetBus:   assetBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})
	approvalapi.Routes(app, approvalapi.Config{
		UserApprovalStatusBus: userApprovalStatusBus,
		AuthClient:            cfg.AuthClient,
		Log:                   cfg.Log,
	})

	commentapi.Routes(app, commentapi.Config{
		Log:                    cfg.Log,
		UserApprovalCommentBus: userApprovalCommentBus,
		AuthClient:             cfg.AuthClient,
	})

	contactinfosapi.Routes(app, contactinfosapi.Config{
		ContactInfosBus: contactInfosBus,
		AuthClient:      cfg.AuthClient,
		Log:             cfg.Log,
	})

	customersapi.Routes(app, customersapi.Config{
		CustomersBus:   customersBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	brandapi.Routes(app, brandapi.Config{
		BrandBus:   brandBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	productcategoryapi.Routes(app, productcategoryapi.Config{
		ProductCategoryBus: productCategoryBus,
		AuthClient:         cfg.AuthClient,
		Log:                cfg.Log,
	})

	warehouseapi.Routes(app, warehouseapi.Config{
		WarehouseBus:   warehouseBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})
	// Permissions routes
	roleapi.Routes(app, roleapi.Config{
		Log:            cfg.Log,
		RoleBus:        roleBus,
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
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	physicalattributeapi.Routes(app, physicalattributeapi.Config{
		PhysicalAttributeBus: physicalAttributeBus,
		AuthClient:           cfg.AuthClient,
		Log:                  cfg.Log,
	})

	productcostapi.Routes(app, productcostapi.Config{
		ProductCostBus: productCostBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
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
		SettingsBus:     settingsBus,
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

	settingsapi.Routes(app, settingsapi.Config{
		Log:            cfg.Log,
		SettingsBus:    settingsBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	// data
	dataapi.Routes(app, dataapi.Config{
		Log:            cfg.Log,
		ConfigStore:    configStore,
		TableStore:     tableStore,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})
}
