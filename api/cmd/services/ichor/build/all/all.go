// Package all binds all the routes into the specified app.
package all

import (
	"time"

	"github.com/timmaaaz/ichor/api/domain/http/assets/approvalstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assetapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assetconditionapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assettagapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/assettypeapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/tagapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/userassetapi"
	"github.com/timmaaaz/ichor/api/domain/http/assets/validassetapi"
	"github.com/timmaaaz/ichor/api/domain/http/core/contactinfoapi"
	"github.com/timmaaaz/ichor/api/domain/http/finance/costhistoryapi"
	"github.com/timmaaaz/ichor/api/domain/http/finance/productcostapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/core/brandapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/core/physicalattributeapi"
	"github.com/timmaaaz/ichor/api/domain/http/inventory/core/productcategoryapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/officeapi"
	"github.com/timmaaaz/ichor/api/domain/http/lots/lottrackingapi"
	"github.com/timmaaaz/ichor/api/domain/http/permissions/roleapi"
	"github.com/timmaaaz/ichor/api/domain/http/permissions/tableaccessapi"
	"github.com/timmaaaz/ichor/api/domain/http/permissions/userroleapi"
	"github.com/timmaaaz/ichor/api/domain/http/quality/metricsapi"
	"github.com/timmaaaz/ichor/api/domain/http/supplier/supplierapi"
	"github.com/timmaaaz/ichor/api/domain/http/supplier/supplierproductapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/reportstoapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/status/approvalapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/status/commentapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/titleapi"
	"github.com/timmaaaz/ichor/api/domain/http/warehouse/warehouseapi"
	"github.com/timmaaaz/ichor/api/domain/http/warehouse/zoneapi"

	"github.com/timmaaaz/ichor/api/domain/http/assets/fulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/checkapi"
	"github.com/timmaaaz/ichor/api/domain/http/homeapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/cityapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/countryapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/regionapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/streetapi"

	"github.com/timmaaaz/ichor/api/domain/http/rawapi"

	"github.com/timmaaaz/ichor/api/domain/http/users/userapi"

	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus/stores/approvalstatusdb"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	validassetdb "github.com/timmaaaz/ichor/business/domain/assets/validassetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus/stores/contactinfodb"
	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus/stores/costhistorydb"
	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus"
	"github.com/timmaaaz/ichor/business/domain/finance/productcostbus/stores/productcostdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus/stores/branddb"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus/stores/physicalattributedb"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus/stores/productcategorydb"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus/stores/lottrackingdb"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus/stores/permissionscache"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus/stores/permissionsdb"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus/stores/rolecache"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus/stores/roledb"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus/stores/tableaccesscache"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus/stores/tableaccessdb"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus/stores/userrolecache"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus/stores/userroledb"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus/stores/metricsdb"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus/stores/supplierdb"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus/stores/supplierproductdb"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus/stores/warehousedb"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus/stores/zonedb"

	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus/stores/assetconditiondb"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus/store/assettagdb"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus/stores/tagdb"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus/stores/userassetdb"
	"github.com/timmaaaz/ichor/business/domain/location/officebus"
	"github.com/timmaaaz/ichor/business/domain/location/officebus/stores/officedb"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus/store/reportstodb"
	"github.com/timmaaaz/ichor/business/domain/users/titlebus"
	"github.com/timmaaaz/ichor/business/domain/users/titlebus/stores/titledb"

	productapi "github.com/timmaaaz/ichor/api/domain/http/inventory/core/productapi"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus/stores/assettypedb"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	fulfillmentstatusdb "github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus/stores"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/homebus/stores/homedb"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	inventoryproductdb "github.com/timmaaaz/ichor/business/domain/inventory/core/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	citydb "github.com/timmaaaz/ichor/business/domain/location/citybus/stores/citydb"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus/stores/countrydb"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus/stores/regiondb"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	streetdb "github.com/timmaaaz/ichor/business/domain/location/streetbus/stores/streetdb"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus/stores/approvaldb"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus/stores/commentdb"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() add {
	return add{}
}

type add struct{}

// Add implements the RouterAdder interface.
func (add) Add(app *web.App, cfg mux.Config) {

	// Construct the business domain packages we need here so we are using the
	// sames instances for the different set of domain apis.
	delegate := delegate.New(cfg.Log)
	userApprovalStatusBus := approvalbus.NewBusiness(cfg.Log, delegate, approvaldb.NewStore(cfg.Log, cfg.DB))
	userBus := userbus.NewBusiness(cfg.Log, delegate, userApprovalStatusBus, usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB), time.Minute))
	userApprovalCommentBus := commentbus.NewBusiness(cfg.Log, delegate, userBus, commentdb.NewStore(cfg.Log, cfg.DB))

	homeBus := homebus.NewBusiness(cfg.Log, userBus, delegate, homedb.NewStore(cfg.Log, cfg.DB))
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
	titleBus := titlebus.NewBusiness(cfg.Log, delegate, titledb.NewStore(cfg.Log, cfg.DB))
	reportsToBus := reportstobus.NewBusiness(cfg.Log, delegate, reportstodb.NewStore(cfg.Log, cfg.DB))
	officeBus := officebus.NewBusiness(cfg.Log, delegate, officedb.NewStore(cfg.Log, cfg.DB))
	userAssetBus := userassetbus.NewBusiness(cfg.Log, delegate, userassetdb.NewStore(cfg.Log, cfg.DB))
	assetBus := assetbus.NewBusiness(cfg.Log, delegate, assetdb.NewStore(cfg.Log, cfg.DB))

	contactInfoBus := contactinfobus.NewBusiness(cfg.Log, delegate, contactinfodb.NewStore(cfg.Log, cfg.DB))
	brandBus := brandbus.NewBusiness(cfg.Log, delegate, branddb.NewStore(cfg.Log, cfg.DB))
	productCategoryBus := productcategorybus.NewBusiness(cfg.Log, delegate, productcategorydb.NewStore(cfg.Log, cfg.DB))
	productBus := productbus.NewBusiness(cfg.Log, delegate, inventoryproductdb.NewStore(cfg.Log, cfg.DB))
	physicalAttributeBus := physicalattributebus.NewBusiness(cfg.Log, delegate, physicalattributedb.NewStore(cfg.Log, cfg.DB))

	productCostBus := productcostbus.NewBusiness(cfg.Log, delegate, productcostdb.NewStore(cfg.Log, cfg.DB))
	costHistoryBus := costhistorybus.NewBusiness(cfg.Log, delegate, costhistorydb.NewStore(cfg.Log, cfg.DB))

	warehouseBus := warehousebus.NewBusiness(cfg.Log, delegate, warehousedb.NewStore(cfg.Log, cfg.DB))
	zoneBus := zonebus.NewBusiness(cfg.Log, delegate, zonedb.NewStore(cfg.Log, cfg.DB))

	supplierBus := supplierbus.NewBusiness(cfg.Log, delegate, supplierdb.NewStore(cfg.Log, cfg.DB))
	supplierProductBus := supplierproductbus.NewBusiness(cfg.Log, delegate, supplierproductdb.NewStore(cfg.Log, cfg.DB))

	metricsBus := metricsbus.NewBusiness(cfg.Log, delegate, metricsdb.NewStore(cfg.Log, cfg.DB))

	lotTrackingBus := lottrackingbus.NewBusiness(cfg.Log, delegate, lottrackingdb.NewStore(cfg.Log, cfg.DB))

	roleBus := rolebus.NewBusiness(cfg.Log, delegate, rolecache.NewStore(cfg.Log, roledb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))
	userRoleBus := userrolebus.NewBusiness(cfg.Log, delegate, userrolecache.NewStore(cfg.Log, userroledb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))
	tableAccessBus := tableaccessbus.NewBusiness(cfg.Log, delegate, tableaccesscache.NewStore(cfg.Log, tableaccessdb.NewStore(cfg.Log, cfg.DB), 60*time.Minute))

	permissionsBus := permissionsbus.NewBusiness(cfg.Log, delegate, permissionscache.NewStore(cfg.Log, permissionsdb.NewStore(cfg.Log, cfg.DB), 60*time.Minute), userRoleBus, tableAccessBus, roleBus)

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homeapi.Routes(app, homeapi.Config{
		Log:            cfg.Log,
		UserBus:        userBus,
		HomeBus:        homeBus,
		AuthClient:     cfg.AuthClient,
		PermissionsBus: permissionsBus,
	})

	rawapi.Routes(app)

	userapi.Routes(app, userapi.Config{
		Log:            cfg.Log,
		UserBus:        userBus,
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

	contactinfoapi.Routes(app, contactinfoapi.Config{
		ContactInfoBus: contactInfoBus,
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

	lottrackingapi.Routes(app, lottrackingapi.Config{
		LotTrackingBus: lotTrackingBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})

	zoneapi.Routes(app, zoneapi.Config{
		ZoneBus:        zoneBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
		PermissionsBus: permissionsBus,
	})
}
