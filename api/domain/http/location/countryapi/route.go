package countryapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/location/countryapp"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	CountryBus *countrybus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(countryapp.NewApp(cfg.CountryBus))

	app.HandlerFunc(http.MethodGet, version, "/location/countries", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/location/countries/{country_id}", api.queryByID, authen)
}
