# Circuit Breaker Improvement Plan

## Problem

The current circuit breaker implementation in the workflow engine has critical design issues for an enterprise ERP system:

1. **Single global breaker**: One circuit breaker controls ALL workflows - email, approvals, inventory, notifications
2. **Very low threshold**: Opens after just 5 failures, trivially easy to trigger
3. **Over-broad blast radius**: A flaky email server blocks inventory updates, approval workflows, everything
4. **Hardcoded values**: Threshold (5) and timeout (30s) are hardcoded despite config fields existing

## Solution

Implement **per-queue-type circuit breakers** with **configurable thresholds**.

### Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Granularity | Per-QueueType (6 breakers) | Natural fit - queue types already exist and represent different reliability domains |
| Default threshold | 50 failures | Tolerant of brief outages; ~10x current value |
| Global fallback | 100 failures | Only trips on catastrophic system-wide failure |
| Timeout | Configurable, default 60s | Longer recovery window for enterprise workloads |

### Queue Types (already defined in rabbitmq)

- `workflow` - General workflow messages
- `approval` - Approval workflows
- `notification` - Notifications
- `inventory` - Inventory operations
- `email` - Email messages (most likely to fail)
- `alert` - Alerts

## Implementation

### Files to Modify

1. **business/sdk/workflow/queue.go** - Main changes

### Step 1: Update CircuitBreaker struct

Add support for per-queue-type breakers:

```go
// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
    mu       sync.RWMutex
    breakers map[rabbitmq.QueueType]*CircuitBreaker
    global   *CircuitBreaker  // Fallback for catastrophic failures
    config   CircuitBreakerConfig
}

type CircuitBreakerConfig struct {
    // Per-queue-type settings (can be customized per queue)
    DefaultThreshold int           // Default: 50
    DefaultTimeout   time.Duration // Default: 60s

    // Global fallback settings
    GlobalThreshold  int           // Default: 100
    GlobalTimeout    time.Duration // Default: 120s

    // Optional per-queue overrides
    QueueOverrides   map[rabbitmq.QueueType]CircuitBreakerSettings
}

type CircuitBreakerSettings struct {
    Threshold int
    Timeout   time.Duration
}
```

### Step 2: Update QueueConfig

Wire up the existing config fields and add new ones:

```go
type QueueConfig struct {
    // ... existing fields ...

    // Circuit breaker settings
    CircuitBreakerThreshold       int           // Per-queue default: 50
    CircuitBreakerTimeout         time.Duration // Per-queue default: 60s
    GlobalCircuitBreakerThreshold int           // Global fallback: 100
    GlobalCircuitBreakerTimeout   time.Duration // Global fallback: 120s
}
```

### Step 3: Update NewQueueManager

Initialize per-queue-type breakers:

```go
func NewQueueManager(...) (*QueueManager, error) {
    // Use config values instead of hardcoded
    cbConfig := CircuitBreakerConfig{
        DefaultThreshold: cfg.CircuitBreakerThreshold,  // Default 50
        DefaultTimeout:   cfg.CircuitBreakerTimeout,    // Default 60s
        GlobalThreshold:  cfg.GlobalCircuitBreakerThreshold, // Default 100
        GlobalTimeout:    cfg.GlobalCircuitBreakerTimeout,   // Default 120s
    }

    cbManager := NewCircuitBreakerManager(cbConfig)

    qm := &QueueManager{
        // ...
        circuitBreakerManager: cbManager,
    }
    return qm, nil
}
```

### Step 4: Update failure/success recording

Track failures per queue type:

```go
func (qm *QueueManager) recordFailure(queueType rabbitmq.QueueType) {
    qm.circuitBreakerManager.RecordFailure(queueType)
}

func (qm *QueueManager) recordSuccess(queueType rabbitmq.QueueType) {
    qm.circuitBreakerManager.RecordSuccess(queueType)
}
```

### Step 5: Update circuit breaker checks

Check appropriate breaker based on context:

```go
func (qm *QueueManager) QueueEvent(ctx context.Context, event TriggerEvent) error {
    queueType := qm.determineQueueType(event)

    // Check queue-specific breaker first, then global
    if qm.circuitBreakerManager.IsOpen(queueType) {
        return fmt.Errorf("circuit breaker is open for queue type %s", queueType)
    }
    // ... rest of method
}
```

### Step 6: Update QueueStatus for observability

Expose per-queue breaker states:

```go
type QueueStatus struct {
    // ... existing fields ...
    CircuitBreakerOn     bool                       // Keep for backwards compat (any breaker open)
    CircuitBreakerStatus map[string]BreakerStatus   // New: per-queue status
}

type BreakerStatus struct {
    State        string // "closed", "open", "half-open"
    FailureCount int32
    LastFailure  *time.Time
}
```

### Step 7: Update DefaultQueueConfig

Set sensible defaults:

```go
func DefaultQueueConfig() QueueConfig {
    return QueueConfig{
        // ... existing ...
        CircuitBreakerThreshold:       50,              // Was 5
        CircuitBreakerTimeout:         60 * time.Second, // Was 30s
        GlobalCircuitBreakerThreshold: 100,
        GlobalCircuitBreakerTimeout:   120 * time.Second,
    }
}
```

## Call Sites to Update

Update all `recordFailure()` and `recordSuccess()` calls to pass queue type:

| Location | Line | Queue Type Source |
|----------|------|-------------------|
| `QueueEvent` | 189 | `qm.determineQueueType(event)` |
| `processAsyncAction` | 330, 337, 352, 362, 382 | Determine from action type (email→email breaker, etc.) |
| `processWorkflowEvent` | 486, 502 | `qm.determineQueueType(event)` from reconstructed event |

### Async Action Queue Type Mapping

For `processAsyncAction`, map the `request_type` from the payload to appropriate queue.

**Registered action types** (from `workflowactions/` handlers):
- `send_email` - SendEmailHandler
- `send_notification` - SendNotificationHandler
- `create_alert` - CreateAlertHandler
- `seek_approval` - SeekApprovalHandler
- `update_field` - UpdateFieldHandler
- `allocate_inventory` - AllocateInventoryHandler

```go
func (qm *QueueManager) getQueueTypeForAction(requestType string) rabbitmq.QueueType {
    switch requestType {
    case "send_email":
        return rabbitmq.QueueTypeEmail
    case "send_notification":
        return rabbitmq.QueueTypeNotification
    case "create_alert":
        return rabbitmq.QueueTypeAlert
    case "allocate_inventory":
        return rabbitmq.QueueTypeInventory
    case "seek_approval":
        return rabbitmq.QueueTypeApproval
    case "update_field":
        // Data operations go to general workflow queue
        return rabbitmq.QueueTypeWorkflow
    default:
        return rabbitmq.QueueTypeWorkflow
    }
}
```

## Testing

**File**: `business/sdk/workflow/queue_test.go`

### Test 1: Update `TestQueueManager_CircuitBreaker` for Per-Queue Isolation

Modify the existing test to verify that email failures only open the email circuit breaker.

```go
func TestQueueManager_CircuitBreaker(t *testing.T) {
    // ... existing setup through line 757 ...

    // Queue events to trigger circuit breaker (need at least 50 failures now)
    eventsQueued := 0
    for i := 0; i < 55; i++ {
        event := workflow.TriggerEvent{
            EventType:  "on_create",
            EntityName: "customers",  // Routes to QueueTypeWorkflow
            EntityID:   uuid.New(),
            Timestamp:  time.Now(),
            RawData: map[string]interface{}{
                "name":  fmt.Sprintf("Customer %d", i),
                "email": fmt.Sprintf("customer%d@example.com", i),
            },
            UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
        }

        if err := qm.QueueEvent(ctx, event); err != nil {
            if strings.Contains(err.Error(), "circuit breaker") {
                t.Logf("Circuit breaker opened after %d events", i)
                break
            }
            t.Logf("Failed to queue event %d: %v", i, err)
        } else {
            eventsQueued++
        }
    }

    // Wait for processing - now checking for 50 failures (new threshold)
    maxWait := 60 * time.Second  // Increase timeout for more events
    // ... rest of polling loop, change threshold check from 5 to 50 ...

    if failuresSinceStart >= 50 {
        status, err := qm.GetQueueStatus(ctx)
        if err != nil {
            t.Fatalf("GetQueueStatus() error = %v", err)
        }

        // Verify per-queue status is exposed
        if status.CircuitBreakerStatus == nil {
            t.Error("Expected CircuitBreakerStatus map to be populated")
        }

        // Email breaker should be open (since send_email actions are failing)
        if emailStatus, ok := status.CircuitBreakerStatus["email"]; ok {
            if emailStatus.State != "open" {
                t.Errorf("Expected email circuit breaker to be open, got %s", emailStatus.State)
            }
        }

        if status.CircuitBreakerOn {
            circuitBreakerOpened = true
            t.Log("✓ Circuit breaker has opened after 50+ failures")
        }
    }

    // ... rest of test ...
}
```

### Test 2: `TestQueueManager_CircuitBreaker_QueueIsolation`

New test verifying email failures don't block other queue types.

```go
func TestQueueManager_CircuitBreaker_QueueIsolation(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
        return otel.GetTraceID(context.Background())
    })

    // Get RabbitMQ container
    container := rabbitmq.GetTestContainer(t)
    client := rabbitmq.NewTestClient(container.URL)
    if err := client.Connect(); err != nil {
        t.Fatalf("connecting to rabbitmq: %s", err)
    }
    defer client.Close()

    // Initialize workflow queue
    queue := rabbitmq.NewTestWorkflowQueue(client, log)
    if err := queue.Initialize(context.Background()); err != nil {
        t.Fatalf("initializing workflow queue: %s", err)
    }

    // Setup database
    db := dbtest.NewDatabase(t, "Test_CircuitBreaker_Isolation")
    ctx := context.Background()

    // Create workflow business layer
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

    // Seed basic data
    _, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
    if err != nil {
        t.Fatalf("seeding workflow: %s", err)
    }

    // Get entities
    entity, err := workflowBus.QueryEntityByName(ctx, "customers")
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

    // === Create TWO rules: one for email (will fail), one for notification (will succeed) ===

    // Rule 1: Email rule that will fail
    emailRule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
        Name:              "Email Failure Rule",
        Description:       "Rule with email action that simulates failure",
        EntityID:          entity.ID,
        EntityTypeID:      entityType.ID,
        TriggerTypeID:     triggerType.ID,
        TriggerConditions: json.RawMessage(`{"field": "type", "operator": "equals", "value": "email_test"}`),
        IsActive:          true,
        CreatedBy:         uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
    })
    if err != nil {
        t.Fatalf("creating email rule: %s", err)
    }

    // Create email template
    emailTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
        Name:        "Failing Email Template",
        Description: "Template for failing emails",
        ActionType:  "send_email",
        DefaultConfig: json.RawMessage(`{
            "recipients": ["default@example.com"],
            "subject": "Default Subject"
        }`),
        CreatedBy: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
    })
    if err != nil {
        t.Fatalf("creating email template: %s", err)
    }

    // Create failing email action
    _, err = workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
        AutomationRuleID: emailRule.ID,
        Name:             "Failing Email Action",
        Description:      "Email action that simulates SMTP failure",
        ActionConfig: json.RawMessage(`{
            "recipients": ["admin@example.com"],
            "subject": "Isolation Test",
            "body": "This email will fail",
            "simulate_failure": true,
            "failure_message": "SMTP connection refused"
        }`),
        ExecutionOrder: 1,
        IsActive:       true,
        TemplateID:     &emailTemplate.ID,
    })
    if err != nil {
        t.Fatalf("creating email action: %s", err)
    }

    // Rule 2: Notification rule that will succeed (routes to notification queue)
    notificationRule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
        Name:              "Notification Success Rule",
        Description:       "Rule with notification action that succeeds",
        EntityID:          entity.ID,
        EntityTypeID:      entityType.ID,
        TriggerTypeID:     triggerType.ID,
        TriggerConditions: json.RawMessage(`{"field": "type", "operator": "equals", "value": "notification_test"}`),
        IsActive:          true,
        CreatedBy:         uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
    })
    if err != nil {
        t.Fatalf("creating notification rule: %s", err)
    }

    // Create notification template
    notificationTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
        Name:        "Success Notification Template",
        Description: "Template for notifications",
        ActionType:  "send_notification",
        DefaultConfig: json.RawMessage(`{
            "title": "Default Title",
            "message": "Default Message"
        }`),
        CreatedBy: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
    })
    if err != nil {
        t.Fatalf("creating notification template: %s", err)
    }

    // Create succeeding notification action
    _, err = workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
        AutomationRuleID: notificationRule.ID,
        Name:             "Success Notification Action",
        Description:      "Notification action that succeeds",
        ActionConfig: json.RawMessage(`{
            "title": "Test Notification",
            "message": "This should succeed even when email breaker is open",
            "priority": "normal"
        }`),
        ExecutionOrder: 1,
        IsActive:       true,
        TemplateID:     &notificationTemplate.ID,
    })
    if err != nil {
        t.Fatalf("creating notification action: %s", err)
    }

    // Create engine and queue manager
    workflow.ResetEngineForTesting()
    engine := workflow.NewEngine(log, db.DB, workflowBus)
    if err := engine.Initialize(ctx, workflowBus); err != nil {
        t.Fatalf("initializing engine: %s", err)
    }

    // Register handlers
    registry := engine.GetRegistry()
    registry.Register(communication.NewSendEmailHandler(log, db.DB))
    registry.Register(communication.NewSendNotificationHandler(log, db.DB))

    // Create queue manager with lower threshold for faster testing
    qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
    if err != nil {
        t.Fatalf("creating queue manager: %s", err)
    }

    // Override config for faster testing (use 10 instead of 50)
    qm.SetCircuitBreakerThreshold(rabbitmq.QueueTypeEmail, 10)

    t.Cleanup(func() {
        qm.ResetCircuitBreaker()
    })

    if err := qm.Initialize(ctx); err != nil {
        t.Fatalf("initializing queue manager: %s", err)
    }
    if err := qm.ClearQueue(ctx); err != nil {
        t.Logf("Warning: could not clear queue: %v", err)
    }
    qm.ResetCircuitBreaker()
    qm.ResetMetrics()

    if err := qm.Start(ctx); err != nil {
        t.Fatalf("starting queue manager: %s", err)
    }
    defer qm.Stop(ctx)

    time.Sleep(500 * time.Millisecond)

    // === Phase 1: Trip the EMAIL circuit breaker ===
    t.Log("Phase 1: Triggering email failures to open email circuit breaker")

    for i := 0; i < 15; i++ {
        event := workflow.TriggerEvent{
            EventType:  "on_create",
            EntityName: "customers",
            EntityID:   uuid.New(),
            Timestamp:  time.Now(),
            RawData: map[string]interface{}{
                "type":  "email_test",  // Matches email rule condition
                "name":  fmt.Sprintf("Email Customer %d", i),
                "email": fmt.Sprintf("email%d@example.com", i),
            },
            UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
        }
        qm.QueueEvent(ctx, event)
    }

    // Wait for email breaker to open
    time.Sleep(10 * time.Second)

    status, err := qm.GetQueueStatus(ctx)
    if err != nil {
        t.Fatalf("GetQueueStatus() error = %v", err)
    }

    // Verify email breaker is open
    emailStatus := status.CircuitBreakerStatus["email"]
    if emailStatus.State != "open" {
        t.Fatalf("Expected email circuit breaker to be open, got %s", emailStatus.State)
    }
    t.Log("✓ Email circuit breaker is open")

    // Verify notification breaker is still closed
    notificationStatus := status.CircuitBreakerStatus["notification"]
    if notificationStatus.State != "closed" {
        t.Errorf("Expected notification circuit breaker to remain closed, got %s", notificationStatus.State)
    }
    t.Log("✓ Notification circuit breaker is still closed")

    // === Phase 2: Verify notification events STILL PROCESS while email is blocked ===
    t.Log("Phase 2: Verifying notification events process while email breaker is open")

    initialProcessed := qm.GetMetrics().TotalProcessed

    // Queue notification events
    for i := 0; i < 5; i++ {
        event := workflow.TriggerEvent{
            EventType:  "on_create",
            EntityName: "customers",
            EntityID:   uuid.New(),
            Timestamp:  time.Now(),
            RawData: map[string]interface{}{
                "type":  "notification_test",  // Matches notification rule condition
                "name":  fmt.Sprintf("Notification Customer %d", i),
                "email": fmt.Sprintf("notify%d@example.com", i),
            },
            UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
        }

        err := qm.QueueEvent(ctx, event)
        if err != nil {
            // Should NOT fail - notification queue is independent of email queue
            t.Errorf("QueueEvent for notification should succeed, got error: %v", err)
        }
    }

    // Wait for processing
    time.Sleep(5 * time.Second)

    finalMetrics := qm.GetMetrics()
    notificationsProcessed := finalMetrics.TotalProcessed - initialProcessed

    if notificationsProcessed < 5 {
        t.Errorf("Expected at least 5 notifications processed while email breaker open, got %d", notificationsProcessed)
    } else {
        t.Logf("✓ Successfully processed %d notification events while email breaker was open", notificationsProcessed)
    }

    // === Phase 3: Verify email events are still blocked ===
    t.Log("Phase 3: Verifying email events are blocked")

    emailEvent := workflow.TriggerEvent{
        EventType:  "on_create",
        EntityName: "customers",
        EntityID:   uuid.New(),
        Timestamp:  time.Now(),
        RawData: map[string]interface{}{
            "type":  "email_test",
            "name":  "Blocked Email Customer",
            "email": "blocked@example.com",
        },
        UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
    }

    // This MAY succeed at queue time (circuit breaker is checked at processing)
    // But processing should be blocked
    qm.QueueEvent(ctx, emailEvent)

    // Re-check status
    status, _ = qm.GetQueueStatus(ctx)
    if status.CircuitBreakerStatus["email"].State != "open" {
        t.Error("Email circuit breaker should still be open")
    } else {
        t.Log("✓ Email circuit breaker correctly remains open")
    }

    t.Log("✓ Queue isolation test completed successfully")
}
```

### Test 3: `TestQueueManager_CircuitBreaker_GlobalFallback`

Test that the global breaker trips when multiple queues accumulate failures.

```go
func TestQueueManager_CircuitBreaker_GlobalFallback(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
        return otel.GetTraceID(context.Background())
    })

    // ... standard RabbitMQ and DB setup ...
    container := rabbitmq.GetTestContainer(t)
    client := rabbitmq.NewTestClient(container.URL)
    if err := client.Connect(); err != nil {
        t.Fatalf("connecting to rabbitmq: %s", err)
    }
    defer client.Close()

    queue := rabbitmq.NewTestWorkflowQueue(client, log)
    if err := queue.Initialize(context.Background()); err != nil {
        t.Fatalf("initializing workflow queue: %s", err)
    }

    db := dbtest.NewDatabase(t, "Test_CircuitBreaker_Global")
    ctx := context.Background()
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

    _, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
    if err != nil {
        t.Fatalf("seeding workflow: %s", err)
    }

    workflow.ResetEngineForTesting()
    engine := workflow.NewEngine(log, db.DB, workflowBus)
    if err := engine.Initialize(ctx, workflowBus); err != nil {
        t.Fatalf("initializing engine: %s", err)
    }

    // Create queue manager with low thresholds for testing
    qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
    if err != nil {
        t.Fatalf("creating queue manager: %s", err)
    }

    // Set low thresholds for faster testing
    // Per-queue: 10, Global: 25 (so we need failures across multiple queues)
    qm.SetCircuitBreakerThreshold(rabbitmq.QueueTypeEmail, 10)
    qm.SetCircuitBreakerThreshold(rabbitmq.QueueTypeNotification, 10)
    qm.SetCircuitBreakerThreshold(rabbitmq.QueueTypeWorkflow, 10)
    qm.SetGlobalCircuitBreakerThreshold(25)

    t.Cleanup(func() {
        qm.ResetCircuitBreaker()
    })

    if err := qm.Initialize(ctx); err != nil {
        t.Fatalf("initializing queue manager: %s", err)
    }
    qm.ResetCircuitBreaker()
    qm.ResetMetrics()

    if err := qm.Start(ctx); err != nil {
        t.Fatalf("starting queue manager: %s", err)
    }
    defer qm.Stop(ctx)

    time.Sleep(500 * time.Millisecond)

    // Simulate failures across multiple queue types by directly calling recordFailure
    // (In real scenario, this would come from actual processing failures)

    // Add 8 failures each to email, notification, and workflow queues
    // Total: 24 failures (just under global threshold of 25)
    for i := 0; i < 8; i++ {
        qm.RecordFailureForTesting(rabbitmq.QueueTypeEmail)
        qm.RecordFailureForTesting(rabbitmq.QueueTypeNotification)
        qm.RecordFailureForTesting(rabbitmq.QueueTypeWorkflow)
    }

    status, _ := qm.GetQueueStatus(ctx)

    // Individual breakers should NOT be open (only 8 failures each, threshold is 10)
    if status.CircuitBreakerStatus["email"].State == "open" {
        t.Error("Email breaker should not be open (only 8 failures)")
    }
    if status.CircuitBreakerStatus["notification"].State == "open" {
        t.Error("Notification breaker should not be open (only 8 failures)")
    }
    if status.CircuitBreakerStatus["workflow"].State == "open" {
        t.Error("Workflow breaker should not be open (only 8 failures)")
    }

    // Global breaker should NOT be open yet (24 failures, threshold is 25)
    if status.CircuitBreakerStatus["global"].State == "open" {
        t.Error("Global breaker should not be open yet (24 failures)")
    }
    t.Log("✓ After 24 total failures: Individual and global breakers still closed")

    // Add 2 more failures to trigger global threshold
    qm.RecordFailureForTesting(rabbitmq.QueueTypeEmail)
    qm.RecordFailureForTesting(rabbitmq.QueueTypeNotification)

    status, _ = qm.GetQueueStatus(ctx)

    // Now global breaker should be open (26 total failures >= 25 threshold)
    if status.CircuitBreakerStatus["global"].State != "open" {
        t.Errorf("Global breaker should be open after 26 failures, got state: %s",
            status.CircuitBreakerStatus["global"].State)
    }
    t.Log("✓ Global breaker opened after 26 total failures")

    // CircuitBreakerOn should be true (for backwards compatibility)
    if !status.CircuitBreakerOn {
        t.Error("CircuitBreakerOn should be true when global breaker is open")
    }
    t.Log("✓ CircuitBreakerOn correctly reports true")

    // All queue operations should now be blocked
    event := workflow.TriggerEvent{
        EventType:  "on_create",
        EntityName: "customers",
        EntityID:   uuid.New(),
        Timestamp:  time.Now(),
        RawData:    map[string]interface{}{"name": "Test"},
        UserID:     uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
    }

    err = qm.QueueEvent(ctx, event)
    if err == nil {
        t.Error("QueueEvent should fail when global breaker is open")
    } else if !strings.Contains(err.Error(), "circuit breaker") {
        t.Errorf("Expected circuit breaker error, got: %v", err)
    } else {
        t.Logf("✓ QueueEvent correctly blocked with global breaker open: %v", err)
    }

    t.Log("✓ Global fallback circuit breaker test completed successfully")
}
```

### Test 4: `TestQueueManager_CircuitBreaker_ConfigOverrides`

Test that per-queue configuration overrides work.

```go
func TestQueueManager_CircuitBreaker_ConfigOverrides(t *testing.T) {
    log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
        return otel.GetTraceID(context.Background())
    })

    // ... standard setup ...
    container := rabbitmq.GetTestContainer(t)
    client := rabbitmq.NewTestClient(container.URL)
    if err := client.Connect(); err != nil {
        t.Fatalf("connecting to rabbitmq: %s", err)
    }
    defer client.Close()

    queue := rabbitmq.NewTestWorkflowQueue(client, log)
    if err := queue.Initialize(context.Background()); err != nil {
        t.Fatalf("initializing workflow queue: %s", err)
    }

    db := dbtest.NewDatabase(t, "Test_CircuitBreaker_Config")
    ctx := context.Background()
    workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

    _, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
    if err != nil {
        t.Fatalf("seeding workflow: %s", err)
    }

    workflow.ResetEngineForTesting()
    engine := workflow.NewEngine(log, db.DB, workflowBus)
    if err := engine.Initialize(ctx, workflowBus); err != nil {
        t.Fatalf("initializing engine: %s", err)
    }

    // Create queue manager
    qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
    if err != nil {
        t.Fatalf("creating queue manager: %s", err)
    }

    // Test: Set different thresholds for different queues
    qm.SetCircuitBreakerThreshold(rabbitmq.QueueTypeEmail, 5)      // Low threshold
    qm.SetCircuitBreakerThreshold(rabbitmq.QueueTypeWorkflow, 100) // High threshold
    qm.SetCircuitBreakerTimeout(rabbitmq.QueueTypeEmail, 10*time.Second)
    qm.SetCircuitBreakerTimeout(rabbitmq.QueueTypeWorkflow, 120*time.Second)

    t.Cleanup(func() {
        qm.ResetCircuitBreaker()
    })

    qm.ResetCircuitBreaker()

    // Add 5 failures to email - should trip
    for i := 0; i < 5; i++ {
        qm.RecordFailureForTesting(rabbitmq.QueueTypeEmail)
    }

    // Add 5 failures to workflow - should NOT trip (threshold is 100)
    for i := 0; i < 5; i++ {
        qm.RecordFailureForTesting(rabbitmq.QueueTypeWorkflow)
    }

    status, _ := qm.GetQueueStatus(ctx)

    // Email should be open (5 failures == 5 threshold)
    emailStatus := status.CircuitBreakerStatus["email"]
    if emailStatus.State != "open" {
        t.Errorf("Email breaker should be open with threshold 5, got state: %s", emailStatus.State)
    }
    t.Log("✓ Email breaker opened at threshold 5")

    // Workflow should still be closed (5 failures < 100 threshold)
    workflowStatus := status.CircuitBreakerStatus["workflow"]
    if workflowStatus.State != "closed" {
        t.Errorf("Workflow breaker should be closed (5 < 100), got state: %s", workflowStatus.State)
    }
    t.Log("✓ Workflow breaker remains closed (threshold 100)")

    // Verify failure counts are tracked correctly
    if emailStatus.FailureCount != 5 {
        t.Errorf("Expected email failure count 5, got %d", emailStatus.FailureCount)
    }
    if workflowStatus.FailureCount != 5 {
        t.Errorf("Expected workflow failure count 5, got %d", workflowStatus.FailureCount)
    }
    t.Log("✓ Failure counts tracked correctly per queue")

    t.Log("✓ Configuration overrides test completed successfully")
}
```

### Test Helper Methods to Add

Add these methods to `QueueManager` for testing purposes:

```go
// RecordFailureForTesting allows tests to simulate failures for a specific queue type.
// This method should only be used in tests.
func (qm *QueueManager) RecordFailureForTesting(queueType rabbitmq.QueueType) {
    qm.recordFailure(queueType)
}

// SetCircuitBreakerThreshold allows tests to set custom thresholds per queue.
func (qm *QueueManager) SetCircuitBreakerThreshold(queueType rabbitmq.QueueType, threshold int) {
    qm.circuitBreakerManager.SetThreshold(queueType, threshold)
}

// SetCircuitBreakerTimeout allows tests to set custom timeouts per queue.
func (qm *QueueManager) SetCircuitBreakerTimeout(queueType rabbitmq.QueueType, timeout time.Duration) {
    qm.circuitBreakerManager.SetTimeout(queueType, timeout)
}

// SetGlobalCircuitBreakerThreshold allows tests to set the global breaker threshold.
func (qm *QueueManager) SetGlobalCircuitBreakerThreshold(threshold int) {
    qm.circuitBreakerManager.SetGlobalThreshold(threshold)
}
```

## Backwards Compatibility

- `CircuitBreakerOn` field in `QueueStatus` remains (true if ANY breaker is open)
- `ResetCircuitBreaker()` resets ALL breakers (for testing)
- Default behavior with no config changes: More tolerant than before (50 vs 5), which is strictly better
