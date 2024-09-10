// Package all binds all the routes into the specified app.
package all

import (
	"time"

	"bitbucket.org/superiortechnologies/ichor/api/domain/http/authapi"
	"bitbucket.org/superiortechnologies/ichor/api/domain/http/checkapi"
	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/mux"
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

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	authapi.Routes(app, authapi.Config{
		UserBus: userBus,
		Auth:    cfg.Auth,
	})
}
