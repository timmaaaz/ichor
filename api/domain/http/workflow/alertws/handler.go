package alertws

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/foundation/logger"
	foundationws "github.com/timmaaaz/ichor/foundation/websocket"
)

// Config holds handler dependencies.
type Config struct {
	Log                *logger.Logger
	AlertHub           *AlertHub
	CORSAllowedOrigins []string
}

// ServeWS returns an http.HandlerFunc that upgrades HTTP to WebSocket.
// NOTE: Returns http.HandlerFunc for use with app.RawHandlerFunc().
// The BearerQueryParam middleware must be applied before this handler.
func ServeWS(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract userID set by BearerQueryParam middleware
		userID, err := mid.GetUserID(ctx)
		if err != nil {
			cfg.Log.Error(ctx, "failed to get user ID from context", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Accept WebSocket upgrade with CORS configuration
		// SECURITY: Use specific origins in production, never wildcard "*"
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: cfg.CORSAllowedOrigins,
		})
		if err != nil {
			cfg.Log.Error(ctx, "websocket upgrade failed", "error", err)
			return
		}

		// Create generic client (foundation layer)
		client := foundationws.NewClient(cfg.AlertHub.Hub(), conn, cfg.Log)

		// Register with AlertHub (fetches roles and registers with string IDs)
		if err := cfg.AlertHub.RegisterClient(ctx, client, userID); err != nil {
			cfg.Log.Error(ctx, "failed to register client", "error", err)
			conn.Close(websocket.StatusInternalError, "registration failed")
			return
		}

		// Start write pump in goroutine (owns connection lifecycle)
		go client.WritePump(ctx)

		// Read pump blocks until disconnect
		client.ReadPump(ctx)
	}
}
