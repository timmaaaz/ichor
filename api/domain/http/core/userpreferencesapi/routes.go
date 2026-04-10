package userpreferencesapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserPreferencesBus *userpreferencesbus.Business
	AuthClient         *authclient.Client
	UserBus            *userbus.Business
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAuthorizeUser := mid.AuthorizeUser(cfg.AuthClient, cfg.UserBus, auth.RuleAdminOrSubject)

	api := newAPI(userpreferencesapp.NewApp(cfg.UserPreferencesBus))

	app.HandlerFunc(http.MethodPut, version, "/users/{user_id}/preferences/{key}", api.set, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodGet, version, "/users/{user_id}/preferences/{key}", api.get, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodGet, version, "/users/{user_id}/preferences", api.getAll, authen, ruleAuthorizeUser)
	app.HandlerFunc(http.MethodDelete, version, "/users/{user_id}/preferences/{key}", api.delete, authen, ruleAuthorizeUser)
}
