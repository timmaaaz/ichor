package productcategoryapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/productcategoryapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log                *logger.Logger
	ProductCategoryBus *productcategorybus.Business
	AuthClient         *authclient.Client
	PermissionsBus     *permissionsbus.Business
}

const (
	routeTable = "product_categories"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(productcategoryapp.NewApp(cfg.ProductCategoryBus))
	app.HandlerFunc(http.MethodGet, version, "/inventory/core/productcategories", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/inventory/core/productcategories/{product_category_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/inventory/core/productcategories", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Create, auth.RuleAny))
	app.HandlerFunc(http.MethodPut, version, "/inventory/core/productcategories/{product_category_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Update, auth.RuleAny))
	app.HandlerFunc(http.MethodDelete, version, "/inventory/core/productcategories/{product_category_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, routeTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
