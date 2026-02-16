package configschemaapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the config schema API routes.
type Config struct {
	AuthClient *authclient.Client
}

// Routes registers the config schema API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI()
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/config/schemas/table-config", api.queryTableConfigSchema, authen)
	app.HandlerFunc(http.MethodGet, version, "/config/schemas/layout", api.queryLayoutSchema, authen)
	app.HandlerFunc(http.MethodGet, version, "/config/schemas/content-types", api.queryContentTypes, authen)
	app.HandlerFunc(http.MethodGet, version, "/config/schemas/page-action-types", api.queryPageActionTypes, authen)
}
