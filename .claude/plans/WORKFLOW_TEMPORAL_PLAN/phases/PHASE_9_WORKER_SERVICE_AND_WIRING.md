# Phase 9: Worker Service & Wiring

**Category**: backend
**Status**: Completed
**Dependencies**: All prior implementation phases (3-8). Phase 9 is the integration point that wires the entire Temporal workflow system into a running service.
- Phase 3 (Core Models) - `WorkflowInput`, `TaskQueue`, activity types
- Phase 4 (Graph Executor) - `GraphExecutor` (used internally by workflow)
- Phase 5 (Workflow Implementation) - `ExecuteGraphWorkflow`, `ExecuteBranchUntilConvergence`
- Phase 6 (Activities & Async) - `Activities` struct (holds `Registry` + `AsyncRegistry`), `ExecuteActionActivity`/`ExecuteAsyncActionActivity` as struct methods, `AsyncRegistry`
- Phase 7 (Trigger System) - `WorkflowTrigger`, `NewWorkflowTrigger`, `EdgeStore` interface
- Phase 8 (Edge Store Adapter) - `edgedb.Store` implementing `EdgeStore`

---

## Overview

Wire all Temporal components (Phases 3-8) into a running system. This phase has two main deliverables:

1. **Rewrite `workflow-worker/main.go`** from a placeholder into a full Temporal worker service that connects to Temporal, registers workflows and activities, initializes action handler registries, and processes workflow executions.

2. **Update `all.go`** to initialize `WorkflowTrigger` alongside the existing engine. Phase 9 creates the trigger instance (conditional on Temporal config) but does **not** wire it to the event pipeline yet — that happens in a future phase after end-to-end integration tests validate the worker.

After this phase, the worker processes workflows dispatched via Temporal, and the trigger is ready to be connected to the event pipeline in a subsequent phase.

## Goals

1. **Rewrite `workflow-worker/main.go`** from placeholder to full Temporal worker service with config parsing, DB connection, Temporal client, action registry (using `workflowactions.RegisterCoreActions`), `Activities` struct creation, worker creation with workflow/activity registration, and graceful shutdown
2. **Initialize `WorkflowTrigger` in `all.go`** by creating Temporal client (conditional on config), `edgedb.Store`, standalone `TriggerProcessor`, and `WorkflowTrigger` instance — ready for event pipeline wiring in a future phase
3. **Implement Temporal logger adapter** bridging `foundation/logger.Logger` to Temporal SDK's `log.Logger` interface

## Prerequisites

- All Phases 3-8 complete and compiling
- Existing `workflowactions.RegisterAll` and `workflowactions.RegisterCoreActions` functions
- Existing `workflow.TriggerProcessor` for rule matching:
  ```go
  // Constructor (business/sdk/workflow/trigger.go):
  func NewTriggerProcessor(log *logger.Logger, db *sqlx.DB, workflowBus *Business) *TriggerProcessor
  // Initialization:
  func (tp *TriggerProcessor) Initialize(ctx context.Context) error
  // Rule matching:
  func (tp *TriggerProcessor) ProcessEvent(ctx context.Context, event TriggerEvent) (*ProcessingResult, error)
  ```
- Existing `workflow.EventPublisher` and `workflow.DelegateHandler` for event pipeline
- Temporal SDK in `go.mod`/`vendor` (from Phase 1)
- Temporal K8s deployment running (from Phase 1)
- Understanding of `all.go` initialization order and RabbitMQ conditional wiring
- Database migrations current: `workflow.rule_actions` (v1.69) and `workflow.action_edges` (v1.992) must exist

---

## Go Package Structure

```
api/cmd/services/workflow-worker/
    main.go                            <- THIS PHASE (Task 1 + Task 2 logger adapter)

api/cmd/services/ichor/build/all/
    all.go                             <- THIS PHASE (Task 3)

business/sdk/workflow/temporal/
    models.go                          <- Phase 3 (COMPLETED)
    graph_executor.go                  <- Phase 4 (COMPLETED)
    workflow.go                        <- Phase 5 (COMPLETED)
    activities.go                      <- Phase 6 (COMPLETED)
    activities_async.go                <- Phase 6 (COMPLETED)
    async_completer.go                 <- Phase 6 (COMPLETED)
    trigger.go                         <- Phase 7 (COMPLETED)
    stores/edgedb/edgedb.go            <- Phase 8 (COMPLETED)
```

---

## Existing Wiring Context (all.go)

### Current Workflow Initialization Order (Lines ~415-575)

```
1. WorkflowBus creation (always)
   workflowStore := workflowdb.NewStore(cfg.Log, cfg.DB)
   workflowBus := workflow.NewBusiness(cfg.Log, delegate, workflowStore)

2. ActionRegistry for cascade visualization (always)
   actionRegistry := workflow.NewActionRegistry()
   workflowactions.RegisterCoreActions(actionRegistry, cfg.Log, cfg.DB)

3. RabbitMQ guard: if cfg.RabbitClient != nil && cfg.RabbitClient.IsConnected()
   3a. WorkflowEngine → Initialize()
   3b. WorkflowQueue → QueueManager → Initialize() → Start()
   3c. EventPublisher (after QueueManager)
   3d. RegisterAll action handlers (after EventPublisher)
   3e. DelegateHandler → RegisterDomain (50+ domains)

4. ActionService for manual execution API
```

### Phase 9 Integration Point

The `WorkflowTrigger` needs to be wired **alongside** the existing RabbitMQ-based pipeline. The trigger replaces `Engine.ExecuteWorkflow()` as the execution entry point but still uses the same `EventPublisher` → `QueueManager` → consumer flow for event delivery.

**Two integration strategies**:
- **Option A: Replace at QueueManager level** - Modify `QueueManager.processWorkflowEvent()` to call `WorkflowTrigger.OnEntityEvent()` instead of `Engine.ExecuteWorkflow()`
- **Option B: Add Temporal client to all.go** - Create `WorkflowTrigger` in `all.go` and pass it to the QueueManager as an alternative executor

For Phase 9, we **initialize** the `WorkflowTrigger` in `all.go` but do **not** wire it to the event pipeline. The existing engine continues to handle all execution. Wiring the trigger to replace the engine's execution path is a future phase task, after integration tests validate the worker.

---

## Task Breakdown

### Task 1: Rewrite workflow-worker main.go

**Status**: Completed

**Description**: Replace the placeholder `main.go` with a full Temporal worker service. The worker connects to Temporal, creates action/async registries, instantiates the `Activities` struct, registers workflows and activities on the worker, and runs until shutdown.

**Notes**:
- Configuration parsing using `conf` package with `ICHOR` prefix
- Database connection using `sqldb.Open`
- Temporal client creation using `client.Dial` with logger adapter
- Action registry using existing `workflow.ActionRegistry` + `workflowactions.RegisterCoreActions`
- Async registry using `temporal.NewAsyncRegistry()` (empty initially — async handlers deferred)
- `Activities` struct instantiated with both registries
- Worker creation with `worker.New` on `temporal.TaskQueue`
- Register workflows as package-level functions: `ExecuteGraphWorkflow`, `ExecuteBranchUntilConvergence`
- Register activities via struct: `w.RegisterActivity(&temporal.Activities{...})` — Temporal resolves struct method names by string
- Graceful shutdown with signal handling

**Files**:
- `api/cmd/services/workflow-worker/main.go`

**Implementation Guide**:

```go
package main

import (
    "context"
    "errors"
    "fmt"
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
    // Shutdown
    // =========================================================================

    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    errCh := make(chan error, 1)
    go func() {
        errCh <- w.Run(worker.InterruptCh())
    }()

    select {
    case err := <-errCh:
        return fmt.Errorf("worker error: %w", err)
    case <-ctx.Done():
        log.Info(context.Background(), "shutting down worker")
        return nil
    }
}
```

**Design Decisions**:

1. **`RegisterCoreActions` initially, not `RegisterAll`** - `RegisterAll` requires `RabbitMQ.WorkflowQueue` and multiple bus dependencies (InventoryItem, InventoryLocation, etc.). The worker doesn't have these wired yet. Start with `RegisterCoreActions` (evaluate_condition, update_field, seek_approval, send_email, send_notification) and expand later.

2. **Default Temporal host uses K8s service name** - `temporal-service.ichor-system.svc.cluster.local:7233` matches the Phase 1 K8s deployment. Override with `ICHOR_TEMPORAL_HOSTPORT` for local dev.

3. **Worker concurrency limits** - 100 concurrent activities and 100 concurrent workflow tasks. These are reasonable defaults. Can be tuned later based on load testing.

4. **`Activities` struct registered with both registries** - Even though `AsyncRegistry` is empty, both activity methods (`ExecuteActionActivity`, `ExecuteAsyncActionActivity`) are registered on the worker via the struct. If an async workflow is dispatched before async handlers are wired, it will fail gracefully with "no handler registered" rather than panic.

---

### Task 2: Implement Temporal Logger Adapter

**Status**: Completed

**Description**: Implement a logger adapter that bridges `foundation/logger.Logger` to Temporal SDK's `log.Logger` interface. This is used by Task 1's `client.Dial` and should be defined in the same `main.go` file.

**Notes**:
- Temporal's `log.Logger` interface: `Debug(msg string, keyvals ...interface{})`, `Info(...)`, `Warn(...)`, `Error(...)`
- Our logger interface: `Debug(ctx context.Context, msg string, args ...any)`
- Adapter passes `context.Background()` for the ctx parameter since Temporal's logger interface doesn't provide context — only message + keyvals
- Keyvals from Temporal are forwarded directly to our logger's variadic args
- Defined as private type in `main.go` (no export needed)
- **Context loss tradeoff**: Temporal's `log.Logger` interface does not accept `context.Context`, so trace IDs and request-scoped values are not propagated through Temporal SDK log calls. This is an unavoidable limitation of the Temporal SDK interface. For observability, rely on Temporal's own workflow/run IDs (automatically included in Temporal logs) rather than our application-level trace IDs.

**Files**:
- `api/cmd/services/workflow-worker/main.go` (in same file)

**Implementation Guide**:

```go
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
```

---

### Task 3: Update all.go for Temporal Trigger Initialization

**Status**: Completed

**Description**: Initialize `WorkflowTrigger` in `all.go` conditional on Temporal configuration. Phase 9 creates the trigger instance but does **not** wire it to the event pipeline — that wiring happens in a future phase after integration tests validate the full flow.

**Notes**:
- Add `TemporalHostPort` field to `mux.Config` struct (empty = Temporal disabled)
- Initialize Temporal client with `client.Dial` (conditional on config, fail gracefully)
- Create `edgedb.Store` with logger and DB
- Create standalone `TriggerProcessor` (separate from engine's internal processor)
- Create `WorkflowTrigger` with all dependencies
- Log success — trigger is initialized and ready for future wiring
- Must not break existing tests that don't have Temporal running

**Files**:
- `api/cmd/services/ichor/build/all/all.go`
- `api/sdk/http/mux/mux.go` (add `TemporalHostPort` to Config)

**Implementation Guide**:

**Step 1: Add Temporal config field to `mux.Config`**

```go
// In api/sdk/http/mux/mux.go, add to Config struct:
type Config struct {
    // ... existing fields ...
    TemporalHostPort string // Empty means Temporal disabled
}
```

**Step 2: Add Temporal block in `all.go` (AFTER workflowBus creation, OUTSIDE RabbitMQ guard)**

The key insight: create a **standalone** `TriggerProcessor` rather than trying to share the engine's internal processor. This keeps the Temporal path decoupled from the existing engine.

```go
// =========================================================================
// Temporal Integration (optional, independent of RabbitMQ engine)
// =========================================================================

if cfg.TemporalHostPort != "" {
    tc, err := client.Dial(client.Options{
        HostPort: cfg.TemporalHostPort,
    })
    if err != nil {
        cfg.Log.Error(context.Background(),
            "temporal client creation failed, Temporal dispatch disabled",
            "error", err,
        )
    } else {
        // Create EdgeStore for loading graph definitions from DB.
        edgeStore := edgedb.NewStore(cfg.Log, cfg.DB)

        // Create standalone TriggerProcessor for Temporal use.
        // This is separate from the engine's internal processor —
        // both load the same automation rules from the DB.
        // Constructor: NewTriggerProcessor(log, db, workflowBus)
        triggerProcessor := workflow.NewTriggerProcessor(cfg.Log, cfg.DB, workflowBus)
        if err := triggerProcessor.Initialize(context.Background()); err != nil {
            cfg.Log.Error(context.Background(),
                "temporal trigger processor rule load failed",
                "error", err,
            )
        } else {
            workflowTrigger := temporal.NewWorkflowTrigger(
                cfg.Log,
                tc,                // WorkflowStarter (client.Client satisfies this)
                triggerProcessor,  // RuleMatcher
                edgeStore,         // EdgeStore
            )

            cfg.Log.Info(context.Background(), "temporal workflow trigger initialized",
                "temporal_host", cfg.TemporalHostPort,
            )

            // Phase 9: Trigger is initialized but NOT wired to event pipeline.
            // Future phase will connect workflowTrigger.OnEntityEvent to
            // the QueueManager's event processing path.
            _ = workflowTrigger
        }
    }
}
```

**Step 3: Pass TemporalHostPort through to all.go**

In `main.go` or wherever `mux.Config` is constructed, add the Temporal host:
```go
cfg.TemporalHostPort = appCfg.Temporal.HostPort  // Empty by default = disabled
```

**Design Decisions**:

1. **Standalone TriggerProcessor** — Creating a separate `TriggerProcessor` instance is cleaner than exposing the engine's internal processor. Both load the same automation rules from the DB. The cost is a duplicate in-memory rule set, which is negligible for the rule counts we handle.

2. **Independent of RabbitMQ** — The Temporal block is placed **outside** the `if cfg.RabbitClient != nil` guard. Temporal can work even without RabbitMQ (for sync-only workflows). This also means the Temporal block can be tested independently.

3. **Phase 9 = Initialize Only** — The trigger is created but `workflowTrigger.OnEntityEvent` is not called from anywhere yet. Wiring to the event pipeline is a separate task (future phase) that should be validated with integration tests first.

4. **Fail Gracefully** — If Temporal client creation fails, log a warning and continue. The existing engine handles all workflow execution. No crash, no degraded state.

---

### Task 4: Implement buildFullActionRegistry (DEFERRED)

**Status**: Pending (may be deferred to a later phase)

**Description**: When the worker needs full handler support (beyond `RegisterCoreActions`), implement a `buildFullActionRegistry` that initializes all business layer dependencies needed by `workflowactions.RegisterAll`. This includes instantiating bus objects for inventory, products, alerts, and the workflow domain.

**Notes**:
- Requires business layer bus initialization (similar to `all.go` lines 350-415)
- Need: InventoryItemBus, InventoryLocationBus, InventoryTransactionBus, ProductBus, WorkflowBus, AlertBus
- `QueueClient` will be nil (worker doesn't have RabbitMQ yet) — `CreateAlertHandler` already handles nil gracefully (guards WebSocket publish with `if h.workflowQueue != nil` at alert.go:197, still creates alerts in DB)
- Can be deferred until integration testing requires full handler set (inventory allocation, alert creation workflows)
- For Phase 9, `RegisterCoreActions` provides all 5 handlers needed for basic workflow testing

**Files**:
- `api/cmd/services/workflow-worker/main.go` (extended)

**Implementation Guide**:

```go
// buildFullActionRegistry creates a registry with ALL action handlers.
// Requires business layer bus initialization.
//
// IMPORTANT: QueueClient is nil — handlers that require it (CreateAlertHandler)
// must handle nil gracefully or be registered separately after the worker
// gains RabbitMQ support.
func buildFullActionRegistry(log *logger.Logger, db *sqlx.DB) *workflow.ActionRegistry {
    registry := workflow.NewActionRegistry()

    delegate := delegate.New(log)

    workflowStore := workflowdb.NewStore(log, db)
    workflowBus := workflow.NewBusiness(log, delegate, workflowStore)

    inventoryItemBus := inventoryitembus.NewBusiness(log, delegate,
        inventoryitemdb.NewStore(log, db))
    inventoryLocationBus := inventorylocationbus.NewBusiness(log, delegate,
        inventorylocationdb.NewStore(log, db))
    inventoryTransactionBus := inventorytransactionbus.NewBusiness(log, delegate,
        inventorytransactiondb.NewStore(log, db))

    productBus := productbus.NewBusiness(log, delegate,
        productdb.NewStore(log, db))

    alertBus := alertbus.NewBusiness(log, delegate,
        alertdb.NewStore(log, db))

    workflowactions.RegisterAll(registry, workflowactions.ActionConfig{
        Log:         log,
        DB:          db,
        QueueClient: nil, // No RabbitMQ in worker yet
        Buses: workflowactions.BusDependencies{
            InventoryItem:        inventoryItemBus,
            InventoryLocation:    inventoryLocationBus,
            InventoryTransaction: inventoryTransactionBus,
            Product:              productBus,
            Workflow:             workflowBus,
            Alert:                alertBus,
        },
    })

    return registry
}
```

---

## Phase 9 vs Future Phase Scope

**Phase 9 (This Phase) — Infrastructure Only:**
- Worker service compiles, starts, and connects to Temporal
- Worker registers workflows and `Activities` struct on the task queue
- `RegisterCoreActions` provides 5 sync handlers for basic testing
- Empty `AsyncRegistry` created (no async handlers registered yet)
- `all.go` initializes `WorkflowTrigger` when Temporal is configured
- Trigger is initialized but **NOT** wired to event pipeline
- Workflows can be started manually via Temporal CLI for testing

**Future Phases — Full Integration:**
- Wire `WorkflowTrigger.OnEntityEvent` to QueueManager event pipeline
- Implement async handler adapters (SendEmailHandler, AllocateInventoryHandler)
- Register async handlers in `AsyncRegistry`
- Add RabbitMQ to worker for queue-backed handlers
- Use `RegisterAll` instead of `RegisterCoreActions`
- Full end-to-end: entity event → Temporal workflow → action execution

---

## Validation Criteria

- [ ] `go build ./api/cmd/services/workflow-worker/...` passes
- [ ] `go build ./api/cmd/services/ichor/...` passes
- [ ] Worker starts and connects to Temporal (manual verification with `make dev-bounce`)
- [ ] Worker registers `ExecuteGraphWorkflow` and `ExecuteBranchUntilConvergence` workflows
- [ ] Worker creates `Activities` struct with `Registry` and `AsyncRegistry`
- [ ] Worker registers `Activities` struct via `w.RegisterActivity(&temporal.Activities{...})`
- [ ] Worker handles SIGINT/SIGTERM gracefully
- [ ] Temporal logger adapter implements `log.Logger` interface (4 methods: Debug, Info, Warn, Error)
- [ ] `all.go` changes don't break existing tests (`make test` passes — conditional Temporal wiring)
- [ ] `all.go` initializes `WorkflowTrigger` when `TemporalHostPort` is configured
- [ ] `all.go` creates standalone `TriggerProcessor` for Temporal (not sharing engine's internal processor)
- [ ] `edgedb.Store` used as `EdgeStore` implementation in `all.go`
- [ ] `all.go` fails gracefully when Temporal is unavailable (log warning, continue with engine)

---

## Deliverables

- `api/cmd/services/workflow-worker/main.go` (rewritten from placeholder)
- `api/cmd/services/ichor/build/all/all.go` (updated with Temporal wiring)

---

## Gotchas & Tips

### Common Pitfalls

- **Don't break existing tests** - Many tests run without Temporal. The `all.go` wiring must be conditional. Use an empty `Temporal.HostPort` as the "disabled" signal.

- **`RegisterAll` requires bus dependencies** - The implementation plan reference shows a simple `buildActionRegistry` but the actual `RegisterAll` needs `BusDependencies` (InventoryItemBus, etc.). Start with `RegisterCoreActions` which only needs `log` and `db`.

- **`RegisterAll` needs `QueueClient`** - `CreateAlertHandler` accepts a `*rabbitmq.WorkflowQueue`. The worker doesn't have RabbitMQ. Passing nil is safe: `CreateAlertHandler.publishAlertToWebSocket()` already guards with `if h.workflowQueue != nil` (alert.go:197). The handler creates alerts in the DB but skips WebSocket publishing when queue is nil. No code changes needed.

- **`TriggerProcessor` is internal to `Engine`** - The existing `TriggerProcessor` is created inside `Engine.Initialize()`. For Phase 9, create a **separate** standalone `TriggerProcessor` instance for `WorkflowTrigger`. This keeps the engine and trigger decoupled:
  ```go
  // Constructor requires (log, db, workflowBus) — NOT just (log)
  triggerProcessor := workflow.NewTriggerProcessor(cfg.Log, cfg.DB, workflowBus)
  if err := triggerProcessor.Initialize(ctx); err != nil { /* handle */ }
  // NOTE: Method is Initialize() NOT LoadRules()
  ```
  The cost is a duplicate in-memory rule set, which is negligible. Future phases can refactor to share a single processor if needed.

- **Temporal client in all.go vs worker** - The main Ichor service (`all.go`) needs a Temporal CLIENT (to start workflows). The workflow-worker needs a Temporal WORKER (to process workflows). These are different roles on the same Temporal server.

- **Worker concurrency vs pod resources** - 100 concurrent activities may be too high for a small pod. Match worker concurrency to K8s resource limits.

- **`worker.InterruptCh()` vs signal handling** - Temporal's `worker.InterruptCh()` already handles SIGINT/SIGTERM. The signal.NotifyContext in the implementation is redundant but provides a clean dual-shutdown path.

- **Temporal client lifecycle in all.go** - The `client.Dial` call in `all.go` creates a Temporal client, but `all.go` has no explicit shutdown hook for cleanup. The client connection is cleaned up on process exit. This is acceptable for now; a future phase can add graceful client shutdown if needed.

- **Database migration prerequisite** - The `edgedb.Store` queries `workflow.rule_actions` (migration v1.69) and `workflow.action_edges` (migration v1.992). If migrations are not current, the worker's `edgedb` queries will fail with SQL errors at runtime. Ensure `make migrate` has been run before starting the worker.

### Tips

- Start with Task 1 + Task 2 (worker main.go + logger adapter) since they're self-contained and testable with `make dev-bounce`
- Task 2 (logger adapter) is trivial — include it in the same file as main.go
- Task 3 (all.go wiring) is the most complex — save for last and test carefully
- Task 4 (buildFullActionRegistry) is deferred — skip unless full handler set is needed for testing
- The worker can be verified by checking Temporal UI for registered workers on the task queue
- **Local development Temporal host**: Use `ICHOR_TEMPORAL_HOSTPORT=localhost:7233` after port-forwarding from K8s:
  ```bash
  kubectl port-forward -n ichor-system svc/temporal-service 7233:7233
  ```
- The `all.go` config field defaults to empty (Temporal disabled) — set `ICHOR_TEMPORAL_HOSTPORT` in dev configmap to enable

### StartAsync Adapters (Phase 6 Assessment Follow-up)

Phase 6 identified that `SendEmailHandler` and `AllocateInventoryHandler` need `StartAsync` adapters. These adapters wrap the existing handler's `Execute` method to:
1. Publish work to RabbitMQ with the Temporal task token
2. Return without waiting for completion

This can be implemented as adapter wrappers:
```go
type asyncAdapter struct {
    workflow.ActionHandler
    queueClient *rabbitmq.WorkflowQueue
}

func (a *asyncAdapter) StartAsync(ctx context.Context, config json.RawMessage,
    execCtx workflow.ActionExecutionContext, taskToken []byte) error {
    // Publish to RabbitMQ with task token for correlation
    // The RabbitMQ consumer will call AsyncCompleter.Complete when done
}
```

Registration pattern for async handlers:
```go
// Register on the AsyncRegistry (NOT the sync ActionRegistry):
asyncRegistry.Register("send_email", &asyncAdapter{
    ActionHandler: communication.NewSendEmailHandler(log, db),
    queueClient:   workflowQueue,
})
asyncRegistry.Register("allocate_inventory", &asyncAdapter{
    ActionHandler: inventory.NewAllocateInventoryHandler(log, db, ...),
    queueClient:   workflowQueue,
})
```

**Phase 9 Note**: Async handler registration is deferred. Phase 9 creates an empty `AsyncRegistry` to satisfy the `Activities` struct. Async handlers will be wired when the worker gains RabbitMQ support.

---

## Testing Strategy

### Manual Verification (This Phase)

Phase 9 is primarily verified manually:

1. **Worker startup**:
   ```bash
   make dev-bounce  # Rebuild and restart pods
   kubectl get pods -n ichor-system | grep workflow-worker
   # Expected: workflow-worker-...  1/1  Running
   ```

2. **Temporal registration**:
   ```bash
   make temporal-ui  # Port-forward to localhost:8280
   # Open http://localhost:8280
   # Navigate to Workers tab → verify worker on "ichor-workflow-queue"
   # Check Workflows: ExecuteGraphWorkflow, ExecuteBranchUntilConvergence
   ```

3. **Log verification**:
   ```bash
   make dev-logs-workflow-worker
   # Expected: "starting workflow worker", "task_queue=ichor-workflow-queue"
   ```

4. **Graceful shutdown**:
   ```bash
   kubectl delete pod -n ichor-system -l app=workflow-worker
   # Check logs for "shutting down worker" message
   ```

5. **Manual workflow start** (optional, validates worker processes tasks):
   ```bash
   # Port-forward Temporal for CLI access
   kubectl port-forward -n ichor-system svc/temporal-service 7233:7233
   # Start a test workflow (will fail with "no actions" but proves worker is processing)
   temporal workflow start \
     --task-queue ichor-workflow-queue \
     --type ExecuteGraphWorkflow \
     --workflow-id test-manual-001
   ```

6. **Empty AsyncRegistry error path** (optional, validates graceful failure):
   If an async action type (e.g., `allocate_inventory`) is dispatched before async handlers are wired, the `ExecuteAsyncActionActivity` method will return an error like `"no handler registered for action type: allocate_inventory"`. Verify this in Temporal UI by checking that the activity fails with an error (not a panic) and the workflow reports the failure cleanly.

### Deployment Configuration

To enable Temporal in the K8s dev environment, set `ICHOR_TEMPORAL_HOSTPORT` in the relevant configmaps:

**Ichor service** (`zarf/k8s/dev/ichor/dev-ichor-configmap.yaml`):
```yaml
data:
  ICHOR_TEMPORAL_HOSTPORT: "temporal-service.ichor-system.svc.cluster.local:7233"
```

**Workflow worker** (`zarf/k8s/dev/workflow-worker/dev-workflow-worker-configmap.yaml`):
```yaml
data:
  ICHOR_TEMPORAL_HOSTPORT: "temporal-service.ichor-system.svc.cluster.local:7233"
```

The worker defaults to the K8s service name. The ichor service defaults to empty (Temporal disabled). Set the configmap value to enable trigger initialization.

### Compile-Time Validation

```bash
go build ./api/cmd/services/workflow-worker/...
go build ./api/cmd/services/ichor/...
```

### Integration Testing (Phase 11)

Full end-to-end trigger-to-execution tests are deferred to Phase 11 (Workflow Integration Tests).

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 9

# Review plan before implementing
/workflow-temporal-plan-review 9

# Review code after implementing
/workflow-temporal-review 9
```
