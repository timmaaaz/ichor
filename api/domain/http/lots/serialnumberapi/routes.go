package serialnumberapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/lots/serialnumberapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/lot/serialnumberbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log             *logger.Logger
	SerialNumberBus *serialnumberbus.Business
	AuthClient      *authclient.Client
	PermissionsBus  *permissionsbus.Business
}

const (
	RouteTable = "serial_numbers"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(serialnumberapp.NewApp(cfg.SerialNumberBus))

	app.HandlerFunc(http.MethodGet, version, "/lots/serialnumber", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/lots/serialnumber/{serial_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/lots/serialnumber", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/lots/serialnumber/{serial_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/lots/serialnumber/{serial_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
