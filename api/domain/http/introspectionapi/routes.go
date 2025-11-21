package introspectionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/introspectionapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/introspectionbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log              *logger.Logger
	IntrospectionBus *introspectionbus.Business
	AuthClient       *authclient.Client
	PermissionsBus   *permissionsbus.Business
}

// RouteTable is the table name used for permissions.
const RouteTable = "introspection.introspection"

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"
	api := newAPI(introspectionapp.NewApp(cfg.IntrospectionBus))
	authen := mid.Authenticate(cfg.AuthClient)

	// GET /v1/introspection/schemas
	app.HandlerFunc(http.MethodGet, version, "/introspection/schemas", api.querySchemas, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))

	// GET /v1/introspection/schemas/{schema}/tables
	app.HandlerFunc(http.MethodGet, version, "/introspection/schemas/{schema}/tables", api.queryTables, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))

	// GET /v1/introspection/tables/{schema}/{table}/columns
	app.HandlerFunc(http.MethodGet, version, "/introspection/tables/{schema}/{table}/columns", api.queryColumns, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))

	// GET /v1/introspection/tables/{schema}/{table}/relationships
	app.HandlerFunc(http.MethodGet, version, "/introspection/tables/{schema}/{table}/relationships", api.queryRelationships, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAdminOnly))
}
