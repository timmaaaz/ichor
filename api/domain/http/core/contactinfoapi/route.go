package contactinfoapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log            *logger.Logger
	ContactInfoBus *contactinfobus.Business
	AuthClient     *authclient.Client
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(contactinfoapp.NewApp(cfg.ContactInfoBus))
	app.HandlerFunc(http.MethodGet, version, "/core/contactinfo", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/core/contactinfo/{contact_info_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/core/contactinfo", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/core/contactinfo/{contact_info_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/core/contactinfo/{contact_info_id}", api.delete, authen, ruleAdmin)

}
