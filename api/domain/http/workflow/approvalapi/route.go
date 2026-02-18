package approvalapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the approval API routes.
type Config struct {
	Log            *logger.Logger
	ApprovalBus    *approvalrequestbus.Business
	UserRoleBus    *userrolebus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
	AsyncCompleter *temporal.AsyncCompleter
}

// RouteTable is the table name used for permission checks.
const RouteTable = "workflow.approval_requests"

// Routes registers the approval API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)

	// Approver endpoints - auth only, business layer filters by approver
	app.HandlerFunc(http.MethodGet, version, "/workflow/approvals/mine", api.queryMine, authen)

	// Resolve endpoint - auth only, handler checks approver/admin
	app.HandlerFunc(http.MethodPost, version, "/workflow/approvals/{id}/resolve", api.resolve, authen)

	// Single lookup
	app.HandlerFunc(http.MethodGet, version, "/workflow/approvals/{id}", api.queryByID, authen)

	// Admin endpoint - requires admin permission
	app.HandlerFunc(http.MethodGet, version, "/workflow/approvals", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
}
