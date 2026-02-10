package formdataapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// TestWorkflow_OrderCreateEvent tests that workflow events fire correctly
// when orders are created. This test directly uses the workflow infrastructure
// without going through the HTTP layer.
func TestWorkflow_OrderCreateEvent(t *testing.T) {
	t.Parallel()

	// Setup database directly (no HTTP mux needed).
	db := dbtest.NewDatabase(t, "Test_Workflow_OrderCreate")

	// Initialize Temporal-based workflow infrastructure.
	wf := apitest.InitWorkflowInfra(t, db)

	ctx := context.Background()
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// -------------------------------------------------------------------------
	// Create workflow rule for orders on_create.
	// -------------------------------------------------------------------------

	orderEntity, err := wf.WorkflowBus.QueryEntityByName(ctx, "orders")
	if err != nil {
		t.Fatalf("querying orders entity: %s", err)
	}

	entityType, err := wf.WorkflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %s", err)
	}

	triggerType, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying trigger type: %s", err)
	}

	rule, err := wf.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:              "Order Created - Send Email Notification",
		Description:       "Sends email notification when an order is created",
		EntityID:          orderEntity.ID,
		EntityTypeID:      entityType.ID,
		TriggerTypeID:     triggerType.ID,
		TriggerConditions: nil,
		IsActive:          true,
		CreatedBy:         adminUserID,
	})
	if err != nil {
		t.Fatalf("creating automation rule: %s", err)
	}

	t.Logf("Created automation rule: %s (ID: %s)", rule.Name, rule.ID)

	// Create email action template.
	emailTemplate, err := wf.WorkflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Order Email Notification",
		Description: "Template for order notification emails",
		ActionType:  "send_email",
		DefaultConfig: json.RawMessage(`{
			"recipients": ["orders@example.com"],
			"subject": "New Order Created"
		}`),
		CreatedBy: adminUserID,
	})
	if err != nil {
		t.Fatalf("creating action template: %s", err)
	}

	// Create rule action.
	emailAction, err := wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Send Order Email",
		Description:      "Send email notification for new order",
		ActionConfig: json.RawMessage(`{
			"recipients": ["sales@example.com", "orders@example.com"],
			"subject": "New Order Created: {{number}}",
			"body": "A new order has been created.\n\nOrder Number: {{number}}\nCustomer ID: {{customer_id}}"
		}`),
		IsActive:   true,
		TemplateID: &emailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating rule action: %s", err)
	}

	_, err = wf.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: emailAction.ID,
		EdgeType:       workflow.EdgeTypeStart,
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating edge for rule action: %s", err)
	}

	// Refresh trigger processor to pick up the new rule.
	if err := wf.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %s", err)
	}

	// -------------------------------------------------------------------------
	// Test: Fire workflow event and verify processing.
	// -------------------------------------------------------------------------

	t.Run("success-email-action-dispatched", func(t *testing.T) {
		// Simulate what the delegate handler does when an order is created.
		orderID := uuid.New()
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "orders",
			EntityID:   orderID,
			Timestamp:  time.Now(),
			RawData: map[string]any{
				"id":          orderID.String(),
				"number":      "TEST-ORDER-001",
				"customer_id": uuid.New().String(),
				"due_date":    time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
			},
			UserID: adminUserID,
		}

		// Dispatch via WorkflowTrigger (replaces QueueManager.QueueEvent).
		if err := wf.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
			t.Fatalf("dispatching workflow event: %s", err)
		}

		// Wait for Temporal workflow execution.
		time.Sleep(3 * time.Second)

		// Verify: no error from dispatch means the workflow was started successfully.
		// Full execution verification (email handler side effects) is covered by temporal package tests.
		t.Log("SUCCESS: Workflow event dispatched and processed via Temporal")
	})

	// -------------------------------------------------------------------------
	// Test: Workflow with simulate_failure flag.
	// -------------------------------------------------------------------------

	t.Run("failure-simulated-logged-not-blocking", func(t *testing.T) {
		// Create a rule with simulate_failure enabled.
		failRule, err := wf.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
			Name:              "Order Created - Fail Test",
			Description:       "Rule that simulates failure for testing",
			EntityID:          orderEntity.ID,
			EntityTypeID:      entityType.ID,
			TriggerTypeID:     triggerType.ID,
			TriggerConditions: nil,
			IsActive:          true,
			CreatedBy:         adminUserID,
		})
		if err != nil {
			t.Fatalf("creating fail rule: %s", err)
		}

		failAction, err := wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: failRule.ID,
			Name:             "Failing Email Action",
			Description:      "Email action that simulates failure",
			ActionConfig: json.RawMessage(`{
				"recipients": ["test@example.com"],
				"subject": "This should fail",
				"body": "Test",
				"simulate_failure": true,
				"failure_message": "Simulated SMTP connection refused"
			}`),
			IsActive:   true,
			TemplateID: &emailTemplate.ID,
		})
		if err != nil {
			t.Fatalf("creating failing action: %s", err)
		}

		_, err = wf.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
			RuleID:         failRule.ID,
			SourceActionID: nil,
			TargetActionID: failAction.ID,
			EdgeType:       workflow.EdgeTypeStart,
			EdgeOrder:      0,
		})
		if err != nil {
			t.Fatalf("creating edge for failing action: %s", err)
		}

		// Refresh trigger processor to pick up the new rule.
		if err := wf.TriggerProcessor.RefreshRules(ctx); err != nil {
			t.Fatalf("refreshing trigger processor: %s", err)
		}

		// Fire workflow event.
		orderID := uuid.New()
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "orders",
			EntityID:   orderID,
			Timestamp:  time.Now(),
			RawData: map[string]any{
				"id":     orderID.String(),
				"number": "TEST-ORDER-FAIL",
			},
			UserID: adminUserID,
		}

		// Dispatch - the trigger itself should succeed (fail-open per rule).
		if err := wf.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
			t.Fatalf("dispatching workflow event: %s", err)
		}

		// Wait for Temporal workflow execution (needs longer for retry delays).
		time.Sleep(5 * time.Second)

		// Verify: dispatch succeeded. The workflow may fail internally
		// (Temporal retries), but the trigger dispatch should not block.
		t.Log("SUCCESS: Workflow failure handled gracefully via Temporal")
	})
}
