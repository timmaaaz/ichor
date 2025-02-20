package streetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	StreetBus  *streetbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(streetapp.NewApp(cfg.StreetBus))
	app.HandlerFunc(http.MethodGet, version, "/location/streets", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/location/streets/{street_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/location/streets", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/location/streets/{street_id}", api.update, authen)
	app.HandlerFunc(http.MethodDelete, version, "/location/streets/{street_id}", api.delete, authen)
}
