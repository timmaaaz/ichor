package supervisorkpiapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/supervisorkpiapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds all dependencies needed by the supervisor KPI API routes.
type Config struct {
	Log                    *logger.Logger
	ApprovalRequestBus     *approvalrequestbus.Business
	InventoryAdjustmentBus *inventoryadjustmentbus.Business
	TransferOrderBus       *transferorderbus.Business
	InspectionBus          *inspectionbus.Business
	PutAwayTaskBus         *putawaytaskbus.Business
	AlertBus               *alertbus.Business
	AuthClient             *authclient.Client
	PermissionsBus         *permissionsbus.Business
}

// RouteTable is the table name used for permission lookups. This is a
// read-only aggregation endpoint with no backing table. It reuses
// inventory.inventory_adjustments as the permission gate — any user with
// read access to adjustments can view the supervisor KPI dashboard.
const RouteTable = "inventory.inventory_adjustments"

// Routes registers supervisor KPI HTTP routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	a := newAPI(supervisorkpiapp.NewApp(
		cfg.Log,
		cfg.ApprovalRequestBus,
		cfg.InventoryAdjustmentBus,
		cfg.TransferOrderBus,
		cfg.InspectionBus,
		cfg.PutAwayTaskBus,
		cfg.AlertBus,
	))

	app.HandlerFunc(http.MethodGet, version, "/inventory/supervisor/kpis", a.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
}
