package inventoryitemapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log              *logger.Logger
	InventoryItemBus *inventoryitembus.Business
	AuthClient       *authclient.Client
	PermissionsBus   *permissionsbus.Business
}

const TableName = "inventory_items"

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(inventoryitemapp.NewApp(cfg.InventoryItemBus))
	app.HandlerFunc(http.MethodGet, version, "/inventory/inventory-items", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/inventory/inventory-items/{item_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/inventory/inventory-items", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/inventory/inventory-items/{item_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/inventory/inventory-items/{item_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, TableName, permissionsbus.Actions.Delete, auth.RuleAny))
}
