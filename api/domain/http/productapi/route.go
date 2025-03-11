package productapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/productapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log            *logger.Logger
	UserBus        *userbus.Business
	ProductBus     *productbus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const (
	routeTable = "products"
)

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAuthorizeProduct := mid.AuthorizeProduct(cfg.AuthClient, cfg.ProductBus)

	api := newAPI(productapp.NewApp(cfg.ProductBus))
	app.HandlerFunc(http.MethodGet, version, "/products", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/products/{product_id}", api.queryByID, authen, ruleAuthorizeProduct,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/products", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Create, auth.RuleAny))
	app.HandlerFunc(http.MethodPut, version, "/products/{product_id}", api.update, authen, ruleAuthorizeProduct,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Update, auth.RuleAny))
	app.HandlerFunc(http.MethodDelete, version, "/products/{product_id}", api.delete, authen, ruleAuthorizeProduct,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
