package ruleapi

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

// Config holds the dependencies for the rule API routes.
type Config struct {
	Log            *logger.Logger
	WorkflowBus    *workflow.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

// RouteTable is the table name used for permission checks.
const RouteTable = "workflow.automation_rules"

// Routes registers the rule API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg.Log, cfg.WorkflowBus)
	authen := mid.Authenticate(cfg.AuthClient)

	// List rules - requires read permission
	app.HandlerFunc(http.MethodGet, version, "/workflow/rules", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// Get single rule - requires read permission
	app.HandlerFunc(http.MethodGet, version, "/workflow/rules/{id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// Create rule - requires create permission
	app.HandlerFunc(http.MethodPost, version, "/workflow/rules", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	// Update rule - requires update permission
	app.HandlerFunc(http.MethodPut, version, "/workflow/rules/{id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	// Delete rule (soft-delete) - requires delete permission
	app.HandlerFunc(http.MethodDelete, version, "/workflow/rules/{id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))

	// Toggle active - requires update permission
	app.HandlerFunc(http.MethodPatch, version, "/workflow/rules/{id}/active", api.toggleActive, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	// ============================================================
	// Rule Actions (Phase 4C)
	// ============================================================

	// List actions for a rule - requires read permission
	app.HandlerFunc(http.MethodGet, version, "/workflow/rules/{id}/actions", api.queryActions, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// Create action for a rule - requires create permission
	app.HandlerFunc(http.MethodPost, version, "/workflow/rules/{id}/actions", api.createAction, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	// Update action - requires update permission
	app.HandlerFunc(http.MethodPut, version, "/workflow/rules/{id}/actions/{action_id}", api.updateAction, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	// Delete action (soft-delete) - requires delete permission
	app.HandlerFunc(http.MethodDelete, version, "/workflow/rules/{id}/actions/{action_id}", api.deleteAction, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))

	// ============================================================
	// Rule Validation (Phase 4C)
	// ============================================================

	// Validate rule configuration - uses Read permission (read-only operation)
	app.HandlerFunc(http.MethodPost, version, "/workflow/rules/{id}/validate", api.validateRule, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// ============================================================
	// Rule Testing/Simulation (Phase 5)
	// ============================================================

	// Test rule with sample data - uses Read permission (read-only simulation)
	app.HandlerFunc(http.MethodPost, version, "/workflow/rules/{id}/test", api.testRule, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	// List execution history for a rule - uses Read permission
	app.HandlerFunc(http.MethodGet, version, "/workflow/rules/{id}/executions", api.queryRuleExecutions, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
}
