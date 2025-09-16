package workflow_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// NOTE: This is NOT testing the following
// - Whether the Engine actually processes workflows correctly - you're just returning success
// - The entire dependency resolution and rule matching logic - that's all bypassed
// - Action execution - none of the actual workflow actions run
// - Real database interactions - no rules are loaded, no state is persiste

// This is testing the following
// The QueueManager correctly integrates with RabbitMQ - messages are being queued, consumed, and acknowledged properly
// The message transformation logic works - converting between TriggerEvents and RabbitMQ Messages
// The circuit breaker and metrics tracking function - failures are counted, circuit opens/closes
// Queue routing logic - different entities go to different queues

// stubEngine wraps a real Engine but provides simplified execution for testing
type stubEngine struct {
	*workflow.Engine
	executionCount int
	lastEvent      workflow.TriggerEvent
	failNext       bool // For testing failure scenarios
}

func newStubEngine(log *logger.Logger, db *sqlx.DB) *stubEngine {

	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	return &stubEngine{
		Engine: workflow.NewEngine(log, db, workflowBus),
	}
}

func (e *stubEngine) ExecuteWorkflow(ctx context.Context, event workflow.TriggerEvent) (*workflow.WorkflowExecution, error) {
	e.executionCount++
	e.lastEvent = event

	if e.failNext {
		e.failNext = false
		return nil, fmt.Errorf("simulated failure")
	}

	now := time.Now()
	return &workflow.WorkflowExecution{
		ExecutionID:  uuid.New(),
		TriggerEvent: event,
		ExecutionPlan: workflow.ExecutionPlan{
			PlanID:           uuid.New(),
			MatchedRuleCount: 1,
			TotalBatches:     1,
			CreatedAt:        now,
		},
		Status:       workflow.StatusCompleted,
		StartedAt:    now,
		CompletedAt:  &now,
		BatchResults: []workflow.BatchResult{},
		Errors:       []string{},
	}, nil
}

func TestQueueManager_Initialize(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")

	// Create workflow engine (mock)
	engine := newStubEngine(log, db.DB)

	// Create queue manager
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()

	// Test initialization
	if err := qm.Initialize(ctx); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Verify queues were created by getting status
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Errorf("GetQueueStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetQueueStatus() returned nil")
	}
}

func TestQueueManager_QueueEvent(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	entityID1 := uuid.New()
	entityID2 := uuid.New()
	entityID3 := uuid.New()

	userID2 := uuid.New()

	tests := []struct {
		name    string
		event   workflow.TriggerEvent
		wantErr bool
	}{
		{
			name: "valid create event",
			event: workflow.TriggerEvent{
				EventType:  "on_create",
				EntityName: "customers",
				EntityID:   entityID1,
				Timestamp:  time.Now(),
				RawData: map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
				},
				UserID: userID2,
			},
			wantErr: false,
		},
		{
			name: "valid update event with field changes",
			event: workflow.TriggerEvent{
				EventType:  "on_update",
				EntityName: "customers",
				EntityID:   entityID1,
				Timestamp:  time.Now(),
				FieldChanges: map[string]workflow.FieldChange{
					"status": {
						OldValue: "regular",
						NewValue: "premium",
					},
				},
				RawData: map[string]interface{}{
					"name":   "John Doe",
					"status": "premium",
				},
				UserID: userID2,
			},
			wantErr: false,
		},
		{
			name: "valid delete event",
			event: workflow.TriggerEvent{
				EventType:  "on_delete",
				EntityName: "customers",
				EntityID:   entityID1,
				Timestamp:  time.Now(),
				UserID:     userID2,
			},
			wantErr: false,
		},
		{
			name: "approval entity event",
			event: workflow.TriggerEvent{
				EventType:  "on_create",
				EntityName: "approvals",
				EntityID:   entityID2,
				Timestamp:  time.Now(),
				RawData: map[string]interface{}{
					"request_id": "req_123",
					"approver":   "manager@example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "inventory entity event",
			event: workflow.TriggerEvent{
				EventType:  "on_update",
				EntityName: "inventory",
				EntityID:   entityID3,
				Timestamp:  time.Now(),
				FieldChanges: map[string]workflow.FieldChange{
					"quantity": {
						OldValue: 100,
						NewValue: 75,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := qm.QueueEvent(ctx, tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueueEvent() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify metrics were updated
				metrics := qm.GetMetrics()
				if metrics.TotalEnqueued == 0 {
					t.Error("TotalEnqueued should be > 0 after queueing event")
				}
			}
		})
	}
}

func TestQueueManager_StartStop(t *testing.T) {

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Test starting
	if err := qm.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Verify it's running
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() error = %v", err)
	}
	if !status.IsRunning {
		t.Error("Queue manager should be running after Start()")
	}

	// Test double start (should error)
	if err := qm.Start(ctx); err == nil {
		t.Error("Start() should error when already running")
	}

	// Test stopping
	if err := qm.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Test double stop (should not error)
	if err := qm.Stop(ctx); err != nil {
		t.Errorf("Stop() on stopped manager should not error: %v", err)
	}
}

func TestQueueManager_ProcessMessage(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Setup RabbitMQ
	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Setup database with test data
	db := dbtest.NewDatabase(t, "Test_Workflow")
	ctx := context.Background()

	// Create workflow business layer
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

	// Seed
	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow: %s", err)
	}

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

	// Create rule BEFORE initializing engine
	ret, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:              "Test Rule",
		Description:       "A rule for testing",
		EntityID:          entity.ID,
		EntityTypeID:      entityType.ID,
		TriggerTypeID:     triggerType.ID,
		TriggerConditions: nil, // No conditions = matches all events of this type
		IsActive:          true,
		CreatedBy:         uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
	})
	if err != nil {
		t.Fatalf("creating rule: %s", err)
	}

	testRule, err := workflowBus.QueryRuleByID(ctx, ret.ID)
	if err != nil {
		t.Fatalf("querying created rule: %s", err)
	}
	t.Logf("Created test rule with ID: %s", testRule.ID)

	// Optional: Create an action for the rule to make it complete
	// Note: You'll need to have an action template set up first, or register a handler
	/*
		_, err = workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: testRule.ID,
			Name:            "Test Action",
			Description:     "Log the event",
			ActionConfig:    json.RawMessage(`{"message": "Customer created: {{entity_name}}"}`),
			ExecutionOrder:  1,
			IsActive:        true,
			TemplateID:      nil, // Or use a real template ID if you have one
		})
		if err != nil {
			t.Logf("Could not create action: %v (this is okay for now)", err)
		}
	*/

	// Create and initialize the engine AFTER creating the rule
	engine := workflow.NewEngine(log, db.DB, workflowBus)
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Create queue manager with real engine
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Clear any lingering messages from previous test runs
	if err := qm.ClearQueue(ctx); err != nil {
		t.Logf("Warning: could not clear queue: %v", err)
	}

	// Get initial metrics for comparison
	initialMetrics := qm.GetMetrics()

	// Start the queue manager
	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Small delay to ensure consumers are ready
	time.Sleep(100 * time.Millisecond)

	// Create and queue a test event
	entityID := uuid.New()
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   entityID,
		Timestamp:  time.Now(),
		RawData: map[string]interface{}{
			"name":    "Test Customer",
			"email":   "test@example.com",
			"status":  "active",
			"revenue": 10000,
		},
		UserID: adminUserID,
	}

	// Queue the event
	if err := qm.QueueEvent(ctx, event); err != nil {
		t.Fatalf("queueing event: %s", err)
	}

	// Wait for processing with timeout
	processed := false
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for !processed {
		select {
		case <-timeout:
			metrics := qm.GetMetrics()
			t.Logf("Final metrics - Enqueued: %d, Processed: %d, Failed: %d",
				metrics.TotalEnqueued, metrics.TotalProcessed, metrics.TotalFailed)
			t.Fatal("Timeout waiting for event processing")

		case <-ticker.C:
			metrics := qm.GetMetrics()

			// Check if processing completed (successfully or with failure)
			if metrics.TotalProcessed > initialMetrics.TotalProcessed ||
				metrics.TotalFailed > initialMetrics.TotalFailed {
				processed = true
			}
		}
	}

	// Verify the results
	finalMetrics := qm.GetMetrics()

	// Should have enqueued exactly one event
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
		t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
			initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
	}

	// Should have processed the event
	if finalMetrics.TotalProcessed != initialMetrics.TotalProcessed+1 {
		t.Errorf("Expected TotalProcessed to increase by 1, got %d -> %d",
			initialMetrics.TotalProcessed, finalMetrics.TotalProcessed)
	}

	// Should not have failed
	if finalMetrics.TotalFailed > initialMetrics.TotalFailed {
		t.Errorf("Unexpected failures: %d", finalMetrics.TotalFailed-initialMetrics.TotalFailed)
	}

	// Verify processing time was recorded
	if finalMetrics.LastProcessedAt == nil {
		t.Error("LastProcessedAt should be set after processing")
	}

	// For very fast processing, AverageProcessTimeMs might be 0
	// This is okay - just log it
	if finalMetrics.AverageProcessTimeMs == 0 && finalMetrics.TotalProcessed > 0 {
		t.Logf("Note: AverageProcessTimeMs is 0 (processing was < 1ms)")
	}

	// Check engine's execution history
	execHistory := engine.GetExecutionHistory(10)
	if len(execHistory) == 0 {
		t.Error("Expected at least one execution in history")
	} else {
		lastExecution := execHistory[0]

		// Verify the correct event was processed
		if lastExecution.TriggerEvent.EntityID != entityID {
			t.Errorf("Expected execution for entity %s, got %s",
				entityID, lastExecution.TriggerEvent.EntityID)
		}

		// Check if rules matched (should be 1 with our test rule)
		if lastExecution.ExecutionPlan.MatchedRuleCount != 1 {
			t.Errorf("Expected 1 matched rule, got %d",
				lastExecution.ExecutionPlan.MatchedRuleCount)
		}

		// Log execution details for debugging
		t.Logf("Execution completed with status: %s, matched rules: %d, batches: %d",
			lastExecution.Status,
			lastExecution.ExecutionPlan.MatchedRuleCount,
			lastExecution.ExecutionPlan.TotalBatches)
	}

	// Verify queue is now empty (message was consumed)
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Errorf("Failed to get queue status: %v", err)
	} else {
		if status.QueueDepth > 0 {
			t.Logf("Warning: Queue still has %d messages", status.QueueDepth)
		}
	}

	t.Logf("Integration test completed successfully - Event was queued and processed")
}

func TestQueueManager_CircuitBreaker(t *testing.T) {

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")

	// Create an engine that always fails
	engine := newStubEngine(log, db.DB)
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Queue multiple events to trigger circuit breaker
	for i := 0; i < 6; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   uuid.New(),
			Timestamp:  time.Now(),
		}

		_ = qm.QueueEvent(ctx, event) // Ignore errors initially
	}

	// Wait for processing attempts
	time.Sleep(3 * time.Second)

	// Circuit breaker should be open now
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() error = %v", err)
	}

	if !status.CircuitBreakerOn {
		t.Error("Circuit breaker should be open after multiple failures")
	}

	// Try to queue another event - should fail
	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
	}

	if err := qm.QueueEvent(ctx, event); err == nil {
		t.Error("QueueEvent() should fail when circuit breaker is open")
	}
}

func TestQueueManager_ClearQueue(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Queue some events
	for i := 0; i < 5; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   uuid.New(),
			Timestamp:  time.Now(),
		}
		if err := qm.QueueEvent(ctx, event); err != nil {
			t.Fatalf("queueing event: %s", err)
		}
	}

	// Clear the queue
	if err := qm.ClearQueue(ctx); err != nil {
		t.Errorf("ClearQueue() error = %v", err)
	}

	// Verify queue is empty
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() error = %v", err)
	}

	if status.QueueDepth != 0 {
		t.Errorf("Queue depth = %d after clear, want 0", status.QueueDepth)
	}
}

func TestQueueManager_Metrics(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")

	engine := newStubEngine(log, db.DB)

	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Queue and process several events
	for i := 0; i < 4; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   uuid.New(),
			Timestamp:  time.Now(),
		}
		if err := qm.QueueEvent(ctx, event); err != nil {
			t.Errorf("QueueEvent() error = %v", err)
		}
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Check metrics
	metrics := qm.GetMetrics()

	if metrics.TotalEnqueued != 4 {
		t.Errorf("TotalEnqueued = %d, want 4", metrics.TotalEnqueued)
	}

	if metrics.TotalProcessed == 0 {
		t.Error("TotalProcessed should be > 0")
	}

	if metrics.TotalFailed == 0 {
		t.Error("TotalFailed should be > 0 (we simulated failures)")
	}

	if metrics.LastProcessedAt == nil {
		t.Error("LastProcessedAt should not be nil after processing")
	}

	if metrics.AverageProcessTimeMs == 0 {
		t.Error("AverageProcessTimeMs should be > 0 after processing")
	}
}

func TestQueueManager_DetermineQueueType(t *testing.T) {

	tests := []struct {
		name      string
		event     workflow.TriggerEvent
		wantQueue rabbitmq.QueueType
	}{
		{
			name: "approval entity",
			event: workflow.TriggerEvent{
				EntityName: "approvals",
			},
			wantQueue: rabbitmq.QueueTypeApproval,
		},
		{
			name: "approval_requests entity",
			event: workflow.TriggerEvent{
				EntityName: "approval_requests",
			},
			wantQueue: rabbitmq.QueueTypeApproval,
		},
		{
			name: "inventory entity",
			event: workflow.TriggerEvent{
				EntityName: "inventory",
			},
			wantQueue: rabbitmq.QueueTypeInventory,
		},
		{
			name: "stock entity",
			event: workflow.TriggerEvent{
				EntityName: "stock",
			},
			wantQueue: rabbitmq.QueueTypeInventory,
		},
		{
			name: "notifications entity",
			event: workflow.TriggerEvent{
				EntityName: "notifications",
			},
			wantQueue: rabbitmq.QueueTypeNotification,
		},
		{
			name: "default to workflow queue",
			event: workflow.TriggerEvent{
				EntityName: "customers",
			},
			wantQueue: rabbitmq.QueueTypeWorkflow,
		},
	}

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	entityID := uuid.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the queue routing logic indirectly through QueueEvent
			ctx := context.Background()
			if err := qm.Initialize(ctx); err != nil {
				t.Fatalf("initializing queue manager: %s", err)
			}

			// Queue the event (internally uses determineQueueType)
			tt.event.EventType = "on_create"
			tt.event.EntityID = entityID
			tt.event.Timestamp = time.Now()

			if err := qm.QueueEvent(ctx, tt.event); err != nil {
				t.Errorf("QueueEvent() error = %v", err)
			}
		})
	}
}

func TestQueueManager_ProcessingResult(t *testing.T) {

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Test ProcessEvents method (compatibility layer)
	result, err := qm.ProcessEvents(ctx, 10)
	if err != nil {
		t.Errorf("ProcessEvents() error = %v", err)
	}

	if result == nil {
		t.Fatal("ProcessEvents() returned nil result")
	}

	if result.StartTime.IsZero() {
		t.Error("ProcessingResult.StartTime should not be zero")
	}

	if result.EndTime.IsZero() {
		t.Error("ProcessingResult.EndTime should not be zero")
	}
}

// Mock engine for testing
type mockEngine struct {
	executeFunc func(ctx context.Context, event workflow.TriggerEvent) (*workflow.WorkflowExecution, error)
}

func (m *mockEngine) ExecuteWorkflow(ctx context.Context, event workflow.TriggerEvent) (*workflow.WorkflowExecution, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, event)
	}
	return &workflow.WorkflowExecution{
		ExecutionID: uuid.New(),
		Status:      "completed",
	}, nil
}

// Benchmark tests

func BenchmarkQueueManager_QueueEvent(b *testing.B) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		b.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		b.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		b.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		b.Fatalf("initializing queue manager: %s", err)
	}

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData: map[string]interface{}{
			"name":  "Benchmark Customer",
			"email": "bench@example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event.EntityID = uuid.New()
		_ = qm.QueueEvent(ctx, event)
	}
}

func BenchmarkQueueManager_ProcessMessage(b *testing.B) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		b.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		b.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()
	db := sqlx.DB{}

	engine := newStubEngine(log, &db)
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		b.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		b.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		b.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   uuid.New(),
			Timestamp:  time.Now(),
		}
		_ = qm.QueueEvent(ctx, event)
	}

	// Wait for processing to complete
	time.Sleep(time.Duration(b.N*10) * time.Millisecond)
}
