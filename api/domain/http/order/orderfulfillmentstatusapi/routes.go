package orderfulfillmentstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/order/orderfulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log                       *logger.Logger
	OrderFulfillmentStatusBus *orderfulfillmentstatusbus.Business
	AuthClient                *authclient.Client
	PermissionsBus            *permissionsbus.Business
}

const (
	RouteTable = "order_fulfillment_statuses"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(orderfulfillmentstatusapp.NewApp(cfg.OrderFulfillmentStatusBus))
	app.HandlerFunc(http.MethodGet, version, "/order/order-fulfillment-statuses", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/order/order-fulfillment-statuses/{order_fulfillment_status_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/order/order-fulfillment-statuses", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))
	app.HandlerFunc(http.MethodPut, version, "/order/order-fulfillment-statuses/{order_fulfillment_status_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))
	app.HandlerFunc(http.MethodDelete, version, "/order/order-fulfillment-statuses/{order_fulfillment_status_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
