package workflowsaveapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ExecutionTestData extends SaveSeedData with workflow execution infrastructure.
type ExecutionTestData struct {
	SaveSeedData
	WF *apitest.WorkflowInfra // Engine, QueueManager, WorkflowBus, Client

	// Action templates for different action types
	CreateAlertTemplate      workflow.ActionTemplate
	SendEmailTemplate        workflow.ActionTemplate
	EvaluateConditionTemplate workflow.ActionTemplate

	// Created workflows for testing (via Save API)
	SimpleWorkflow    *workflowsaveapp.SaveWorkflowResponse
	SequenceWorkflow  *workflowsaveapp.SaveWorkflowResponse
	BranchingWorkflow *workflowsaveapp.SaveWorkflowResponse
}

// insertExecutionSeedData initializes workflow infrastructure and creates test workflows.
// It accepts the already-created SaveSeedData to avoid duplicate role creation.
func insertExecutionSeedData(t *testing.T, test *apitest.Test, sd SaveSeedData) ExecutionTestData {
	t.Helper()
	ctx := context.Background()

	// 2. Use existing workflow infrastructure helper
	wf := apitest.InitWorkflowInfra(t, test.DB)

	// 3. Reset for clean test state
	workflow.ResetEngineForTesting()
	wf.QueueManager.ResetMetrics()

	// 4. Re-initialize engine after reset
	if err := wf.Engine.Initialize(ctx, wf.WorkflowBus); err != nil {
		t.Fatalf("initializing workflow engine: %v", err)
	}

	// 5. Create action templates for different action types
	createAlertTemplate, err := wf.WorkflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Create Alert Template",
		Description: "Template for create_alert actions",
		ActionType:  "create_alert",
		DefaultConfig: json.RawMessage(`{
			"alert_type": "default",
			"severity": "info",
			"title": "Alert",
			"message": "Default message"
		}`),
		CreatedBy: sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating create_alert template: %v", err)
	}

	sendEmailTemplate, err := wf.WorkflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Send Email Template",
		Description: "Template for send_email actions",
		ActionType:  "send_email",
		DefaultConfig: json.RawMessage(`{
			"recipients": ["test@example.com"],
			"subject": "Default Subject",
			"body": "Default body"
		}`),
		CreatedBy: sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating send_email template: %v", err)
	}

	evaluateConditionTemplate, err := wf.WorkflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Evaluate Condition Template",
		Description: "Template for evaluate_condition actions",
		ActionType:  "evaluate_condition",
		DefaultConfig: json.RawMessage(`{
			"conditions": []
		}`),
		CreatedBy: sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating evaluate_condition template: %v", err)
	}

	// 6. Create test workflows via direct business layer
	// We create these directly rather than via HTTP to avoid test interdependencies
	simpleWorkflow := createSimpleWorkflowDirect(t, wf.WorkflowBus, sd, createAlertTemplate.ID)
	sequenceWorkflow := createSequenceWorkflowDirect(t, wf.WorkflowBus, sd, createAlertTemplate.ID)
	branchingWorkflow := createBranchingWorkflowDirect(t, wf.WorkflowBus, sd, createAlertTemplate.ID, evaluateConditionTemplate.ID)

	// 7. Re-initialize engine after creating workflows to pick up new rules
	if err := wf.Engine.Initialize(ctx, wf.WorkflowBus); err != nil {
		t.Fatalf("re-initializing workflow engine: %v", err)
	}

	return ExecutionTestData{
		SaveSeedData:              sd,
		WF:                        wf,
		CreateAlertTemplate:       createAlertTemplate,
		SendEmailTemplate:         sendEmailTemplate,
		EvaluateConditionTemplate: evaluateConditionTemplate,
		SimpleWorkflow:            simpleWorkflow,
		SequenceWorkflow:          sequenceWorkflow,
		BranchingWorkflow:         branchingWorkflow,
	}
}

// createSimpleWorkflowDirect creates a workflow with 1 action via the business layer.
func createSimpleWorkflowDirect(t *testing.T, wfBus *workflow.Business, sd SaveSeedData, alertTemplateID uuid.UUID) *workflowsaveapp.SaveWorkflowResponse {
	t.Helper()
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data for simple workflow")
	}

	// Create rule
	rule, err := wfBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Simple Test Workflow",
		Description:   "A simple workflow with 1 action",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating simple workflow rule: %v", err)
	}

	// Create action with template reference
	action, err := wfBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Create Alert",
		Description:      "Creates an alert",
		ActionConfig:     json.RawMessage(`{"alert_type":"simple_test","severity":"info","title":"Simple Test Alert","message":"Test message from simple workflow"}`),
		ExecutionOrder:   1,
		IsActive:         true,
		TemplateID:       &alertTemplateID,
	})
	if err != nil {
		t.Fatalf("creating simple workflow action: %v", err)
	}

	// Create start edge
	edge, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating simple workflow edge: %v", err)
	}

	return &workflowsaveapp.SaveWorkflowResponse{
		ID:            rule.ID.String(),
		Name:          rule.Name,
		Description:   rule.Description,
		IsActive:      rule.IsActive,
		EntityID:      rule.EntityID.String(),
		TriggerTypeID: rule.TriggerTypeID.String(),
		Actions: []workflowsaveapp.SaveActionResponse{
			{
				ID:             action.ID.String(),
				Name:           action.Name,
				Description:    action.Description,
				ActionType:     "create_alert",
				ActionConfig:   action.ActionConfig,
				ExecutionOrder: action.ExecutionOrder,
				IsActive:       action.IsActive,
			},
		},
		Edges: []workflowsaveapp.SaveEdgeResponse{
			{
				ID:             edge.ID.String(),
				SourceActionID: "",
				TargetActionID: action.ID.String(),
				EdgeType:       edge.EdgeType,
				EdgeOrder:      edge.EdgeOrder,
			},
		},
	}
}

// createSequenceWorkflowDirect creates a workflow with 3 sequential actions.
func createSequenceWorkflowDirect(t *testing.T, wfBus *workflow.Business, sd SaveSeedData, alertTemplateID uuid.UUID) *workflowsaveapp.SaveWorkflowResponse {
	t.Helper()
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data for sequence workflow")
	}

	// Use on_update trigger type if available
	triggerTypeID := sd.TriggerTypes[0].ID
	if len(sd.TriggerTypes) > 1 {
		triggerTypeID = sd.TriggerTypes[1].ID // on_update
	}

	// Create rule
	rule, err := wfBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Sequence Test Workflow",
		Description:   "A workflow with 3 sequential actions",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: triggerTypeID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating sequence workflow rule: %v", err)
	}

	// Create 3 actions
	actions := make([]workflow.RuleAction, 3)
	for i := 0; i < 3; i++ {
		stepNum := string(rune('1' + i))
		action, err := wfBus.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: rule.ID,
			Name:             "Sequence Action " + stepNum,
			Description:      "Step " + stepNum + " of sequence",
			ActionConfig:     json.RawMessage(`{"alert_type":"sequence_step","severity":"info","title":"Step ` + stepNum + `","message":"Sequence step ` + stepNum + `"}`),
			ExecutionOrder:   i + 1,
			IsActive:         true,
			TemplateID:       &alertTemplateID,
		})
		if err != nil {
			t.Fatalf("creating sequence workflow action %d: %v", i, err)
		}
		actions[i] = action
	}

	// Create edges: start -> action[0] -> action[1] -> action[2]
	edges := make([]workflow.ActionEdge, 3)

	// Start edge
	startEdge, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: actions[0].ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating sequence start edge: %v", err)
	}
	edges[0] = startEdge

	// Sequence edges
	for i := 0; i < 2; i++ {
		sourceID := actions[i].ID
		edge, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
			RuleID:         rule.ID,
			SourceActionID: &sourceID,
			TargetActionID: actions[i+1].ID,
			EdgeType:       "sequence",
			EdgeOrder:      i + 1,
		})
		if err != nil {
			t.Fatalf("creating sequence edge %d: %v", i, err)
		}
		edges[i+1] = edge
	}

	// Build response
	actionResponses := make([]workflowsaveapp.SaveActionResponse, len(actions))
	for i, a := range actions {
		actionResponses[i] = workflowsaveapp.SaveActionResponse{
			ID:             a.ID.String(),
			Name:           a.Name,
			Description:    a.Description,
			ActionType:     "create_alert",
			ActionConfig:   a.ActionConfig,
			ExecutionOrder: a.ExecutionOrder,
			IsActive:       a.IsActive,
		}
	}

	edgeResponses := make([]workflowsaveapp.SaveEdgeResponse, len(edges))
	for i, e := range edges {
		sourceID := ""
		if e.SourceActionID != nil {
			sourceID = e.SourceActionID.String()
		}
		edgeResponses[i] = workflowsaveapp.SaveEdgeResponse{
			ID:             e.ID.String(),
			SourceActionID: sourceID,
			TargetActionID: e.TargetActionID.String(),
			EdgeType:       e.EdgeType,
			EdgeOrder:      e.EdgeOrder,
		}
	}

	return &workflowsaveapp.SaveWorkflowResponse{
		ID:            rule.ID.String(),
		Name:          rule.Name,
		Description:   rule.Description,
		IsActive:      rule.IsActive,
		EntityID:      rule.EntityID.String(),
		TriggerTypeID: rule.TriggerTypeID.String(),
		Actions:       actionResponses,
		Edges:         edgeResponses,
	}
}

// createBranchingWorkflowDirect creates a workflow with evaluate_condition branching.
func createBranchingWorkflowDirect(t *testing.T, wfBus *workflow.Business, sd SaveSeedData, alertTemplateID, conditionTemplateID uuid.UUID) *workflowsaveapp.SaveWorkflowResponse {
	t.Helper()
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data for branching workflow")
	}

	// Create rule
	rule, err := wfBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Branching Test Workflow",
		Description:   "A workflow with conditional branching",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating branching workflow rule: %v", err)
	}

	// Create condition action
	conditionAction, err := wfBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Evaluate Amount",
		Description:      "Evaluates if amount > 1000",
		ActionConfig:     json.RawMessage(`{"conditions":[{"field":"amount","operator":"greater_than","value":1000}]}`),
		ExecutionOrder:   1,
		IsActive:         true,
		TemplateID:       &conditionTemplateID,
	})
	if err != nil {
		t.Fatalf("creating branching condition action: %v", err)
	}

	// Create true branch action (alert for high value)
	trueBranchAction, err := wfBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "High Value Alert",
		Description:      "Alert for high value items",
		ActionConfig:     json.RawMessage(`{"alert_type":"high_value","severity":"warning","title":"High Value Alert","message":"Amount exceeds threshold"}`),
		ExecutionOrder:   2,
		IsActive:         true,
		TemplateID:       &alertTemplateID,
	})
	if err != nil {
		t.Fatalf("creating true branch action: %v", err)
	}

	// Create false branch action (alert for normal value)
	falseBranchAction, err := wfBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Normal Value Alert",
		Description:      "Alert for normal value items",
		ActionConfig:     json.RawMessage(`{"alert_type":"normal_value","severity":"info","title":"Normal Value Alert","message":"Standard processing"}`),
		ExecutionOrder:   2,
		IsActive:         true,
		TemplateID:       &alertTemplateID,
	})
	if err != nil {
		t.Fatalf("creating false branch action: %v", err)
	}

	// Create edges
	// Start edge to condition
	startEdge, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: conditionAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating branching start edge: %v", err)
	}

	// True branch edge
	conditionID := conditionAction.ID
	trueBranchEdge, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &conditionID,
		TargetActionID: trueBranchAction.ID,
		EdgeType:       "true_branch",
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating true branch edge: %v", err)
	}

	// False branch edge
	falseBranchEdge, err := wfBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &conditionID,
		TargetActionID: falseBranchAction.ID,
		EdgeType:       "false_branch",
		EdgeOrder:      2,
	})
	if err != nil {
		t.Fatalf("creating false branch edge: %v", err)
	}

	return &workflowsaveapp.SaveWorkflowResponse{
		ID:            rule.ID.String(),
		Name:          rule.Name,
		Description:   rule.Description,
		IsActive:      rule.IsActive,
		EntityID:      rule.EntityID.String(),
		TriggerTypeID: rule.TriggerTypeID.String(),
		Actions: []workflowsaveapp.SaveActionResponse{
			{
				ID:             conditionAction.ID.String(),
				Name:           conditionAction.Name,
				Description:    conditionAction.Description,
				ActionType:     "evaluate_condition",
				ActionConfig:   conditionAction.ActionConfig,
				ExecutionOrder: conditionAction.ExecutionOrder,
				IsActive:       conditionAction.IsActive,
			},
			{
				ID:             trueBranchAction.ID.String(),
				Name:           trueBranchAction.Name,
				Description:    trueBranchAction.Description,
				ActionType:     "create_alert",
				ActionConfig:   trueBranchAction.ActionConfig,
				ExecutionOrder: trueBranchAction.ExecutionOrder,
				IsActive:       trueBranchAction.IsActive,
			},
			{
				ID:             falseBranchAction.ID.String(),
				Name:           falseBranchAction.Name,
				Description:    falseBranchAction.Description,
				ActionType:     "create_alert",
				ActionConfig:   falseBranchAction.ActionConfig,
				ExecutionOrder: falseBranchAction.ExecutionOrder,
				IsActive:       falseBranchAction.IsActive,
			},
		},
		Edges: []workflowsaveapp.SaveEdgeResponse{
			{
				ID:             startEdge.ID.String(),
				SourceActionID: "",
				TargetActionID: conditionAction.ID.String(),
				EdgeType:       startEdge.EdgeType,
				EdgeOrder:      startEdge.EdgeOrder,
			},
			{
				ID:             trueBranchEdge.ID.String(),
				SourceActionID: conditionAction.ID.String(),
				TargetActionID: trueBranchAction.ID.String(),
				EdgeType:       trueBranchEdge.EdgeType,
				EdgeOrder:      trueBranchEdge.EdgeOrder,
			},
			{
				ID:             falseBranchEdge.ID.String(),
				SourceActionID: conditionAction.ID.String(),
				TargetActionID: falseBranchAction.ID.String(),
				EdgeType:       falseBranchEdge.EdgeType,
				EdgeOrder:      falseBranchEdge.EdgeOrder,
			},
		},
	}
}

// waitForProcessing waits for at least one event to be processed by the queue manager.
func waitForProcessing(t *testing.T, qm *workflow.QueueManager, initial workflow.QueueMetrics, timeout time.Duration) bool {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			metrics := qm.GetMetrics()
			t.Logf("Timeout waiting for processing - Enqueued: %d, Processed: %d, Failed: %d",
				metrics.TotalEnqueued, metrics.TotalProcessed, metrics.TotalFailed)
			return false
		case <-ticker.C:
			metrics := qm.GetMetrics()
			if metrics.TotalProcessed > initial.TotalProcessed ||
				metrics.TotalFailed > initial.TotalFailed {
				return true
			}
		}
	}
}

// createTriggerEvent creates a TriggerEvent for testing workflow execution.
func createTriggerEvent(entityName string, eventType string, userID uuid.UUID, rawData map[string]any) workflow.TriggerEvent {
	return workflow.TriggerEvent{
		EventType:  eventType,
		EntityName: entityName,
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    rawData,
		UserID:     userID,
	}
}
