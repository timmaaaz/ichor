package reportstoapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/users/reportstoapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log          *logger.Logger
	ReportsToBus *reportstobus.Business
	AuthClient   *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(reportstoapp.NewApp(cfg.ReportsToBus))
	app.HandlerFunc(http.MethodGet, version, "/users/reportsto", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/users/reportsto/{reports_to_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/users/reportsto", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/users/reportsto/{reports_to_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/users/reportsto/{reports_to_id}", api.delete, authen, ruleAdmin)
}
