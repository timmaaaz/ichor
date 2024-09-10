// Package mux provides support to bind domain level routes
// to the application mux.
package mux

import (
	"context"
	"embed"
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/mid"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/auth"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/authclient"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
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
	Build      string
	Log        *logger.Logger
	Auth       *auth.Auth
	AuthClient *authclient.Client
	DB         *sqlx.DB
	Tracer     trace.Tracer
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
