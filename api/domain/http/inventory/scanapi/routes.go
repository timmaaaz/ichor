package scanapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/scanapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the scan API.
type Config struct {
	Log            *logger.Logger
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
	ScanApp        *scanapp.App
}

// RouteTable is the permission table used to authorize scan reads.
const RouteTable = "inventory.inventory_items"

// Routes registers the barcode scan endpoint.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(cfg.ScanApp)

	app.HandlerFunc(http.MethodGet, version, "/inventory/scan", api.scan, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
}
