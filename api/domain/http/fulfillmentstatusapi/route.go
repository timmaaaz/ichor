package fulfillmentstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/fulfillmentstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers
type Config struct {
	Log                  *logger.Logger
	FulfillmentStatusBus *fulfillmentstatusbus.Business
	AuthClient           *authclient.Client
}

// Routes adds routes to the group
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(fulfillmentstatusapp.NewApp(cfg.FulfillmentStatusBus))

	app.HandlerFunc(http.MethodGet, version, "/fulfillmentstatus", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/fulfillmentstatus/{fulfillment_status_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/fulfillmentstatus", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/fulfillmentstatus/{fulfillment_status_id}", api.update, authen)
	app.HandlerFunc(http.MethodDelete, version, "/fulfillmentstatus/{fulfillment_status_id}", api.delete, authen)
}
