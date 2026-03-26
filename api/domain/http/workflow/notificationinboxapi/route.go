package notificationinboxapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the notification inbox API.
type Config struct {
	Log             *logger.Logger
	NotificationBus *notificationbus.Business
	AuthClient      *authclient.Client
}

// Routes registers the notification inbox API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/workflow/notifications", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/notifications/count", api.count, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/notifications/{notification_id}/read", api.markAsRead, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/notifications/read-all", api.markAllAsRead, authen)
}
