// Package all binds all the routes into the specified app.
package all

import (
	"time"

	"github.com/timmaaaz/ichor/api/domain/http/checkapi"
	"github.com/timmaaaz/ichor/api/domain/http/homeapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/countryapi"
	"github.com/timmaaaz/ichor/api/domain/http/location/regionapi"
	"github.com/timmaaaz/ichor/api/domain/http/productapi"
	"github.com/timmaaaz/ichor/api/domain/http/rawapi"
	"github.com/timmaaaz/ichor/api/domain/http/tranapi"
	"github.com/timmaaaz/ichor/api/domain/http/userapi"
	"github.com/timmaaaz/ichor/api/domain/http/vproductapi"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/homebus/stores/homedb"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus/stores/countrydb"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus/stores/regiondb"
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
}
