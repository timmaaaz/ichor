package brandapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"

	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log        *logger.Logger
	BrandBus   *brandbus.Business
	AuthClient *authclient.Client
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(brandapp.NewApp(cfg.BrandBus))
	app.HandlerFunc(http.MethodGet, version, "/inventory/core/brands", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/inventory/core/brands/{brand_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/inventory/core/brands", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/inventory/core/brands/{brand_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/inventory/core/brands/{brand_id}", api.delete, authen, ruleAdmin)

}
