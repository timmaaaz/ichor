package edgeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the edge API routes.
type Config struct {
	Log            *logger.Logger
	WorkflowBus    *workflow.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

// RouteTable is the table name used for permission checks.
// Uses the same table as rules since edges are part of rule configuration.
const RouteTable = "workflow.automation_rules"

// Routes registers the edge API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg.Log, cfg.WorkflowBus)
	authen := mid.Authenticate(cfg.AuthClient)

	// List edges for a rule - requires read permission
	app.HandlerFunc(http.MethodGet, version, "/workflow/rules/{ruleID}/edges", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// Get single edge - requires read permission
	app.HandlerFunc(http.MethodGet, version, "/workflow/rules/{ruleID}/edges/{edgeID}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// Create edge - requires create permission
	app.HandlerFunc(http.MethodPost, version, "/workflow/rules/{ruleID}/edges", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	// Delete single edge - requires delete permission
	app.HandlerFunc(http.MethodDelete, version, "/workflow/rules/{ruleID}/edges/{edgeID}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))

	// Delete all edges for a rule - requires delete permission
	// Note: Using a different path to avoid conflict with single edge delete
	app.HandlerFunc(http.MethodDelete, version, "/workflow/rules/{ruleID}/edges-all", api.deleteAll, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
