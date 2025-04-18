package inventoryadjustmentapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/movement/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/movement/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log                    *logger.Logger
	InventoryAdjustmentBus *inventoryadjustmentbus.Business
	AuthClient             *authclient.Client
	PermissionsBus         *permissionsbus.Business
}

const (
	RouteTable = "inventory_adjustments"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(inventoryadjustmentapp.NewApp(cfg.InventoryAdjustmentBus))

	app.HandlerFunc(http.MethodGet, version, "/movement/inventoryadjustment", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/movement/inventoryadjustment/{adjustment_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/movement/inventoryadjustment", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/movement/inventoryadjustment/{adjustment_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/movement/inventoryadjustment/{adjustment_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
