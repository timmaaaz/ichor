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

	// Setup database directly (no HTTP mux needed)
	db := dbtest.NewDatabase(t, "Test_Workflow_OrderCreate")

	// Initialize workflow infrastructure (optional - only for tests that need it)
	wf := apitest.InitWorkflowInfra(t, db)

	ctx := context.Background()
	adminUserID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// -------------------------------------------------------------------------
	// Create workflow rule for orders on_create

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

	// Create automation rule for orders
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

	// Create email action template
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

	// Create rule action
	emailAction, err := wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Send Order Email",
		Description:      "Send email notification for new order",
		ActionConfig: json.RawMessage(`{
			"recipients": ["sales@example.com", "orders@example.com"],
			"subject": "New Order Created: {{number}}",
			"body": "A new order has been created.\n\nOrder Number: {{number}}\nCustomer ID: {{customer_id}}"
		}`),

		IsActive:       true,
		TemplateID:     &emailTemplate.ID,
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

	// Re-initialize engine to pick up the new rule
	if err := wf.Engine.Initialize(ctx, wf.WorkflowBus); err != nil {
		t.Fatalf("re-initializing engine: %s", err)
	}

	// -------------------------------------------------------------------------
	// Test: Fire workflow event and verify processing

	t.Run("success-email-action-executes", func(t *testing.T) {
		initialMetrics := wf.QueueManager.GetMetrics()

		// Simulate what EventPublisher will do after formdata creates an order
		orderID := uuid.New()
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "orders",
			EntityID:   orderID,
			Timestamp:  time.Now(),
			RawData: map[string]interface{}{
				"id":          orderID.String(),
				"number":      "TEST-ORDER-001",
				"customer_id": uuid.New().String(),
				"due_date":    time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
			},
			UserID: adminUserID,
		}

		if err := wf.QueueManager.QueueEvent(ctx, event); err != nil {
			t.Fatalf("queueing workflow event: %s", err)
		}

		// Wait for processing
		waitForProcessing(t, wf.QueueManager, initialMetrics, 5*time.Second)

		// Verify results
		finalMetrics := wf.QueueManager.GetMetrics()

		if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
			t.Errorf("Expected 1 event enqueued, got %d",
				finalMetrics.TotalEnqueued-initialMetrics.TotalEnqueued)
		}

		if finalMetrics.TotalProcessed != initialMetrics.TotalProcessed+1 {
			t.Errorf("Expected 1 event processed, got %d",
				finalMetrics.TotalProcessed-initialMetrics.TotalProcessed)
		}

		if finalMetrics.TotalFailed > initialMetrics.TotalFailed {
			t.Errorf("Unexpected failures: %d",
				finalMetrics.TotalFailed-initialMetrics.TotalFailed)
		}

		// Verify email action was executed
		execHistory := wf.Engine.GetExecutionHistory(10)
		if len(execHistory) == 0 {
			t.Error("Expected at least one execution in history")
		} else {
			verifyEmailActionExecuted(t, execHistory[0])
		}

		t.Log("SUCCESS: Workflow event processed and email action executed")
	})

	// -------------------------------------------------------------------------
	// Test: Workflow failure with simulate_failure flag

	t.Run("failure-simulated-logged-not-blocking", func(t *testing.T) {
		// Create a rule with simulate_failure enabled
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

		// Create action with simulate_failure
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

			IsActive:       true,
			TemplateID:     &emailTemplate.ID,
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

		// Re-initialize engine to pick up the new rule
		if err := wf.Engine.Initialize(ctx, wf.WorkflowBus); err != nil {
			t.Fatalf("re-initializing engine: %s", err)
		}

		initialMetrics := wf.QueueManager.GetMetrics()

		// Fire workflow event
		orderID := uuid.New()
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "orders",
			EntityID:   orderID,
			Timestamp:  time.Now(),
			RawData: map[string]interface{}{
				"id":     orderID.String(),
				"number": "TEST-ORDER-FAIL",
			},
			UserID: adminUserID,
		}

		wf.QueueManager.QueueEvent(ctx, event)

		// Wait for processing (needs longer timeout due to retry delays)
		waitForProcessing(t, wf.QueueManager, initialMetrics, 10*time.Second)

		// Event should be processed (even if actions fail)
		finalMetrics := wf.QueueManager.GetMetrics()
		if finalMetrics.TotalProcessed <= initialMetrics.TotalProcessed {
			t.Error("Expected event to be processed")
		}

		// Check execution history for failures
		execHistory := wf.Engine.GetExecutionHistory(10)
		if len(execHistory) > 0 {
			lastExec := execHistory[0]
			t.Logf("Matched rules: %d", lastExec.ExecutionPlan.MatchedRuleCount)

			for _, batch := range lastExec.BatchResults {
				for _, ruleResult := range batch.RuleResults {
					for _, actionResult := range ruleResult.ActionResults {
						if actionResult.Status == "failed" {
							t.Logf("EXPECTED FAILURE: %s - %s",
								actionResult.ActionType, actionResult.ErrorMessage)
						}
					}
				}
			}
		}

		t.Log("SUCCESS: Workflow failure logged, processing continued")
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

// verifyEmailActionExecuted checks that an email action was executed in the workflow execution.
func verifyEmailActionExecuted(t *testing.T, exec *workflow.WorkflowExecution) {
	t.Helper()

	if exec.ExecutionPlan.MatchedRuleCount < 1 {
		t.Errorf("Expected at least 1 matched rule, got %d", exec.ExecutionPlan.MatchedRuleCount)
		return
	}

	emailExecuted := false
	for _, batch := range exec.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			for _, actionResult := range ruleResult.ActionResults {
				if actionResult.ActionType == "send_email" {
					emailExecuted = true
					if actionResult.Status != "success" {
						t.Errorf("Email action failed: %s", actionResult.ErrorMessage)
					} else {
						t.Log("EMAIL SENT TO recipients (logged by handler)")
					}
				}
			}
		}
	}

	if !emailExecuted {
		t.Error("Expected email action to execute")
	}
}
