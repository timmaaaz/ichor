package contactinfoapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log            *logger.Logger
	ContactInfoBus *contactinfobus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	routeTable = "contact_info"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(contactinfoapp.NewApp(cfg.ContactInfoBus))
	app.HandlerFunc(http.MethodGet, version, "/core/contactinfo", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/core/contactinfo/{contact_info_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/core/contactinfo", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Create, auth.RuleAny))
	app.HandlerFunc(http.MethodPut, version, "/core/contactinfo/{contact_info_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Update, auth.RuleAny))
	app.HandlerFunc(http.MethodDelete, version, "/core/contactinfo/{contact_info_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
