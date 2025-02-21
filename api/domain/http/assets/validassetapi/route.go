package validassetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log           *logger.Logger
	ValidAssetBus *validassetbus.Business
	AuthClient    *authclient.Client
	UserBus       *userbus.Business
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)
	// checkPermissions := mid.AuthorizeCheckPermissions(cfg.AuthClient, cfg.UserBus, auth.RuleAdminOnly)

	api := newAPI(validassetapp.NewApp(cfg.ValidAssetBus))
	app.HandlerFunc(http.MethodGet, version, "/assets/validassets", api.query, authen, mid.AuthorizeTable(cfg.AuthClient, "valid_assets", "read", auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/assets/validassets/{valid_asset_id}", api.queryByID, authen, mid.AuthorizeTable(cfg.AuthClient, "valid_assets", "read", auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/assets/validassets", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/assets/validassets/{valid_asset_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/assets/validassets/{valid_asset_id}", api.delete, authen, ruleAdmin)
}
