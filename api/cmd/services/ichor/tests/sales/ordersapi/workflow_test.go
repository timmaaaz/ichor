package ordersapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// TestWorkflow_OrdersDelegateEvents tests that workflow events fire correctly
// when orders are created, updated, and deleted via ordersbus.
// This validates the Phase 2 delegate pattern integration.
func TestWorkflow_OrdersDelegateEvents(t *testing.T) {
	t.Parallel()

	// Setup database
	db := dbtest.NewDatabase(t, "Test_Workflow_OrdersDelegate")

	// Initialize workflow infrastructure
	wf := apitest.InitWorkflowInfra(t, db)

	ctx := context.Background()
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// -------------------------------------------------------------------------
	// Create workflow rule for orders on_create
	// -------------------------------------------------------------------------

	// Query for orders entity (must exist in seed data)
	// Note: Entity name in workflow.entities is just the table name, not schema-qualified
	orderEntity, err := wf.WorkflowBus.QueryEntityByName(ctx, "orders")
	if err != nil {
		t.Fatalf("querying orders entity: %s", err)
	}

	entityType, err := wf.WorkflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %s", err)
	}

	triggerTypeCreate, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %s", err)
	}

	triggerTypeUpdate, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_update")
	if err != nil {
		t.Fatalf("querying on_update trigger type: %s", err)
	}

	triggerTypeDelete, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_delete")
	if err != nil {
		t.Fatalf("querying on_delete trigger type: %s", err)
	}

	// Create email action template
	emailTemplate, err := wf.WorkflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Order Event Notification",
		Description: "Template for order event notifications",
		ActionType:  "send_email",
		DefaultConfig: json.RawMessage(`{
			"recipients": ["orders@example.com"],
			"subject": "Order Event"
		}`),
		CreatedBy: adminUserID,
	})
	if err != nil {
		t.Fatalf("creating action template: %s", err)
	}

	// Create automation rule for order creation
	ruleCreate, err := wf.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:              "Order Created - Delegate Pattern Test",
		Description:       "Tests delegate pattern fires on_create events",
		EntityID:          orderEntity.ID,
		EntityTypeID:      entityType.ID,
		TriggerTypeID:     triggerTypeCreate.ID,
		TriggerConditions: nil,
		IsActive:          true,
		CreatedBy:         adminUserID,
	})
	if err != nil {
		t.Fatalf("creating on_create rule: %s", err)
	}

	_, err = wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: ruleCreate.ID,
		Name:             "Send Create Email",
		ActionConfig: json.RawMessage(`{
			"recipients": ["sales@example.com"],
			"subject": "Order Created: {{Number}}",
			"body": "Order {{ID}} was created"
		}`),
		ExecutionOrder: 1,
		IsActive:       true,
		TemplateID:     &emailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating on_create rule action: %s", err)
	}

	// Create automation rule for order update
	ruleUpdate, err := wf.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:              "Order Updated - Delegate Pattern Test",
		Description:       "Tests delegate pattern fires on_update events",
		EntityID:          orderEntity.ID,
		EntityTypeID:      entityType.ID,
		TriggerTypeID:     triggerTypeUpdate.ID,
		TriggerConditions: nil,
		IsActive:          true,
		CreatedBy:         adminUserID,
	})
	if err != nil {
		t.Fatalf("creating on_update rule: %s", err)
	}

	_, err = wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: ruleUpdate.ID,
		Name:             "Send Update Email",
		ActionConfig: json.RawMessage(`{
			"recipients": ["sales@example.com"],
			"subject": "Order Updated: {{Number}}",
			"body": "Order {{ID}} was updated"
		}`),
		ExecutionOrder: 1,
		IsActive:       true,
		TemplateID:     &emailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating on_update rule action: %s", err)
	}

	// Create automation rule for order delete
	ruleDelete, err := wf.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:              "Order Deleted - Delegate Pattern Test",
		Description:       "Tests delegate pattern fires on_delete events",
		EntityID:          orderEntity.ID,
		EntityTypeID:      entityType.ID,
		TriggerTypeID:     triggerTypeDelete.ID,
		TriggerConditions: nil,
		IsActive:          true,
		CreatedBy:         adminUserID,
	})
	if err != nil {
		t.Fatalf("creating on_delete rule: %s", err)
	}

	_, err = wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: ruleDelete.ID,
		Name:             "Send Delete Email",
		ActionConfig: json.RawMessage(`{
			"recipients": ["sales@example.com"],
			"subject": "Order Deleted: {{Number}}",
			"body": "Order {{ID}} was deleted"
		}`),
		ExecutionOrder: 1,
		IsActive:       true,
		TemplateID:     &emailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating on_delete rule action: %s", err)
	}

	// Re-initialize engine to pick up new rules
	if err := wf.Engine.Initialize(ctx, wf.WorkflowBus); err != nil {
		t.Fatalf("re-initializing engine: %s", err)
	}

	// -------------------------------------------------------------------------
	// Create EventPublisher and wire up DelegateHandler
	// -------------------------------------------------------------------------

	eventPublisher := workflow.NewEventPublisher(db.Log, wf.QueueManager)
	delegateHandler := workflow.NewDelegateHandler(db.Log, eventPublisher)
	delegateHandler.RegisterDomain(db.BusDomain.Delegate, ordersbus.DomainName, ordersbus.EntityName)

	// Use the existing ordersbus from BusDomain (shares delegate with delegate handler)
	ordersBus := db.BusDomain.Order

	// -------------------------------------------------------------------------
	// Seed FK dependencies for orders
	// -------------------------------------------------------------------------

	// 1. Seed order fulfillment statuses
	oflStatuses, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, db.BusDomain.OrderFulfillmentStatus)
	if err != nil {
		t.Fatalf("seeding order fulfillment statuses: %s", err)
	}
	fulfillmentStatusID := oflStatuses[0].ID

	// 2. Seed minimal customer dependencies
	regions, err := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		t.Fatalf("querying regions: %s", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, db.BusDomain.City)
	if err != nil {
		t.Fatalf("seeding cities: %s", err)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 1, []uuid.UUID{cities[0].ID}, db.BusDomain.Street)
	if err != nil {
		t.Fatalf("seeding streets: %s", err)
	}

	tzs, err := db.BusDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		t.Fatalf("querying timezones: %s", err)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 1, []uuid.UUID{streets[0].ID}, []uuid.UUID{tzs[0].ID}, db.BusDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contact infos: %s", err)
	}

	customers, err := customersbus.TestSeedCustomers(ctx, 1, []uuid.UUID{streets[0].ID}, []uuid.UUID{contactInfos[0].ID}, []uuid.UUID{adminUserID}, db.BusDomain.Customers)
	if err != nil {
		t.Fatalf("seeding customers: %s", err)
	}
	customerID := customers[0].ID

	// 3. Seed currencies for orders
	currencies, err := currencybus.TestSeedCurrencies(ctx, 1, db.BusDomain.Currency)
	if err != nil {
		t.Fatalf("seeding currencies: %s", err)
	}
	currencyID := currencies[0].ID

	// -------------------------------------------------------------------------
	// Test: Create order fires on_create event
	// -------------------------------------------------------------------------

	t.Run("create-fires-on_create-event", func(t *testing.T) {
		initialMetrics := wf.QueueManager.GetMetrics()

		// Create an order through ordersbus
		newOrder := ordersbus.NewOrder{
			Number:              "DELEGATE-TEST-001",
			CustomerID:          customerID,
			DueDate:             time.Now().Add(7 * 24 * time.Hour),
			FulfillmentStatusID: fulfillmentStatusID,
			CurrencyID:          currencyID,
			CreatedBy:           adminUserID,
		}

		order, err := ordersBus.Create(ctx, newOrder)
		if err != nil {
			t.Fatalf("creating order: %s", err)
		}

		t.Logf("Created order: %s (ID: %s)", order.Number, order.ID)

		// Wait for async event processing
		waitForProcessing(t, wf.QueueManager, initialMetrics, 5*time.Second)

		// Verify event was enqueued and processed
		finalMetrics := wf.QueueManager.GetMetrics()
		if finalMetrics.TotalEnqueued <= initialMetrics.TotalEnqueued {
			t.Errorf("Expected event to be enqueued, but TotalEnqueued did not increase")
		}

		if finalMetrics.TotalProcessed <= initialMetrics.TotalProcessed {
			t.Errorf("Expected event to be processed, but TotalProcessed did not increase")
		}

		// Verify email action was executed
		execHistory := wf.Engine.GetExecutionHistory(10)
		if len(execHistory) == 0 {
			t.Error("Expected at least one execution in history")
		} else {
			verifyActionExecuted(t, execHistory[0], "send_email", "on_create")
		}

		t.Log("SUCCESS: ordersbus.Create() fired on_create event via delegate")
	})

	// -------------------------------------------------------------------------
	// Test: Update order fires on_update event
	// -------------------------------------------------------------------------

	t.Run("update-fires-on_update-event", func(t *testing.T) {
		// First create an order to update
		newOrder := ordersbus.NewOrder{
			Number:              "DELEGATE-TEST-002",
			CustomerID:          customerID,
			DueDate:             time.Now().Add(7 * 24 * time.Hour),
			FulfillmentStatusID: fulfillmentStatusID,
			CurrencyID:          currencyID,
			CreatedBy:           adminUserID,
		}

		order, err := ordersBus.Create(ctx, newOrder)
		if err != nil {
			t.Fatalf("creating order for update test: %s", err)
		}

		// Wait for create event to process
		time.Sleep(500 * time.Millisecond)

		initialMetrics := wf.QueueManager.GetMetrics()

		// Update the order
		updatedNumber := "DELEGATE-TEST-002-UPDATED"
		updateOrder := ordersbus.UpdateOrder{
			Number:    &updatedNumber,
			UpdatedBy: &adminUserID,
		}

		_, err = ordersBus.Update(ctx, order, updateOrder)
		if err != nil {
			t.Fatalf("updating order: %s", err)
		}

		t.Logf("Updated order: %s -> %s", order.Number, updatedNumber)

		// Wait for async event processing
		waitForProcessing(t, wf.QueueManager, initialMetrics, 5*time.Second)

		// Verify event was enqueued and processed
		finalMetrics := wf.QueueManager.GetMetrics()
		if finalMetrics.TotalEnqueued <= initialMetrics.TotalEnqueued {
			t.Errorf("Expected event to be enqueued, but TotalEnqueued did not increase")
		}

		// Verify execution history shows on_update
		execHistory := wf.Engine.GetExecutionHistory(10)
		foundUpdate := false
		for _, exec := range execHistory {
			if exec.ExecutionPlan.TriggerEvent.EventType == "on_update" {
				foundUpdate = true
				verifyActionExecuted(t, exec, "send_email", "on_update")
				break
			}
		}
		if !foundUpdate {
			t.Error("Expected on_update event in execution history")
		}

		t.Log("SUCCESS: ordersbus.Update() fired on_update event via delegate")
	})

	// -------------------------------------------------------------------------
	// Test: Delete order fires on_delete event
	// -------------------------------------------------------------------------

	t.Run("delete-fires-on_delete-event", func(t *testing.T) {
		// First create an order to delete
		newOrder := ordersbus.NewOrder{
			Number:              "DELEGATE-TEST-003",
			CustomerID:          customerID,
			DueDate:             time.Now().Add(7 * 24 * time.Hour),
			FulfillmentStatusID: fulfillmentStatusID,
			CurrencyID:          currencyID,
			CreatedBy:           adminUserID,
		}

		order, err := ordersBus.Create(ctx, newOrder)
		if err != nil {
			t.Fatalf("creating order for delete test: %s", err)
		}

		// Wait for create event to process
		time.Sleep(500 * time.Millisecond)

		initialMetrics := wf.QueueManager.GetMetrics()

		// Delete the order
		err = ordersBus.Delete(ctx, order)
		if err != nil {
			t.Fatalf("deleting order: %s", err)
		}

		t.Logf("Deleted order: %s (ID: %s)", order.Number, order.ID)

		// Wait for async event processing
		waitForProcessing(t, wf.QueueManager, initialMetrics, 5*time.Second)

		// Verify event was enqueued and processed
		finalMetrics := wf.QueueManager.GetMetrics()
		if finalMetrics.TotalEnqueued <= initialMetrics.TotalEnqueued {
			t.Errorf("Expected event to be enqueued, but TotalEnqueued did not increase")
		}

		// Verify execution history shows on_delete
		execHistory := wf.Engine.GetExecutionHistory(10)
		foundDelete := false
		for _, exec := range execHistory {
			if exec.ExecutionPlan.TriggerEvent.EventType == "on_delete" {
				foundDelete = true
				verifyActionExecuted(t, exec, "send_email", "on_delete")
				break
			}
		}
		if !foundDelete {
			t.Error("Expected on_delete event in execution history")
		}

		t.Log("SUCCESS: ordersbus.Delete() fired on_delete event via delegate")
	})
}

// waitForProcessing waits for at least one event to be processed.
func waitForProcessing(t *testing.T, qm *workflow.QueueManager, initial workflow.QueueMetrics, timeout time.Duration) {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			metrics := qm.GetMetrics()
			t.Fatalf("Timeout waiting for processing - Enqueued: %d, Processed: %d, Failed: %d",
				metrics.TotalEnqueued, metrics.TotalProcessed, metrics.TotalFailed)
		case <-ticker.C:
			metrics := qm.GetMetrics()
			if metrics.TotalProcessed > initial.TotalProcessed ||
				metrics.TotalFailed > initial.TotalFailed {
				return
			}
		}
	}
}

// verifyActionExecuted checks that an action was executed for the given event type.
func verifyActionExecuted(t *testing.T, exec *workflow.WorkflowExecution, actionType, eventType string) {
	t.Helper()

	if exec.ExecutionPlan.TriggerEvent.EventType != eventType {
		t.Logf("Execution event type: %s (expected: %s)", exec.ExecutionPlan.TriggerEvent.EventType, eventType)
	}

	if exec.ExecutionPlan.MatchedRuleCount < 1 {
		t.Errorf("Expected at least 1 matched rule for %s, got %d", eventType, exec.ExecutionPlan.MatchedRuleCount)
		return
	}

	actionExecuted := false
	for _, batch := range exec.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			for _, actionResult := range ruleResult.ActionResults {
				if actionResult.ActionType == actionType {
					actionExecuted = true
					if actionResult.Status != "success" {
						t.Errorf("%s action failed for %s: %s", actionType, eventType, actionResult.ErrorMessage)
					} else {
						t.Logf("%s action executed successfully for %s", actionType, eventType)
					}
				}
			}
		}
	}

	if !actionExecuted {
		t.Errorf("Expected %s action to execute for %s", actionType, eventType)
	}
}
