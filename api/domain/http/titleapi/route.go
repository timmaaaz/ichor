package titleapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/titleapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/titlebus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers
type Config struct {
	Log        *logger.Logger
	TitleBus   *titlebus.Business
	AuthClient *authclient.Client
}

// Routes adds routes to the group
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(titleapp.NewApp(cfg.TitleBus))

	app.HandlerFunc(http.MethodGet, version, "/titles", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/titles/{title_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/titles", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/titles/{title_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/titles/{title_id}", api.delete, authen, ruleAdmin)
}
