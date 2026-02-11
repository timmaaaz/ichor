package workflowsaveapi

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the workflow save API routes.
type Config struct {
	Log            *logger.Logger
	DB             *sqlx.DB
	WorkflowBus    *workflow.Business
	Delegate       *delegate.Delegate
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
	ActionRegistry *workflow.ActionRegistry
}

// RouteTable is the table name used for permission checks.
const RouteTable = "workflow.automation_rules"

// Routes registers the workflow save API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	workflowApp := workflowsaveapp.NewApp(cfg.Log, cfg.DB, cfg.WorkflowBus, cfg.Delegate, cfg.ActionRegistry)
	api := newAPI(workflowApp)
	authen := mid.Authenticate(cfg.AuthClient)

	// Update workflow (full save) - requires update permission
	app.HandlerFunc(http.MethodPut, version, "/workflow/rules/{id}/full", api.save, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAdminOnly))

	// Create workflow (full save) - requires create permission
	app.HandlerFunc(http.MethodPost, version, "/workflow/rules/full", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))

	// Action type discovery - returns all registered action types with output ports
	if cfg.ActionRegistry != nil {
		discovery := newDiscoveryAPI(cfg.ActionRegistry)
		app.HandlerFunc(http.MethodGet, version, "/workflow/action-types", discovery.queryActionTypes, authen)
	}
}
