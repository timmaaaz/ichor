package putawaytaskapi

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/putawaytaskapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds all dependencies needed by the put-away task API routes.
type Config struct {
	Log               *logger.Logger
	PutAwayTaskBus    *putawaytaskbus.Business
	InvTransactionBus *inventorytransactionbus.Business
	InvItemBus        *inventoryitembus.Business
	DB                *sqlx.DB
	AuthClient        *authclient.Client
	PermissionsBus    *permissionsbus.Business
}

const (
	// RouteTable is the table name used for permission lookups.
	RouteTable = "inventory.put_away_tasks"
)

// Routes registers all put-away task HTTP routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	a := newAPI(putawaytaskapp.NewApp(
		cfg.PutAwayTaskBus,
		cfg.InvTransactionBus,
		cfg.InvItemBus,
		cfg.DB,
	))

	app.HandlerFunc(http.MethodGet, version, "/inventory/put-away-tasks", a.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/inventory/put-away-tasks/{task_id}", a.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/inventory/put-away-tasks", a.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/inventory/put-away-tasks/{task_id}", a.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/inventory/put-away-tasks/{task_id}", a.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
