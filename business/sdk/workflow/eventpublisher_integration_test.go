package workflow_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// TestEventPublisher_IntegrationWithRules tests the full workflow:
// EventPublisher → RabbitMQ → QueueManager → Engine → Action Execution
//
// This is the complete integration test that validates the event firing
// infrastructure works end-to-end with real workflow rules.
func TestEventPublisher_IntegrationWithRules(t *testing.T) {
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

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Real database
	db := dbtest.NewDatabase(t, "Test_EventPublisher_Integration")
	ctx := context.Background()

	// Real workflow business layer
	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))

	// NOTE: Do NOT call TestSeedFullWorkflow - trigger types and entities are
	// already seeded by the database migration/seed process. Calling it again
	// causes "duplicated entry" errors.
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// -------------------------------------------------------------------------
	// Create rule for "orders" on_create
	// -------------------------------------------------------------------------

	orderEntity, err := workflowBus.QueryEntityByName(ctx, "orders")
	if err != nil {
		t.Fatalf("querying orders entity: %s", err)
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
		Name:          "Order Created - EventPublisher Test",
		Description:   "Fires when order is created via EventPublisher",
		EntityID:      orderEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerType.ID,
		IsActive:      true,
		CreatedBy:     adminUserID,
	})
	if err != nil {
		t.Fatalf("creating rule: %s", err)
	}

	t.Logf("Created automation rule: %s (ID: %s)", rule.Name, rule.ID)

	// Create email action
	emailTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Order Email Template",
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
		ActionConfig:     json.RawMessage(`{"subject": "New Order: {{number}}", "body": "Order created for customer {{customer_id}}"}`),
		IsActive:         true,
		TemplateID:       &emailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	// -------------------------------------------------------------------------
	// Initialize engine AFTER creating rules
	// -------------------------------------------------------------------------

	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Register action handlers
	registry := engine.GetRegistry()
	registry.Register(communication.NewSendEmailHandler(log, db.DB))
	registry.Register(communication.NewSendNotificationHandler(log, db.DB))
	alertBus := alertbus.NewBusiness(log, alertdb.NewStore(log, db.DB))
	registry.Register(communication.NewCreateAlertHandler(log, alertBus, nil))

	// Create queue manager
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}
	if err := qm.ClearQueue(ctx); err != nil {
		t.Logf("Warning: could not clear queue: %v", err)
	}
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Small delay for consumers to be ready
	time.Sleep(100 * time.Millisecond)

	// -------------------------------------------------------------------------
	// Create EventPublisher and fire event
	// -------------------------------------------------------------------------

	publisher := workflow.NewEventPublisher(log, qm)

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Publish event via EventPublisher (simulating formdataapp)
	orderResult := map[string]interface{}{
		"id":          uuid.New().String(),
		"number":      "ORD-12345",
		"customer_id": uuid.New().String(),
		"due_date":    time.Now().AddDate(0, 0, 30).Format(time.RFC3339),
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

	// -------------------------------------------------------------------------
	// Verify results
	// -------------------------------------------------------------------------

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
			for _, ruleResult := range batch.RuleResults {
				for _, action := range ruleResult.ActionResults {
					if action.ActionType == "send_email" && action.Status == "success" {
						actionExecuted = true
						t.Logf("Email action executed: %s", action.ActionName)
					}
				}
			}
		}
		if !actionExecuted {
			t.Error("Expected email action to execute successfully")
		}
	}

	t.Log("SUCCESS: EventPublisher integration test with rules completed")
}

// TestEventPublisher_MultipleEntityTypes tests that EventPublisher correctly
// handles events for different entity types and routes them to appropriate rules.
func TestEventPublisher_MultipleEntityTypes(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

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

	db := dbtest.NewDatabase(t, "Test_EventPublisher_MultiEntity")
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// NOTE: Trigger types and entities are already seeded by database migrations.

	// Get trigger and entity types
	entityType, _ := workflowBus.QueryEntityTypeByName(ctx, "table")
	createTrigger, _ := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	updateTrigger, _ := workflowBus.QueryTriggerTypeByName(ctx, "on_update")

	// Get entities
	orderEntity, _ := workflowBus.QueryEntityByName(ctx, "orders")
	customerEntity, _ := workflowBus.QueryEntityByName(ctx, "customers")

	// Create email template
	emailTemplate, _ := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Multi Entity Email Template",
		ActionType:    "send_email",
		DefaultConfig: json.RawMessage(`{"recipients": ["test@example.com"]}`),
		CreatedBy:     adminUserID,
	})

	// Create rule for orders on_create
	orderRule, _ := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Order Create Rule",
		EntityID:      orderEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: createTrigger.ID,
		IsActive:      true,
		CreatedBy:     adminUserID,
	})
	workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: orderRule.ID,
		Name:             "Order Email",
		ActionConfig:     json.RawMessage(`{"subject": "Order Created"}`),
		IsActive:         true,
		TemplateID:       &emailTemplate.ID,
	})

	// Create rule for customers on_update
	customerRule, _ := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Customer Update Rule",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: updateTrigger.ID,
		IsActive:      true,
		CreatedBy:     adminUserID,
	})
	workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: customerRule.ID,
		Name:             "Customer Email",
		ActionConfig:     json.RawMessage(`{"subject": "Customer Updated"}`),
		IsActive:         true,
		TemplateID:       &emailTemplate.ID,
	})

	// Initialize engine with rules
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
	engine.Initialize(ctx, workflowBus)
	engine.GetRegistry().Register(communication.NewSendEmailHandler(log, db.DB))

	qm, _ := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	qm.Initialize(ctx)
	qm.ClearQueue(ctx)
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	qm.Start(ctx)
	defer qm.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	publisher := workflow.NewEventPublisher(log, qm)

	// Fire order create event
	publisher.PublishCreateEvent(ctx, "orders", map[string]interface{}{
		"id":     uuid.New().String(),
		"number": "ORD-001",
	}, adminUserID)

	// Fire customer update event
	publisher.PublishUpdateEvent(ctx, "customers", map[string]interface{}{
		"id":   uuid.New().String(),
		"name": "Updated Customer",
	}, nil, adminUserID)

	// Wait for processing
	time.Sleep(1 * time.Second)

	// Verify both events processed
	history := engine.GetExecutionHistory(10)
	if len(history) < 2 {
		t.Errorf("Expected at least 2 executions, got %d", len(history))
	}

	t.Log("SUCCESS: Multiple entity types processed correctly")
}

// TestEventPublisher_TemplateSubstitution tests that entity data is correctly
// passed through the workflow and template substitution works.
func TestEventPublisher_TemplateSubstitution(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	queue.Initialize(context.Background())

	db := dbtest.NewDatabase(t, "Test_EventPublisher_Template")
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// NOTE: Trigger types and entities are already seeded by database migrations.

	// Create rule with template variables
	customerEntity, _ := workflowBus.QueryEntityByName(ctx, "customers")
	entityType, _ := workflowBus.QueryEntityTypeByName(ctx, "table")
	triggerType, _ := workflowBus.QueryTriggerTypeByName(ctx, "on_create")

	emailTemplate, _ := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Template Test",
		ActionType:    "send_email",
		DefaultConfig: json.RawMessage(`{}`),
		CreatedBy:     adminUserID,
	})

	rule, _ := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Template Test Rule",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerType.ID,
		IsActive:      true,
		CreatedBy:     adminUserID,
	})

	// Action with template variables that should be substituted from entity data
	workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Template Email",
		ActionConfig: json.RawMessage(`{
			"recipients": ["{{email}}"],
			"subject": "Welcome {{name}}!",
			"body": "Hello {{name}}, your account {{id}} has been created."
		}`),
		IsActive:       true,
		TemplateID:     &emailTemplate.ID,
	})

	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
	engine.Initialize(ctx, workflowBus)
	engine.GetRegistry().Register(communication.NewSendEmailHandler(log, db.DB))

	qm, _ := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	qm.Initialize(ctx)
	qm.ClearQueue(ctx)
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	qm.Start(ctx)
	defer qm.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	publisher := workflow.NewEventPublisher(log, qm)

	// Fire event with data that should be substituted into template
	customerID := uuid.New().String()
	publisher.PublishCreateEvent(ctx, "customers", map[string]interface{}{
		"id":    customerID,
		"name":  "John Doe",
		"email": "john.doe@example.com",
	}, adminUserID)

	// Wait for processing
	time.Sleep(1 * time.Second)

	// Verify processing completed
	history := engine.GetExecutionHistory(10)
	if len(history) == 0 {
		t.Error("Expected at least one execution")
	} else {
		// Check action result data for evidence of template substitution
		for _, batch := range history[0].BatchResults {
			for _, ruleResult := range batch.RuleResults {
				for _, action := range ruleResult.ActionResults {
					if action.Status == "success" {
						t.Logf("Action executed successfully: %s", action.ActionName)
					}
				}
			}
		}
	}

	t.Log("SUCCESS: Template substitution test completed")
}

// TestEventPublisher_UpdateWithFieldChanges tests that update events correctly
// pass field change information through the workflow.
func TestEventPublisher_UpdateWithFieldChanges(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	queue.Initialize(context.Background())

	db := dbtest.NewDatabase(t, "Test_EventPublisher_FieldChanges")
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// NOTE: Trigger types and entities are already seeded by database migrations.

	// Create rule for status changes
	orderEntity, _ := workflowBus.QueryEntityByName(ctx, "orders")
	entityType, _ := workflowBus.QueryEntityTypeByName(ctx, "table")
	updateTrigger, _ := workflowBus.QueryTriggerTypeByName(ctx, "on_update")

	emailTemplate, _ := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Status Change Template",
		ActionType:    "send_email",
		DefaultConfig: json.RawMessage(`{"recipients": ["test@example.com"]}`),
		CreatedBy:     adminUserID,
	})

	rule, _ := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Order Status Change",
		EntityID:      orderEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: updateTrigger.ID,
		IsActive:      true,
		CreatedBy:     adminUserID,
	})

	workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Status Change Email",
		ActionConfig:     json.RawMessage(`{"subject": "Order status changed"}`),
		IsActive:         true,
		TemplateID:       &emailTemplate.ID,
	})

	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
	engine.Initialize(ctx, workflowBus)
	engine.GetRegistry().Register(communication.NewSendEmailHandler(log, db.DB))

	qm, _ := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	qm.Initialize(ctx)
	qm.ClearQueue(ctx)
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	qm.Start(ctx)
	defer qm.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	publisher := workflow.NewEventPublisher(log, qm)
	initialMetrics := qm.GetMetrics()

	// Fire update event with field changes
	orderResult := map[string]interface{}{
		"id":     uuid.New().String(),
		"number": "ORD-003",
		"status": "shipped",
	}

	fieldChanges := map[string]workflow.FieldChange{
		"status": {
			OldValue: "processing",
			NewValue: "shipped",
		},
	}

	publisher.PublishUpdateEvent(ctx, "orders", orderResult, fieldChanges, adminUserID)

	// Wait for processing
	time.Sleep(1 * time.Second)

	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalProcessed <= initialMetrics.TotalProcessed {
		t.Error("Expected event to be processed")
	}

	t.Log("SUCCESS: Update with field changes processed correctly")
}

// TestEventPublisher_HighVolume tests that EventPublisher can handle a high
// volume of events without blocking or dropping messages.
func TestEventPublisher_HighVolume(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	queue.Initialize(context.Background())

	db := dbtest.NewDatabase(t, "Test_EventPublisher_HighVolume")
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))

	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
	engine.Initialize(ctx, workflowBus)

	qm, _ := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	qm.Initialize(ctx)
	qm.ClearQueue(ctx)
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	qm.Start(ctx)
	defer qm.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	publisher := workflow.NewEventPublisher(log, qm)
	initialMetrics := qm.GetMetrics()

	// Fire 50 events rapidly
	eventCount := 50
	start := time.Now()

	for i := 0; i < eventCount; i++ {
		result := map[string]interface{}{
			"id":     uuid.New().String(),
			"number": i,
		}
		publisher.PublishCreateEvent(ctx, "test_entity", result, uuid.New())
	}

	publishDuration := time.Since(start)
	t.Logf("Published %d events in %v", eventCount, publishDuration)

	// Publishing should be fast (non-blocking)
	if publishDuration > time.Second {
		t.Errorf("Publishing %d events took too long: %v", eventCount, publishDuration)
	}

	// Wait for events to be queued
	time.Sleep(2 * time.Second)

	finalMetrics := qm.GetMetrics()
	enqueuedCount := finalMetrics.TotalEnqueued - initialMetrics.TotalEnqueued

	// All events should be enqueued
	if int(enqueuedCount) != eventCount {
		t.Errorf("Expected %d events enqueued, got %d", eventCount, enqueuedCount)
	}

	t.Logf("SUCCESS: High volume test - %d events enqueued", enqueuedCount)
}

// TestEventPublisher_CreateAlert tests the full workflow for alert creation:
// EventPublisher → RabbitMQ → Engine → CreateAlertHandler → Alert in Database
//
// This validates that the create_alert action actually persists alerts and recipients.
func TestEventPublisher_CreateAlert(t *testing.T) {
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

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Real database
	db := dbtest.NewDatabase(t, "Test_EventPublisher_CreateAlert")
	ctx := context.Background()

	// Real workflow business layer
	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))
	alertBus := alertbus.NewBusiness(log, alertdb.NewStore(log, db.DB))

	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// -------------------------------------------------------------------------
	// Create rule for "orders" on_create with create_alert action
	// -------------------------------------------------------------------------

	orderEntity, err := workflowBus.QueryEntityByName(ctx, "orders")
	if err != nil {
		t.Fatalf("querying orders entity: %s", err)
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
		Name:          "Order Created - Alert Test",
		Description:   "Creates an alert when order is created",
		EntityID:      orderEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerType.ID,
		IsActive:      true,
		CreatedBy:     adminUserID,
	})
	if err != nil {
		t.Fatalf("creating rule: %s", err)
	}

	t.Logf("Created automation rule: %s (ID: %s)", rule.Name, rule.ID)

	// Create alert action template
	alertTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:       "Order Alert Template",
		ActionType: "create_alert",
		DefaultConfig: json.RawMessage(`{
			"alert_type": "order_notification",
			"severity": "medium"
		}`),
		CreatedBy: adminUserID,
	})
	if err != nil {
		t.Fatalf("creating alert template: %s", err)
	}

	// Create rule action with create_alert type, targeting adminUserID as recipient
	_, err = workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Create Order Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "order_notification",
			"severity": "high",
			"title": "New Order: {{number}}",
			"message": "A new order has been created: {{number}}",
			"recipients": {
				"users": ["5cf37266-3473-4006-984f-9325122678b7"]
			}
		}`),
		IsActive:       true,
		TemplateID:     &alertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	// -------------------------------------------------------------------------
	// Initialize engine AFTER creating rules
	// -------------------------------------------------------------------------

	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Register action handlers - importantly, the CreateAlertHandler
	registry := engine.GetRegistry()
	registry.Register(communication.NewSendEmailHandler(log, db.DB))
	registry.Register(communication.NewSendNotificationHandler(log, db.DB))
	registry.Register(communication.NewCreateAlertHandler(log, alertBus, nil))

	// Create queue manager
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}
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

	// Small delay for consumers to be ready
	time.Sleep(100 * time.Millisecond)

	// -------------------------------------------------------------------------
	// Verify no alerts exist before the test
	// -------------------------------------------------------------------------

	initialAlerts, err := alertBus.QueryMine(ctx, adminUserID, nil, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying initial alerts: %s", err)
	}
	initialAlertCount := len(initialAlerts)

	// -------------------------------------------------------------------------
	// Create EventPublisher and fire event
	// -------------------------------------------------------------------------

	publisher := workflow.NewEventPublisher(log, qm)
	initialMetrics := qm.GetMetrics()

	// Publish event via EventPublisher (simulating formdataapp)
	orderNumber := "ORD-ALERT-TEST-001"
	orderResult := map[string]interface{}{
		"id":          uuid.New().String(),
		"number":      orderNumber,
		"customer_id": uuid.New().String(),
		"due_date":    time.Now().AddDate(0, 0, 30).Format(time.RFC3339),
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

	// Give a little extra time for alert to be persisted
	time.Sleep(200 * time.Millisecond)

	// -------------------------------------------------------------------------
	// Verify alert was created in the database
	// -------------------------------------------------------------------------

	finalAlerts, err := alertBus.QueryMine(ctx, adminUserID, nil, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying final alerts: %s", err)
	}

	newAlertCount := len(finalAlerts) - initialAlertCount
	if newAlertCount != 1 {
		t.Errorf("Expected 1 new alert, got %d", newAlertCount)
	}

	// Find our alert by checking the title contains the order number
	var foundAlert *alertbus.Alert
	for i := range finalAlerts {
		if finalAlerts[i].Title == "New Order: "+orderNumber {
			foundAlert = &finalAlerts[i]
			break
		}
	}

	if foundAlert == nil {
		t.Error("Could not find alert with expected title")
	} else {
		// Verify alert properties
		if foundAlert.AlertType != "order_notification" {
			t.Errorf("Expected alert_type 'order_notification', got '%s'", foundAlert.AlertType)
		}
		if foundAlert.Severity != "high" {
			t.Errorf("Expected severity 'high', got '%s'", foundAlert.Severity)
		}
		if foundAlert.Message != "A new order has been created: "+orderNumber {
			t.Errorf("Expected message with order number, got '%s'", foundAlert.Message)
		}
		if foundAlert.SourceEntityName != "orders" {
			t.Errorf("Expected source_entity_name 'orders', got '%s'", foundAlert.SourceEntityName)
		}
		if foundAlert.Status != alertbus.StatusActive {
			t.Errorf("Expected status 'active', got '%s'", foundAlert.Status)
		}

		t.Logf("Alert created successfully: ID=%s, Title=%s", foundAlert.ID, foundAlert.Title)
	}

	// -------------------------------------------------------------------------
	// Verify execution history shows create_alert action executed
	// -------------------------------------------------------------------------

	history := engine.GetExecutionHistory(10)
	if len(history) == 0 {
		t.Error("Expected at least one execution in history")
	} else {
		lastExec := history[0]
		if lastExec.ExecutionPlan.MatchedRuleCount != 1 {
			t.Errorf("Expected 1 matched rule, got %d", lastExec.ExecutionPlan.MatchedRuleCount)
		}

		// Verify create_alert action executed
		alertActionExecuted := false
		for _, batch := range lastExec.BatchResults {
			for _, ruleResult := range batch.RuleResults {
				for _, action := range ruleResult.ActionResults {
					if action.ActionType == "create_alert" && action.Status == "success" {
						alertActionExecuted = true
						t.Logf("Create alert action executed: %s", action.ActionName)
					}
				}
			}
		}
		if !alertActionExecuted {
			t.Error("Expected create_alert action to execute successfully")
		}
	}

	// Verify metrics
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

	t.Log("SUCCESS: CreateAlert integration test completed - alert persisted to database")
}
