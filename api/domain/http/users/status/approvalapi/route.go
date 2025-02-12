package approvalapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/users/status/approvalapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers
type Config struct {
	Log                   *logger.Logger
	UserApprovalStatusBus *approvalbus.Business
	AuthClient            *authclient.Client
}

// Routes adds routes to the group
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(approvalapp.NewApp(cfg.UserApprovalStatusBus))

	app.HandlerFunc(http.MethodGet, version, "/users/status/approvals", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/users/status/approvals/{user_approval_status_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/users/status/approvals", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/users/status/approvals/{user_approval_status_id}", api.update, authen)
	app.HandlerFunc(http.MethodDelete, version, "/users/status/approvals/{user_approval_status_id}", api.delete, authen)
}
