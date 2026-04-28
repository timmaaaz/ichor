// Package paperworkapi maintains the web-based API for paperwork.
package paperworkapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/paperwork/paperworkapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Per-route table_access constants — paperwork endpoints render data from
// three different domain tables, and authorization piggybacks on existing
// table-level Read permissions for each. Tests reference these constants
// when downgrading non-admin role permissions to assert 403.
const (
	RouteTablePickSheet     = "sales.orders"
	RouteTableReceiveCover  = "procurement.purchase_orders"
	RouteTableTransferSheet = "inventory.transfer_orders"
)

// Config carries the dependencies for the paperwork API.
type Config struct {
	Log            *logger.Logger
	PaperworkBus   *paperworkbus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

// Routes registers paperwork endpoints behind Authenticate + Authorize.
// Each endpoint maps to the Read permission on the underlying domain table.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(paperworkapp.NewApp(cfg.PaperworkBus))

	app.HandlerFunc(http.MethodGet, version, "/paperwork/pick-sheet", api.pickSheet, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTablePickSheet, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/paperwork/receive-cover", api.receiveCover, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTableReceiveCover, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/paperwork/transfer-sheet", api.transferSheet, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTableTransferSheet, permissionsbus.Actions.Read, auth.RuleAny))
}
