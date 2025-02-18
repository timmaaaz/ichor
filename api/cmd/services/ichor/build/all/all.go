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
	"github.com/timmaaaz/ichor/api/domain/http/location/officeapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/reportstoapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/status/approvalapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/status/commentapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/titleapi"

	"github.com/timmaaaz/ichor/api/domain/http/assets/fulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/checkapi"
	"github.com/timmaaaz/ichor/api/domain/http/homeapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/cityapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/countryapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/regionapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/streetapi"
	"github.com/timmaaaz/ichor/api/domain/http/productapi"
	"github.com/timmaaaz/ichor/api/domain/http/rawapi"
	"github.com/timmaaaz/ichor/api/domain/http/tranapi"
	"github.com/timmaaaz/ichor/api/domain/http/users/userapi"
	"github.com/timmaaaz/ichor/api/domain/http/vproductapi"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus/stores/approvalstatusdb"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	validassetdb "github.com/timmaaaz/ichor/business/domain/assets/validassetbus/stores/assetdb"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus/stores/contactinfodb"

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

	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus/stores/assettypedb"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	fulfillmentstatusdb "github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus/stores"
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
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus/stores/approvaldb"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus/stores/commentdb"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/domain/vproductbus"
	"github.com/timmaaaz/ichor/business/domain/vproductbus/stores/vproductdb"
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
	productBus := productbus.NewBusiness(cfg.Log, userBus, delegate, productdb.NewStore(cfg.Log, cfg.DB))
	homeBus := homebus.NewBusiness(cfg.Log, userBus, delegate, homedb.NewStore(cfg.Log, cfg.DB))
	vproductBus := vproductbus.NewBusiness(vproductdb.NewStore(cfg.Log, cfg.DB))
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

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homeapi.Routes(app, homeapi.Config{
		Log:        cfg.Log,
		UserBus:    userBus,
		HomeBus:    homeBus,
		AuthClient: cfg.AuthClient,
	})

	productapi.Routes(app, productapi.Config{
		Log:        cfg.Log,
		UserBus:    userBus,
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
	})

	rawapi.Routes(app)

	tranapi.Routes(app, tranapi.Config{
		Log:        cfg.Log,
		DB:         cfg.DB,
		UserBus:    userBus,
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
	})

	userapi.Routes(app, userapi.Config{
		Log:        cfg.Log,
		UserBus:    userBus,
		AuthClient: cfg.AuthClient,
	})

	vproductapi.Routes(app, vproductapi.Config{
		Log:         cfg.Log,
		UserBus:     userBus,
		VProductBus: vproductBus,
		AuthClient:  cfg.AuthClient,
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
		ValidAssetBus: validAssetBus,
		AuthClient:    cfg.AuthClient,
		Log:           cfg.Log,
	})

	tagapi.Routes(app, tagapi.Config{
		TagBus:     tagBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	titleapi.Routes(app, titleapi.Config{
		TitleBus:   titleBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

	assettagapi.Routes(app, assettagapi.Config{
		AssetTagBus: assetTagBus,
		AuthClient:  cfg.AuthClient,
		Log:         cfg.Log,
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

	contactinfoapi.Routes(app, contactinfoapi.Config{
		ContactInfoBus: contactInfoBus,
		AuthClient:     cfg.AuthClient,
		Log:            cfg.Log,
	})
}
