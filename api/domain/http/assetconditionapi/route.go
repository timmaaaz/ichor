package assetconditionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/assetconditionapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log               *logger.Logger
	AssetConditionBus *assetconditionbus.Business
	AuthClient        *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(assetconditionapp.NewApp(cfg.AssetConditionBus))
	app.HandlerFunc(http.MethodGet, version, "/assetconditions", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/assetconditions/{asset_condition_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/assetconditions", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/assetconditions/{asset_condition_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/assetconditions/{asset_condition_id}", api.delete, authen, ruleAdmin)
}
