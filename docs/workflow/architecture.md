# Workflow Engine Architecture

This document describes the architecture of the Ichor workflow automation engine, powered by [Temporal](https://temporal.io).

## System Overview

The workflow engine enables event-driven automation. When business entities change, events flow through the system to trigger configured actions. The system is split across two services:

- **ichor** (main API): Captures entity events and dispatches workflows to Temporal
- **workflow-worker**: Picks up workflow tasks from Temporal and executes action graphs

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API Layer (ichor service)                          │
│  ordersapi / formdataapi / other apis...                                    │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │ delegate.Call()
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Trigger Side (ichor service)                          │
│  TemporalDelegateHandler → WorkflowTrigger → TriggerProcessor               │
│                                   │                                          │
│                                   │ client.ExecuteWorkflow()                │
└───────────────────────────────────┼─────────────────────────────────────────┘
                                    │ (Temporal task queue)
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Execution Side (workflow-worker service)                │
│  ExecuteGraphWorkflow → GraphExecutor → Activities → ActionHandlers         │
│  ├── Linear: executeSingleAction                                            │
│  ├── Parallel: executeParallelWithConvergence                               │
│  └── Fire-and-forget: child workflows + PARENT_CLOSE_POLICY_ABANDON         │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### TemporalDelegateHandler

**Location**: `business/sdk/workflow/temporal/delegatehandler.go`

Bridges the delegate pattern to the Temporal workflow system. Listens for domain events and dispatches them to the WorkflowTrigger.

```go
type DelegateHandler struct {
    log     *logger.Logger
    trigger *WorkflowTrigger
}

func (h *DelegateHandler) RegisterDomain(delegate, domainName, entityName)
```

**Key behaviors:**
- Dispatches events in a goroutine (non-blocking, fail-open)
- Extracts entity data via JSON marshaling and reflection
- Same `RegisterDomain` interface as old handler — domain registration unchanged

**Event mapping:**
| Domain Action | Workflow Event Type |
|---------------|---------------------|
| `created` | `on_create` |
| `updated` | `on_update` |
| `deleted` | `on_delete` |

### WorkflowTrigger

**Location**: `business/sdk/workflow/temporal/trigger.go`

Receives TriggerEvents, evaluates matching rules via TriggerProcessor, and dispatches Temporal workflows for each match.

```go
type WorkflowTrigger struct {
    log       *logger.Logger
    matcher   RuleMatcher
    edgeStore EdgeStore
    starter   WorkflowStarter
}

func (wt *WorkflowTrigger) OnEntityEvent(ctx, event TriggerEvent) error
```

**Key behaviors:**
- Loads graph definition (actions + edges) from EdgeStore for each matched rule
- Dispatches workflow to Temporal with `WorkflowInput` containing the full graph
- Workflow ID format: `workflow-{ruleID}-{entityID}-{executionID}`
- Fail-open per rule: individual rule failures logged and skipped, other rules still dispatched

### GraphExecutor

**Location**: `business/sdk/workflow/temporal/graph_executor.go`

Deterministic graph traversal engine that interprets visual workflow graphs at runtime.

```go
type GraphExecutor struct {
    actions    map[uuid.UUID]ActionNode
    edges      []ActionEdge
    bySource   map[uuid.UUID][]ActionEdge
    byTarget   map[uuid.UUID][]ActionEdge
}

func NewGraphExecutor(graph GraphDefinition) *GraphExecutor
func (ge *GraphExecutor) GetStartActions() []ActionNode
func (ge *GraphExecutor) GetNextActions(actionID uuid.UUID, result map[string]any) []ActionNode
func (ge *GraphExecutor) FindConvergencePoint(actionIDs []uuid.UUID) uuid.UUID
```

**Determinism guarantees:**
- Sorted map iteration (by SortOrder, then UUID tie-breaking)
- No non-deterministic operations (no `time.Now()`, `rand`, or bare goroutines)
- Safe for Temporal workflow replay

**Supports all 5 edge types:**
| Type | Description |
|------|-------------|
| `start` | Entry point (source_action_id is null) |
| `sequence` | Unconditional linear flow |
| `true_branch` | Followed when condition evaluates to true |
| `false_branch` | Followed when condition evaluates to false |
| `always` | Always-execute path from a condition |

**Convergence detection:**
- BFS-based reachability analysis from all parallel branches
- Identifies the nearest node reachable from all branches (convergence point)
- `uuid.Nil` convergence point signals fire-and-forget (no waiting)

### Workflow Implementation

**Location**: `business/sdk/workflow/temporal/workflow.go`

The main Temporal workflow that interprets action graphs.

**Key functions:**
- `ExecuteGraphWorkflow` — main entry point, validates input, traverses graph
- `executeSingleAction` — executes one action as a Temporal activity
- `executeParallelWithConvergence` — forks child workflows, waits for all at convergence
- `executeFireAndForget` — child workflows with `PARENT_CLOSE_POLICY_ABANDON`
- `ExecuteBranchUntilConvergence` — child workflow for a single branch

**Continue-As-New:**
- Triggers at `HistoryLengthThreshold` (10,000 history events)
- Preserves full `MergedContext` via `ContinuationState` on `WorkflowInput`
- Prevents unbounded history growth for long-running workflows

**Versioning:**
- `workflow.GetVersion(ctx, "graph-interpreter", DefaultVersion, 1)`
- Enables safe schema evolution without breaking in-flight workflows

### Activities

**Location**: `business/sdk/workflow/temporal/activities.go`

Activities bridge Temporal's activity system to action handlers.

```go
type Activities struct {
    Registry      *workflowactions.ActionRegistry
    AsyncRegistry *AsyncRegistry
}

func (a *Activities) ExecuteActionActivity(ctx, input ActionActivityInput) (ActionActivityOutput, error)
func (a *Activities) ExecuteAsyncActionActivity(ctx, input ActionActivityInput) (ActionActivityOutput, error)
```

**Activity routing:**
- Regular actions → `ExecuteActionActivity` (MaximumAttempts=3)
- Async actions (send_email, allocate_inventory) → `ExecuteAsyncActionActivity` (MaximumAttempts=1)
- Human actions (manager_approval, manual_review, etc.) → `ExecuteAsyncActionActivity` (MaximumAttempts=1)

### TriggerProcessor

**Location**: `business/sdk/workflow/trigger.go`

Evaluates trigger events against automation rules. Unchanged from the old architecture.

```go
type TriggerProcessor struct {
    log          *logger.Logger
    db           *sqlx.DB
    workflowBus  *Business
    activeRules  []AutomationRuleView
    lastLoadTime time.Time
    cacheTimeout time.Duration
}

func (tp *TriggerProcessor) Initialize(ctx) error
func (tp *TriggerProcessor) ProcessEvent(ctx, event) (*ProcessingResult, error)
func (tp *TriggerProcessor) RefreshRules(ctx) error
```

> **Note**: Rules are loaded via `Initialize()` and cached. Use `RefreshRules()` to force a reload. Cache invalidation is registered via `RegisterCacheInvalidation(delegate)`.

### ActionRegistry

**Location**: `business/sdk/workflow/workflowactions/register.go`

Registry of action handlers by type.

```go
type ActionRegistry struct {
    handlers map[string]ActionHandler
}

func (r *ActionRegistry) Register(handler ActionHandler)
func (r *ActionRegistry) Get(actionType string) (ActionHandler, bool)
```

**Registered handlers:**
- `create_alert` — CreateAlertHandler
- `update_field` — UpdateFieldHandler
- `send_email` — SendEmailHandler
- `send_notification` — SendNotificationHandler
- `seek_approval` — SeekApprovalHandler
- `allocate_inventory` — AllocateInventoryHandler
- `evaluate_condition` — ConditionHandler (branching support)

### ActionHandler Interface

**Location**: `business/sdk/workflow/interfaces.go`

```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
    SupportsManualExecution() bool
    IsAsync() bool
    GetDescription() string
}
```

All 7 action handlers implement this interface. The interface is unchanged from the old architecture.

### EntityModifier Interface

**Location**: `business/sdk/workflow/interfaces.go`

Optional interface for action handlers that modify database entities. Used for cascade visualization.

```go
type EntityModifier interface {
    GetEntityModifications(config json.RawMessage) []EntityModification
}
```

Currently implemented by `UpdateFieldHandler`.

See [cascade-visualization.md](cascade-visualization.md) for details.

## Event Flow

### Complete Event Lifecycle

```
1. API Request
   └── ordersapi.create(ctx, request)

2. App Layer
   └── ordersapp.Create(ctx, app.NewOrder)

3. Business Layer
   └── ordersbus.Create(ctx, bus.NewOrder)
       ├── Saves to database
       └── delegate.Call(ctx, ActionCreatedData(order))

4. Delegate System
   └── Dispatches to registered handlers
       └── TemporalDelegateHandler receives "order/created"

5. TemporalDelegateHandler (goroutine)
   └── Extracts entity data
       └── workflowTrigger.OnEntityEvent(ctx, triggerEvent)

6. WorkflowTrigger
   ├── triggerProcessor.ProcessEvent(event)
   │   └── Returns matched rules
   └── For each matched rule:
       ├── edgeStore.QueryActionsByRule / QueryEdgesByRule
       └── temporalClient.ExecuteWorkflow(ctx, workflowInput)

7. Temporal Server
   └── Task queued on "ichor-workflow" task queue

8. workflow-worker
   └── Picks up task
       └── ExecuteGraphWorkflow(ctx, workflowInput)

9. GraphExecutor
   ├── GetStartActions() → first actions
   └── For each action:
       ├── Execute as Temporal activity
       └── GetNextActions() → follow edges

10. ActionHandler
    └── Executes action (create alert, send email, etc.)
```

### TriggerEvent Structure

```go
type TriggerEvent struct {
    EventType    string                    // "on_create", "on_update", "on_delete"
    EntityName   string                    // "orders", "customers", etc.
    EntityID     uuid.UUID                 // The entity's UUID
    FieldChanges map[string]FieldChange    // For on_update events
    Timestamp    time.Time                 // When event occurred
    RawData      map[string]interface{}    // Entity data snapshot
    UserID       uuid.UUID                 // User who triggered
}
```

## Initialization

### Application Startup (all.go)

```go
// 1. Create workflow store and business layer
workflowStore := workflowdb.NewStore(cfg.Log, cfg.DB)
workflowBus := workflow.NewBusiness(cfg.Log, workflowStore)

// 2. Connect to Temporal (conditional on TemporalHostPort)
if cfg.TemporalHostPort != "" {
    temporalClient, _ := client.Dial(client.Options{HostPort: cfg.TemporalHostPort})

    // 3. Create edge store for loading graph definitions
    edgeStore := edgedb.NewStore(cfg.Log, cfg.DB)

    // 4. Create and initialize trigger processor
    triggerProcessor := workflow.NewTriggerProcessor(cfg.Log, cfg.DB, workflowBus)
    triggerProcessor.Initialize(ctx)
    triggerProcessor.RegisterCacheInvalidation(delegate)

    // 5. Create workflow trigger
    workflowTrigger := temporal.NewWorkflowTrigger(cfg.Log, triggerProcessor, edgeStore, temporalClient)

    // 6. Create delegate handler and register ~60 domains
    delegateHandler := temporal.NewDelegateHandler(cfg.Log, workflowTrigger)
    delegateHandler.RegisterDomain(delegate, ordersbus.DomainName, ordersbus.EntityName)
    // ... register other domains
}
```

### Worker Startup (workflow-worker/main.go)

```go
// 1. Connect to Temporal and database
temporalClient, _ := client.Dial(client.Options{HostPort: cfg.Temporal.HostPort})
db := sqldb.Open(cfg.DB)

// 2. Build action registry
registry := workflowactions.RegisterCoreActions(log, db)

// 3. Create worker
w := worker.New(temporalClient, temporal.TaskQueue, worker.Options{})
w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)
w.RegisterActivity(&temporal.Activities{Registry: registry, AsyncRegistry: temporal.NewAsyncRegistry()})

// 4. Start worker (w.Start() + signal wait + w.Stop() for health check support)
w.Start()
// ... signal handling ...
w.Stop()
```

## Error Handling

### Non-Blocking Philosophy

Workflow failures should **never** block the primary operation. A failed email notification should not prevent an order from being created.

**Implementation:**
1. TemporalDelegateHandler dispatches in a goroutine
2. Errors are logged but not returned to the business layer
3. Primary operation always completes
4. Temporal provides durable retry for failed workflows

### Retry Strategy

Temporal handles retries automatically based on activity type:

| Activity Type | MaximumAttempts | Rationale |
|---------------|-----------------|-----------|
| Regular (evaluate_condition, update_field, etc.) | 3 | Safe to retry |
| Async (send_email, allocate_inventory) | 1 | Prevent duplicate side effects |
| Human (manager_approval, manual_review, etc.) | 1 | Prevent duplicate side effects |

Per-rule fail-open: if one rule's workflow dispatch fails, other matched rules still execute.

### Execution Tracking

All executions are recorded in `workflow.automation_executions`:
- Event details
- Matched rules
- Action results (success/failure)
- Error messages
- Execution timing

Temporal also provides built-in workflow visibility via the Temporal UI (port 8280).

## Performance Considerations

### Async Processing

Events are dispatched to Temporal and processed by the workflow-worker:
- API requests return immediately
- Workflow processing doesn't add latency to the primary operation
- Scalable by adding more workflow-worker replicas

### Graph-Based Execution

All rules with actions use graph-based traversal via directed edges. Rules without actions are saved as inactive drafts.

**Execution patterns:**
- **Linear**: Actions execute sequentially via `sequence` edges
- **Branching**: Condition actions route to `true_branch` or `false_branch`
- **Parallel**: Multiple next actions fork into child workflows that converge
- **Fire-and-forget**: Parallel branches with no convergence point

### Caching

- Rules are loaded by TriggerProcessor on startup and cached (5-minute TTL)
- Cache invalidation registered via delegate events (rule create/update/delete)
- Template processor caches parsed templates

## Configuration

### Temporal Settings

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `ICHOR_TEMPORAL_HOSTPORT` | Temporal server address | `""` (disabled) |

When `TemporalHostPort` is empty, the Temporal trigger system is disabled and events are not dispatched to workflows.

### Task Queue

The system uses a single task queue: `ichor-workflow`

All workflow types and activities are registered on this queue. Multiple workers can process the same queue for horizontal scaling.

## Related Documentation

- [Temporal Integration](temporal.md) — Detailed Temporal architecture and design
- [Worker Deployment](worker-deployment.md) — Worker service operations
- [Migration from RabbitMQ](migration-from-rabbitmq.md) — What changed and why
- [Database Schema](database-schema.md) — Workflow table definitions
- [Event Infrastructure](event-infrastructure.md) — Delegate pattern and workflow dispatch
- [Actions Overview](actions/overview.md) — Action handler interface and registry
- [Adding Domains](adding-domains.md) — How to add workflow events to domains
- [Testing](testing.md) — Testing patterns and examples
