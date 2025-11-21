package pageactionapi

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the configuration for the pageaction routes.
type Config struct {
	Log            *logger.Logger
	PageActionBus  *pageactionbus.Business
	DB             *sqlx.DB
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	RouteTablePageActions        = "config.page_actions"
	RouteTablePageActionButtons  = "config.page_action_buttons"
	RouteTablePageActionDropdowns = "config.page_action_dropdowns"
)

// Routes binds all the page action routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(pageactionapp.NewAppWithDB(cfg.PageActionBus, cfg.DB))
	authen := mid.Authenticate(cfg.AuthClient)

	// Query routes
	app.HandlerFunc(http.MethodGet, version, "/config/page-actions", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActions, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/config/page-actions/{action_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActions, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/config/page-configs/actions/{page_config_id}", api.queryByPageConfigID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActions, permissionsbus.Actions.Read, auth.RuleAny))

	// Button routes
	app.HandlerFunc(http.MethodPost, version, "/config/page-actions/buttons", api.createButton, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActionButtons, permissionsbus.Actions.Create, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodPut, version, "/config/page-actions/buttons/{action_id}", api.updateButton, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActionButtons, permissionsbus.Actions.Update, auth.RuleAdminOnly))

	// Dropdown routes
	app.HandlerFunc(http.MethodPost, version, "/config/page-actions/dropdowns", api.createDropdown, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActionDropdowns, permissionsbus.Actions.Create, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodPut, version, "/config/page-actions/dropdowns/{action_id}", api.updateDropdown, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActionDropdowns, permissionsbus.Actions.Update, auth.RuleAdminOnly))

	// Separator routes
	app.HandlerFunc(http.MethodPost, version, "/config/page-actions/separators", api.createSeparator, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActions, permissionsbus.Actions.Create, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodPut, version, "/config/page-actions/separators/{action_id}", api.updateSeparator, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActions, permissionsbus.Actions.Update, auth.RuleAdminOnly))

	// Delete route (works for all action types)
	app.HandlerFunc(http.MethodDelete, version, "/config/page-actions/{action_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActions, permissionsbus.Actions.Delete, auth.RuleAdminOnly))

	// Batch operations
	app.HandlerFunc(http.MethodPost, version, "/config/page-configs/actions/batch/{page_config_id}", api.batchCreate, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePageActions, permissionsbus.Actions.Create, auth.RuleAdminOnly))
}
