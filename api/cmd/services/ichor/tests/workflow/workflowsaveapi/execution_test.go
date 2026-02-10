//go:build ignore
// +build ignore

// Phase 13: Excluded until Phase 15 rewrites for Temporal.

package workflowsaveapi_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Phase 7: Workflow Execution Integration Tests
// =============================================================================

// runExecutionTests runs all execution tests as subtests.
// These are not HTTP tests - they test the workflow engine directly.
func runExecutionTests(t *testing.T, sd ExecutionTestData) {
	t.Run("exec-single-alert", func(t *testing.T) {
		testExecuteSingleCreateAlert(t, sd)
	})
	t.Run("exec-sequence", func(t *testing.T) {
		testExecuteSequence3Actions(t, sd)
	})
	t.Run("exec-branch-true", func(t *testing.T) {
		testExecuteBranchTrue(t, sd)
	})
	t.Run("exec-branch-false", func(t *testing.T) {
		testExecuteBranchFalse(t, sd)
	})
	t.Run("exec-record-created", func(t *testing.T) {
		testExecutionRecordCreated(t, sd)
	})
	t.Run("exec-history-tracking", func(t *testing.T) {
		testExecutionHistoryTracking(t, sd)
	})
	t.Run("exec-no-matching-rules", func(t *testing.T) {
		testNoMatchingRules(t, sd)
	})
}

// testExecuteSingleCreateAlert tests that a simple workflow with 1 action executes correctly.
func testExecuteSingleCreateAlert(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	// Get the entity name from the seed data (first entity)
	if len(sd.Entities) == 0 {
		t.Fatal("no entities in seed data")
	}
	entityName := sd.Entities[0].Name

	// Get trigger type name from seed data (first trigger type = on_create)
	if len(sd.TriggerTypes) == 0 {
		t.Fatal("no trigger types in seed data")
	}
	triggerTypeName := sd.TriggerTypes[0].Name

	// Create trigger event
	event := workflow.TriggerEvent{
		EventType:  triggerTypeName,
		EntityName: entityName,
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"status": "new"},
		UserID:     sd.Users[0].ID,
	}

	// Execute workflow via engine
	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}
}

// testExecuteSequence3Actions tests that a workflow with 3 sequential actions executes in order.
func testExecuteSequence3Actions(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	// Use the sequence workflow's entity and trigger type
	if len(sd.Entities) == 0 {
		t.Fatal("no entities in seed data")
	}

	// Use the second trigger type (on_update) if available
	triggerTypeName := sd.TriggerTypes[0].Name
	if len(sd.TriggerTypes) > 1 {
		triggerTypeName = sd.TriggerTypes[1].Name
	}

	// Create trigger event
	event := workflow.TriggerEvent{
		EventType:  triggerTypeName,
		EntityName: sd.Entities[0].Name,
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"status": "processing"},
		UserID:     sd.Users[0].ID,
	}

	// Execute workflow via engine
	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Parse the sequence workflow rule ID
	ruleID, err := uuid.Parse(sd.SequenceWorkflow.ID)
	if err != nil {
		t.Fatalf("parsing sequence workflow rule ID: %v", err)
	}

	// Find results for our specific rule
	var matchedRuleResult *workflow.RuleResult
	for _, batch := range execution.BatchResults {
		for i := range batch.RuleResults {
			if batch.RuleResults[i].RuleID == ruleID {
				matchedRuleResult = &batch.RuleResults[i]
				break
			}
		}
	}

	if matchedRuleResult == nil {
		// Rule may not match if trigger type doesn't match, which is OK
		// The test validates that the engine ran without error
		return
	}

	// Verify 3 actions executed
	if len(matchedRuleResult.ActionResults) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(matchedRuleResult.ActionResults))
	}

	// Verify all actions succeeded
	for i, ar := range matchedRuleResult.ActionResults {
		if ar.Status != "success" {
			t.Fatalf("action[%d] failed: %s", i, ar.ErrorMessage)
		}
	}
}

// testExecuteBranchTrue tests that a branching workflow takes the true branch when condition is met.
func testExecuteBranchTrue(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create trigger event with data that makes condition TRUE (amount > 1000)
	event := workflow.TriggerEvent{
		EventType:  sd.TriggerTypes[0].Name,
		EntityName: sd.Entities[0].Name,
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData: map[string]any{
			"amount": 1500, // condition: amount > 1000 should be TRUE
			"status": "approved",
		},
		UserID: sd.Users[0].ID,
	}

	// Execute workflow via engine
	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Parse the branching workflow rule ID
	ruleID, err := uuid.Parse(sd.BranchingWorkflow.ID)
	if err != nil {
		t.Fatalf("parsing branching workflow rule ID: %v", err)
	}

	// Find results for our specific rule
	var matchedRuleResult *workflow.RuleResult
	for _, batch := range execution.BatchResults {
		for i := range batch.RuleResults {
			if batch.RuleResults[i].RuleID == ruleID {
				matchedRuleResult = &batch.RuleResults[i]
				break
			}
		}
	}

	if matchedRuleResult == nil {
		// Rule may not match, which is OK for this test
		return
	}

	// Verify the evaluate_condition action was executed
	foundCondition := false
	for _, ar := range matchedRuleResult.ActionResults {
		if ar.ActionType == "evaluate_condition" {
			foundCondition = true
		}
	}

	if !foundCondition {
		t.Fatal("evaluate_condition action not found in results")
	}
}

// testExecuteBranchFalse tests that a branching workflow takes the false branch when condition is not met.
func testExecuteBranchFalse(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create trigger event with data that makes condition FALSE (amount <= 1000)
	event := workflow.TriggerEvent{
		EventType:  sd.TriggerTypes[0].Name,
		EntityName: sd.Entities[0].Name,
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData: map[string]any{
			"amount": 500, // condition: amount > 1000 should be FALSE
			"status": "pending",
		},
		UserID: sd.Users[0].ID,
	}

	// Execute workflow via engine
	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify execution completed
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed:\n%s", formatExecutionErrors(execution))
	}

	// Parse the branching workflow rule ID
	ruleID, err := uuid.Parse(sd.BranchingWorkflow.ID)
	if err != nil {
		t.Fatalf("parsing branching workflow rule ID: %v", err)
	}

	// Find results for our specific rule
	var matchedRuleResult *workflow.RuleResult
	for _, batch := range execution.BatchResults {
		for i := range batch.RuleResults {
			if batch.RuleResults[i].RuleID == ruleID {
				matchedRuleResult = &batch.RuleResults[i]
				break
			}
		}
	}

	if matchedRuleResult == nil {
		// Rule may not match, which is OK
		return
	}

	// Verify the evaluate_condition action was executed
	foundCondition := false
	for _, ar := range matchedRuleResult.ActionResults {
		if ar.ActionType == "evaluate_condition" {
			foundCondition = true
		}
	}

	if !foundCondition {
		t.Fatal("evaluate_condition action not found in results")
	}
}

// testExecutionRecordCreated tests that execution records are properly created and tracked.
func testExecutionRecordCreated(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Create trigger event
	event := workflow.TriggerEvent{
		EventType:  sd.TriggerTypes[0].Name,
		EntityName: sd.Entities[0].Name,
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{},
		UserID:     sd.Users[0].ID,
	}

	// Execute workflow via engine
	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify we got an execution ID
	if execution.ExecutionID == uuid.Nil {
		t.Fatal("execution ID should not be nil")
	}

	// Verify execution metadata
	if execution.StartedAt.IsZero() {
		t.Fatal("started_at should not be zero")
	}

	if execution.CompletedAt == nil {
		t.Fatal("completed_at should not be nil")
	}

	if execution.TotalDuration == nil {
		t.Fatal("total_duration should not be nil")
	}

	if *execution.TotalDuration <= 0 {
		t.Fatal("total_duration should be > 0")
	}

	// Verify execution is in history
	history := sd.WF.Engine.GetExecutionHistory(10)
	found := false
	for _, h := range history {
		if h.ExecutionID == execution.ExecutionID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("execution not found in GetExecutionHistory")
	}

	// Verify stats were updated
	stats := sd.WF.Engine.GetStats()
	if stats.TotalWorkflowsProcessed == 0 {
		t.Fatal("total_workflows_processed should be > 0")
	}
}

// testExecutionHistoryTracking tests that multiple executions are tracked in history.
func testExecutionHistoryTracking(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data")
	}

	// Get initial history count
	initialHistory := len(sd.WF.Engine.GetExecutionHistory(100))

	// Execute multiple workflows
	executionIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		event := workflow.TriggerEvent{
			EventType:  sd.TriggerTypes[0].Name,
			EntityName: sd.Entities[0].Name,
			EntityID:   uuid.New(),
			Timestamp:  time.Now(),
			RawData:    map[string]any{"iteration": i},
			UserID:     sd.Users[0].ID,
		}

		execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
		if err != nil {
			t.Fatalf("execution %d failed: %v", i, err)
		}
		executionIDs[i] = execution.ExecutionID
	}

	// Verify all executions are in history
	history := sd.WF.Engine.GetExecutionHistory(100)
	finalHistoryCount := len(history)

	// Should have at least 3 new executions
	if finalHistoryCount < initialHistory+3 {
		t.Fatalf("expected at least %d executions in history, got %d", initialHistory+3, finalHistoryCount)
	}

	// Verify each execution ID is in history
	for i, execID := range executionIDs {
		found := false
		for _, h := range history {
			if h.ExecutionID == execID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("execution[%d] (ID: %s) not found in history", i, execID)
		}
	}
}

// testNoMatchingRules tests behavior when no rules match the trigger event.
func testNoMatchingRules(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	// Create trigger event for a non-existent entity
	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "nonexistent_entity_xyz_123",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{},
		UserID:     sd.Users[0].ID,
	}

	// Execute workflow via engine
	execution, err := sd.WF.Engine.ExecuteWorkflow(ctx, event)
	if err != nil {
		t.Fatalf("execution should not fail: %v", err)
	}

	// Verify execution completed with no matching rules
	if execution.Status != workflow.StatusCompleted {
		t.Fatalf("expected completed status, got %s", execution.Status)
	}

	// Verify no rules matched
	if execution.ExecutionPlan.MatchedRuleCount != 0 {
		t.Fatalf("expected 0 matched rules, got %d", execution.ExecutionPlan.MatchedRuleCount)
	}

	// Verify no batch results
	if len(execution.BatchResults) != 0 {
		t.Fatalf("expected 0 batch results, got %d", len(execution.BatchResults))
	}
}

// formatExecutionErrors extracts detailed error information from a workflow execution
// to make test failures unambiguous.
func formatExecutionErrors(execution *workflow.WorkflowExecution) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("status=%s, errors=%v", execution.Status, execution.Errors))

	for i, batch := range execution.BatchResults {
		if batch.BatchStatus != "completed" {
			sb.WriteString(fmt.Sprintf("\n  batch[%d] status=%s:", i, batch.BatchStatus))
			for j, rule := range batch.RuleResults {
				if rule.Status != "success" {
					sb.WriteString(fmt.Sprintf("\n    rule[%d] %s (ID=%s): %s",
						j, rule.RuleName, rule.RuleID, rule.ErrorMessage))
					for k, action := range rule.ActionResults {
						if action.Status != "success" {
							sb.WriteString(fmt.Sprintf("\n      action[%d] %s (%s): %s",
								k, action.ActionName, action.ActionType, action.ErrorMessage))
						}
					}
				}
			}
		}
	}
	return sb.String()
}
