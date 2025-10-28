package pageapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log            *logger.Logger
	PageBus        *pagebus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	RouteTable = "pages"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(pageapp.NewApp(cfg.PageBus))
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/core/pages", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodGet, version, "/core/pages/{page_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
	// QueryByUserID only requires authentication - any user can query their accessible pages
	app.HandlerFunc(http.MethodGet, version, "/core/pages/user/{user_id}", api.queryByUserID, authen)
	app.HandlerFunc(http.MethodPost, version, "/core/pages", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodPut, version, "/core/pages/{page_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodDelete, version, "/core/pages/{page_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAdminOnly))
}
