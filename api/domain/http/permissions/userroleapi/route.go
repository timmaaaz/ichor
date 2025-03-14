package userroleapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/permissions/userroleapp.go"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log            *logger.Logger
	UserRoleBus    *userrolebus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	RouteTable = "user_roles"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(userroleapp.NewApp(cfg.UserRoleBus))

	app.HandlerFunc(http.MethodPost, version, "/permissions/user_roles", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodDelete, version, "/permissions/user_roles/{user_role_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodGet, version, "/permissions/user_roles", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/permissions/user_roles/{user_role_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
}
