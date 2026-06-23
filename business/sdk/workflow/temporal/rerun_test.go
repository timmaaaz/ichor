package temporal_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// newRerunTrigger builds a trigger wired to the supplied fakes, mirroring how the
// existing trigger tests construct it. The matcher is irrelevant for rerun (the
// rule is supplied directly from the loaded execution), so a permissive stub is fine.
func newRerunTrigger(starter temporal.WorkflowStarter, edges temporal.EdgeStore, exec temporal.ExecutionStore) *temporal.WorkflowTrigger {
	matcher := &mockRuleMatcher{result: &workflow.ProcessingResult{}}
	return temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edges, exec)
}

func TestReconstructTriggerEvent_ReversesBuildTriggerData(t *testing.T) {
	entityID := uuid.New()
	userID := uuid.New()
	orig := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "order_line_items",
		EntityID:   entityID,
		UserID:     userID,
		RawData:    map[string]any{"product_id": "p1", "quantity": float64(7), "order_id": "o1"},
		FieldChanges: map[string]workflow.FieldChange{
			"quantity": {OldValue: float64(0), NewValue: float64(7)},
		},
	}

	// Round-trip through the persisted shape.
	td := temporal.BuildTriggerData(orig)
	td[temporal.CascadeLineageKey] = temporal.WorkflowLineage{} // present in persisted JSON; must be ignored
	raw, _ := json.Marshal(td)

	got, err := temporal.ReconstructTriggerEvent(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.EntityID != entityID || got.EntityName != "order_line_items" || got.EventType != "on_create" {
		t.Fatalf("metadata wrong: %+v", got)
	}
	if got.UserID != userID {
		t.Fatalf("user id wrong: %v", got.UserID)
	}
	if got.EventID != uuid.Nil {
		t.Fatalf("EventID must be zero for a fresh dedup key, got %v", got.EventID)
	}
	if got.RawData["product_id"] != "p1" || got.RawData["quantity"] != float64(7) {
		t.Fatalf("raw data not recovered: %+v", got.RawData)
	}
	if _, ok := got.RawData[temporal.CascadeLineageKey]; ok {
		t.Fatal("CascadeLineageKey leaked into RawData")
	}
	if fc, ok := got.FieldChanges["quantity"]; !ok || fc.NewValue != float64(7) {
		t.Fatalf("field changes not recovered: %+v", got.FieldChanges)
	}
}

func TestRerunExecution_FreshIDAndDispatch(t *testing.T) {
	ruleID := uuid.New()
	origExecID := uuid.New()
	td, _ := json.Marshal(temporal.BuildTriggerData(workflow.TriggerEvent{
		EventType: "on_create", EntityName: "order_line_items", EntityID: uuid.New(),
		RawData: map[string]any{"product_id": "p1", "quantity": float64(7)},
	}))

	execStore := &mockExecutionStore{
		byID: map[uuid.UUID]workflow.AutomationExecution{
			origExecID: {ID: origExecID, AutomationRuleID: &ruleID, RuleName: "Granular Inventory Pipeline", EntityType: "order_line_items", TriggerData: td},
		},
	}
	starter := newMockWorkflowStarter()

	// The graph must be non-empty for the rule or startWorkflowForRule skips dispatch.
	edges := newMockEdgeStore()
	actions, edgeDefs := testGraph(ruleID)
	edges.actions[ruleID] = actions
	edges.edges[ruleID] = edgeDefs

	tr := newRerunTrigger(starter, edges, execStore)

	newID, err := tr.RerunExecution(context.Background(), origExecID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newID == origExecID || newID == uuid.Nil {
		t.Fatalf("expected a fresh execution id, got %v (orig %v)", newID, origExecID)
	}
	if len(starter.calls) != 1 {
		t.Fatalf("expected ExecuteWorkflow to be called once, got %d", len(starter.calls))
	}
	if execStore.lastCreated.ID != newID {
		t.Fatalf("execution record not created with the fresh id: %v", execStore.lastCreated.ID)
	}
	if execStore.lastCreated.AutomationRuleID == nil || *execStore.lastCreated.AutomationRuleID != ruleID {
		t.Fatalf("rerun must re-fire the originating rule, got %+v", execStore.lastCreated.AutomationRuleID)
	}
}

func TestRerunExecution_NoRule_Errors(t *testing.T) {
	origExecID := uuid.New()
	execStore := &mockExecutionStore{byID: map[uuid.UUID]workflow.AutomationExecution{
		origExecID: {ID: origExecID, AutomationRuleID: nil}, // manual execution
	}}
	tr := newRerunTrigger(newMockWorkflowStarter(), newMockEdgeStore(), execStore)

	_, err := tr.RerunExecution(context.Background(), origExecID)
	if !errors.Is(err, temporal.ErrExecutionNotRerunnable) {
		t.Fatalf("err = %v, want ErrExecutionNotRerunnable", err)
	}
}

func TestRerunExecution_LoadError(t *testing.T) {
	execStore := &mockExecutionStore{queryErr: errors.New("db down")}
	tr := newRerunTrigger(newMockWorkflowStarter(), newMockEdgeStore(), execStore)

	_, err := tr.RerunExecution(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected an error when the execution cannot be loaded")
	}
}
