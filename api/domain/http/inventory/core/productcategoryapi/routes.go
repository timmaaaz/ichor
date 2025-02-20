package productcategoryapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/core/productcategoryapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log                *logger.Logger
	ProductCategoryBus *productcategorybus.Business
	AuthClient         *authclient.Client
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(productcategoryapp.NewApp(cfg.ProductCategoryBus))
	app.HandlerFunc(http.MethodGet, version, "/inventory/core/productcategories", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/inventory/core/productcategories/{product_category_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/inventory/core/productcategories", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/inventory/core/productcategories/{product_category_id}", api.update, authen, ruleAdmin)
	app.HandlerFunc(http.MethodDelete, version, "/inventory/core/productcategories/{product_category_id}", api.delete, authen, ruleAdmin)

}
