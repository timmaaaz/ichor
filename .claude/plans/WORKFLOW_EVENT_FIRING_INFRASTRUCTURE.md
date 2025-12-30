# Workflow Event Firing Infrastructure

**Status**: Ready for Implementation
**Priority**: High - Required before Default Status Management Phase 2
**Category**: backend (cross-cutting infrastructure)

---

## Overview

Wire the workflow event system into the application so that `TriggerEvent`s are fired when entities are created, updated, or deleted. This is foundational infrastructure that enables automation rules to execute.

**Current State**: The workflow engine is fully implemented and tested (see `TestQueueManager_ProcessMessage` in `queue_test.go:323`), but NO production code calls `QueueEvent()` to fire events. The engine exists but has no triggers.

**Goal**: When an entity is created/updated/deleted via formdata (or domain bus layers), fire a `TriggerEvent` so automation rules can execute.

---

## User Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Entry Points | **Phased** | Start with formdata (Phase 1), add delegate pattern for comprehensive coverage (Phase 2) |
| Error Handling | **Non-blocking** | Log errors, don't fail primary operation. Future: configurable notifications on failure |
| RabbitMQ | **Required** | Service fails to start without it |

---

## How It Works Today (Test Reference)

Reference: `business/sdk/workflow/queue_test.go:323` - `TestQueueManager_ProcessMessage`

```go
// 1. Setup workflow infrastructure
workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))
engine := workflow.NewEngine(log, db.DB, workflowBus)
engine.Initialize(ctx, workflowBus)

// 2. Register action handlers
registry := engine.GetRegistry()
registry.Register(communication.NewSendEmailHandler(log, db.DB))

// 3. Create queue manager and start consumers
qm, _ := workflow.NewQueueManager(log, db.DB, engine, client)
qm.Initialize(ctx)
qm.Start(ctx)

// 4. Fire an event (THIS IS WHAT'S MISSING IN PRODUCTION)
event := workflow.TriggerEvent{
    EventType:  "on_create",
    EntityName: "customers",
    EntityID:   entityID,
    Timestamp:  time.Now(),
    RawData: map[string]interface{}{
        "name":  "Test Customer",
        "email": "test@example.com",
    },
    UserID: adminUserID,
}
qm.QueueEvent(ctx, event)

// 5. Event flows through:
//    RabbitMQ → Consumer → Engine.ExecuteWorkflow() → TriggerProcessor → ActionExecutor
```

---

## Files to Modify

| File | Change |
|------|--------|
| `api/cmd/services/ichor/main.go` | Add RabbitMQ config and initialization |
| `api/sdk/http/mux/mux.go` | Add RabbitMQ client to Config struct |
| `api/cmd/services/ichor/build/all/all.go` | Wire workflow infrastructure, pass to formdata |
| `app/domain/formdata/formdataapp/formdataapp.go` | Add event publisher, fire events post-commit |
| **NEW**: `business/sdk/workflow/eventpublisher.go` | Non-blocking event publishing service |

---

## Phase 1: Implementation Steps

**Execution Order** (optimized for testability):

| Order | Step | Description | Command |
|-------|------|-------------|---------|
| 1 | Step 3 | Create EventPublisher | `implement step 3` |
| 2 | Step 2 | Extend mux.Config | `implement step 2` |
| 3 | Step 1 | Add RabbitMQ to main.go | `implement step 1` |
| 4 | Step 4 | Wire in all.go | `implement step 4` |
| 5 | Step 5 | Modify formdataapp | `implement step 5` |

To execute a step, ask: **"implement step N from WORKFLOW_EVENT_FIRING_INFRASTRUCTURE"**

---

### Step 1: Add RabbitMQ Configuration to main.go

**Status**: [x] Complete

**File**: `api/cmd/services/ichor/main.go`

Add RabbitMQ config to the cfg struct (after OAuth section ~line 127):

```go
RabbitMQ struct {
    URL           string        `conf:"default:amqp://guest:guest@rabbitmq-service:5672/"`
    MaxRetries    int           `conf:"default:5"`
    RetryDelay    time.Duration `conf:"default:5s"`
    PrefetchCount int           `conf:"default:10"`
}
```

Initialize RabbitMQ client after database support (~line 179):

```go
// -------------------------------------------------------------------------
// Initialize RabbitMQ Support

log.Info(ctx, "startup", "status", "initializing RabbitMQ support")

rabbitConfig := rabbitmq.Config{
    URL:           cfg.RabbitMQ.URL,
    MaxRetries:    cfg.RabbitMQ.MaxRetries,
    RetryDelay:    cfg.RabbitMQ.RetryDelay,
    PrefetchCount: cfg.RabbitMQ.PrefetchCount,
}

rabbitClient := rabbitmq.NewClient(log, rabbitConfig)

if err := rabbitClient.WaitForConnection(30 * time.Second); err != nil {
    return fmt.Errorf("connecting to RabbitMQ: %w", err)
}
defer rabbitClient.Close()

log.Info(ctx, "startup", "status", "RabbitMQ connected")
```

Pass to cfgMux (~line 279):

```go
cfgMux := mux.Config{
    Build:        build,
    Log:          log,
    AuthClient:   authClient,
    DB:           db,
    Tracer:       tracer,
    RabbitClient: rabbitClient,  // NEW
}
```

---

### Step 2: Extend mux.Config

**Status**: [x] Complete (implemented as part of Step 1)

**File**: `api/sdk/http/mux/mux.go`

Add import:
```go
"github.com/timmaaaz/ichor/foundation/rabbitmq"
```

Add to Config struct (~line 42):
```go
type Config struct {
    Build        string
    Log          *logger.Logger
    Auth         *auth.Auth
    AuthClient   *authclient.Client
    DB           *sqlx.DB
    Tracer       trace.Tracer
    RabbitClient *rabbitmq.Client  // NEW
}
```

---

### Step 3: Create EventPublisher Service

**Status**: [x] Complete

**NEW File**: `business/sdk/workflow/eventpublisher.go`

```go
package workflow

import (
    "context"
    "encoding/json"
    "fmt"
    "reflect"
    "time"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/foundation/logger"
)

// EventPublisher provides non-blocking workflow event publishing.
// Events are queued asynchronously - failures are logged but never block
// the primary operation.
type EventPublisher struct {
    log          *logger.Logger
    queueManager *QueueManager
}

// NewEventPublisher creates a new event publisher.
func NewEventPublisher(log *logger.Logger, qm *QueueManager) *EventPublisher {
    return &EventPublisher{
        log:          log,
        queueManager: qm,
    }
}

// PublishCreateEvent fires an on_create event for the given entity.
func (ep *EventPublisher) PublishCreateEvent(ctx context.Context, entityName string, result interface{}, userID uuid.UUID) {
    ep.publishEvent(ctx, "on_create", entityName, result, nil, userID)
}

// PublishUpdateEvent fires an on_update event with optional field changes.
func (ep *EventPublisher) PublishUpdateEvent(ctx context.Context, entityName string, result interface{}, fieldChanges map[string]FieldChange, userID uuid.UUID) {
    ep.publishEvent(ctx, "on_update", entityName, result, fieldChanges, userID)
}

// PublishDeleteEvent fires an on_delete event.
func (ep *EventPublisher) PublishDeleteEvent(ctx context.Context, entityName string, entityID uuid.UUID, userID uuid.UUID) {
    event := TriggerEvent{
        EventType:  "on_delete",
        EntityName: entityName,
        EntityID:   entityID,
        Timestamp:  time.Now().UTC(),
        UserID:     userID,
    }
    ep.queueEventNonBlocking(ctx, event)
}

func (ep *EventPublisher) publishEvent(ctx context.Context, eventType, entityName string, result interface{}, fieldChanges map[string]FieldChange, userID uuid.UUID) {
    entityID, rawData, err := ep.extractEntityData(result)
    if err != nil {
        ep.log.Error(ctx, "workflow event: extract entity data failed",
            "entityName", entityName,
            "eventType", eventType,
            "error", err)
        return
    }

    event := TriggerEvent{
        EventType:    eventType,
        EntityName:   entityName,
        EntityID:     entityID,
        FieldChanges: fieldChanges,
        Timestamp:    time.Now().UTC(),
        RawData:      rawData,
        UserID:       userID,
    }

    ep.queueEventNonBlocking(ctx, event)
}

func (ep *EventPublisher) queueEventNonBlocking(ctx context.Context, event TriggerEvent) {
    // Fire in goroutine to avoid blocking the primary operation
    go func() {
        queueCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := ep.queueManager.QueueEvent(queueCtx, event); err != nil {
            ep.log.Error(queueCtx, "workflow event: queue failed",
                "entityName", event.EntityName,
                "entityID", event.EntityID,
                "eventType", event.EventType,
                "error", err)
            // Future: fire notification action for admin alerting
        } else {
            ep.log.Info(queueCtx, "workflow event queued",
                "entityName", event.EntityName,
                "entityID", event.EntityID,
                "eventType", event.EventType)
        }
    }()
}

// extractEntityData extracts ID and raw data from entity result.
func (ep *EventPublisher) extractEntityData(result interface{}) (uuid.UUID, map[string]interface{}, error) {
    if result == nil {
        return uuid.Nil, nil, fmt.Errorf("nil result")
    }

    // JSON marshal/unmarshal to get map representation
    data, err := json.Marshal(result)
    if err != nil {
        return uuid.Nil, nil, fmt.Errorf("marshal result: %w", err)
    }

    var rawData map[string]interface{}
    if err := json.Unmarshal(data, &rawData); err != nil {
        return uuid.Nil, nil, fmt.Errorf("unmarshal to map: %w", err)
    }

    // Extract ID from JSON (app layer uses string IDs)
    var entityID uuid.UUID
    if id, ok := rawData["id"].(string); ok {
        if parsed, err := uuid.Parse(id); err == nil {
            entityID = parsed
        }
    }

    // Fallback: reflection for struct field ID
    if entityID == uuid.Nil {
        entityID = ep.extractIDViaReflection(result)
    }

    return entityID, rawData, nil
}

func (ep *EventPublisher) extractIDViaReflection(result interface{}) uuid.UUID {
    val := reflect.ValueOf(result)
    if val.Kind() == reflect.Ptr {
        val = val.Elem()
    }

    if val.Kind() != reflect.Struct {
        return uuid.Nil
    }

    idField := val.FieldByName("ID")
    if !idField.IsValid() {
        return uuid.Nil
    }

    switch id := idField.Interface().(type) {
    case uuid.UUID:
        return id
    case string:
        if parsed, err := uuid.Parse(id); err == nil {
            return parsed
        }
    }

    return uuid.Nil
}
```

---

### Step 4: Wire Workflow Infrastructure in all.go

**Status**: [x] Complete

**File**: `api/cmd/services/ichor/build/all/all.go`

Add imports:
```go
import (
    // ... existing imports ...
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
)
```

In the `Add()` method, after delegate initialization (~line 300), add workflow infrastructure:

```go
// =========================================================================
// Initialize Workflow Infrastructure
// =========================================================================

var eventPublisher *workflow.EventPublisher

if cfg.RabbitClient != nil && cfg.RabbitClient.IsConnected() {
    workflowStore := workflowdb.NewStore(cfg.Log, cfg.DB)
    workflowBus := workflow.NewBusiness(cfg.Log, workflowStore)

    workflowEngine := workflow.NewEngine(cfg.Log, cfg.DB, workflowBus)
    if err := workflowEngine.Initialize(context.Background(), workflowBus); err != nil {
        cfg.Log.Error(context.Background(), "workflow engine init failed", "error", err)
    } else {
        queueManager, err := workflow.NewQueueManager(cfg.Log, cfg.DB, workflowEngine, cfg.RabbitClient)
        if err != nil {
            cfg.Log.Error(context.Background(), "queue manager creation failed", "error", err)
        } else {
            if err := queueManager.Initialize(context.Background()); err != nil {
                cfg.Log.Error(context.Background(), "queue manager init failed", "error", err)
            } else if err := queueManager.Start(context.Background()); err != nil {
                cfg.Log.Error(context.Background(), "queue manager start failed", "error", err)
            } else {
                eventPublisher = workflow.NewEventPublisher(cfg.Log, queueManager)
                cfg.Log.Info(context.Background(), "workflow event infrastructure initialized")
            }
        }
    }
}
```

Update formdata initialization (~line 840) to pass eventPublisher:

```go
formDataApp := formdataapp.NewApp(formDataRegistry, cfg.DB, formBus, formFieldBus)
formDataApp.SetEventPublisher(eventPublisher)  // NEW
```

---

### Step 5: Modify formdataapp to Fire Events

**Status**: [x] Complete

**File**: `app/domain/formdata/formdataapp/formdataapp.go`

Add field to App struct:
```go
type App struct {
    registry       *formdataregistry.Registry
    db             *sqlx.DB
    formBus        *formbus.Business
    formFieldBus   *formfieldbus.Business
    templateProc   *workflow.TemplateProcessor
    eventPublisher *workflow.EventPublisher  // NEW
}
```

Add setter method:
```go
// SetEventPublisher sets the workflow event publisher (optional).
func (a *App) SetEventPublisher(ep *workflow.EventPublisher) {
    a.eventPublisher = ep
}
```

Add pending event type:
```go
type pendingEvent struct {
    entityName string
    operation  formdataregistry.EntityOperation
    result     any
}
```

Modify `UpsertFormData()` to collect and fire events after commit:

In the operations loop (~line 148), collect events:
```go
// Collect events to fire after commit
var pendingEvents []pendingEvent

for _, step := range plan {
    // ... existing execution code ...

    result, err := a.executeOperation(ctx, step, entityData, templateContext, entityFields, fields)
    if err != nil {
        return FormDataResponse{}, ...
    }

    results[step.EntityName] = result
    templateContext[step.EntityName] = result

    // Queue event for post-commit firing
    if a.eventPublisher != nil {
        pendingEvents = append(pendingEvents, pendingEvent{
            entityName: step.EntityName,
            operation:  step.Operation,
            result:     result,
        })
    }
}
```

After `tx.Commit()` (~line 175), fire events:
```go
// 5. Commit transaction
if err := tx.Commit(); err != nil {
    return FormDataResponse{}, errs.Newf(errs.Internal, "commit: %s", err)
}

// 6. Fire workflow events AFTER successful commit
if a.eventPublisher != nil {
    userID, _ := mid.GetUserID(ctx)
    for _, pe := range pendingEvents {
        switch pe.operation {
        case formdataregistry.OperationCreate:
            a.eventPublisher.PublishCreateEvent(ctx, pe.entityName, pe.result, userID)
        case formdataregistry.OperationUpdate:
            a.eventPublisher.PublishUpdateEvent(ctx, pe.entityName, pe.result, nil, userID)
        }
    }
}

return FormDataResponse{
    Success: true,
    Results: results,
}, nil
```

---

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Event firing location | After tx.Commit() | Ensures data consistency - events only fire for committed data |
| Error handling | Non-blocking (goroutine) | Workflow failures should never block entity creation |
| Multi-entity transactions | Fire for ALL entities | Allows workflows to react to any entity, selective via rule matching |
| FieldChanges for updates | nil (Phase 1) | Full change tracking requires old vs new comparison - deferred |
| EventPublisher injection | Setter method | Keeps constructor simple, allows optional dependency |

---

## Testing Strategy

### Overview

All tests use **real infrastructure** (no mocks):
- Real PostgreSQL via testcontainers (existing `dbtest.NewDatabase`)
- Real RabbitMQ via testcontainers (existing `rabbitmq.GetTestContainer`)
- Real workflow engine, queue manager, and event publisher

This follows the existing patterns in `queue_test.go:323` (TestQueueManager_ProcessMessage).

---

### Test 1: EventPublisher Unit Tests (Real QueueManager)

**File**: `business/sdk/workflow/eventpublisher_test.go`

Tests EventPublisher with real RabbitMQ and QueueManager.

```go
package workflow_test

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/dbtest"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/otel"
    "github.com/timmaaaz/ichor/foundation/rabbitmq"
)

func TestEventPublisher_PublishCreateEvent(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
        return otel.GetTraceID(context.Background())
    })

    // Real RabbitMQ container
    container := rabbitmq.GetTestContainer(t)
    client := rabbitmq.NewTestClient(container.URL)
    if err := client.Connect(); err != nil {
        t.Fatalf("connecting to rabbitmq: %s", err)
    }
    defer client.Close()

    queue := rabbitmq.NewWorkflowQueue(client, log)
    if err := queue.Initialize(context.Background()); err != nil {
        t.Fatalf("initializing queue: %s", err)
    }

    // Real database
    db := dbtest.NewDatabase(t, "Test_EventPublisher_Create")
    ctx := context.Background()

    // Real workflow business layer
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

    // Real workflow engine (no rules needed - just testing event queuing)
    engine := workflow.NewEngine(log, db.DB, workflowBus)
    if err := engine.Initialize(ctx, workflowBus); err != nil {
        t.Fatalf("initializing engine: %s", err)
    }

    // Real queue manager
    qm, err := workflow.NewQueueManager(log, db.DB, engine, client)
    if err != nil {
        t.Fatalf("creating queue manager: %s", err)
    }
    if err := qm.Initialize(ctx); err != nil {
        t.Fatalf("initializing queue manager: %s", err)
    }
    if err := qm.ClearQueue(ctx); err != nil {
        t.Logf("Warning: could not clear queue: %v", err)
    }
    if err := qm.Start(ctx); err != nil {
        t.Fatalf("starting queue manager: %s", err)
    }
    defer qm.Stop(ctx)

    // Create EventPublisher
    publisher := workflow.NewEventPublisher(log, qm)

    // Get initial metrics
    initialMetrics := qm.GetMetrics()

    // Test entity result (simulates app layer response)
    orderResult := struct {
        ID         string `json:"id"`
        Number     string `json:"number"`
        CustomerID string `json:"customer_id"`
    }{
        ID:         uuid.New().String(),
        Number:     "ORD-001",
        CustomerID: uuid.New().String(),
    }

    // Publish create event
    publisher.PublishCreateEvent(ctx, "sales.orders", orderResult, uuid.New())

    // Wait for async event to be queued
    time.Sleep(200 * time.Millisecond)

    // Verify event was enqueued
    finalMetrics := qm.GetMetrics()
    if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
        t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
            initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
    }
}

func TestEventPublisher_PublishUpdateEvent(t *testing.T) {
    // Similar structure - tests update event with FieldChanges
}

func TestEventPublisher_ExtractEntityID(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", nil)

    // Test various result formats
    tests := []struct {
        name     string
        result   interface{}
        expected uuid.UUID
        wantErr  bool
    }{
        {
            name:     "string ID in JSON",
            result:   struct{ ID string `json:"id"` }{ID: "550e8400-e29b-41d4-a716-446655440000"},
            expected: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
        },
        {
            name:     "uuid.UUID field",
            result:   struct{ ID uuid.UUID }{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")},
            expected: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
        },
        {
            name:     "map with string ID",
            result:   map[string]interface{}{"id": "550e8400-e29b-41d4-a716-446655440000"},
            expected: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
        },
        {
            name:    "nil result",
            result:  nil,
            wantErr: true,
        },
    }

    publisher := &workflow.EventPublisher{}
    // Note: extractEntityData needs to be exported or test needs to be in same package

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // This test validates the ID extraction logic
            // Implementation depends on whether extractEntityData is exported
        })
    }
}
```

---

### Test 2: EventPublisher Integration with Workflow Rules

**File**: `business/sdk/workflow/eventpublisher_integration_test.go`

Tests the full flow: EventPublisher → RabbitMQ → Engine → Action Execution

```go
package workflow_test

import (
    "context"
    "encoding/json"
    "os"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/sdk/dbtest"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
    "github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/otel"
    "github.com/timmaaaz/ichor/foundation/rabbitmq"
)

func TestEventPublisher_IntegrationWithRules(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
        return otel.GetTraceID(context.Background())
    })

    // Real RabbitMQ
    container := rabbitmq.GetTestContainer(t)
    client := rabbitmq.NewTestClient(container.URL)
    if err := client.Connect(); err != nil {
        t.Fatalf("connecting to rabbitmq: %s", err)
    }
    defer client.Close()

    queue := rabbitmq.NewWorkflowQueue(client, log)
    queue.Initialize(context.Background())

    // Real database
    db := dbtest.NewDatabase(t, "Test_EventPublisher_Integration")
    ctx := context.Background()

    // Real workflow business layer
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

    // Seed workflow infrastructure
    adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")
    _, err := workflow.TestSeedFullWorkflow(ctx, adminUserID, workflowBus)
    if err != nil {
        t.Fatalf("seeding workflow: %s", err)
    }

    // Create rule for "orders" on_create
    entity, err := workflowBus.QueryEntityByName(ctx, "orders")
    if err != nil {
        t.Fatalf("querying entity: %s", err)
    }

    entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
    if err != nil {
        t.Fatalf("querying entity type: %s", err)
    }

    triggerType, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
    if err != nil {
        t.Fatalf("querying trigger type: %s", err)
    }

    rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
        Name:          "Order Created Rule",
        Description:   "Fires when order is created",
        EntityID:      entity.ID,
        EntityTypeID:  entityType.ID,
        TriggerTypeID: triggerType.ID,
        IsActive:      true,
        CreatedBy:     adminUserID,
    })
    if err != nil {
        t.Fatalf("creating rule: %s", err)
    }

    // Create email action
    emailTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
        Name:          "Order Email",
        ActionType:    "send_email",
        DefaultConfig: json.RawMessage(`{"recipients": ["test@example.com"]}`),
        CreatedBy:     adminUserID,
    })
    if err != nil {
        t.Fatalf("creating template: %s", err)
    }

    _, err = workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
        AutomationRuleID: rule.ID,
        Name:             "Send Order Notification",
        ActionConfig:     json.RawMessage(`{"subject": "New Order: {{number}}", "body": "Order created"}`),
        ExecutionOrder:   1,
        IsActive:         true,
        TemplateID:       &emailTemplate.ID,
    })
    if err != nil {
        t.Fatalf("creating rule action: %s", err)
    }

    // Initialize engine AFTER creating rules
    engine := workflow.NewEngine(log, db.DB, workflowBus)
    if err := engine.Initialize(ctx, workflowBus); err != nil {
        t.Fatalf("initializing engine: %s", err)
    }

    // Register action handlers
    engine.GetRegistry().Register(communication.NewSendEmailHandler(log, db.DB))

    // Create queue manager
    qm, err := workflow.NewQueueManager(log, db.DB, engine, client)
    if err != nil {
        t.Fatalf("creating queue manager: %s", err)
    }
    qm.Initialize(ctx)
    qm.ClearQueue(ctx)
    qm.Start(ctx)
    defer qm.Stop(ctx)

    // Small delay for consumers
    time.Sleep(100 * time.Millisecond)

    // Create EventPublisher
    publisher := workflow.NewEventPublisher(log, qm)

    // Get initial metrics
    initialMetrics := qm.GetMetrics()

    // Publish event via EventPublisher (simulating formdataapp)
    orderResult := map[string]interface{}{
        "id":          uuid.New().String(),
        "number":      "ORD-12345",
        "customer_id": uuid.New().String(),
        "due_date":    time.Now().Format(time.RFC3339),
    }

    publisher.PublishCreateEvent(ctx, "orders", orderResult, adminUserID)

    // Wait for async processing
    processed := false
    timeout := time.After(5 * time.Second)
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for !processed {
        select {
        case <-timeout:
            metrics := qm.GetMetrics()
            t.Fatalf("Timeout - Enqueued: %d, Processed: %d, Failed: %d",
                metrics.TotalEnqueued, metrics.TotalProcessed, metrics.TotalFailed)
        case <-ticker.C:
            metrics := qm.GetMetrics()
            if metrics.TotalProcessed > initialMetrics.TotalProcessed {
                processed = true
            }
        }
    }

    // Verify results
    finalMetrics := qm.GetMetrics()

    if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
        t.Errorf("Expected 1 event enqueued, got %d", finalMetrics.TotalEnqueued-initialMetrics.TotalEnqueued)
    }

    if finalMetrics.TotalProcessed != initialMetrics.TotalProcessed+1 {
        t.Errorf("Expected 1 event processed, got %d", finalMetrics.TotalProcessed-initialMetrics.TotalProcessed)
    }

    if finalMetrics.TotalFailed > initialMetrics.TotalFailed {
        t.Errorf("Unexpected failures: %d", finalMetrics.TotalFailed-initialMetrics.TotalFailed)
    }

    // Verify execution history
    history := engine.GetExecutionHistory(10)
    if len(history) == 0 {
        t.Error("Expected at least one execution in history")
    } else {
        lastExec := history[0]
        if lastExec.ExecutionPlan.MatchedRuleCount != 1 {
            t.Errorf("Expected 1 matched rule, got %d", lastExec.ExecutionPlan.MatchedRuleCount)
        }

        // Verify email action executed
        actionExecuted := false
        for _, batch := range lastExec.BatchResults {
            for _, rule := range batch.RuleResults {
                for _, action := range rule.ActionResults {
                    if action.ActionType == "send_email" && action.Status == "success" {
                        actionExecuted = true
                    }
                }
            }
        }
        if !actionExecuted {
            t.Error("Expected email action to execute successfully")
        }
    }

    t.Log("EventPublisher integration test with rules completed successfully")
}
```

---

### Test 3: FormData App with EventPublisher

**File**: `app/domain/formdata/formdataapp/formdataapp_workflow_test.go`

Tests formdataapp.UpsertFormData firing workflow events.

```go
package formdataapp_test

import (
    "context"
    "encoding/json"
    "os"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/app/domain/formdata/formdataapp"
    "github.com/timmaaaz/ichor/business/sdk/dbtest"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
    "github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/rabbitmq"
)

func TestFormdataApp_FiresWorkflowEvents(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", nil)

    // Real RabbitMQ
    container := rabbitmq.GetTestContainer(t)
    client := rabbitmq.NewTestClient(container.URL)
    client.Connect()
    defer client.Close()

    queue := rabbitmq.NewWorkflowQueue(client, log)
    queue.Initialize(context.Background())

    // Real database with full BusDomain
    db := dbtest.NewDatabase(t, "Test_FormData_Workflow")
    ctx := context.Background()

    // Setup workflow infrastructure
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))
    adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")
    workflow.TestSeedFullWorkflow(ctx, adminUserID, workflowBus)

    // Create workflow rule (similar to above)
    // ... create rule for entity that will be tested ...

    // Setup engine and queue manager
    engine := workflow.NewEngine(log, db.DB, workflowBus)
    engine.Initialize(ctx, workflowBus)
    engine.GetRegistry().Register(communication.NewSendEmailHandler(log, db.DB))

    qm, _ := workflow.NewQueueManager(log, db.DB, engine, client)
    qm.Initialize(ctx)
    qm.ClearQueue(ctx)
    qm.Start(ctx)
    defer qm.Stop(ctx)

    // Create EventPublisher
    publisher := workflow.NewEventPublisher(log, qm)

    // Create formdataapp with EventPublisher
    // Use db.BusDomain which has all business layer instances
    formDataApp := formdataapp.NewApp(
        nil, // registry - would need to build for this test
        db.DB,
        db.BusDomain.Form,
        db.BusDomain.FormField,
    )
    formDataApp.SetEventPublisher(publisher)

    // Get initial metrics
    initialMetrics := qm.GetMetrics()

    // Call UpsertFormData
    // This requires setting up a form and formdata request
    // The exact implementation depends on your test data setup

    // Wait for async processing
    time.Sleep(500 * time.Millisecond)

    // Verify workflow event was fired
    finalMetrics := qm.GetMetrics()
    if finalMetrics.TotalEnqueued <= initialMetrics.TotalEnqueued {
        t.Error("Expected workflow event to be enqueued after formdata operation")
    }

    t.Log("FormData workflow integration test completed")
}
```

---

### Test Files Summary

| File | Location | What It Tests |
|------|----------|---------------|
| `eventpublisher_test.go` | `business/sdk/workflow/` | EventPublisher with real RabbitMQ (basic queuing) |
| `eventpublisher_integration_test.go` | `business/sdk/workflow/` | Full flow: Publisher → Queue → Engine → Action |
| `formdataapp_workflow_test.go` | `app/domain/formdata/formdataapp/` | FormData calling EventPublisher |

All tests use:
- Real PostgreSQL (via `dbtest.NewDatabase`)
- Real RabbitMQ (via `rabbitmq.GetTestContainer`)
- Real workflow engine and queue manager
- Real action handlers

---

### Manual Testing

```bash
# 1. Start services with RabbitMQ
make dev-up

# 2. Create order via formdata
curl -X POST /v1/formdata/{form_id}/upsert \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": {"sales.orders": {"operation": "create", "order": 1}},
    "data": {"sales.orders": {"number": "ORD-001", ...}}
  }'

# 3. Check logs for workflow event
make dev-logs | grep "workflow event"

# 4. Check RabbitMQ management UI
# http://localhost:15672 (guest/guest)

# 5. Query automation_executions table
make pgcli
SELECT * FROM workflow.automation_executions ORDER BY executed_at DESC LIMIT 5;
```

---

## Validation Checklist

- [ ] RabbitMQ client connects on application startup
- [ ] Workflow engine initializes with queue manager
- [ ] Queue manager starts and consumers are active
- [ ] `go build ./...` passes
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Creating entity via formdata fires `on_create` event
- [ ] Event appears in RabbitMQ queue (check logs or UI)
- [ ] If automation rule exists, it matches and executes

---

## Phase 2 (Future): Delegate Pattern Integration

After Phase 1 is stable, add comprehensive coverage via delegate pattern:

1. Create `business/sdk/workflow/delegatehandler.go`
2. Register workflow handler with delegate in all.go
3. Add event.go to domain bus layers (ordersbus, etc.)
4. Call `delegate.Call()` in Create/Update/Delete methods

This provides event firing from ALL entry points (direct API, formdata, internal calls).

---

## Related Plans

- **Default Status Management Phase 2**: Depends on this infrastructure being complete
- **Default Status Management Phase 3 (Alerts)**: Uses `create_alert` action handler

---

## Notes

- Event firing should be **non-blocking** - don't fail the request if workflow fails
- Log errors but allow the primary operation to succeed
- The `FieldChanges` field in `TriggerEvent` is used for `on_update` events to detect what changed - this requires tracking old vs new values (deferred to Phase 2)
