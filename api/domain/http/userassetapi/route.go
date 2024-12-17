package userassetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/userassetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log          *logger.Logger
	UserAssetBus *userassetbus.Business
	AuthClient   *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(userassetapp.NewApp(cfg.UserAssetBus))
	app.HandlerFunc(http.MethodGet, version, "/userassets", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/userassets/{user_asset_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/userassets", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/userassets/{user_asset_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/userassets/{user_asset_id}", api.delete, authen, ruleAdmin)
}
