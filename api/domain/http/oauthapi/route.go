// api/domain/http/oauthapi/route.go
package oauthapi

import (
	"net/http"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the configuration for the OAuth app.
type Config struct {
	Auth            *auth.Auth
	Log             *logger.Logger
	TokenKey        string
	GoogleKey       string
	GoogleSecret    string
	Callback        string
	StoreKey        string
	UIAdminRedirect string
	UILoginRedirect string
	Environment     string
	EnableDevAuth   bool
	TokenExpiration time.Duration
}

// Routes adds the OAuth routes to the web.App.
func Routes(app *web.App, cfg Config) {
	api := newAPI(cfg)

	// Add OAuth routes using RawHandlerFunc
	app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}", api.authenticate)
	app.RawHandlerFunc(http.MethodGet, "", "/api/auth/{provider}/callback", api.authCallback)
	app.RawHandlerFunc(http.MethodGet, "", "/api/logout/{provider}", api.logout)
}
