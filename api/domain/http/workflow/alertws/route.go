package alertws

import (
	"net/http"

	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// RouteConfig holds route registration dependencies.
type RouteConfig struct {
	Log                *logger.Logger
	AlertHub           *AlertHub
	CORSAllowedOrigins []string
}

// Routes registers the WebSocket routes for real-time alerts.
func Routes(app *web.App, cfg RouteConfig, wsAuth web.MidFunc) {
	const version = "v1"

	handler := ServeWS(Config{
		Log:                cfg.Log,
		AlertHub:           cfg.AlertHub,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
	})

	app.RawHandlerFunc(http.MethodGet, version, "/workflow/alerts/ws", handler, wsAuth)
}
