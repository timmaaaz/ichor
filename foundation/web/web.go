// Package web contains a small web framework extension.
package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"regexp"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Encoder defines behavior that can encode a data model and provide
// the content type for that encoding.
type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

// HandlerFunc represents a function that handles a http request within our own
// little mini framework.
type HandlerFunc func(ctx context.Context, r *http.Request) Encoder

// Logger represents a function that will be called to add information
// to the logs.
type Logger func(ctx context.Context, msg string, args ...any)

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	log    Logger
	tracer trace.Tracer
	mux    *http.ServeMux // Standard HTTP routes (wrapped by OTEL)
	rawMux *http.ServeMux // WebSocket/streaming routes (bypass OTEL)
	otmux  http.Handler   // OTEL-wrapped version of mux
	mw     []MidFunc
	origins []string
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(log Logger, tracer trace.Tracer, mw ...MidFunc) *App {
	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/trusted.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/

	mux := http.NewServeMux()
	rawMux := http.NewServeMux()

	return &App{
		log:    log,
		tracer: tracer,
		mux:    mux,
		rawMux: rawMux,
		otmux:  otelhttp.NewHandler(mux, "request"),
		mw:     mw,
	}
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic. Routes registered via RawHandlerFunc bypass OTEL tracing
// to avoid interference with WebSocket upgrades. All other routes go through
// the OpenTelemetry handler for tracing.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if this route is registered on the raw mux (e.g., WebSocket handlers).
	// These bypass OTEL wrapping because otelhttp writes a 200 OK response after
	// the handler returns, which interferes with WebSocket's 101 Switching Protocols.
	if _, pattern := a.rawMux.Handler(r); pattern != "" {
		a.rawMux.ServeHTTP(w, r)
		return
	}

	// All other routes go through OTEL for tracing.
	a.otmux.ServeHTTP(w, r)
}

// EnableCORS enables CORS preflight requests to work in the middleware. It
// prevents the MethodNotAllowedHandler from being called. This must be enabled
// for the CORS middleware to work.
func (a *App) EnableCORS(origins []string) {
	a.origins = origins

	handler := func(ctx context.Context, r *http.Request) Encoder {
		return cors{Status: "OK"}
	}
	handler = wrapMiddleware([]MidFunc{a.corsHandler}, handler)

	a.HandlerFuncNoMid(http.MethodOptions, "", "/", handler)
}

func (a *App) corsHandler(webHandler HandlerFunc) HandlerFunc {
	h := func(ctx context.Context, r *http.Request) Encoder {
		w := GetWriter(ctx)

		for _, origin := range a.origins {
			w.Header().Add("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		return webHandler(ctx, r)
	}

	return h
}

// HandlerFuncNoMid sets a handler function for a given HTTP method and path
// pair to the application server mux. Does not include the application
// middleware or OTEL tracing.
func (a *App) HandlerFuncNoMid(method string, group string, path string, handlerFunc HandlerFunc) {
	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := setWriter(r.Context(), w)

		resp := handlerFunc(ctx, r)

		if err := Respond(ctx, w, resp); err != nil {
			a.log(ctx, "web-respond", "ERROR", err)
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, h)
}

// HandlerFunc sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) HandlerFunc(method string, group string, path string, handlerFunc HandlerFunc, mw ...MidFunc) {
	handlerFunc = wrapMiddleware(mw, handlerFunc)
	handlerFunc = wrapMiddleware(a.mw, handlerFunc)

	if a.origins != nil {
		handlerFunc = wrapMiddleware([]MidFunc{a.corsHandler}, handlerFunc)
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := setTracer(r.Context(), a.tracer)
		ctx = setWriter(ctx, w)

		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

		resp := handlerFunc(ctx, r)

		if err := Respond(ctx, w, resp); err != nil {
			a.log(ctx, "web-respond", "ERROR", err)
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, h)
}

// RawHandlerFunc sets a raw handler function for a given HTTP method and path
// pair to the application server mux. This is designed for WebSocket handlers
// and other handlers that need direct control over the HTTP response.
//
// IMPORTANT: All WebSocket endpoints MUST use this method, not HandlerFunc.
// The standard HandlerFunc routes through otelhttp which writes HTTP 200 after
// handlers return, breaking WebSocket's required 101 Switching Protocols response.
//
// Key differences from HandlerFunc:
//   - Bypasses otelhttp wrapper (required for WebSocket 101 response)
//   - Does NOT apply CORS middleware (WebSocket uses websocket.AcceptOptions.OriginPatterns)
//   - Does NOT inject OTEL trace headers into response
//   - Does NOT call web.Respond() - the raw handler writes its own response
//   - Still applies authentication and other passed middleware
//
// For WebSocket authentication, use mid.BearerQueryParam which extracts JWT
// from the ?token= query parameter since browsers cannot set headers on
// WebSocket upgrade requests.
func (a *App) RawHandlerFunc(method string, group string, path string, rawHandlerFunc http.HandlerFunc, mw ...MidFunc) {
	handlerFunc := func(ctx context.Context, r *http.Request) Encoder {
		r = r.WithContext(ctx)
		rawHandlerFunc(GetWriter(ctx), r)
		return nil
	}

	handlerFunc = wrapMiddleware(mw, handlerFunc)
	handlerFunc = wrapMiddleware(a.mw, handlerFunc)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := setTracer(r.Context(), a.tracer)
		ctx = setWriter(ctx, w)

		handlerFunc(ctx, r)
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	// Register directly on the raw mux to bypass otelhttp wrapper.
	// WebSocket connections hijack the underlying connection, which conflicts
	// with otelhttp's response writer wrapping. The otelhttp handler writes
	// a 200 OK after the handler returns, even if the connection was hijacked
	// for WebSocket upgrade (which should be 101 Switching Protocols).
	a.rawMux.HandleFunc(finalPath, h)
}

// FileServerReact starts a file server based on the specified file system and
// directory inside that file system for a statically built react webapp.
func (a *App) FileServerReact(static embed.FS, dir string) error {
	fileMatcher := regexp.MustCompile(`\.[a-zA-Z]*$`)

	fSys, err := fs.Sub(static, dir)
	if err != nil {
		return fmt.Errorf("switching to static folder: %w", err)
	}

	fileServer := http.FileServer(http.FS(fSys))

	h := func(w http.ResponseWriter, r *http.Request) {
		if !fileMatcher.MatchString(r.URL.Path) {
			p, err := static.ReadFile(fmt.Sprintf("%s/index.html", dir))
			if err != nil {
				return
			}

			w.Write(p)
			return
		}

		fileServer.ServeHTTP(w, r)
	}

	a.mux.HandleFunc("/", h)

	return nil
}

// FileServer starts a file server based on the specified file system and
// directory inside that file system.
func (a *App) FileServer(static embed.FS, dir string, notFoundHandler http.HandlerFunc) error {
	fSys, err := fs.Sub(static, dir)
	if err != nil {
		return fmt.Errorf("switching to static folder: %w", err)
	}

	fileServer := http.FileServer(http.FS(fSys))

	h := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		if path == "" {
			path = "index.html"
		}

		f, err := fSys.Open(path)
		if err != nil {
			notFoundHandler(w, r)
			return
		}
		defer f.Close()

		fileServer.ServeHTTP(w, r)
	}

	a.mux.HandleFunc("/", h)

	return nil
}
