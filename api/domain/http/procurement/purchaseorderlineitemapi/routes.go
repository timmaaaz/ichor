package purchaseorderlineitemapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log                      *logger.Logger
	PurchaseOrderLineItemBus *purchaseorderlineitembus.Business
	AuthClient               *authclient.Client
	PermissionsBus           *permissionsbus.Business
}

const (
	RouteTable = "purchase_order_line_items"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(purchaseorderlineitemapp.NewApp(cfg.PurchaseOrderLineItemBus))

	app.HandlerFunc(http.MethodGet, version, "/procurement/purchase-order-line-items", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/procurement/purchase-order-line-items/batch", api.queryByIDs, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/procurement/purchase-order-line-items/purchase-order/{purchase_order_id}", api.queryByPurchaseOrderID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/procurement/purchase-order-line-items/{purchase_order_line_item_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/procurement/purchase-order-line-items", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/procurement/purchase-order-line-items/{purchase_order_line_item_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/procurement/purchase-order-line-items/{purchase_order_line_item_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/procurement/purchase-order-line-items/{purchase_order_line_item_id}/receive-quantity", api.receiveQuantity, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))
}
