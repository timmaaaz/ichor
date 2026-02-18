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
	"time"

	"github.com/ardanlabs/conf/v3"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus/stores/inventoryitemdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/stores/inventorylocationdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
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
		RabbitMQ struct {
			URL           string        `conf:"default:amqp://guest:guest@rabbitmq-service:5672/"`
			MaxRetries    int           `conf:"default:5"`
			RetryDelay    time.Duration `conf:"default:5s"`
			PrefetchCount int           `conf:"default:10"`
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
	// RabbitMQ (for WebSocket notification fan-out ONLY - not workflow orchestration)
	// Temporal handles all workflow logic. RabbitMQ just broadcasts to browser UIs.
	// =========================================================================

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

	workflowQueue := rabbitmq.NewWorkflowQueue(rabbitClient, log)
	if err := workflowQueue.Initialize(context.Background()); err != nil {
		return fmt.Errorf("initializing workflow queue: %w", err)
	}

	log.Info(context.Background(), "RabbitMQ connected for WebSocket notifications")

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
	// Business Layer Dependencies
	// =========================================================================

	// Delegate for UUID generation and timestamps (testing seams).
	del := delegate.New(log)

	// Inventory domain buses - required for allocate_inventory action.
	inventoryItemBus := inventoryitembus.NewBusiness(log, del, inventoryitemdb.NewStore(log, db))
	inventoryLocationBus := inventorylocationbus.NewBusiness(log, del, inventorylocationdb.NewStore(log, db))
	inventoryTransactionBus := inventorytransactionbus.NewBusiness(log, del, inventorytransactiondb.NewStore(log, db))

	// Product bus - required for allocation validation.
	productBus := productbus.NewBusiness(log, del, productdb.NewStore(log, db))

	// Workflow bus - for idempotency tracking.
	workflowBus := workflow.NewBusiness(log, del, workflowdb.NewStore(log, db))

	// Alert bus - required for create_alert action.
	alertBus := alertbus.NewBusiness(log, alertdb.NewStore(log, db))

	// Approval request bus - required for seek_approval action.
	approvalRequestBus := approvalrequestbus.NewBusiness(log, approvalrequestdb.NewStore(log, db))

	// =========================================================================
	// Action Registry
	// =========================================================================

	// All actions now run through the sync activity (ExecuteActionActivity).
	// Temporal handles retries, timeouts, and failure recovery natively.
	// RabbitMQ is used ONLY for real-time WebSocket notifications to browser UIs.
	actionRegistry := workflow.NewActionRegistry()
	workflowactions.RegisterAll(actionRegistry, workflowactions.ActionConfig{
		Log:         log,
		DB:          db,
		QueueClient: workflowQueue, // Enables real-time WebSocket notifications
		Buses: workflowactions.BusDependencies{
			InventoryItem:        inventoryItemBus,
			InventoryLocation:    inventoryLocationBus,
			InventoryTransaction: inventoryTransactionBus,
			Product:              productBus,
			Workflow:             workflowBus,
			Alert:                alertBus,
			ApprovalRequest:      approvalRequestBus,
		},
	})

	// Create async registry for human-in-the-loop actions.
	asyncRegistry := temporal.NewAsyncRegistry()
	asyncRegistry.Register("seek_approval", approval.NewSeekApprovalHandler(log, db, approvalRequestBus, alertBus))

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
	// Temporal resolves struct method names by string: "ExecuteActionActivity" / "ExecuteAsyncActionActivity".
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
