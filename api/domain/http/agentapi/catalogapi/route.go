package catalogapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the catalog API routes.
type Config struct {
	AuthClient *authclient.Client
}

// Routes registers the agent catalog API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI()
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/agent/catalog", api.queryCatalog, authen)
}
