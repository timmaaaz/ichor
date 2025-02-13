package commentapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/users/status/commentapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers
type Config struct {
	Log                    *logger.Logger
	UserApprovalCommentBus *commentbus.Business
	AuthClient             *authclient.Client
}

// Routes adds routes to the group
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(commentapp.NewApp(cfg.UserApprovalCommentBus))

	app.HandlerFunc(http.MethodGet, version, "/users/status/comments", api.query, authen, ruleAdmin)
	app.HandlerFunc(http.MethodGet, version, "/users/status/comments/{user_status_comment_id}", api.queryByID, authen)
	app.HandlerFunc(http.MethodPost, version, "/users/status/comments", api.create, authen, ruleAdmin)
	app.HandlerFunc(http.MethodPut, version, "/users/status/comments/{user_status_comment_id}", api.update, authen)
	app.HandlerFunc(http.MethodDelete, version, "/users/status/comments/{user_status_comment_id}", api.delete, authen)
}
