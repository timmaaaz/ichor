// Package all binds all the routes into the specified app.
package all

import (
	"time"

	"github.com/timmaaaz/ichor/api/domain/http/approvalstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/assetconditionapi"
	"github.com/timmaaaz/ichor/api/domain/http/assettypeapi"
	"github.com/timmaaaz/ichor/api/domain/http/checkapi"
	"github.com/timmaaaz/ichor/api/domain/http/fulfillmentstatusapi"
	"github.com/timmaaaz/ichor/api/domain/http/homeapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/cityapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/countryapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/regionapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/streetapi"
	"github.com/timmaaaz/ichor/api/domain/http/productapi"
	"github.com/timmaaaz/ichor/api/domain/http/rawapi"
	"github.com/timmaaaz/ichor/api/domain/http/tranapi"
	"github.com/timmaaaz/ichor/api/domain/http/userapi"
	"github.com/timmaaaz/ichor/api/domain/http/vproductapi"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus/stores/approvalstatusdb"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	assetconditiondb "github.com/timmaaaz/ichor/business/domain/assetconditionbus/stores"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/assettypebus/stores/assettypedb"
	"github.com/timmaaaz/ichor/business/domain/fulfillmentstatusbus"
	fulfillmentstatusdb "github.com/timmaaaz/ichor/business/domain/fulfillmentstatusbus/stores"
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
	userBus := userbus.NewBusiness(cfg.Log, delegate, usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB), time.Minute))
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
	assetConditionBus := assetconditionbus.NewBusiness(cfg.Log, delegate, assetconditiondb.NewStore(cfg.Log, cfg.DB))

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

}
