package assettypeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/assets/assettypebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log            *logger.Logger
	AssetTypeBus   *assettypebus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(assettypeapp.NewApp(cfg.AssetTypeBus))
	app.HandlerFunc(http.MethodGet, version, "/assets/assettypes", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "asset_types", permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/assets/assettypes/{asset_type_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "asset_types", permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/assets/assettypes", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "asset_types", permissionsbus.Actions.Create, auth.RuleAny))
	app.HandlerFunc(http.MethodPut, version, "/assets/assettypes/{asset_type_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "asset_types", permissionsbus.Actions.Update, auth.RuleAny))
	app.HandlerFunc(http.MethodDelete, version, "/assets/assettypes/{asset_type_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, "asset_types", permissionsbus.Actions.Delete, auth.RuleAny))
}
