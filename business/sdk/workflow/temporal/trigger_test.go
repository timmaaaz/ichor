package temporal_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Mock EdgeStore
// =============================================================================

type mockEdgeStore struct {
	actions map[uuid.UUID][]temporal.ActionNode
	edges   map[uuid.UUID][]temporal.ActionEdge
	err     error // If set, all calls return this error
}

func newMockEdgeStore() *mockEdgeStore {
	return &mockEdgeStore{
		actions: make(map[uuid.UUID][]temporal.ActionNode),
		edges:   make(map[uuid.UUID][]temporal.ActionEdge),
	}
}

func (m *mockEdgeStore) QueryActionsByRule(_ context.Context, ruleID uuid.UUID) ([]temporal.ActionNode, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.actions[ruleID], nil
}

func (m *mockEdgeStore) QueryEdgesByRule(_ context.Context, ruleID uuid.UUID) ([]temporal.ActionEdge, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.edges[ruleID], nil
}

// =============================================================================
// Mock RuleMatcher
// =============================================================================

type mockRuleMatcher struct {
	result *workflow.ProcessingResult
	err    error
}

func (m *mockRuleMatcher) ProcessEvent(_ context.Context, _ workflow.TriggerEvent) (*workflow.ProcessingResult, error) {
	return m.result, m.err
}

// =============================================================================
// Mock WorkflowStarter
// =============================================================================

type mockWorkflowRun struct {
	id    string
	runID string
}

func (m *mockWorkflowRun) GetID() string    { return m.id }
func (m *mockWorkflowRun) GetRunID() string { return m.runID }
func (m *mockWorkflowRun) Get(_ context.Context, _ any) error {
	return nil
}
func (m *mockWorkflowRun) GetWithOptions(_ context.Context, _ any, _ client.WorkflowRunGetOptions) error {
	return nil
}

type workflowStartCall struct {
	options  client.StartWorkflowOptions
	workflow any
	args     []any
}

type mockWorkflowStarter struct {
	calls []workflowStartCall
	err   error // If set, all calls return this error
}

func newMockWorkflowStarter() *mockWorkflowStarter {
	return &mockWorkflowStarter{}
}

func (m *mockWorkflowStarter) ExecuteWorkflow(_ context.Context, options client.StartWorkflowOptions, wf any, args ...any) (client.WorkflowRun, error) {
	m.calls = append(m.calls, workflowStartCall{
		options:  options,
		workflow: wf,
		args:     args,
	})
	if m.err != nil {
		return nil, m.err
	}
	return &mockWorkflowRun{id: options.ID, runID: "test-run-id"}, nil
}

// =============================================================================
// Test Helpers
// =============================================================================

func testLogger() *logger.Logger {
	return logger.New(io.Discard, logger.LevelError, "TEST", func(context.Context) string { return "" })
}

func testEvent() workflow.TriggerEvent {
	return workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "orders",
		EntityID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Timestamp:  time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC),
		UserID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		RawData: map[string]any{
			"status": "pending",
			"total":  100.50,
		},
	}
}

func testUpdateEvent() workflow.TriggerEvent {
	return workflow.TriggerEvent{
		EventType:  "on_update",
		EntityName: "orders",
		EntityID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Timestamp:  time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC),
		UserID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		FieldChanges: map[string]workflow.FieldChange{
			"status": {OldValue: "pending", NewValue: "shipped"},
			"amount": {OldValue: 100, NewValue: 200},
		},
		RawData: map[string]any{
			"status": "shipped",
			"total":  200.00,
		},
	}
}

func testRuleID() uuid.UUID {
	return uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
}

func testRuleID2() uuid.UUID {
	return uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
}

func testGraph(ruleID uuid.UUID) ([]temporal.ActionNode, []temporal.ActionEdge) {
	actionID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	actions := []temporal.ActionNode{
		{
			ID:         actionID,
			Name:       "send_notification",
			ActionType: "send_notification",
			Config:     []byte(`{"message": "Order created"}`),
			IsActive:   true,
		},
	}
	edges := []temporal.ActionEdge{
		{
			ID:             uuid.New(),
			TargetActionID: actionID,
			EdgeType:       "start",
			SortOrder:      1,
		},
	}
	_ = ruleID // Used for context only
	return actions, edges
}

func matchedResult(ruleID uuid.UUID, ruleName string) *workflow.ProcessingResult {
	return &workflow.ProcessingResult{
		TotalRulesEvaluated: 1,
		MatchedRules: []workflow.RuleMatchResult{
			{
				Rule:    workflow.AutomationRuleView{ID: ruleID, Name: ruleName},
				Matched: true,
			},
		},
	}
}

// =============================================================================
// buildTriggerData Tests
// =============================================================================

func TestBuildTriggerData_BasicEvent(t *testing.T) {
	event := testEvent()

	// buildTriggerData is not exported, so we test it indirectly
	// by verifying WorkflowInput.TriggerData through OnEntityEvent.
	// However, the phase doc has it as an exported-style function.
	// Let's test via the full flow.

	// For a direct test, we need the function to be accessible.
	// Since it's in the same package, we test the behavior through OnEntityEvent.
	// We'll verify the data through the mock starter's captured args.

	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	actions, edges := testGraph(ruleID)
	edgeStore.actions[ruleID] = actions
	edgeStore.edges[ruleID] = edges

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "test-rule")}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 1 {
		t.Fatalf("expected 1 workflow start, got %d", len(starter.calls))
	}

	// Verify the WorkflowInput was passed correctly.
	call := starter.calls[0]
	if call.options.TaskQueue != temporal.TaskQueue {
		t.Errorf("expected task queue %q, got %q", temporal.TaskQueue, call.options.TaskQueue)
	}

	// Verify workflow ID format: workflow-{ruleID}-{entityID}-{executionID}
	wfID := call.options.ID
	if len(wfID) == 0 {
		t.Error("workflow ID should not be empty")
	}

	// Should start with "workflow-{ruleID}-{entityID}-"
	prefix := "workflow-" + ruleID.String() + "-" + event.EntityID.String() + "-"
	if len(wfID) < len(prefix) || wfID[:len(prefix)] != prefix {
		t.Errorf("workflow ID should start with %q, got %q", prefix, wfID)
	}
}

func TestBuildTriggerData_UpdateEventWithFieldChanges(t *testing.T) {
	event := testUpdateEvent()

	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	actions, edges := testGraph(ruleID)
	edgeStore.actions[ruleID] = actions
	edgeStore.edges[ruleID] = edges

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "test-rule")}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 1 {
		t.Fatalf("expected 1 workflow start, got %d", len(starter.calls))
	}

	// Verify workflow was started (input contains trigger data with field changes).
	if len(starter.calls[0].args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(starter.calls[0].args))
	}

	input, ok := starter.calls[0].args[0].(temporal.WorkflowInput)
	if !ok {
		t.Fatalf("expected WorkflowInput arg, got %T", starter.calls[0].args[0])
	}

	// Verify trigger data contains field_changes.
	if input.TriggerData == nil {
		t.Fatal("trigger data should not be nil")
	}
	changes, ok := input.TriggerData["field_changes"]
	if !ok {
		t.Fatal("trigger data should contain field_changes")
	}
	changesMap, ok := changes.(map[string]any)
	if !ok {
		t.Fatalf("field_changes should be map[string]any, got %T", changes)
	}
	if _, ok := changesMap["status"]; !ok {
		t.Error("field_changes should contain 'status'")
	}
	if _, ok := changesMap["amount"]; !ok {
		t.Error("field_changes should contain 'amount'")
	}
}

func TestBuildTriggerData_EmptyRawData(t *testing.T) {
	event := workflow.TriggerEvent{
		EventType:  "on_delete",
		EntityName: "orders",
		EntityID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Timestamp:  time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC),
		UserID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		// RawData is nil
	}

	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	actions, edges := testGraph(ruleID)
	edgeStore.actions[ruleID] = actions
	edgeStore.edges[ruleID] = edges

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "test-rule")}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 1 {
		t.Fatalf("expected 1 workflow start, got %d", len(starter.calls))
	}

	input := starter.calls[0].args[0].(temporal.WorkflowInput)
	if input.TriggerData == nil {
		t.Fatal("trigger data should not be nil even with empty raw data")
	}
	if input.TriggerData["event_type"] != "on_delete" {
		t.Errorf("expected event_type 'on_delete', got %v", input.TriggerData["event_type"])
	}
	if input.TriggerData["entity_name"] != "orders" {
		t.Errorf("expected entity_name 'orders', got %v", input.TriggerData["entity_name"])
	}
}

// =============================================================================
// OnEntityEvent Tests
// =============================================================================

func TestOnEntityEvent_Success(t *testing.T) {
	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	actions, edges := testGraph(ruleID)
	edgeStore.actions[ruleID] = actions
	edgeStore.edges[ruleID] = edges

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "test-rule")}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 1 {
		t.Fatalf("expected 1 workflow start, got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_NoMatches(t *testing.T) {
	edgeStore := newMockEdgeStore()
	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{
		result: &workflow.ProcessingResult{
			TotalRulesEvaluated: 5,
			MatchedRules:        nil,
		},
	}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 0 {
		t.Errorf("expected 0 workflow starts for no matches, got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_UnmatchedRulesSkipped(t *testing.T) {
	edgeStore := newMockEdgeStore()
	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{
		result: &workflow.ProcessingResult{
			TotalRulesEvaluated: 2,
			MatchedRules: []workflow.RuleMatchResult{
				{Rule: workflow.AutomationRuleView{ID: testRuleID()}, Matched: false},
				{Rule: workflow.AutomationRuleView{ID: testRuleID2()}, Matched: false},
			},
		},
	}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 0 {
		t.Errorf("expected 0 workflow starts for unmatched rules, got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_EmptyGraph(t *testing.T) {
	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	// No actions or edges registered for this rule.

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "empty-rule")}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 0 {
		t.Errorf("expected 0 workflow starts for empty graph, got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_EdgeStoreError(t *testing.T) {
	ruleID1 := testRuleID()
	ruleID2 := testRuleID2()

	// EdgeStore fails for all calls (simulating DB error).
	edgeStore := newMockEdgeStore()
	edgeStore.err = errors.New("database connection failed")

	// But add a second rule with its own edge store that works.
	// Since our mock returns error for all, both rules will fail.
	// The trigger should still not return an error (fail-open per rule).

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{
		result: &workflow.ProcessingResult{
			TotalRulesEvaluated: 2,
			MatchedRules: []workflow.RuleMatchResult{
				{Rule: workflow.AutomationRuleView{ID: ruleID1, Name: "rule-1"}, Matched: true},
				{Rule: workflow.AutomationRuleView{ID: ruleID2, Name: "rule-2"}, Matched: true},
			},
		},
	}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("expected no error (fail-open), got: %v", err)
	}

	// No workflows should have started since EdgeStore failed for both.
	if len(starter.calls) != 0 {
		t.Errorf("expected 0 workflow starts when EdgeStore fails, got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_TemporalError(t *testing.T) {
	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	actions, edges := testGraph(ruleID)
	edgeStore.actions[ruleID] = actions
	edgeStore.edges[ruleID] = edges

	starter := newMockWorkflowStarter()
	starter.err = errors.New("temporal unavailable")

	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "test-rule")}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("expected no error (fail-open per rule), got: %v", err)
	}

	// The starter was called but returned an error.
	if len(starter.calls) != 1 {
		t.Errorf("expected 1 attempt, got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_ProcessEventError(t *testing.T) {
	edgeStore := newMockEdgeStore()
	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{
		err: errors.New("rule matching failed"),
	}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err == nil {
		t.Fatal("expected error when ProcessEvent fails")
	}

	if len(starter.calls) != 0 {
		t.Errorf("expected 0 workflow starts when ProcessEvent fails, got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_MultipleRules(t *testing.T) {
	ruleID1 := testRuleID()
	ruleID2 := testRuleID2()

	edgeStore := newMockEdgeStore()
	actions1, edges1 := testGraph(ruleID1)
	edgeStore.actions[ruleID1] = actions1
	edgeStore.edges[ruleID1] = edges1

	actionID2 := uuid.New()
	edgeStore.actions[ruleID2] = []temporal.ActionNode{
		{ID: actionID2, Name: "create_alert", ActionType: "create_alert", Config: []byte(`{}`), IsActive: true},
	}
	edgeStore.edges[ruleID2] = []temporal.ActionEdge{
		{ID: uuid.New(), TargetActionID: actionID2, EdgeType: "start", SortOrder: 1},
	}

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{
		result: &workflow.ProcessingResult{
			TotalRulesEvaluated: 2,
			MatchedRules: []workflow.RuleMatchResult{
				{Rule: workflow.AutomationRuleView{ID: ruleID1, Name: "rule-1"}, Matched: true},
				{Rule: workflow.AutomationRuleView{ID: ruleID2, Name: "rule-2"}, Matched: true},
			},
		},
	}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 2 {
		t.Fatalf("expected 2 workflow starts, got %d", len(starter.calls))
	}

	// Verify workflow IDs are unique.
	if starter.calls[0].options.ID == starter.calls[1].options.ID {
		t.Error("workflow IDs should be unique for different rules")
	}

	// Verify each workflow has the correct task queue.
	for i, call := range starter.calls {
		if call.options.TaskQueue != temporal.TaskQueue {
			t.Errorf("call %d: expected task queue %q, got %q", i, temporal.TaskQueue, call.options.TaskQueue)
		}
	}
}

func TestOnEntityEvent_MultipleRulesPartialFailure(t *testing.T) {
	ruleID1 := testRuleID()
	ruleID2 := testRuleID2()

	edgeStore := newMockEdgeStore()
	// Rule 1: no graph (empty) - will be skipped.
	// Rule 2: has graph - should succeed.
	actionID2 := uuid.New()
	edgeStore.actions[ruleID2] = []temporal.ActionNode{
		{ID: actionID2, Name: "create_alert", ActionType: "create_alert", Config: []byte(`{}`), IsActive: true},
	}
	edgeStore.edges[ruleID2] = []temporal.ActionEdge{
		{ID: uuid.New(), TargetActionID: actionID2, EdgeType: "start", SortOrder: 1},
	}

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{
		result: &workflow.ProcessingResult{
			TotalRulesEvaluated: 2,
			MatchedRules: []workflow.RuleMatchResult{
				{Rule: workflow.AutomationRuleView{ID: ruleID1, Name: "empty-rule"}, Matched: true},
				{Rule: workflow.AutomationRuleView{ID: ruleID2, Name: "valid-rule"}, Matched: true},
			},
		},
	}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), testEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the second rule should have started a workflow.
	if len(starter.calls) != 1 {
		t.Fatalf("expected 1 workflow start (rule 1 empty), got %d", len(starter.calls))
	}
}

func TestOnEntityEvent_WorkflowInputContents(t *testing.T) {
	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	actions, edges := testGraph(ruleID)
	edgeStore.actions[ruleID] = actions
	edgeStore.edges[ruleID] = edges

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "test-rule")}

	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	event := testEvent()
	err := trigger.OnEntityEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(starter.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(starter.calls))
	}

	input := starter.calls[0].args[0].(temporal.WorkflowInput)

	// Verify RuleID.
	if input.RuleID != ruleID {
		t.Errorf("expected rule_id %s, got %s", ruleID, input.RuleID)
	}

	// Verify RuleName.
	if input.RuleName != "test-rule" {
		t.Errorf("expected rule_name 'test-rule', got %s", input.RuleName)
	}

	// Verify ExecutionID is set.
	if input.ExecutionID == uuid.Nil {
		t.Error("execution_id should not be nil")
	}

	// Verify graph.
	if len(input.Graph.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(input.Graph.Actions))
	}
	if len(input.Graph.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(input.Graph.Edges))
	}

	// Verify trigger data contains event metadata.
	if input.TriggerData["event_type"] != "on_create" {
		t.Errorf("expected event_type 'on_create', got %v", input.TriggerData["event_type"])
	}
	if input.TriggerData["entity_name"] != "orders" {
		t.Errorf("expected entity_name 'orders', got %v", input.TriggerData["entity_name"])
	}
	if input.TriggerData["entity_id"] != event.EntityID.String() {
		t.Errorf("expected entity_id %s, got %v", event.EntityID, input.TriggerData["entity_id"])
	}
	if input.TriggerData["user_id"] != event.UserID.String() {
		t.Errorf("expected user_id %s, got %v", event.UserID, input.TriggerData["user_id"])
	}

	// Verify raw data was merged.
	if input.TriggerData["status"] != "pending" {
		t.Errorf("expected status 'pending' from raw data, got %v", input.TriggerData["status"])
	}
	if input.TriggerData["total"] != 100.50 {
		t.Errorf("expected total 100.50 from raw data, got %v", input.TriggerData["total"])
	}

	// Verify ContinuationState is nil (first execution).
	if input.ContinuationState != nil {
		t.Error("ContinuationState should be nil for initial execution")
	}
}

func TestWorkflowIDFormat(t *testing.T) {
	edgeStore := newMockEdgeStore()
	ruleID := testRuleID()
	actions, edges := testGraph(ruleID)
	edgeStore.actions[ruleID] = actions
	edgeStore.edges[ruleID] = edges

	starter := newMockWorkflowStarter()
	matcher := &mockRuleMatcher{result: matchedResult(ruleID, "test-rule")}

	event := testEvent()
	trigger := temporal.NewWorkflowTrigger(testLogger(), starter, matcher, edgeStore)
	err := trigger.OnEntityEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wfID := starter.calls[0].options.ID

	// Format: workflow-{ruleID}-{entityID}-{executionID}
	// Verify prefix.
	expectedPrefix := "workflow-" + ruleID.String() + "-" + event.EntityID.String() + "-"
	if len(wfID) <= len(expectedPrefix) {
		t.Fatalf("workflow ID too short: %s", wfID)
	}
	if wfID[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("workflow ID prefix mismatch: expected %q, got %q", expectedPrefix, wfID[:len(expectedPrefix)])
	}

	// Verify the suffix is a valid UUID (executionID).
	suffix := wfID[len(expectedPrefix):]
	if _, err := uuid.Parse(suffix); err != nil {
		t.Errorf("workflow ID suffix should be a valid UUID, got %q: %v", suffix, err)
	}
}
