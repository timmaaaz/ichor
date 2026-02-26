package notificationsapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the notifications summary API.
type Config struct {
	Log         *logger.Logger
	AlertBus    *alertbus.Business
	ApprovalBus *approvalrequestbus.Business
	UserRoleBus *userrolebus.Business
	AuthClient  *authclient.Client
}

// Routes registers the notifications API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/workflow/notifications/summary", api.summary, authen)
}
