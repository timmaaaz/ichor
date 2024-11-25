package approvalstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/approvalstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/approvalstatusbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers
type Config struct {
	Log               *logger.Logger
	ApprovalStatusBus *approvalstatusbus.Business
	AuthClient        *authclient.Client
}

// Routes adds routes to the group
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(approvalstatusapp.NewApp(cfg.ApprovalStatusBus))

	app.HandlerFunc(http.MethodGet, version, "/approvalstatus", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/approvalstatus/{approval_status_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/approvalstatus", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/approvalstatus/{approval_status_id}", api.update, authen)
	app.HandlerFunc(http.MethodDelete, version, "/approvalstatus/{approval_status_id}", api.delete, authen)
}
