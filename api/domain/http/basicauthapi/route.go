package basicauthapi

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// TODO: These should take {provider} path arguments like the oauthapi routes do.

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log             *logger.Logger
	Auth            *auth.Auth
	DB              *sqlx.DB
	TokenKey        string
	TokenExpiration time.Duration
	UserBus         *userbus.Business
}

func Routes(app *web.App, cfg Config) {
	api := NewAPI(cfg)

	app.RawHandlerFunc(http.MethodPost, "", "/api/auth/basic/login", api.login)
	app.RawHandlerFunc(http.MethodPost, "", "/api/auth/basic/refresh", api.refresh)
	app.RawHandlerFunc(http.MethodPost, "", "/api/auth/basic/logout", api.logout)
}
