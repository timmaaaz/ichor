package assettagapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/assets/assettagapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log         *logger.Logger
	AssetTagBus *assettagbus.Business
	AuthClient  *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(assettagapp.NewApp(cfg.AssetTagBus))
	app.HandlerFunc(http.MethodGet, version, "/assets/assettags", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/assets/assettags/{asset_tag_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/assets/assettags", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/assets/assettags/{asset_tag_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/assets/assettags/{asset_tag_id}", api.delete, authen, ruleAdmin)
}
