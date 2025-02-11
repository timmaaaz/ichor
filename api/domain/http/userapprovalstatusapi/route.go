package userapprovalstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/userapprovalstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers
type Config struct {
	Log                   *logger.Logger
	UserApprovalStatusBus *userapprovalstatusbus.Business
	AuthClient            *authclient.Client
}

// Routes adds routes to the group
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(userapprovalstatusapp.NewApp(cfg.UserApprovalStatusBus))

	app.HandlerFunc(http.MethodGet, version, "/userapprovalstatus", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/userapprovalstatus/{user_approval_status_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/userapprovalstatus", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/userapprovalstatus/{user_approval_status_id}", api.update, authen)
	app.HandlerFunc(http.MethodDelete, version, "/userapprovalstatus/{user_approval_status_id}", api.delete, authen)
}
