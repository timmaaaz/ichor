// Package main is the entry point for the workflow-worker service.
// The workflow-worker connects to Temporal and processes workflow executions.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf/v3"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
	"github.com/timmaaaz/ichor/foundation/logger"
)

var build = "develop"

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo, "WORKFLOW-WORKER",
		func(context.Context) string { return "" })

	if err := run(log); err != nil {
		log.Error(context.Background(), "startup", "error", err)
		os.Exit(1)
	}
}

func run(log *logger.Logger) error {

	// =========================================================================
	// Configuration
	// =========================================================================

	cfg := struct {
		conf.Version
		Temporal struct {
			HostPort  string `conf:"default:temporal-service.ichor-system.svc.cluster.local:7233"`
			Namespace string `conf:"default:default"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,mask"`
			Host       string `conf:"default:database-service.ichor-system.svc.cluster.local"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:true"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "Workflow Worker Service",
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

	log.Info(context.Background(), "starting service", "version", cfg.Build)

	// =========================================================================
	// Database
	// =========================================================================

	db, err := sqldb.Open(sqldb.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer db.Close()

	// =========================================================================
	// Temporal Client
	// =========================================================================

	tc, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
		Logger:    newTemporalLogger(log),
	})
	if err != nil {
		return fmt.Errorf("creating temporal client: %w", err)
	}
	defer tc.Close()

	// =========================================================================
	// Action Registries
	// =========================================================================

	// Sync action registry — RegisterCoreActions provides 5 handlers:
	// evaluate_condition, update_field, seek_approval, send_email, send_notification.
	// Full RegisterAll (with inventory/alert handlers) deferred until worker
	// gains RabbitMQ + bus dependencies.
	actionRegistry := workflow.NewActionRegistry()
	workflowactions.RegisterCoreActions(actionRegistry, log, db)

	// Async action registry — empty for now. Async handler adapters
	// (SendEmailHandler, AllocateInventoryHandler) will be registered
	// when the full async completion flow is implemented.
	asyncRegistry := temporal.NewAsyncRegistry()

	// =========================================================================
	// Temporal Worker
	// =========================================================================

	// Start with conservative concurrency limits.
	// Tune based on pod resource limits and load testing.
	w := worker.New(tc, temporal.TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     100,
		MaxConcurrentWorkflowTaskExecutionSize: 100,
	})

	// Register workflows (package-level functions).
	// Temporal resolves by name: "ExecuteGraphWorkflow", "ExecuteBranchUntilConvergence".
	w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)

	// Register activities via Activities struct.
	// Temporal resolves struct method names by string: "ExecuteActionActivity",
	// "ExecuteAsyncActionActivity". Both registries are passed to the struct
	// so the activity methods can dispatch to the correct handler.
	w.RegisterActivity(&temporal.Activities{
		Registry:      actionRegistry,
		AsyncRegistry: asyncRegistry,
	})

	log.Info(context.Background(), "starting workflow worker",
		"task_queue", temporal.TaskQueue,
		"temporal_host", cfg.Temporal.HostPort,
		"build", cfg.Build,
	)

	// =========================================================================
	// Health Server
	// =========================================================================

	ready := make(chan struct{})
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		mux.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-ready:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			default:
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("not ready"))
			}
		})
		if err := http.ListenAndServe(":4001", mux); err != nil {
			log.Error(context.Background(), "health server error", "error", err)
		}
	}()

	// =========================================================================
	// Start Worker & Shutdown
	// =========================================================================

	if err := w.Start(); err != nil {
		return fmt.Errorf("starting worker: %w", err)
	}
	close(ready)

	log.Info(context.Background(), "worker started and ready")

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Info(context.Background(), "shutting down worker")
	w.Stop()

	return nil
}

// =============================================================================
// Temporal Logger Adapter
// =============================================================================

// temporalLogger adapts foundation/logger to Temporal's log.Logger interface.
type temporalLogger struct {
	log *logger.Logger
}

func newTemporalLogger(log *logger.Logger) *temporalLogger {
	return &temporalLogger{log: log}
}

func (l *temporalLogger) Debug(msg string, keyvals ...interface{}) {
	l.log.Debug(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Info(msg string, keyvals ...interface{}) {
	l.log.Info(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.log.Warn(context.Background(), msg, keyvals...)
}

func (l *temporalLogger) Error(msg string, keyvals ...interface{}) {
	l.log.Error(context.Background(), msg, keyvals...)
}
