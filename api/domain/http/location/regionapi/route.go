package regionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/location/regionapp"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Constants for the regionapi package.
type Config struct {
	Log        *logger.Logger
	RegionBus  *regionbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(regionapp.NewApp(cfg.RegionBus))

	app.HandlerFunc(http.MethodGet, version, "/location/regions", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/location/regions/{region_id}", api.queryByID, authen)
}
