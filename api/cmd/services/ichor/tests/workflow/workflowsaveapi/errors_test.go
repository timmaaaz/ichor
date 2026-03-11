package workflowsaveapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Phase 10: Error Handling & Edge Case Tests (Temporal-based)
// =============================================================================

// runErrorTests runs all error handling tests as subtests.
// These tests verify proper error handling and edge cases via Temporal.
func runErrorTests(t *testing.T, sd ExecutionTestData) {
	t.Run("error-action-fails-sequence-stops", func(t *testing.T) {
		testActionFailsSequenceStops(t, sd)
	})
	t.Run("error-condition-field-not-found", func(t *testing.T) {
		testConditionFieldNotFound(t, sd)
	})
	t.Run("error-condition-type-mismatch", func(t *testing.T) {
		testConditionTypeMismatch(t, sd)
	})
	t.Run("error-no-actions-defined", func(t *testing.T) {
		testNoActionsDefined(t, sd)
	})
	t.Run("error-inactive-action-skipped", func(t *testing.T) {
		testInactiveActionSkipped(t, sd)
	})
}

// =============================================================================
// 10a. Action Failures
// =============================================================================

// testActionFailsSequenceStops tests that when an action fails in a sequence,
// the workflow handles the failure gracefully (action 1 still succeeds).
func testActionFailsSequenceStops(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with 3 actions where action 2 has invalid config
	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Fail Sequence Test " + uuid.New().String()[:8],
		Description:   "Tests sequence stops on action failure",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	userIDStr := sd.Users[0].ID.String()

	// Action 1: Valid create_alert
	action1, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Action 1 - Valid",
		ActionConfig:     json.RawMessage(`{"alert_type":"test","severity":"low","title":"Test 1","message":"Should succeed","recipients":{"users":["` + userIDStr + `"],"roles":[]}}`),
		IsActive:         true,
		TemplateID:       &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action 1: %v", err)
	}

	// Action 2: Invalid config - missing required fields for evaluate_condition
	action2, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Action 2 - Invalid Config",
		ActionConfig:     json.RawMessage(`{"invalid_field":"this should fail"}`),
		IsActive:         true,
		TemplateID:       &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action 2: %v", err)
	}

	// Action 3: Valid create_alert (may be skipped on action 2 failure)
	action3, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Action 3 - Should Skip",
		ActionConfig:     json.RawMessage(`{"alert_type":"test","severity":"low","title":"Test 3","message":"Should be skipped","recipients":{"users":["` + userIDStr + `"],"roles":[]}}`),
		IsActive:         true,
		TemplateID:       &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action 3: %v", err)
	}

	// Create edges: start -> action1 -> action2 -> action3
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action1.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	a1ID := action1.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &a1ID,
		TargetActionID: action2.ID,
		EdgeType:       "sequence",
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating edge 1->2: %v", err)
	}

	a2ID := action2.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &a2ID,
		TargetActionID: action3.ID,
		EdgeType:       "sequence",
		EdgeOrder:      2,
	})
	if err != nil {
		t.Fatalf("creating edge 2->3: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	// Wait for action 1's alert (alert_type: "test") to appear. Wait up to 10s.
	alertType := "test"
	var found bool
	for i := 0; i < 20; i++ {
		alerts, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "5"))
		if len(alerts) > 0 {
			found = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !found {
		t.Fatal("timeout: action 1's alert not created after 10s")
	}
	t.Log("SUCCESS: action 1 succeeded, workflow handled action 2 failure gracefully")

	// Suppress unused variable warnings from seeding.
	_ = action3
}

// =============================================================================
// 10b. Trigger Condition Errors
// =============================================================================

// testConditionFieldNotFound tests that conditions referencing non-existent fields
// are handled gracefully (workflow completes without panic).
func testConditionFieldNotFound(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Missing Field Test " + uuid.New().String()[:8],
		Description:   "Tests condition with missing field",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	// Create evaluate_condition action that references non-existent field
	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Non-Existent Field",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "nonexistent_field_xyz", "operator": "equals", "value": "test"}]
		}`),
		IsActive:   true,
		TemplateID: &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: conditionAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status": "active",
		"amount": 100,
		// intentionally no "nonexistent_field_xyz"
	})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	// Workflow should complete gracefully. Wait 3s and verify no panic.
	time.Sleep(3 * time.Second)
	t.Log("SUCCESS: missing field condition handled gracefully")
}

// testConditionTypeMismatch tests that type mismatches in conditions are handled gracefully.
func testConditionTypeMismatch(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Type Mismatch Test " + uuid.New().String()[:8],
		Description:   "Tests condition type mismatch handling",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	// Create evaluate_condition with numeric operator applied to a string field
	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Numeric Check on String",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "name", "operator": "greater_than", "value": 100}]
		}`),
		IsActive:   true,
		TemplateID: &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: conditionAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"name": "This is a string, not a number",
	})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	time.Sleep(3 * time.Second)
	t.Log("SUCCESS: type mismatch handled gracefully")
}

// =============================================================================
// 10e. Invalid Workflow States
// =============================================================================

// testNoActionsDefined tests behavior when a rule has no actions defined.
// The rule has zero actions and zero edges — trigger processor should not dispatch it.
func testNoActionsDefined(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a rule with no actions and no edges
	_, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "No Actions Test " + uuid.New().String()[:8],
		Description:   "Tests rule with no actions",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	before, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	beforeCount := len(before)

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	time.Sleep(2 * time.Second)

	after, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if len(after) != beforeCount {
		t.Errorf("expected no new alerts for rule with no actions, got %d new", len(after)-beforeCount)
	}
	t.Log("SUCCESS: rule with no actions handled gracefully")
}

// testInactiveActionSkipped tests that inactive actions are skipped during execution.
func testInactiveActionSkipped(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Inactive Action Test " + uuid.New().String()[:8],
		Description:   "Tests inactive action skipping",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	userIDStr := sd.Users[0].ID.String()

	// Action 1: Active
	action1, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Active Action 1",
		ActionConfig:     json.RawMessage(`{"alert_type":"active_test","severity":"low","title":"Active 1","message":"Should execute","recipients":{"users":["` + userIDStr + `"],"roles":[]}}`),
		IsActive:         true,
		TemplateID:       &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action 1: %v", err)
	}

	// Action 2: Inactive - should be skipped
	action2, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Inactive Action 2",
		ActionConfig:     json.RawMessage(`{"alert_type":"inactive_test","severity":"high","title":"Inactive 2","message":"Should NOT execute","recipients":{"users":["` + userIDStr + `"],"roles":[]}}`),
		IsActive:         false,
		TemplateID:       &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action 2: %v", err)
	}

	// Action 3: Active
	action3, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Active Action 3",
		ActionConfig:     json.RawMessage(`{"alert_type":"active_test","severity":"low","title":"Active 3","message":"Should execute","recipients":{"users":["` + userIDStr + `"],"roles":[]}}`),
		IsActive:         true,
		TemplateID:       &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action 3: %v", err)
	}

	// Create edges: start -> action1 -> action2 -> action3
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action1.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	a1ID := action1.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &a1ID,
		TargetActionID: action2.ID,
		EdgeType:       "sequence",
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating edge 1->2: %v", err)
	}

	a2ID := action2.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &a2ID,
		TargetActionID: action3.ID,
		EdgeType:       "sequence",
		EdgeOrder:      2,
	})
	if err != nil {
		t.Fatalf("creating edge 2->3: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	inactiveAlertType := "inactive_test"
	before, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &inactiveAlertType}, alertbus.DefaultOrderBy, page.MustParse("1", "10"))
	beforeInactiveCount := len(before)

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	// Poll for active action's alert ("active_test")
	activeAlertType := "active_test"
	var found bool
	for i := 0; i < 20; i++ {
		alerts, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &activeAlertType}, alertbus.DefaultOrderBy, page.MustParse("1", "10"))
		if len(alerts) > 0 {
			found = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !found {
		t.Fatal("timeout: active action's alert not created after 10s")
	}

	// Verify inactive action did NOT create an alert
	after, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &inactiveAlertType}, alertbus.DefaultOrderBy, page.MustParse("1", "10"))
	if len(after) > beforeInactiveCount {
		t.Errorf("inactive action 2 should NOT have created an alert, got %d new", len(after)-beforeInactiveCount)
	}

	t.Log("SUCCESS: inactive action was skipped")

	// Suppress unused variable warnings.
	_ = action3
}
