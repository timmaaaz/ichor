package timezoneapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/geography/timezoneapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log            *logger.Logger
	TimezoneBus    *timezonebus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	RouteTable = "geography.timezones"
)

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(timezoneapp.NewApp(cfg.TimezoneBus))
	app.HandlerFunc(http.MethodGet, version, "/geography/timezones", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/geography/timezones/{timezone_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/geography/timezones", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodPut, version, "/geography/timezones/{timezone_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAdminOnly))
	app.HandlerFunc(http.MethodDelete, version, "/geography/timezones/{timezone_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAdminOnly))
}
