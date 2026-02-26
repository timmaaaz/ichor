package inventoryadjustmentapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
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
	RouteTable = "inventory.inventory_adjustments"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(inventoryadjustmentapp.NewApp(cfg.InventoryAdjustmentBus))

	app.HandlerFunc(http.MethodGet, version, "/inventory/inventory-adjustments", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/inventory/inventory-adjustments/{adjustment_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/inventory/inventory-adjustments", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/inventory/inventory-adjustments/{adjustment_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/inventory/inventory-adjustments/{adjustment_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/inventory/inventory-adjustments/{adjustment_id}/approve", api.approve, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/inventory/inventory-adjustments/{adjustment_id}/reject", api.reject, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))
}
