package tranapi

import (
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/mid"
	"bitbucket.org/superiortechnologies/ichor/app/domain/tranapp"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/auth"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/authclient"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/sqldb"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	DB         *sqlx.DB
	UserBus    *userbus.Business
	ProductBus *productbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	transaction := mid.BeginCommitRollback(cfg.Log, sqldb.NewBeginner(cfg.DB))
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(tranapp.NewApp(cfg.UserBus, cfg.ProductBus))
	app.HandlerFunc(http.MethodPost, version, "/tranexample", api.create, authen, ruleAdmin, transaction)
}
