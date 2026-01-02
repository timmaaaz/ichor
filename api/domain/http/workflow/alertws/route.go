// Package alertws provides WebSocket endpoints for real-time alert delivery.
//
// # Architecture Note: WebSocket Route Registration
//
// WebSocket handlers MUST use app.RawHandlerFunc() instead of app.HandlerFunc().
// This is critical because:
//
//  1. The standard HandlerFunc routes go through otelhttp.NewHandler which wraps
//     the response writer and writes HTTP 200 OK after the handler returns.
//
//  2. WebSocket upgrades require HTTP 101 Switching Protocols. The otelhttp wrapper
//     interferes with this by writing its own status code after connection hijacking.
//
//  3. RawHandlerFunc registers routes on a separate mux that bypasses OTEL wrapping,
//     allowing the WebSocket upgrade to complete properly.
//
// # CORS Handling
//
// WebSocket CORS is handled differently than HTTP CORS:
//   - HTTP CORS uses Access-Control-* response headers
//   - WebSocket CORS validates the Origin request header server-side
//
// The websocket.Accept() function handles origin validation via OriginPatterns.
// Do NOT apply HTTP CORS middleware to WebSocket routes.
//
// # Authentication
//
// WebSocket connections cannot use Authorization headers during the upgrade.
// Use the BearerQueryParam middleware which extracts JWT from ?token= parameter.
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
//
// IMPORTANT: Uses RawHandlerFunc to bypass OTEL wrapping. See package docs.
func Routes(app *web.App, cfg RouteConfig, wsAuth web.MidFunc) {
	const version = "v1"

	handler := ServeWS(Config{
		Log:                cfg.Log,
		AlertHub:           cfg.AlertHub,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
	})

	// RawHandlerFunc is required for WebSocket - see package documentation.
	app.RawHandlerFunc(http.MethodGet, version, "/workflow/alerts/ws", handler, wsAuth)
}
