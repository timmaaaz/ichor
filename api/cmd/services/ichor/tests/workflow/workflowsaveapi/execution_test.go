package workflowsaveapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Workflow Execution Integration Tests (Temporal-based)
// =============================================================================

// runExecutionTests runs all execution tests as subtests.
// These are not HTTP tests — they test the workflow engine via Temporal directly.
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
	t.Run("exec-no-matching-rules", func(t *testing.T) {
		testNoMatchingRules(t, sd)
	})
}

// testExecuteSingleCreateAlert tests that a simple workflow with 1 create_alert
// action produces a new alert record when its trigger fires.
func testExecuteSingleCreateAlert(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data for simple workflow test")
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	// SimpleWorkflow creates alerts with alert_type "simple_test" — filter specifically
	// to avoid counting alerts created by BranchingWorkflow (which shares TriggerTypes[0]).
	alertType := "simple_test"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	// SimpleWorkflow uses TriggerTypes[0] and Entities[0]
	event := createTriggerEvent(
		sd.Entities[0].Name,
		sd.TriggerTypes[0].Name,
		sd.Users[0].ID,
		map[string]any{},
	)

	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: single create_alert workflow executed via Temporal")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no new simple_test alert after 15s — SimpleWorkflow may not have dispatched")
}

// testExecuteSequence3Actions tests that a workflow with 3 sequential create_alert
// actions produces at least 3 new alert records.
func testExecuteSequence3Actions(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data for sequence workflow test")
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	// SequenceWorkflow uses TriggerTypes[1] if available, else TriggerTypes[0]
	triggerTypeName := sd.TriggerTypes[0].Name
	if len(sd.TriggerTypes) > 1 {
		triggerTypeName = sd.TriggerTypes[1].Name
	}

	event := createTriggerEvent(
		sd.Entities[0].Name,
		triggerTypeName,
		sd.Users[0].ID,
		map[string]any{},
	)

	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying alerts after: %v", err)
		}
		if len(after)-beforeCount >= 3 {
			t.Log("SUCCESS: sequence workflow created >=3 alerts")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: expected >=3 new alerts from sequence workflow after 15s")
}

// testExecuteBranchTrue fires the BranchingWorkflow with a high amount (>1000),
// which should route the true branch and produce a "high_value" alert.
func testExecuteBranchTrue(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data for branch-true test")
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	// Capture beforeCount so the test isn't satisfied by a pre-existing high_value alert
	// from a prior test run or concurrent workflow execution.
	alertType := "high_value"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying high_value alerts before: %v", err)
	}
	beforeCount := len(before)

	// BranchingWorkflow uses TriggerTypes[0] and Entities[0]
	event := createTriggerEvent(
		sd.Entities[0].Name,
		sd.TriggerTypes[0].Name,
		sd.Users[0].ID,
		map[string]any{"amount": float64(1500)},
	)

	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying high_value alerts: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: branch-true path created high_value alert")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no new high_value alert after 15s")
}

// testExecuteBranchFalse fires the BranchingWorkflow with a low amount (<=1000),
// which should route the false branch and produce a "normal_value" alert.
func testExecuteBranchFalse(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		t.Fatal("insufficient seed data for branch-false test")
	}

	if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	// Capture beforeCount so the test isn't satisfied by a pre-existing normal_value alert
	// from a prior test run or concurrent workflow execution.
	alertType := "normal_value"
	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying normal_value alerts before: %v", err)
	}
	beforeCount := len(before)

	// BranchingWorkflow uses TriggerTypes[0] and Entities[0]
	event := createTriggerEvent(
		sd.Entities[0].Name,
		sd.TriggerTypes[0].Name,
		sd.Users[0].ID,
		map[string]any{"amount": float64(500)},
	)

	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger: %v", err)
	}

	for i := 0; i < 30; i++ {
		after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{AlertType: &alertType}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
		if err != nil {
			t.Fatalf("querying normal_value alerts: %v", err)
		}
		if len(after) > beforeCount {
			t.Log("SUCCESS: branch-false path created normal_value alert")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout: no new normal_value alert after 15s")
}

// testNoMatchingRules fires an event for a non-existent entity and verifies
// that no new alerts are created (no rules match).
func testNoMatchingRules(t *testing.T, sd ExecutionTestData) {
	ctx := context.Background()

	before, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts before: %v", err)
	}
	beforeCount := len(before)

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "nonexistent_entity_xyz_" + uuid.New().String()[:8],
		EntityID:   uuid.New(),
		UserID:     sd.Users[0].ID,
	}

	if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("unexpected error for no-match event: %v", err)
	}

	time.Sleep(2 * time.Second)

	after, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying alerts after: %v", err)
	}
	if len(after) != beforeCount {
		t.Errorf("expected no new alerts, got %d new", len(after)-beforeCount)
	}
	t.Log("SUCCESS: no matching rules fired no workflows")
}
