package vproductapi

import (
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/mid"
	"bitbucket.org/superiortechnologies/ichor/app/domain/vproductapp"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/auth"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/authclient"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/vproductbus"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log         *logger.Logger
	UserBus     *userbus.Business
	VProductBus *vproductbus.Business
	AuthClient  *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(vproductapp.NewApp(cfg.VProductBus))
	app.HandlerFunc(http.MethodGet, version, "/vproducts", api.query, authen, ruleAdmin)
}
