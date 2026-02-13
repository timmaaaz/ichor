package chatapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/sdk/agenttools"
	"github.com/timmaaaz/ichor/business/sdk/llm"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the agent chat API routes.
type Config struct {
	Log                *logger.Logger
	LLMProvider        llm.Provider
	ToolExecutor       *agenttools.Executor
	AuthClient         *authclient.Client
	CORSAllowedOrigins []string
}

// Routes registers the agent chat API routes.
//
// IMPORTANT: Uses RawHandlerFunc to bypass OTEL and WriteTimeout for SSE
// streaming. See package-level documentation for rationale.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)
	cors := corsMiddleware(cfg.CORSAllowedOrigins)

	// POST handler with CORS middleware wrapping auth (so CORS headers are
	// present even on auth failures).
	app.RawHandlerFunc(http.MethodPost, version, "/agent/chat", api.chat, cors, authen)

	// OPTIONS preflight handler. RawHandlerFunc registers on a separate mux
	// from the framework's catch-all OPTIONS handler, so we need an explicit
	// one here to prevent Go 1.22's method-aware routing from returning 405.
	app.RawHandlerFunc(http.MethodOptions, version, "/agent/chat", corsPreflight(cfg.CORSAllowedOrigins))
}

// corsMiddleware returns a web.MidFunc that sets CORS headers on every
// response — including error responses from inner middleware (e.g. auth).
func corsMiddleware(origins []string) web.MidFunc {
	return func(handler web.HandlerFunc) web.HandlerFunc {
		return func(ctx context.Context, r *http.Request) web.Encoder {
			w := web.GetWriter(ctx)
			for _, origin := range origins {
				w.Header().Add("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			return handler(ctx, r)
		}
	}
}

// corsPreflight returns an http.HandlerFunc that responds to OPTIONS preflight
// requests with the appropriate CORS headers.
func corsPreflight(origins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, origin := range origins {
			w.Header().Add("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusNoContent)
	}
}
