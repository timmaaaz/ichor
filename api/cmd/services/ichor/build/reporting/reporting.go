// Package reporting binds the reporting domain set of routes into the specified app.
package reporting

import (
	"github.com/timmaaaz/ichor/api/domain/http/checkapi"

	"github.com/timmaaaz/ichor/api/sdk/http/mux"

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

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

}
