package alertapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the alert API routes.
type Config struct {
	Log            *logger.Logger
	AlertBus       *alertbus.Business
	UserBus        *userbus.Business
	RoleBus        *rolebus.Business
	UserRoleBus    *userrolebus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
	WorkflowQueue  *rabbitmq.WorkflowQueue
}

// RouteTable is the table name used for permission checks.
const RouteTable = "workflow.alerts"

// Routes registers the alert API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg.Log, cfg.AlertBus, cfg.UserBus, cfg.RoleBus, cfg.UserRoleBus, cfg.WorkflowQueue)
	authen := mid.Authenticate(cfg.AuthClient)

	// User endpoints - authentication only, business layer handles recipient filtering
	// Everyone can access alerts, but they only see alerts they're recipients of
	app.HandlerFunc(http.MethodGet, version, "/workflow/alerts/mine", api.queryMine, authen)

	// Bulk action endpoints (must come before /{id} routes to avoid path conflicts)
	app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/acknowledge-selected", api.acknowledgeSelected, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/acknowledge-all", api.acknowledgeAll, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/dismiss-selected", api.dismissSelected, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/dismiss-all", api.dismissAll, authen)

	// Test endpoint - creates a test alert for the authenticated user (for E2E WebSocket testing)
	app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/test", api.testAlert, authen)

	// Single alert endpoints
	app.HandlerFunc(http.MethodGet, version, "/workflow/alerts/{id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/{id}/acknowledge", api.acknowledge, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/alerts/{id}/dismiss", api.dismiss, authen)

	// Admin endpoint - requires read permission on workflow.alerts table
	app.HandlerFunc(http.MethodGet, version, "/workflow/alerts", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
}
