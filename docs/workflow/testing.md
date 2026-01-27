# Testing

Testing patterns and examples for the workflow system.

## Overview

All workflow tests use **real infrastructure** (no mocks):
- Real PostgreSQL via testcontainers (`dbtest.NewDatabase`)
- Real RabbitMQ via testcontainers (`rabbitmq.GetTestContainer`)
- Real workflow engine, queue manager, and event publisher
- Real action handlers

This follows the existing patterns in `queue_test.go`.

## Test Categories

| Category | Location | Purpose |
|----------|----------|---------|
| Unit Tests | `business/sdk/workflow/*_test.go` | Component testing |
| Integration Tests | `business/sdk/workflow/eventpublisher_integration_test.go` | Full flow testing |
| API Tests | `api/cmd/services/ichor/tests/workflow/alertapi/` | HTTP endpoint testing |
| Action Tests | `business/sdk/workflow/workflowactions/*/` | Action handler testing |

## Test Setup Pattern

### Basic Setup

```go
func TestWorkflow(t *testing.T) {
    // Logger
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
    db := dbtest.NewDatabase(t, "Test_Workflow")
    ctx := context.Background()

    // Real workflow business layer
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

    // ... test code
}
```

### With Engine and Queue Manager

```go
// Create engine
engine := workflow.NewEngine(log, db.DB, workflowBus)
if err := engine.Initialize(ctx, workflowBus); err != nil {
    t.Fatalf("initializing engine: %s", err)
}

// Register action handlers
registry := engine.GetRegistry()
registry.Register(communication.NewSendEmailHandler(log, db.DB))
registry.Register(communication.NewCreateAlertHandler(log, db.DB, alertBus))

// Create queue manager
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
```

## EventPublisher Unit Tests

**File**: `business/sdk/workflow/eventpublisher_test.go`

### Test Create Event

```go
func TestEventPublisher_PublishCreateEvent(t *testing.T) {
    // Setup (as above)
    publisher := workflow.NewEventPublisher(log, qm)

    // Get initial metrics
    initialMetrics := qm.GetMetrics()

    // Test entity result
    orderResult := struct {
        ID         string `json:"id"`
        Number     string `json:"number"`
        CustomerID string `json:"customer_id"`
    }{
        ID:         uuid.New().String(),
        Number:     "ORD-001",
        CustomerID: uuid.New().String(),
    }

    // Publish event
    publisher.PublishCreateEvent(ctx, "orders", orderResult, uuid.New())

    // Wait for async processing
    time.Sleep(200 * time.Millisecond)

    // Verify
    finalMetrics := qm.GetMetrics()
    if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
        t.Errorf("Expected TotalEnqueued to increase by 1")
    }
}
```

### Test ID Extraction

```go
func TestEventPublisher_ExtractEntityID(t *testing.T) {
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
            name:    "nil result",
            result:  nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test ID extraction logic
        })
    }
}
```

## Integration Tests

**File**: `business/sdk/workflow/eventpublisher_integration_test.go`

### Full Flow Test

```go
func TestEventPublisher_IntegrationWithRules(t *testing.T) {
    // Setup (as above)

    // Seed workflow data
    adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")
    workflow.TestSeedFullWorkflow(ctx, adminUserID, workflowBus)

    // Create entity in workflow.entities
    entity, _ := workflowBus.QueryEntityByName(ctx, "orders")
    entityType, _ := workflowBus.QueryEntityTypeByName(ctx, "table")
    triggerType, _ := workflowBus.QueryTriggerTypeByName(ctx, "on_create")

    // Create rule
    rule, _ := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
        Name:          "Order Created Rule",
        EntityID:      entity.ID,
        EntityTypeID:  entityType.ID,
        TriggerTypeID: triggerType.ID,
        IsActive:      true,
        CreatedBy:     adminUserID,
    })

    // Create action template
    template, _ := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
        Name:          "Order Email",
        ActionType:    "send_email",
        DefaultConfig: json.RawMessage(`{"recipients": ["test@example.com"]}`),
        CreatedBy:     adminUserID,
    })

    // Create rule action
    workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
        AutomationRuleID: rule.ID,
        Name:             "Send Order Notification",
        ActionConfig:     json.RawMessage(`{"subject": "New Order: {{number}}"}`),
        ExecutionOrder:   1,
        IsActive:         true,
        TemplateID:       &template.ID,
    })

    // Initialize engine AFTER creating rules
    engine := workflow.NewEngine(log, db.DB, workflowBus)
    engine.Initialize(ctx, workflowBus)
    engine.GetRegistry().Register(communication.NewSendEmailHandler(log, db.DB))

    // Create queue manager and publisher
    qm, _ := workflow.NewQueueManager(log, db.DB, engine, client)
    qm.Initialize(ctx)
    qm.ClearQueue(ctx)
    qm.Start(ctx)
    defer qm.Stop(ctx)

    publisher := workflow.NewEventPublisher(log, qm)

    // Get initial metrics
    initialMetrics := qm.GetMetrics()

    // Publish event
    orderResult := map[string]interface{}{
        "id":     uuid.New().String(),
        "number": "ORD-12345",
    }
    publisher.PublishCreateEvent(ctx, "orders", orderResult, adminUserID)

    // Wait for processing
    timeout := time.After(5 * time.Second)
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-timeout:
            t.Fatal("Timeout waiting for event processing")
        case <-ticker.C:
            metrics := qm.GetMetrics()
            if metrics.TotalProcessed > initialMetrics.TotalProcessed {
                // Verify results
                finalMetrics := qm.GetMetrics()
                if finalMetrics.TotalFailed > initialMetrics.TotalFailed {
                    t.Error("Unexpected failures")
                }

                // Check execution history
                history := engine.GetExecutionHistory(10)
                if len(history) == 0 {
                    t.Error("Expected execution history")
                }
                return
            }
        }
    }
}
```

## Action Handler Tests

### Alert Handler Test

**File**: `business/sdk/workflow/workflowactions/communication/alert_test.go`

```go
func TestCreateAlertHandler_Validate(t *testing.T) {
    handler := NewCreateAlertHandler(log, db, alertBus)

    tests := []struct {
        name    string
        config  string
        wantErr bool
    }{
        {
            name: "valid config",
            config: `{
                "message": "Test alert",
                "recipients": {"users": ["5cf37266-3473-4006-984f-9325122678b7"]}
            }`,
            wantErr: false,
        },
        {
            name:    "missing message",
            config:  `{"recipients": {"users": ["uuid"]}}`,
            wantErr: true,
        },
        {
            name:    "no recipients",
            config:  `{"message": "Test"}`,
            wantErr: true,
        },
        {
            name: "invalid severity",
            config: `{
                "message": "Test",
                "severity": "invalid",
                "recipients": {"users": ["uuid"]}
            }`,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := handler.Validate(json.RawMessage(tt.config))
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Alert Handler Execution Test

```go
func TestCreateAlertHandler_Execute(t *testing.T) {
    // Setup with real database and alertBus
    handler := NewCreateAlertHandler(log, db, alertBus)

    config := json.RawMessage(`{
        "alert_type": "test",
        "severity": "high",
        "title": "Test: {{number}}",
        "message": "Order {{number}} created",
        "recipients": {"users": ["5cf37266-3473-4006-984f-9325122678b7"]}
    }`)

    execCtx := workflow.ExecutionContext{
        Event: workflow.TriggerEvent{
            EntityName: "orders",
            EntityID:   uuid.New(),
        },
        RawData: map[string]interface{}{
            "number": "ORD-001",
        },
    }

    result, err := handler.Execute(ctx, config, execCtx)
    if err != nil {
        t.Fatalf("Execute() error: %v", err)
    }

    if result.Status != "success" {
        t.Errorf("Expected success, got %s: %s", result.Status, result.Error)
    }

    // Verify alert was created
    alertID := result.Data["alert_id"].(uuid.UUID)
    alert, err := alertBus.QueryByID(ctx, alertID)
    if err != nil {
        t.Fatalf("QueryByID() error: %v", err)
    }

    if alert.Title != "Test: ORD-001" {
        t.Errorf("Expected title 'Test: ORD-001', got %s", alert.Title)
    }
}
```

## Delegate Handler Tests

**File**: `business/sdk/workflow/delegatehandler_test.go`

```go
func TestDelegateHandler_OrdersCreated(t *testing.T) {
    // Setup
    publisher := workflow.NewEventPublisher(log, qm)
    handler := workflow.NewDelegateHandler(log, publisher)

    // Register domain
    delegate := delegate.New(log)
    handler.RegisterDomain(delegate, ordersbus.DomainName, ordersbus.EntityName)

    // Get initial metrics
    initialMetrics := qm.GetMetrics()

    // Fire delegate event
    order := ordersbus.Order{
        ID:        uuid.New(),
        Number:    "ORD-001",
        CreatedBy: uuid.New(),
    }
    delegate.Call(ctx, ordersbus.ActionCreatedData(order))

    // Wait for async
    time.Sleep(200 * time.Millisecond)

    // Verify
    finalMetrics := qm.GetMetrics()
    if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
        t.Error("Expected event to be enqueued")
    }
}
```

## Running Tests

### Run All Workflow Tests

```bash
go test -v ./business/sdk/workflow/...
```

### Run Specific Test

```bash
go test -v ./business/sdk/workflow/... -run TestEventPublisher
```

### Run with Race Detector

```bash
go test -race -v ./business/sdk/workflow/...
```

### Run Integration Tests Only

```bash
go test -v ./business/sdk/workflow/... -run Integration
```

### Run Alert API Tests

```bash
go test -v ./api/cmd/services/ichor/tests/workflow/alertapi/...
```

## Manual Testing

### Start Services

```bash
make dev-up
```

### Create Order via FormData

```bash
curl -X POST /v1/formdata/{form_id}/upsert \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": {"sales.orders": {"operation": "create", "order": 1}},
    "data": {"sales.orders": {"number": "ORD-001"}}
  }'
```

### Check Logs

```bash
make dev-logs | grep "workflow event"
```

### Check RabbitMQ

Open http://localhost:15672 (guest/guest)

### Query Executions

```bash
make pgcli
SELECT * FROM workflow.automation_executions ORDER BY executed_at DESC LIMIT 5;
```

## Test Data Seeding

### Seed Workflow Infrastructure

```go
adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")
workflow.TestSeedFullWorkflow(ctx, adminUserID, workflowBus)
```

This creates:
- Standard trigger types (on_create, on_update, on_delete, scheduled)
- Standard entity types (table, view)
- Common entities (orders, customers, products, etc.)

### Create Test Rule

```go
rule, _ := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
    Name:          "Test Rule",
    EntityID:      entityID,
    EntityTypeID:  entityTypeID,
    TriggerTypeID: triggerTypeID,
    IsActive:      true,
    CreatedBy:     adminUserID,
})
```

## Metrics

The QueueManager provides metrics for testing:

```go
metrics := qm.GetMetrics()
// metrics.TotalEnqueued - Events queued
// metrics.TotalProcessed - Events processed
// metrics.TotalFailed - Events failed
```

## Troubleshooting Tests

### Test Timeout

Increase timeout or check RabbitMQ connection:
```go
timeout := time.After(10 * time.Second) // Increase from 5s
```

### Events Not Processing

1. Check queue manager started: `qm.Start(ctx)`
2. Check engine initialized: `engine.Initialize(ctx, workflowBus)`
3. Check action handlers registered
4. Check rules exist and are active

### Database Errors

1. Ensure migrations are current
2. Check foreign key constraints
3. Verify test data seeding completed
