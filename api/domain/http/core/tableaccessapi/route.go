package tableaccessapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/core/tableaccessapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log            *logger.Logger
	TableAccessBus *tableaccessbus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	RouteTable = "table_access"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(tableaccessapp.NewApp(cfg.TableAccessBus))
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/core/table-access", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/core/table-access/{table_access_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/core/table-access", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodPut, version, "/core/table-access/{table_access_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodDelete, version, "/core/table-access/{table_access_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAdminOnly))

}
