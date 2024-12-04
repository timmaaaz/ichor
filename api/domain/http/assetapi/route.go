package assetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/assetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	AssetBus   *assetbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(assetapp.NewApp(cfg.AssetBus))
	app.HandlerFunc(http.MethodGet, version, "/assets", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/assets/{asset_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/assets", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/assets/{asset_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/assets/{asset_id}", api.delete, authen, ruleAdmin)
}
