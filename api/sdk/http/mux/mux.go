// Package mux provides support to bind domain level routes
// to the application mux.
package mux

import (
	"context"
	"embed"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/web"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin []string
	static     *embed.FS
	staticDir  string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origins []string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origins
	}
}

// WithFileServer provides configuration options for file server.
func WithFileServer(static embed.FS, dir string) func(opts *Options) {
	return func(opts *Options) {
		opts.static = &static
		opts.staticDir = dir
	}
}

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build          string
	Log            *logger.Logger
	Auth           *auth.Auth
	AuthClient     *authclient.Client
	DB             *sqlx.DB
	Tracer         trace.Tracer
	RabbitClient   *rabbitmq.Client
	TemporalClient client.Client // nil means Temporal disabled

	// LLM provider configuration for agent chat.
	LLMProvider       string
	LLMAPIKey         string
	LLMModel          string
	LLMMaxTokens      int
	LLMBaseURL        string
	LLMHost           string
	LLMThinkingEffort string

	// Resend email delivery configuration.
	// ResendAPIKey empty means email delivery is disabled (graceful degradation).
	ResendAPIKey  string
	ResendFrom    string

	// CORSAllowedOrigins for WebSocket and SSE upgrade routes.
	// Defaults to "*" if empty (open — set from ICHOR_WEB_CORS_ALLOWED_ORIGINS).
	CORSAllowedOrigins []string
}

// RouteAdder defines behavior that sets the routes to bind for an instance
// of the service.
type RouteAdder interface {
	Add(app *web.App, cfg Config)
}

// WebAPI constructs a http.Handler with all application routes bound.
func WebAPI(cfg Config, routeAdder RouteAdder, options ...func(opts *Options)) http.Handler {
	logger := func(ctx context.Context, msg string, args ...any) {
		cfg.Log.Info(ctx, msg, args...)
	}

	app := web.NewApp(
		logger,
		cfg.Tracer,
		mid.Otel(cfg.Tracer),
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	var opts Options
	for _, option := range options {
		option(&opts)
	}

	if len(opts.corsOrigin) > 0 {
		app.EnableCORS(opts.corsOrigin)
	}

	routeAdder.Add(app, cfg)

	if opts.static != nil {
		app.FileServer(*opts.static, opts.staticDir, http.NotFound)
	}

	return app
}
