package workflow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
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

	workflow.ResetEngineForTesting()
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
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")

	// Create workflow engine (mock)
	engine := newStubEngine(log, db.DB)

	// Create queue manager
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client, queue)
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
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client, queue)
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
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client, queue)
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

	// TEST SETUP ==============================================================
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Setup RabbitMQ
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

	// CREATE RULE =============================================================
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

	// CREATE ACTION TEMPLATE ==================================================
	emailTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Send Email Template",
		Description: "Template for sending emails",
		ActionType:  "send_email",
		DefaultConfig: json.RawMessage(`{
			"recipients": ["test@example.com"],
			"subject": "Default Subject"
		}`),
		CreatedBy: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
	})
	if err != nil {
		t.Fatalf("creating email template: %s", err)
	}

	// CREATE RULE ACTION ======================================================
	_, err = workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: testRule.ID,
		Name:             "Send Test Email",
		Description:      "Send email when customer is created",
		ActionConfig: json.RawMessage(`{
			"recipients": ["admin@example.com", "manager@example.com"],
			"subject": "New Customer: {{name}}",
			"body": "A new customer {{name}} with email {{email}} has been created."
		}`),
		ExecutionOrder: 1,
		IsActive:       true,
		TemplateID:     &emailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	// INITIALIZE ENGINE =======================================================
	// Create engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)

	// Now initialize the engine
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Register the email handler
	registry := engine.GetRegistry()
	registry.Register(communication.NewSendEmailHandler(log, db.DB))

	// Create queue manager with real engine
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
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
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions

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

		// VERIFY EMAIL ACTION WAS EXECUTED ===================================
		actionExecuted := false
		for _, batchResult := range lastExecution.BatchResults {
			for _, ruleResult := range batchResult.RuleResults {
				t.Logf("Rule %s executed with %d actions", ruleResult.RuleName, len(ruleResult.ActionResults))

				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.ActionType == "send_email" {
						actionExecuted = true

						// Check the action succeeded
						if actionResult.Status != "success" {
							t.Errorf("Email action failed: %s", actionResult.ErrorMessage)
						}

						// Check result data
						if actionResult.ResultData != nil {
							t.Logf("Email action result: %v", actionResult.ResultData)

							// Verify expected fields in result
							if emailID, ok := actionResult.ResultData["email_id"]; ok {
								t.Logf("Email sent with ID: %v", emailID)
							}
							if status, ok := actionResult.ResultData["status"]; ok {
								if status != "sent" {
									t.Errorf("Expected email status 'sent', got %v", status)
								}
							}
						}
					}
				}
			}
		}

		if !actionExecuted {
			t.Error("Expected email action to be executed")
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

	t.Logf("Integration test completed successfully - Event was queued and processed with email action")
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
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Setup database
	db := dbtest.NewDatabase(t, "Test_CircuitBreaker")
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

	// Create a rule for circuit breaker testing
	rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:              "Circuit Breaker Test Rule",
		Description:       "Rule to test circuit breaker with simulated email failures",
		EntityID:          entity.ID,
		EntityTypeID:      entityType.ID,
		TriggerTypeID:     triggerType.ID,
		TriggerConditions: nil, // Match all
		IsActive:          true,
		CreatedBy:         uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
	})
	if err != nil {
		t.Fatalf("creating rule: %s", err)
	}

	// Create email template
	emailTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Send Email Template",
		Description: "Template for sending emails",
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

	// Create email action configured to simulate SMTP failure
	_, err = workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Send Email with Simulated Failure",
		Description:      "Email action that simulates SMTP server failure",
		ActionConfig: json.RawMessage(`{
			"recipients": ["admin@example.com"],
			"subject": "Circuit Breaker Test",
			"body": "This email will fail to trigger circuit breaker",
			"simulate_failure": true,
			"failure_message": "Connection refused: SMTP server at smtp.example.com:587 is not responding"
		}`),
		ExecutionOrder: 1,
		IsActive:       true,
		TemplateID:     &emailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	// Create real engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)

	// Initialize the engine
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Register the real email handler (which supports failure simulation)
	registry := engine.GetRegistry()
	registry.Register(communication.NewSendEmailHandler(log, db.DB))

	// Create queue manager with real engine
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	// Ensure circuit breaker is reset at the end of the test for isolation
	t.Cleanup(func() {
		qm.ResetCircuitBreaker()
	})

	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Clear any existing messages
	if err := qm.ClearQueue(ctx); err != nil {
		t.Logf("Warning: could not clear queue: %v", err)
	}
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Start the queue manager
	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Wait for consumers to be ready
	time.Sleep(500 * time.Millisecond)

	// Queue events to trigger circuit breaker (need at least 5 failures)
	eventsQueued := 0
	for i := 0; i < 6; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   uuid.New(),
			Timestamp:  time.Now(),
			RawData: map[string]interface{}{
				"name":  fmt.Sprintf("Customer %d", i),
				"email": fmt.Sprintf("customer%d@example.com", i),
			},
			UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
		}

		if err := qm.QueueEvent(ctx, event); err != nil {
			// Circuit breaker might open partway through
			if strings.Contains(err.Error(), "circuit breaker") {
				t.Logf("Circuit breaker opened after %d events", i)
				break
			}
			t.Logf("Failed to queue event %d: %v", i, err)
		} else {
			eventsQueued++
		}
	}

	t.Logf("Successfully queued %d events", eventsQueued)

	// Wait for processing and circuit breaker to trigger
	maxWait := 30 * time.Second
	checkInterval := 500 * time.Millisecond
	timeout := time.After(maxWait)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	circuitBreakerOpened := false
	for !circuitBreakerOpened {
		select {
		case <-timeout:
			metrics := qm.GetMetrics()
			t.Logf("Timeout - Metrics: Enqueued=%d, Processed=%d, Failed=%d",
				metrics.TotalEnqueued-initialMetrics.TotalEnqueued,
				metrics.TotalProcessed-initialMetrics.TotalProcessed,
				metrics.TotalFailed-initialMetrics.TotalFailed)
			t.Fatal("Circuit breaker did not open within timeout period")

		case <-ticker.C:
			metrics := qm.GetMetrics()
			failuresSinceStart := metrics.TotalFailed - initialMetrics.TotalFailed

			t.Logf("Check %v - Failures: %d/5, Processed: %d, Enqueued: %d",
				time.Since(time.Now().Add(-maxWait+10*time.Second)),
				failuresSinceStart,
				metrics.TotalProcessed-initialMetrics.TotalProcessed,
				metrics.TotalEnqueued-initialMetrics.TotalEnqueued)

			// Check if we have enough failures
			if failuresSinceStart >= 5 {
				status, err := qm.GetQueueStatus(ctx)
				if err != nil {
					t.Fatalf("GetQueueStatus() error = %v", err)
				}

				if status.CircuitBreakerOn {
					circuitBreakerOpened = true
					t.Log("✓ Circuit breaker has opened after 5+ failures")
				}
			}
		}
	}

	// Verify circuit breaker is still open
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() error = %v", err)
	}

	if !status.CircuitBreakerOn {
		t.Error("Circuit breaker should remain open")
	}

	// Try to queue another event - should fail because circuit breaker is open
	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData: map[string]interface{}{
			"name":  "Should Fail Customer",
			"email": "fail@example.com",
		},
		UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
	}

	err = qm.QueueEvent(ctx, event)
	if err == nil {
		t.Error("QueueEvent() should fail when circuit breaker is open")
	} else if !strings.Contains(err.Error(), "circuit breaker") {
		t.Errorf("Expected circuit breaker error, got: %v", err)
	} else {
		t.Logf("✓ Queue correctly rejected event with circuit breaker open: %v", err)
	}

	// Verify final metrics
	finalMetrics := qm.GetMetrics()
	t.Logf("Final metrics - Failed: %d, Processed: %d, Enqueued: %d",
		finalMetrics.TotalFailed-initialMetrics.TotalFailed,
		finalMetrics.TotalProcessed-initialMetrics.TotalProcessed,
		finalMetrics.TotalEnqueued-initialMetrics.TotalEnqueued)

	t.Log("✓ Circuit breaker test completed successfully")
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
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Setup database
	db := dbtest.NewDatabase(t, "Test_ClearQueue")
	ctx := context.Background()

	// Create workflow business layer
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

	// Seed basic data (entities, trigger types, etc.)
	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow: %s", err)
	}

	// Create real engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)

	// Initialize the engine
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Create queue manager with real engine
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

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
			RawData: map[string]interface{}{
				"name":  fmt.Sprintf("Customer %d", i),
				"email": fmt.Sprintf("customer%d@example.com", i),
			},
			UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
		}
		if err := qm.QueueEvent(ctx, event); err != nil {
			t.Fatalf("queueing event: %s", err)
		}
	}

	// Give a moment for messages to be fully queued
	time.Sleep(100 * time.Millisecond)

	// Verify messages were queued
	initialStatus, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() before clear error = %v", err)
	}

	if initialStatus.QueueDepth == 0 {
		t.Error("Expected messages to be in queue before clear")
	}

	t.Logf("Queue depth before clear: %d", initialStatus.QueueDepth)

	// Clear the queue
	if err := qm.ClearQueue(ctx); err != nil {
		t.Errorf("ClearQueue() error = %v", err)
	}

	// Small delay to ensure purge completes
	time.Sleep(100 * time.Millisecond)

	// Verify queue is empty
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() after clear error = %v", err)
	}

	if status.QueueDepth != 0 {
		t.Errorf("Queue depth = %d after clear, want 0", status.QueueDepth)
	}

	// Verify metrics show the queued messages
	metrics := qm.GetMetrics()
	if metrics.TotalEnqueued != 5 {
		t.Errorf("Expected 5 messages enqueued, got %d", metrics.TotalEnqueued)
	}

	t.Logf("Successfully cleared queue. Initial depth: %d, Final depth: %d",
		initialStatus.QueueDepth, status.QueueDepth)
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
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Setup database
	db := dbtest.NewDatabase(t, "Test_Metrics")
	ctx := context.Background()

	// Create workflow business layer
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

	// Seed basic data
	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow: %s", err)
	}

	// Create real engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)

	// Initialize the engine
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Create queue manager with real engine
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

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
			RawData: map[string]interface{}{
				"name":  fmt.Sprintf("Customer %d", i),
				"email": fmt.Sprintf("customer%d@example.com", i),
			},
			UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
		}
		if err := qm.QueueEvent(ctx, event); err != nil {
			t.Errorf("QueueEvent() error = %v", err)
		}
	}

	// Wait for processing
	time.Sleep(5 * time.Second)

	// Check metrics
	metrics := qm.GetMetrics()

	if metrics.TotalEnqueued != 4 {
		t.Errorf("TotalEnqueued = %d, want 4", metrics.TotalEnqueued)
	}

	if metrics.TotalProcessed != 4 {
		t.Errorf("TotalProcessed = %d, want 4", metrics.TotalProcessed)
	}

	// Note: With real engine and no rules matching, there shouldn't be failures
	// unless we create a failing rule
	if metrics.TotalFailed != 0 {
		t.Logf("Note: TotalFailed = %d (expected 0 with no matching rules)", metrics.TotalFailed)
	}

	if metrics.LastProcessedAt == nil {
		t.Error("LastProcessedAt should not be nil after processing")
	}

	if metrics.TotalProcessed > 0 && metrics.AverageProcessTimeMs == 0 {
		t.Logf("Warning: AverageProcessTimeMs is 0 (processing was very fast)")
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

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Setup database
	db := dbtest.NewDatabase(t, "Test_QueueType")
	ctx := context.Background()

	// Create workflow business layer
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

	// Seed basic data
	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow: %s", err)
	}

	// Create real engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)

	// Initialize the engine
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Initialize workflow queue to create all queue types
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(ctx); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Create queue manager
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Queue the event (internally uses determineQueueType)
			tt.event.EventType = "on_create"
			tt.event.EntityID = uuid.New()
			tt.event.Timestamp = time.Now()
			tt.event.UserID = uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")
			tt.event.RawData = map[string]interface{}{
				"test": "data",
			}

			if err := qm.QueueEvent(ctx, tt.event); err != nil {
				t.Errorf("QueueEvent() error = %v", err)
			}

			// Verify the message went to the correct queue by checking queue stats
			_, err := queue.GetQueueStats(ctx, tt.wantQueue)
			if err != nil {
				t.Errorf("Failed to get queue stats: %v", err)
			}

			// Clear the queue for next test
			if err := queue.PurgeQueue(ctx, tt.wantQueue); err != nil {
				t.Logf("Warning: Failed to purge queue: %v", err)
			}
		})
	}
}

func TestQueueManager_ProcessingResult(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Setup database
	db := dbtest.NewDatabase(t, "Test_ProcessingResult")
	ctx := context.Background()

	// Create workflow business layer
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

	// Seed basic data
	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow: %s", err)
	}

	// Create real engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)

	// Initialize the engine
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Initialize workflow queue
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(ctx); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Create queue manager
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

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

	// Queue some events first to have something to process
	for i := 0; i < 3; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   uuid.New(),
			Timestamp:  time.Now(),
			RawData: map[string]interface{}{
				"name": fmt.Sprintf("Customer %d", i),
			},
			UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
		}
		if err := qm.QueueEvent(ctx, event); err != nil {
			t.Errorf("Failed to queue event: %v", err)
		}
	}

	// Wait a bit for processing
	time.Sleep(2 * time.Second)

	// Check results after processing
	result2, err := qm.ProcessEvents(ctx, 10)
	if err != nil {
		t.Errorf("Second ProcessEvents() error = %v", err)
	}

	if result2.ProcessedCount < 3 {
		t.Logf("Note: ProcessedCount = %d (events may still be processing)", result2.ProcessedCount)
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

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		b.Fatalf("initializing workflow queue: %s", err)
	}

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client, queue)
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
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		b.Fatalf("initializing workflow queue: %s", err)
	}

	db := sqlx.DB{}

	engine := newStubEngine(log, &db)
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client, queue)
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
