package main

import (
	"context"
	"embed"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all"
	"github.com/timmaaaz/ichor/api/cmd/services/ichor/build/crud"
	"github.com/timmaaaz/ichor/api/cmd/services/ichor/build/reporting"
	"github.com/timmaaaz/ichor/api/domain/http/basicauthapi"
	"github.com/timmaaaz/ichor/api/domain/http/oauthapi"
	"github.com/timmaaaz/ichor/api/sdk/http/debug"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/keystore"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/web"
	"go.temporal.io/sdk/client"
)

/*
	Need to figure out timeouts for http service.
*/

//go:embed static
var static embed.FS

var build = "develop"
var routes = "all" // go build -ldflags "-X main.routes=crud"

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return otel.GetTraceID(ctx)
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "ICHOR", traceIDFn, events)

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout        time.Duration `conf:"default:5s"`
			WriteTimeout       time.Duration `conf:"default:10s"`
			IdleTimeout        time.Duration `conf:"default:120s"`
			ShutdownTimeout    time.Duration `conf:"default:20s"`
			APIHost            string        `conf:"default:0.0.0.0:8080"`
			DebugHost          string        `conf:"default:0.0.0.0:8090"`
			CORSAllowedOrigins []string      `conf:"default:*"`
		}
		Auth struct {
			Host       string `conf:"default:http://auth-service:6000"`
			KeysEnvVar string `conf:"default:ICHOR_KEYS"`
			KeysFolder string `conf:"default:zarf/keys/"`
			ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
			Issuer     string `conf:"default:service project"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:database-service"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Tempo struct {
			Host        string  `conf:"default:tempo:4317"`
			ServiceName string  `conf:"default:ichor"`
			Probability float64 `conf:"default:0.05"`
			// Shouldn't use a high Probability value in non-developer systems.
			// 0.05 should be enough for most systems. Some might want to have
			// this even lower.
		}
		OAuth struct {
			Environment        string        `conf:"default:development"`
			GoogleKey          string        `conf:"default:abc-123,mask"`
			GoogleSecret       string        `conf:"default:abc-123,mask"`
			Callback           string        `conf:"default:http://localhost:3000"`
			StoreKey           string        `conf:"default:dev-session-key-32-bytes-long!!!,mask"`
			TokenKey           string        `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1,mask"`
			UIAdminRedirect    string        `conf:"default:http://localhost:3001/admin?token="`
			UILoginRedirect    string        `conf:"default:http://localhost:3001/login"`
			TokenExpiration    time.Duration `conf:"default:20m"`
			DevTokenExpiration time.Duration `conf:"default:8h"`
		}
		RabbitMQ struct {
			URL           string        `conf:"default:amqp://guest:guest@rabbitmq-service:5672/"`
			MaxRetries    int           `conf:"default:5"`
			RetryDelay    time.Duration `conf:"default:5s"`
			PrefetchCount int           `conf:"default:10"`
		}
		Temporal struct {
			HostPort string `conf:"default:temporal-service.ichor-system.svc.cluster.local:7233"`
		}
		LLM struct {
			Provider  string `conf:"default:ollama"`
			APIKey    string `conf:"mask"`
			Model     string `conf:"default:qwen2.5:latest"`
			MaxTokens int    `conf:"default:4096"`
			BaseURL   string `conf:"default:http://localhost:8080"`
			Host      string `conf:"default:http://host.docker.internal:11434"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "Ichor",
		},
	}

	const prefix = "ICHOR"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", cfg.Build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	log.BuildInfo(ctx)

	expvar.NewString("build").Set(cfg.Build)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "hostport", cfg.DB.Host)

	db, err := sqldb.Open(sqldb.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer db.Close()

	// -------------------------------------------------------------------------
	// Initialize RabbitMQ Support

	log.Info(ctx, "startup", "status", "initializing RabbitMQ support")

	rabbitConfig := rabbitmq.Config{
		URL:                cfg.RabbitMQ.URL,
		MaxRetries:         cfg.RabbitMQ.MaxRetries,
		RetryDelay:         cfg.RabbitMQ.RetryDelay,
		PrefetchCount:      cfg.RabbitMQ.PrefetchCount,
		PrefetchSize:       0,
		PublisherConfirms:  true,
		ExchangeName:       "workflow",
		ExchangeType:       "topic",
		DeadLetterExchange: "workflow.dlx",
	}

	rabbitClient := rabbitmq.NewClient(log, rabbitConfig)

	if err := rabbitClient.WaitForConnection(30 * time.Second); err != nil {
		return fmt.Errorf("connecting to RabbitMQ: %w", err)
	}
	defer rabbitClient.Close()

	log.Info(ctx, "startup", "status", "RabbitMQ connected")

	// -------------------------------------------------------------------------
	// Initialize Temporal Client

	var temporalClient client.Client
	if cfg.Temporal.HostPort != "" {
		tc, err := client.Dial(client.Options{
			HostPort: cfg.Temporal.HostPort,
		})
		if err != nil {
			log.Error(ctx, "temporal: client creation failed, workflow dispatch disabled", "error", err)
		} else {
			temporalClient = tc
			defer temporalClient.Close()
			log.Info(ctx, "startup", "status", "Temporal client connected", "host", cfg.Temporal.HostPort)
		}
	}

	// -------------------------------------------------------------------------
	// Configure OAuth Providers based on environment

	log.Info(ctx, "startup", "status", "configuring OAuth providers")

	if cfg.OAuth.Environment == "production" {
		if cfg.OAuth.GoogleKey == "" || cfg.OAuth.GoogleSecret == "" {
			return errors.New("Google OAuth credentials required in production")
		}
		goth.UseProviders(
			google.New(cfg.OAuth.GoogleKey, cfg.OAuth.GoogleSecret, cfg.OAuth.Callback),
		)
	} else {
		// Development/Staging - add dev provider
		providers := []goth.Provider{
			oauthapi.NewDevelopmentProvider(cfg.OAuth.Callback),
		}
		if cfg.OAuth.GoogleKey != "" && cfg.OAuth.GoogleSecret != "" {
			providers = append(providers,
				google.New(cfg.OAuth.GoogleKey, cfg.OAuth.GoogleSecret, cfg.OAuth.Callback))
		}
		goth.UseProviders(providers...)
	}

	// -------------------------------------------------------------------------
	// Initialize authentication support

	log.Info(ctx, "startup", "status", "initializing authentication support")

	authClient := authclient.New(log, cfg.Auth.Host)

	ks := keystore.New()

	n1, err := ks.LoadByEnv(cfg.Auth.KeysEnvVar)
	if err != nil {
		return fmt.Errorf("loading keys by env: %w", err)
	}

	n2, err := ks.LoadByFileSystem(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		return fmt.Errorf("loading keys by fs: %w", err)
	}

	if n1+n2 == 0 {
		return fmt.Errorf("no keys exist: %w", err)
	}

	oauthAuth, err := auth.New(auth.Config{
		Log:       log,
		DB:        db,
		KeyLookup: ks,
		Issuer:    cfg.Auth.Issuer,
	})
	if err != nil {
		return fmt.Errorf("constructing OAuth auth: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Tracing Support

	log.Info(ctx, "startup", "status", "initializing tracing support")

	traceProvider, err := otel.InitTracing(otel.Config{
		ServiceName: cfg.Tempo.ServiceName,
		Host:        cfg.Tempo.Host,
		ExcludedRoutes: map[string]struct{}{
			"/v1/liveness":  {},
			"/v1/readiness": {},
		},
		Probability: cfg.Tempo.Probability,
	})
	if err != nil {
		return fmt.Errorf("starting tracing: %w", err)
	}

	defer traceProvider.Shutdown(context.Background())

	tracer := traceProvider.Tracer(cfg.Tempo.ServiceName)

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := mux.Config{
		Build:          build,
		Log:            log,
		Auth:           oauthAuth,
		AuthClient:     authClient,
		DB:             db,
		Tracer:         tracer,
		RabbitClient:   rabbitClient,
		TemporalClient: temporalClient,
		LLMProvider:    cfg.LLM.Provider,
		LLMAPIKey:      cfg.LLM.APIKey,
		LLMModel:       cfg.LLM.Model,
		LLMMaxTokens:   cfg.LLM.MaxTokens,
		LLMBaseURL:     cfg.LLM.BaseURL,
		LLMHost:        cfg.LLM.Host,
	}

	routes, userBus := buildRoutes(cfgMux)

	log.Info(ctx, "startup", "status", "binding V1 API routes", "userbus valid")

	webAPI := mux.WebAPI(cfgMux,
		routes,
		mux.WithCORS(cfg.Web.CORSAllowedOrigins),
		mux.WithFileServer(static, "static"),
	)

	// Add OAuth routes to the webAPI (assuming webAPI is a *web.App)
	oauthCfg := oauthapi.Config{
		Auth:            oauthAuth,
		Log:             log,
		TokenKey:        cfg.OAuth.TokenKey,
		StoreKey:        cfg.OAuth.StoreKey,
		UIAdminRedirect: cfg.OAuth.UIAdminRedirect,
		UILoginRedirect: cfg.OAuth.UILoginRedirect,
	}

	// Set token expiration based on environment
	if cfg.OAuth.Environment == "production" {
		oauthCfg.TokenExpiration = cfg.OAuth.TokenExpiration // 20m from config
	} else {
		oauthCfg.TokenExpiration = cfg.OAuth.DevTokenExpiration // 8h from config
	}

	basicAuthCfg := basicauthapi.Config{
		Log:             log,
		Auth:            oauthAuth,
		DB:              db,
		TokenKey:        cfg.OAuth.TokenKey,
		TokenExpiration: cfg.OAuth.TokenExpiration,
		UserBus:         userBus,
	}

	// Cast webAPI to *web.App to add routes
	if app, ok := webAPI.(*web.App); ok {
		oauthapi.Routes(app, oauthCfg)
		basicauthapi.Routes(app, basicAuthCfg)
	} else {
		return errors.New("failed to add OAuth routes: webAPI is not *web.App")
	}

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      webAPI,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func buildRoutes(cfgMux mux.Config) (mux.RouteAdder, *userbus.Business) {

	// The idea here is that we can build different versions of the binary
	// with different sets of exposed web APIs. By default we build a single
	// an instance with all the web APIs.
	//
	// Here is the scenario. It would be nice to build two binaries, one for the
	// transactional APIs (CRUD) and one for the reporting APIs. This would allow
	// the system to run two instances of the database. One instance tuned for the
	// transactional database calls and the other tuned for the reporting calls.
	// Tuning meaning indexing and memory requirements. The two databases can be
	// kept in sync with replication.

	switch routes {
	case "crud":
		r := crud.Routes()
		r.InitializeDependencies(cfgMux) // Initialize before returning
		return r, r.UserBus

	case "reporting":
		return reporting.Routes(), nil

	default:
		r := all.Routes()
		r.InitializeDependencies(cfgMux) // Initialize before returning
		return r, r.UserBus
	}
}
