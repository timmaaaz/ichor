package directedworkapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the directed-work API routes.
type Config struct {
	Log               *logger.Logger
	PickTaskBus       *picktaskbus.Business
	PutAwayTaskBus    *putawaytaskbus.Business
	CycleCountItemBus *cyclecountitembus.Business
	InspectionBus     *inspectionbus.Business
	TransferOrderBus  *transferorderbus.Business
	OrdersBus         *ordersbus.Business
	AuthClient        *authclient.Client
}

// Routes registers the directed-work endpoint.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)

	// GET /v1/floor/work/next — returns the single best next work item
	// for the authenticated worker, or {"work_item": null} if none.
	app.HandlerFunc(http.MethodGet, version, "/floor/work/next", api.queryNext, authen)
}
