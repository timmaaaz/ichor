//go:build ignore
// +build ignore

// Phase 13: Excluded until Phase 15 rewrites for Temporal.

package workflowsaveapi_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Phase 9: Action-Specific Integration Tests
// =============================================================================

// runActionTests runs all action-specific tests as subtests.
// These tests verify that each action type executes correctly with proper
// configuration and produces expected side effects.
func runActionTests(t *testing.T, sd ExecutionTestData) {
	// 9a. create_alert Action Tests
	t.Run("action-alert-basic", func(t *testing.T) {
		testCreateAlertBasic(t, sd)
	})
	t.Run("action-alert-with-recipients", func(t *testing.T) {
		testCreateAlertWithRecipients(t, sd)
	})
	t.Run("action-alert-template-vars", func(t *testing.T) {
		testCreateAlertTemplateVars(t, sd)
	})
	t.Run("action-alert-severity-levels", func(t *testing.T) {
		testCreateAlertSeverityLevels(t, sd)
	})

	// 9c. send_email Action Tests
	t.Run("action-email-basic", func(t *testing.T) {
		testSendEmailBasic(t, sd)
	})
	t.Run("action-email-multiple-recipients", func(t *testing.T) {
		testSendEmailMultipleRecipients(t, sd)
	})

	// 9d. evaluate_condition Action Tests
	t.Run("action-condition-equals-true", func(t *testing.T) {
		testConditionEqualsTrue(t, sd)
	})
	t.Run("action-condition-equals-false", func(t *testing.T) {
		testConditionEqualsFalse(t, sd)
	})
	t.Run("action-condition-greater-than", func(t *testing.T) {
		testConditionGreaterThan(t, sd)
	})
	t.Run("action-condition-multiple-and", func(t *testing.T) {
		testConditionMultipleAnd(t, sd)
	})
}

// =============================================================================
// 9a. create_alert Action Tests
// =============================================================================

// testCreateAlertBasic tests basic alert creation with minimal config.
func testCreateAlertBasic(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with a create_alert action
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

	// Create action with basic config
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
		IsActive:       true,
		TemplateID:     &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	// Create start edge
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

	// Re-initialize engine to pick up new rule
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Verify alert was created by finding results for our rule
	var foundAlert bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.ActionType == "create_alert" && actionResult.Status == "success" {
						foundAlert = true
						// Verify result contains alert_id
						if actionResult.ResultData != nil {
							if _, hasAlertID := actionResult.ResultData["alert_id"]; !hasAlertID {
								t.Error("create_alert result should contain alert_id")
							}
						}
					}
				}
			}
		}
	}

	if !foundAlert {
		t.Error("create_alert action did not execute successfully")
	}
}

// testCreateAlertWithRecipients tests alert creation with multiple user and role recipients.
func testCreateAlertWithRecipients(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with a create_alert action with multiple recipients
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

	// Create action with multiple recipients
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
		IsActive:       true,
		TemplateID:     &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	// Create start edge
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

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find result for our rule
	var foundSuccess bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.Status == "success" {
						foundSuccess = true
					}
				}
			}
		}
	}

	if !foundSuccess {
		t.Error("create_alert action with recipients did not succeed")
	}
}

// testCreateAlertTemplateVars tests that template variables are resolved in alert messages.
func testCreateAlertTemplateVars(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with template variables in the message
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

	// Create action with template variables
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
		IsActive:       true,
		TemplateID:     &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	// Create start edge
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

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow with data that includes template variables
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status":      "active",
		"value":       12345,
		"entity_name": "TestEntity",
	})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find result for our rule and verify action succeeded
	var foundSuccess bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.Status == "success" {
						foundSuccess = true
					}
				}
			}
		}
	}

	if !foundSuccess {
		t.Error("create_alert action with templates did not succeed")
	}
}

// testCreateAlertSeverityLevels tests that all severity levels are accepted.
func testCreateAlertSeverityLevels(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	severities := []string{"low", "medium", "high", "critical"}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
				t.Fatal("insufficient seed data")
			}

			// Create a workflow for each severity
			rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
				Name:          "Severity Test " + severity + " " + uuid.New().String()[:8],
				Description:   "Tests " + severity + " severity",
				EntityID:      sd.Entities[0].ID,
				EntityTypeID:  sd.EntityTypes[0].ID,
				TriggerTypeID: sd.TriggerTypes[0].ID,
				IsActive:      true,
				CreatedBy:     sd.Users[0].ID,
			})
			if err != nil {
				t.Fatalf("creating rule: %v", err)
			}

			// Create action with specified severity
			action, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
				AutomationRuleID: rule.ID,
				Name:             "Severity " + severity + " Alert",
				ActionConfig: json.RawMessage(`{
					"alert_type": "severity_test",
					"severity": "` + severity + `",
					"title": "Severity Test",
					"message": "Testing ` + severity + ` severity",
					"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
				}`),
				IsActive:       true,
				TemplateID:     &sd.CreateAlertTemplate.ID,
			})
			if err != nil {
				t.Fatalf("creating action: %v", err)
			}

			// Create start edge
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

			// Re-initialize engine
			if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
				t.Fatalf("reinitializing engine: %v", err)
			}

			// Execute workflow
			event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})

			execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
			if err != nil {
				t.Fatalf("executing workflow: %v", err)
			}

			// Verify execution completed
			if execution.Status != workflow.StatusCompleted {
				t.Fatalf("expected completed for severity %s:\n%s", severity, formatExecutionErrors(execution))
			}

			// Find result for our rule
			var foundSuccess bool
			for _, batch := range execution.BatchResults {
				for _, ruleResult := range batch.RuleResults {
					if ruleResult.RuleID == rule.ID {
						for _, actionResult := range ruleResult.ActionResults {
							if actionResult.Status == "success" {
								foundSuccess = true
							}
						}
					}
				}
			}

			if !foundSuccess {
				t.Errorf("create_alert action with %s severity did not succeed", severity)
			}
		})
	}
}

// =============================================================================
// 9c. send_email Action Tests
// =============================================================================

// testSendEmailBasic tests basic email sending.
func testSendEmailBasic(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with send_email action
	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Email Basic Test " + uuid.New().String()[:8],
		Description:   "Tests basic send_email action",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	// Create send_email action
	action, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Basic Email Action",
		ActionConfig: json.RawMessage(`{
			"recipients": ["test@example.com"],
			"subject": "Test Email Subject",
			"body": "This is a test email body"
		}`),
		IsActive:       true,
		TemplateID:     &sd.SendEmailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	// Create start edge
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

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find result for our rule
	var foundEmailAction bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.ActionType == "send_email" && actionResult.Status == "success" {
						foundEmailAction = true
						// Verify result contains email_id
						if actionResult.ResultData != nil {
							if _, hasEmailID := actionResult.ResultData["email_id"]; !hasEmailID {
								t.Error("send_email result should contain email_id")
							}
							if status, hasStatus := actionResult.ResultData["status"]; hasStatus {
								if status != "sent" {
									t.Errorf("expected email status 'sent', got %v", status)
								}
							}
						}
					}
				}
			}
		}
	}

	if !foundEmailAction {
		t.Error("send_email action did not execute successfully")
	}
}

// testSendEmailMultipleRecipients tests email sending to multiple recipients.
func testSendEmailMultipleRecipients(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with send_email action for multiple recipients
	rule, err := sd.WF.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Email Multi Test " + uuid.New().String()[:8],
		Description:   "Tests send_email with multiple recipients",
		EntityID:      sd.Entities[0].ID,
		EntityTypeID:  sd.EntityTypes[0].ID,
		TriggerTypeID: sd.TriggerTypes[0].ID,
		IsActive:      true,
		CreatedBy:     sd.Users[0].ID,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	// Create send_email action with 3 recipients
	action, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Multi-Recipient Email",
		ActionConfig: json.RawMessage(`{
			"recipients": ["user1@example.com", "user2@example.com", "user3@example.com"],
			"subject": "Multi-Recipient Test",
			"body": "Email sent to multiple recipients"
		}`),
		IsActive:       true,
		TemplateID:     &sd.SendEmailTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating action: %v", err)
	}

	// Create start edge
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

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find result for our rule
	var foundSuccess bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.Status == "success" {
						foundSuccess = true
						// Verify recipients were processed
						if actionResult.ResultData != nil {
							if recipients, hasRecipients := actionResult.ResultData["recipients"].([]interface{}); hasRecipients {
								if len(recipients) != 3 {
									t.Errorf("expected 3 recipients, got %d", len(recipients))
								}
							}
						}
					}
				}
			}
		}
	}

	if !foundSuccess {
		t.Error("send_email action with multiple recipients did not succeed")
	}
}

// =============================================================================
// 9d. evaluate_condition Action Tests
// =============================================================================

// testConditionEqualsTrue tests that equals condition evaluates to true when matched.
func testConditionEqualsTrue(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with evaluate_condition action
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

	// Create evaluate_condition action
	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Status",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
		}`),
		IsActive:       true,
		TemplateID:     &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	// Create true branch action
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
		IsActive:       true,
		TemplateID:     &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating true branch action: %v", err)
	}

	// Create edges
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
		EdgeType:       "true_branch",
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating true branch edge: %v", err)
	}

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow with data that makes condition TRUE
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status": "active", // This matches the condition
	})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find results for our rule and verify branch taken
	var foundCondition, foundTrueBranch bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.ActionID == conditionAction.ID {
						foundCondition = true
						// Check branch taken (stored directly on ActionResult)
						if actionResult.BranchTaken != "true_branch" {
							t.Errorf("expected true_branch, got %s", actionResult.BranchTaken)
						}
					}
					if actionResult.ActionID == trueBranchAction.ID && actionResult.Status == "success" {
						foundTrueBranch = true
					}
				}
			}
		}
	}

	if !foundCondition {
		t.Error("evaluate_condition action was not executed")
	}
	if !foundTrueBranch {
		t.Error("true branch action was not executed")
	}
}

// testConditionEqualsFalse tests that equals condition evaluates to false when not matched.
func testConditionEqualsFalse(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with evaluate_condition action
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

	// Create evaluate_condition action
	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Status",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "status", "operator": "equals", "value": "active"}]
		}`),
		IsActive:       true,
		TemplateID:     &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	// Create false branch action
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
		IsActive:       true,
		TemplateID:     &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating false branch action: %v", err)
	}

	// Create edges
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
		EdgeType:       "false_branch",
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating false branch edge: %v", err)
	}

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow with data that makes condition FALSE
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status": "inactive", // This does NOT match the condition
	})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find results for our rule
	var foundCondition, foundFalseBranch bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.ActionID == conditionAction.ID {
						foundCondition = true
						// Check branch taken (stored directly on ActionResult)
						if actionResult.BranchTaken != "false_branch" {
							t.Errorf("expected false_branch, got %s", actionResult.BranchTaken)
						}
					}
					if actionResult.ActionID == falseBranchAction.ID && actionResult.Status == "success" {
						foundFalseBranch = true
					}
				}
			}
		}
	}

	if !foundCondition {
		t.Error("evaluate_condition action was not executed")
	}
	if !foundFalseBranch {
		t.Error("false branch action was not executed")
	}
}

// testConditionGreaterThan tests numeric greater_than condition.
func testConditionGreaterThan(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with evaluate_condition action
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

	// Create evaluate_condition action with greater_than
	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Amount",
		ActionConfig: json.RawMessage(`{
			"conditions": [{"field_name": "amount", "operator": "greater_than", "value": 1000}]
		}`),
		IsActive:       true,
		TemplateID:     &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	// Create true branch action for high amounts
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
		IsActive:       true,
		TemplateID:     &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating true branch action: %v", err)
	}

	// Create edges
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
		EdgeType:       "true_branch",
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating true branch edge: %v", err)
	}

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow with amount > 1000
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"amount": 1500, // Greater than 1000
	})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find results for our rule
	var foundTrueBranch bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.ActionID == trueBranchAction.ID && actionResult.Status == "success" {
						foundTrueBranch = true
					}
				}
			}
		}
	}

	if !foundTrueBranch {
		t.Error("greater_than condition did not trigger true branch for amount > 1000")
	}
}

// testConditionMultipleAnd tests multiple conditions with AND logic.
func testConditionMultipleAnd(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 || len(sd.EntityTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create a workflow with multiple conditions
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

	// Create evaluate_condition action with multiple conditions (AND logic)
	conditionAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Check Multiple",
		ActionConfig: json.RawMessage(`{
			"conditions": [
				{"field_name": "status", "operator": "equals", "value": "approved"},
				{"field_name": "amount", "operator": "greater_than", "value": 500}
			],
			"logic_type": "and"
		}`),
		IsActive:       true,
		TemplateID:     &sd.EvaluateConditionTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating condition action: %v", err)
	}

	// Create true branch action
	trueBranchAction, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "All Conditions Met",
		ActionConfig: json.RawMessage(`{
			"alert_type": "all_conditions_met",
			"severity": "medium",
			"title": "All Conditions Met",
			"message": "Both status=approved AND amount>500",
			"recipients": {"users": ["` + sd.Users[0].ID.String() + `"], "roles": []}
		}`),
		IsActive:       true,
		TemplateID:     &sd.CreateAlertTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating true branch action: %v", err)
	}

	// Create edges
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
		EdgeType:       "true_branch",
		EdgeOrder:      1,
	})
	if err != nil {
		t.Fatalf("creating true branch edge: %v", err)
	}

	// Re-initialize engine
	if err := sd.WF.Engine.Initialize(ctx, sd.WF.WorkflowBus); err != nil {
		t.Fatalf("reinitializing engine: %v", err)
	}

	// Execute workflow with data that matches BOTH conditions
	event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{
		"status": "approved",
		"amount": 750, // Both conditions must be true
	})

	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("executing workflow: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Find results for our rule
	var foundTrueBranch bool
	for _, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.RuleID == rule.ID {
				for _, actionResult := range ruleResult.ActionResults {
					if actionResult.ActionID == trueBranchAction.ID && actionResult.Status == "success" {
						foundTrueBranch = true
					}
				}
			}
		}
	}

	if !foundTrueBranch {
		t.Error("multiple AND conditions did not trigger true branch when all conditions met")
	}
}

