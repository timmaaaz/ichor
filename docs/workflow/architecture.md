# Workflow Engine Architecture

This document describes the architecture of the Ichor workflow automation engine.

## System Overview

The workflow engine enables event-driven automation. When business entities change, events flow through the system to trigger configured actions.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API Layer                                          │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │   ordersapi      │    │  formdataapi     │    │  other apis...   │       │
│  └────────┬─────────┘    └────────┬─────────┘    └────────┬─────────┘       │
└───────────┼──────────────────────┼──────────────────────────┼───────────────┘
            │                      │                          │
            ▼                      ▼                          ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          App Layer                                           │
│  ┌──────────────────┐    ┌──────────────────┐                               │
│  │   ordersapp      │    │  formdataapp     │◄───── EventPublisher          │
│  └────────┬─────────┘    │   (Phase 1)      │       (fires after tx)        │
│           │              └──────────────────┘                               │
└───────────┼─────────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Business Layer                                        │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │   ordersbus      │    │  customersbus    │    │  other buses...  │       │
│  │  ┌────────────┐  │    │                  │    │                  │       │
│  │  │  event.go  │  │    │                  │    │                  │       │
│  │  └────────────┘  │    │                  │    │                  │       │
│  └────────┬─────────┘    └──────────────────┘    └──────────────────┘       │
│           │                                                                  │
│           ▼                                                                  │
│  ┌──────────────────────────────────────────────────────────────────┐       │
│  │                      delegate.Delegate                            │       │
│  │   ┌─────────────────────────────────────────────────────────┐    │       │
│  │   │          DelegateHandler                                 │    │       │
│  │   │   - Listens to: "order/created", "order/updated", etc.  │    │       │
│  │   │   - Publishes to: EventPublisher                         │    │       │
│  │   └─────────────────────────────────────────────────────────┘    │       │
│  └──────────────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Workflow Infrastructure                               │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐       │
│  │  EventPublisher  │───▶│   QueueManager   │───▶│   RabbitMQ       │       │
│  └──────────────────┘    └────────┬─────────┘    └──────────────────┘       │
│                                   │                                          │
│                                   ▼                                          │
│                          ┌──────────────────┐                               │
│                          │  WorkflowEngine  │                               │
│                          │  - TriggerProc   │                               │
│                          │  - ActionExec    │                               │
│                          └──────────────────┘                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### EventPublisher

**Location**: `business/sdk/workflow/eventpublisher.go`

The EventPublisher provides non-blocking workflow event publishing. Events are queued asynchronously - failures are logged but never block the primary operation.

```go
type EventPublisher struct {
    log          *logger.Logger
    queueManager *QueueManager
}

// Standard event methods (non-blocking):
func (ep *EventPublisher) PublishCreateEvent(ctx, entityName, result, userID)
func (ep *EventPublisher) PublishUpdateEvent(ctx, entityName, result, fieldChanges, userID)
func (ep *EventPublisher) PublishDeleteEvent(ctx, entityName, entityID, userID)

// Blocking methods for batch operations (e.g., FormData):
func (ep *EventPublisher) PublishCreateEventsBlocking(ctx, entityName, results []any, userID)
func (ep *EventPublisher) PublishUpdateEventsBlocking(ctx, entityName, results []any, userID)

// Custom event method for async action handlers:
func (ep *EventPublisher) PublishCustomEvent(ctx, event TriggerEvent)
```

**Key behaviors:**
- Events are queued in a goroutine (non-blocking)
- 5-second timeout for queue operations
- Extracts entity ID from result via JSON or reflection
- Logs errors but never fails the primary operation

### DelegateHandler

**Location**: `business/sdk/workflow/delegatehandler.go`

Bridges the delegate pattern to the workflow event system. Listens for domain events and converts them to workflow TriggerEvents.

```go
type DelegateHandler struct {
    log       *logger.Logger
    publisher *EventPublisher
}

// Registration:
func (h *DelegateHandler) RegisterDomain(delegate, domainName, entityName)
```

The entity name is passed directly to `RegisterDomain` rather than stored in a mapping. This simplifies the handler while allowing multiple domains to be registered.

**Event mapping:**
| Domain Action | Workflow Event Type |
|---------------|---------------------|
| `created` | `on_create` |
| `updated` | `on_update` |
| `deleted` | `on_delete` |

### QueueManager

**Location**: `business/sdk/workflow/queue.go`

Manages the RabbitMQ queue for workflow events. Handles publishing, consuming, and processing.

```go
// Simplified view - see queue.go for full implementation
type QueueManager struct {
    log    *logger.Logger
    db     *sqlx.DB
    engine *Engine
    client *rabbitmq.Client
    queue  *rabbitmq.WorkflowQueue
    config QueueConfig
    // ... state management, metrics, circuit breakers
}

// Key methods:
func (qm *QueueManager) QueueEvent(ctx, event TriggerEvent) error
func (qm *QueueManager) Start(ctx) error  // Starts consumers
func (qm *QueueManager) Stop(ctx) error   // Stops consumers
```

> **Note**: The struct shown above is simplified. The actual implementation includes additional fields for state management (`mu`, `isRunning`, `consumers`, `stopChan`, `processingWG`), metrics tracking (`metrics`, `metricsLock`), circuit breaker management (`circuitBreakerManager`), and WebSocket handler registry (`handlerRegistry`).

**Features:**
- Configurable consumer count
- Metrics tracking (enqueued, processed, failed)
- Dead letter queue for failed messages
- Graceful shutdown
- Per-queue-type circuit breakers with global fallback
- WebSocket handler registry for real-time message delivery

### Engine (WorkflowEngine)

**Location**: `business/sdk/workflow/engine.go`

The core engine that evaluates events against rules and executes actions. Implemented as a singleton.

```go
// Simplified view - see engine.go for full implementation
type Engine struct {
    log              *logger.Logger
    db               *sqlx.DB
    workflowBus      *Business

    // Sub-components
    triggerProcessor *TriggerProcessor
    dependencies     *DependencyResolver
    executor         *ActionExecutor

    // State management (mu, isInitialized, activeExecutions, executionHistory, stats, config)
}

// Key methods:
func (e *Engine) ExecuteWorkflow(ctx, event TriggerEvent) (*WorkflowExecution, error)
func (e *Engine) GetRegistry() *ActionRegistry  // Returns registry from executor
func (e *Engine) Initialize(ctx, workflowBus) error
```

> **Note**: The action registry is accessed through the `ActionExecutor` component (`e.executor.GetRegistry()`), not stored directly on the Engine.

**Execution flow:**
1. Receive TriggerEvent
2. TriggerProcessor evaluates which rules match
3. Group actions by execution order (same order = parallel batch)
4. Execute each batch through ActionExecutor
5. Record execution history

### TriggerProcessor

**Location**: `business/sdk/workflow/trigger.go`

Evaluates trigger events against automation rules.

```go
type TriggerProcessor struct {
    log          *logger.Logger
    db           *sqlx.DB
    workflowBus  *Business

    // Cached data
    activeRules  []AutomationRuleView  // Note: uses view model, not base model
    lastLoadTime time.Time
    cacheTimeout time.Duration         // Default: 5 minutes
}

// Key methods:
func (tp *TriggerProcessor) Initialize(ctx) error           // Loads rules on startup
func (tp *TriggerProcessor) ProcessEvent(ctx, event) (*ProcessingResult, error)  // Returns matched rules
func (tp *TriggerProcessor) RefreshRules(ctx) error         // Forces cache reload
```

> **Note**: Rules are loaded internally via `Initialize()` and cached. There is no public `LoadRules()` method - use `RefreshRules()` to force a reload.

**Condition evaluation:**
- Supports 8 operators (equals, not_equals, changed_from, etc.)
- Multiple conditions use AND logic
- Field values extracted from event's RawData
- Previous values for `changed_from` come from FieldChanges

### ActionRegistry

**Location**: `business/sdk/workflow/workflowactions/register.go`

Registry of action handlers by type.

```go
type ActionRegistry struct {
    handlers map[string]ActionHandler
}

// Registration:
func (r *ActionRegistry) Register(handler ActionHandler)
func (r *ActionRegistry) Get(actionType string) (ActionHandler, bool)
```

**Registered handlers:**
- `create_alert` - CreateAlertHandler
- `update_field` - UpdateFieldHandler
- `send_email` - SendEmailHandler
- `send_notification` - SendNotificationHandler
- `seek_approval` - SeekApprovalHandler
- `allocate_inventory` - AllocateInventoryHandler

### ActionHandler Interface

**Location**: `business/sdk/workflow/interfaces.go`

```go
type ActionHandler interface {
    Execute(ctx context.Context, config json.RawMessage, context ActionExecutionContext) (any, error)
    Validate(config json.RawMessage) error
    GetType() string
}
```

All action handlers must implement this interface. The `Execute` method returns `any` to allow different handlers to return different result types - this trades compile-time type safety for runtime flexibility, which is necessary for a plugin/registry system.

### AsyncActionHandler Interface

**Location**: `business/sdk/workflow/interfaces.go`

For actions that queue work asynchronously (like inventory allocation), handlers can implement the extended `AsyncActionHandler` interface:

```go
type AsyncActionHandler interface {
    ActionHandler

    // ProcessQueued processes a queued message asynchronously.
    // payload contains the serialized QueuedPayload (use json.Unmarshal)
    // publisher is used to fire result events for downstream workflow rules
    ProcessQueued(ctx context.Context, payload json.RawMessage, publisher *EventPublisher) error
}
```

The handler is responsible for deserializing the payload, performing the async work, and firing result events via the publisher.

## Event Flow

### Complete Event Lifecycle

```
1. API Request
   └── ordersapi.create(ctx, request)

2. App Layer
   └── ordersapp.Create(ctx, app.NewOrder)
       └── Validates and calls business layer

3. Business Layer
   └── ordersbus.Create(ctx, bus.NewOrder)
       ├── Saves to database
       └── delegate.Call(ctx, ActionCreatedData(order))

4. Delegate System
   └── Dispatches to registered handlers
       └── DelegateHandler receives "order/created"

5. DelegateHandler
   └── Extracts entity data from params
       └── eventPublisher.PublishCreateEvent(ctx, "orders", entity, userID)

6. EventPublisher (goroutine)
   └── Constructs TriggerEvent
       └── queueManager.QueueEvent(ctx, event)

7. RabbitMQ
   └── Message queued to workflow queue

8. QueueManager Consumer
   └── Picks up message
       └── engine.ExecuteWorkflow(ctx, event)

9. WorkflowEngine
   ├── triggerProcessor.EvaluateEvent(event)
   │   └── Returns matched rules
   └── For each matched rule:
       └── actionExecutor.Execute(actions, execCtx)

10. ActionHandler
    └── Executes action (create alert, send email, etc.)

11. Execution History
    └── Records results in automation_executions
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

type FieldChange struct {
    OldValue interface{}
    NewValue interface{}
}
```

### ExecutionContext Structure

```go
type ExecutionContext struct {
    Event       TriggerEvent
    Rule        AutomationRule
    ExecutionID uuid.UUID
    RawData     map[string]interface{}  // Template variable context
}
```

## Initialization

### Application Startup (all.go)

```go
// 1. Create workflow store and business layer
workflowStore := workflowdb.NewStore(cfg.Log, cfg.DB)
workflowBus := workflow.NewBusiness(cfg.Log, workflowStore)

// 2. Create and initialize engine
workflowEngine := workflow.NewEngine(cfg.Log, cfg.DB, workflowBus)
workflowEngine.Initialize(ctx, workflowBus)

// 3. Register action handlers
registry := workflowEngine.GetRegistry()
registry.Register(communication.NewCreateAlertHandler(log, db, alertBus))
registry.Register(communication.NewSendEmailHandler(log, db))
registry.Register(communication.NewSendNotificationHandler(log, db))
registry.Register(data.NewUpdateFieldHandler(log, db))
registry.Register(approval.NewSeekApprovalHandler(log, db))
registry.Register(inventory.NewAllocateInventoryHandler(log, db, ...))

// 4. Create queue manager and start consumers
queueManager, _ := workflow.NewQueueManager(log, db, workflowEngine, rabbitClient)
queueManager.Initialize(ctx)
queueManager.Start(ctx)

// 5. Create event publisher
eventPublisher := workflow.NewEventPublisher(log, queueManager)

// 6. Create delegate handler and register domains
delegateHandler := workflow.NewDelegateHandler(log, eventPublisher)
delegateHandler.RegisterDomain(delegate, ordersbus.DomainName, ordersbus.EntityName)
// ... register other domains
```

## Error Handling

### Non-Blocking Philosophy

Workflow failures should **never** block the primary operation. A failed email notification should not prevent an order from being created.

**Implementation:**
1. EventPublisher queues events in a goroutine
2. Errors are logged but not returned
3. Primary operation always completes
4. Failed events go to dead letter queue for retry/investigation

### Retry Strategy

- RabbitMQ handles message redelivery
- Dead letter queue for messages that fail repeatedly
- Configurable retry count and backoff

### Execution Tracking

All executions are recorded in `workflow.automation_executions`:
- Event details
- Matched rules
- Action results (success/failure)
- Error messages
- Execution timing

## Performance Considerations

### Async Processing

Events are queued to RabbitMQ and processed asynchronously:
- API requests return immediately
- Workflow processing doesn't add latency
- Scalable by adding more consumers

### Parallel Execution

Actions with the same `execution_order` run in parallel:
```
execution_order=1: [email, alert]  ← run in parallel
execution_order=2: [update_field]  ← waits for order 1
execution_order=3: [approval]      ← waits for order 2
```

### Caching

- Rules are loaded once during engine initialization
- Rule changes require engine restart or reload
- Template processor caches parsed templates

## Configuration

### RabbitMQ Settings

```go
RabbitMQ struct {
    URL           string        // amqp://guest:guest@rabbitmq:5672/
    MaxRetries    int           // default: 5
    RetryDelay    time.Duration // default: 5s
    PrefetchCount int           // default: 10
}
```

### Queue Types

The system uses multiple specialized queues for different action types:

| Queue Type | Purpose |
|------------|---------|
| `QueueTypeWorkflow` | General workflow events and data operations |
| `QueueTypeApproval` | Approval request processing |
| `QueueTypeNotification` | Push notification delivery |
| `QueueTypeInventory` | Inventory allocation operations |
| `QueueTypeEmail` | Email sending |
| `QueueTypeAlert` | Alert creation and WebSocket delivery |

**Source**: `business/sdk/workflow/queue.go:289-296`

### Queue Settings

- Durable: Yes
- Auto-delete: No
- Dead letter exchange: `workflow_dlx`

### Circuit Breaker

The queue manager implements circuit breakers to handle downstream service failures gracefully.

**Per-Queue Circuit Breakers:**
- Each queue type has its own circuit breaker
- Failures in one queue type don't affect others
- Default threshold: 50 failures
- Default timeout: 60 seconds

**Global Circuit Breaker:**
- Fallback for catastrophic failures across all queues
- Default threshold: 100 failures
- Default timeout: 120 seconds

**Circuit Breaker States:**
| State | Description |
|-------|-------------|
| `closed` | Normal operation, requests flow through |
| `open` | Too many failures, requests are rejected immediately |
| `half-open` | Recovery mode, allowing test requests through |

**Behavior:**
- When failures exceed the threshold, the breaker opens
- While open, new events for that queue type are rejected
- After the timeout, the breaker transitions to half-open
- A successful request in half-open state closes the breaker
- A failed request in half-open state reopens the breaker

**Source**: `business/sdk/workflow/queue.go:86-130`
