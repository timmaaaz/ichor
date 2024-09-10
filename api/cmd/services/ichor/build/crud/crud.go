// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"time"

	"bitbucket.org/superiortechnologies/ichor/api/domain/http/checkapi"
	"bitbucket.org/superiortechnologies/ichor/api/domain/http/homeapi"
	"bitbucket.org/superiortechnologies/ichor/api/domain/http/location/countryapi"
	"bitbucket.org/superiortechnologies/ichor/api/domain/http/productapi"
	"bitbucket.org/superiortechnologies/ichor/api/domain/http/tranapi"
	"bitbucket.org/superiortechnologies/ichor/api/domain/http/userapi"
	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/mux"
	"bitbucket.org/superiortechnologies/ichor/business/domain/homebus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/homebus/stores/homedb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus/stores/countrydb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus/stores/productdb"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus/stores/usercache"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus/stores/userdb"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/delegate"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
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
	countryBus := countrybus.NewBusiness(cfg.Log, delegate, countrydb.NewStore(cfg.Log, cfg.DB))

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homeapi.Routes(app, homeapi.Config{
		UserBus:    userBus,
		HomeBus:    homeBus,
		AuthClient: cfg.AuthClient,
	})

	productapi.Routes(app, productapi.Config{
		UserBus:    userBus,
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
	})

	tranapi.Routes(app, tranapi.Config{
		UserBus:    userBus,
		ProductBus: productBus,
		Log:        cfg.Log,
		AuthClient: cfg.AuthClient,
		DB:         cfg.DB,
	})

	userapi.Routes(app, userapi.Config{
		UserBus:    userBus,
		AuthClient: cfg.AuthClient,
	})

	countryapi.Routes(app, countryapi.Config{
		CountryBus: countryBus,
		AuthClient: cfg.AuthClient,
		Log:        cfg.Log,
	})

}
