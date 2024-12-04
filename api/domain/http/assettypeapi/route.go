package assettypeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/assettypeapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log          *logger.Logger
	AssetTypeBus *assettypebus.Business
	AuthClient   *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(assettypeapp.NewApp(cfg.AssetTypeBus))
	app.HandlerFunc(http.MethodGet, version, "/assettypes", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/assettypes/{asset_type_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/assettypes", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/assettypes/{asset_type_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/assettypes/{asset_type_id}", api.delete, authen, ruleAdmin)
}
