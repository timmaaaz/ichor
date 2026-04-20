package scenarioapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/scenarioapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/scenariobus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// RouteTable is the table_access table name for scenario authorization.
const RouteTable = "inventory.scenarios"

// Config carries the dependencies for the scenario API.
type Config struct {
	Log            *logger.Logger
	ScenarioBus    *scenariobus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

// Routes registers the scenario API routes. All routes require ADMIN access
// (auth.RuleAdminOnly) — scenario management is an admin-only operation
// for floor testing setup. No supervisor-only split is required at this stage.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	authorize := func(action permissionsbus.Action) web.MidFunc {
		return mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, action, auth.RuleAdminOnly)
	}

	scenarioApp := scenarioapp.NewApp(cfg.ScenarioBus)
	a := newAPI(scenarioApp)

	// Scenario CRUD
	app.HandlerFunc(http.MethodGet, version, "/scenarios", a.query, authen,
		authorize(permissionsbus.Actions.Read))

	app.HandlerFunc(http.MethodGet, version, "/scenarios/{id}", a.queryByID, authen,
		authorize(permissionsbus.Actions.Read))

	app.HandlerFunc(http.MethodPost, version, "/scenarios", a.create, authen,
		authorize(permissionsbus.Actions.Create))

	app.HandlerFunc(http.MethodPut, version, "/scenarios/{id}", a.update, authen,
		authorize(permissionsbus.Actions.Update))

	app.HandlerFunc(http.MethodDelete, version, "/scenarios/{id}", a.delete, authen,
		authorize(permissionsbus.Actions.Delete))

	// Active scenario query
	app.HandlerFunc(http.MethodGet, version, "/scenarios/active", a.active, authen,
		authorize(permissionsbus.Actions.Read))

	// Runtime operations — Load and Reset mutate floor data, so they require Update.
	app.HandlerFunc(http.MethodPost, version, "/scenarios/{id}/load", a.load, authen,
		authorize(permissionsbus.Actions.Update))

	app.HandlerFunc(http.MethodPost, version, "/scenarios/active/reset", a.reset, authen,
		authorize(permissionsbus.Actions.Update))
}
