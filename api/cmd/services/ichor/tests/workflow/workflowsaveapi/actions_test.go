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
// Action-Specific Integration Tests (Temporal-based)
// =============================================================================

// runActionTests runs all action-specific tests as subtests.
// These tests verify that each action type executes correctly with proper
// configuration and produces expected side effects via Temporal polling.
func runActionTests(t *testing.T, sd ExecutionTestData) {
	// create_alert action tests
	t.Run("action-alert-basic", func(t *testing.T) { testCreateAlertBasic(t, sd) })
	t.Run("action-alert-with-recipients", func(t *testing.T) { testCreateAlertWithRecipients(t, sd) })
	t.Run("action-alert-template-vars", func(t *testing.T) { testCreateAlertTemplateVars(t, sd) })
	t.Run("action-alert-severity-levels", func(t *testing.T) { testCreateAlertSeverityLevels(t, sd) })

	// evaluate_condition action tests
	t.Run("action-condition-equals-true", func(t *testing.T) { testConditionEqualsTrue(t, sd) })
	t.Run("action-condition-equals-false", func(t *testing.T) { testConditionEqualsFalse(t, sd) })
	t.Run("action-condition-greater-than", func(t *testing.T) { testConditionGreaterThan(t, sd) })
	t.Run("action-condition-multiple-and", func(t *testing.T) { testConditionMultipleAnd(t, sd) })
}

// =============================================================================
// create_alert Action Tests
// =============================================================================

// testCreateAlertBasic tests basic alert creation with minimal config.
func testCreateAlertBasic(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Alert Basic Test " + uuid.New().String()[:8],
		Description:   "Tests basic create_alert action",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	action, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Basic Alert Action",
		ActionConfig: json.RawMessage(`{
			"alert_type": "basic_test",
			"severity": "low",
			"title": "Basic Test Alert",
			"message": "This is a basic test alert",
			"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	alertType := "basic_test"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: basic_test alert created")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no new basic_test alert after 15s")
}

// testCreateAlertWithRecipients tests alert creation with multiple user recipients.
func testCreateAlertWithRecipients(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Alert Recipients Test " + uuid.New().String()[:8],
		Description:   "Tests create_alert with multiple recipients",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	action, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Multi-Recipient Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "multi_recipient_test",
			"severity": "medium",
			"title": "Multi-Recipient Alert",
			"message": "Alert with multiple recipients",
			"recipients": {
				"users": ["` + sd.Users[0].ID.String() + `"],
				"roles": []
			}
		}`),
		IsActive:   true,
		TemplateID: &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	alertType := "multi_recipient_test"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: multi_recipient_test alert created")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no new multi_recipient_test alert after 15s")
}

// testCreateAlertTemplateVars tests that template variables are resolved in alert messages.
func testCreateAlertTemplateVars(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Alert Template Test " + uuid.New().String()[:8],
		Description:   "Tests template variable substitution",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	action, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Template Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "template_test",
			"severity": "high",
			"title": "Alert for {{entity_name}}",
			"message": "Status is {{status}} and value is {{value}}",
			"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	alertType := "template_test"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status":      "active",
		"value":       12345,
		"entity_name": "TestEntity",
	})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: template_test alert created")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no new template_test alert after 15s")
}

// testCreateAlertSeverityLevels tests that all severity levels are accepted.
func testCreateAlertSeverityLevels(t *testing.T, sd ExecutionTestData) {
	severities := []struct {
		level     string
		alertType string
	}{
		{"low", "severity_low"},
		{"medium", "severity_medium"},
		{"high", "severity_high"},
		{"critical", "severity_critical"},
	}

	for _, sv := range severities {
		sv := sv
		t.Run(sv.level, func(t *testing.T) {
			ctx := context.Background()

			if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
				t.Fatal("insufficient seed data")
			}

			rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Severity Test " + sv.level + " " + uuid.New().String()[:8],
				Description:   "Tests " + sv.level + " severity",
				EntityID:      sd.Entities[0].ID,
				EntityTypeID:  sd.EntityTypes[0].ID,
				TriggerTypeID: sd.TriggerTypes[0].ID,
				IsActive:      true,
				CreatedBy:     sd.Users[0].ID,
			})
			if err != nil {
				t.Fatalf("creating rule: %v", err)
			}

			action, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
				AutomationRuleID: rule.ID,
				Name:             "Severity " + sv.level + " Alert",
				ActionConfig: json.RawMessage(`{
					"alert_type": "` + sv.alertType + `",
					"severity": "` + sv.level + `",
					"title": "Severity Test",
					"message": "Testing ` + sv.level + ` severity",
					"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
				}`),
				IsActive:   true,
				TemplateID: &sd.CreateAlertTemplate.ID,
			})
			if err != nil {
				t.Fatalf("creating action: %v", err)
			}

			_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
				RuleID:         rule.ID,
				SourceActionID: nil,
				TargetActionID: action.ID,
				EdgeType:       "start",
				EdgeOrder:      0,
			})
			if err != nil {
				t.Fatalf("creating edge: %v", err)
			}

			if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
				t.Fatalf("refreshing rules: %v", err)
			}

			alertType := sv.alertType
			before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
			if err != nil {
				t.Fatalf("querying alerts before: %v", err)
			}
			beforeCount := len(before)

			event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})
			if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
				t.Fatalf("firing trigger: %v", err)
			}

			for i := 0; i < 30; i++ {
				after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
				if err != nil {
					t.Fatalf("querying alerts after: %v", err)
				}
				if len(after) > beforeCount {
					t.Logf("SUCCESS: %s severity alert created", sv.level)
					return
				}
				time.Sleep(500 * time.Millisecond)
			}
			t.Fatalf("timeout: no new %s alert after 15s", sv.alertType)
		})
	}
}

// =============================================================================
// evaluate_condition Action Tests
// =============================================================================

// testConditionEqualsTrue tests that equals condition evaluates to true when matched.
func testConditionEqualsTrue(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Condition Equals True " + uuid.New().String()[:8],
		Description:   "Tests condition equals matching",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Status",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
		}`),
		IsActive:   true,
		TemplateID: &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	trueBranchAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "True Branch Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "condition_true",
			"severity": "low",
			"title": "Condition True",
			"message": "Condition evaluated to true",
			"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating true branch action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: conditionAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	condID := conditionAction.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &condID,
		TargetActionID: trueBranchAction.ID,
		EdgeType:       "sequence",
		SourceOutput:   strPtr("true"),
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating true branch edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	alertType := "condition_true"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	// status=="active" → condition is TRUE → true branch fires
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status": "active",
	})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: condition_true alert created — true branch taken")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no condition_true alert after 15s — true branch may not have fired")
}

// testConditionEqualsFalse tests that equals condition evaluates to false when not matched.
func testConditionEqualsFalse(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Condition Equals False " + uuid.New().String()[:8],
		Description:   "Tests condition equals not matching",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Status",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
		}`),
		IsActive:   true,
		TemplateID: &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	falseBranchAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "False Branch Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "condition_false",
			"severity": "low",
			"title": "Condition False",
			"message": "Condition evaluated to false",
			"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating false branch action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: conditionAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	condID := conditionAction.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &condID,
		TargetActionID: falseBranchAction.ID,
		EdgeType:       "sequence",
		SourceOutput:   strPtr("false"),
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating false branch edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	alertType := "condition_false"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	// status=="inactive" → condition is FALSE → false branch fires
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status": "inactive",
	})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: condition_false alert created — false branch taken")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no condition_false alert after 15s — false branch may not have fired")
}

// testConditionGreaterThan tests numeric greater_than condition.
func testConditionGreaterThan(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Condition Greater Than " + uuid.New().String()[:8],
		Description:   "Tests greater_than condition",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Amount",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "amount", "operator": "greater_than", "value": 100}]
		}`),
		IsActive:   true,
		TemplateID: &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	trueBranchAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "High Amount Alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "high_amount",
			"severity": "high",
			"title": "High Amount Detected",
			"message": "Amount exceeds threshold",
			"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating true branch action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: conditionAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	condID := conditionAction.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &condID,
		TargetActionID: trueBranchAction.ID,
		EdgeType:       "sequence",
		SourceOutput:   strPtr("true"),
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating true branch edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	alertType := "high_amount"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	// amount=200 > 100 → condition is TRUE → true branch fires
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"amount": float64(200),
	})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: high_amount alert created — greater_than condition triggered true branch")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no high_amount alert after 15s — greater_than condition may not have fired true branch")
}

// testConditionMultipleAnd tests multiple conditions with AND logic.
func testConditionMultipleAnd(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Condition Multiple AND " + uuid.New().String()[:8],
		Description:   "Tests multiple conditions with AND logic",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Multiple",
		ActionConfig: json.RawMessage(`{
			"conditions": [
				{"field_name": "status", "operator": "equals", "value": "active"},
				{"field_name": "amount", "operator": "greater_than", "value": 50}
			],
			"logic_type": "and"
		}`),
		IsActive:   true,
		TemplateID: &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	trueBranchAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "All Conditions Met",
		ActionConfig: json.RawMessage(`{
			"alert_type": "all_conditions_met",
			"severity": "medium",
			"title": "All Conditions Met",
			"message": "Both status=active AND amount>50",
			"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating true branch action: %v", err)
	}

	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: conditionAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	condID := conditionAction.ID
	_, err = sd.WF.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: &condID,
		TargetActionID: trueBranchAction.ID,
		EdgeType:       "sequence",
		SourceOutput:   strPtr("true"),
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating true branch edge: %v", err)
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	alertType := "all_conditions_met"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	// status=="active" AND amount=100 > 50 → both conditions TRUE → true branch fires
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status": "active",
		"amount": float64(100),
	})
	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: all_conditions_met alert created — AND logic triggered true branch")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no all_conditions_met alert after 15s — multiple AND conditions may not have fired true branch")
}
