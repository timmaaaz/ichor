package temporal

import (
	"testing"

	"github.com/google/uuid"
)

// =============================================================================
// Multiple Always Edges
// =============================================================================

func TestGetNextActions_MultipleAlwaysEdges(t *testing.T) {
	t.Parallel()

	// Source -> A (true_branch) + B (always) + C (always)
	source := deterministicUUID("multi-always-source")
	a := deterministicUUID("multi-always-a")
	b := deterministicUUID("multi-always-b")
	c := deterministicUUID("multi-always-c")

	srcRef := source
	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: source, Name: "source", ActionType: "evaluate_condition"},
			{ID: a, Name: "true_action", ActionType: "send_email"},
			{ID: b, Name: "always_1", ActionType: "set_field"},
			{ID: c, Name: "always_2", ActionType: "create_alert"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("ma-start"), SourceActionID: nil, TargetActionID: source, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("ma-true"), SourceActionID: &srcRef, TargetActionID: a, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("true"), SortOrder: 1},
			{ID: deterministicUUID("ma-always1"), SourceActionID: &srcRef, TargetActionID: b, EdgeType: EdgeTypeAlways, SortOrder: 2},
			{ID: deterministicUUID("ma-always2"), SourceActionID: &srcRef, TargetActionID: c, EdgeType: EdgeTypeAlways, SortOrder: 3},
		},
	}

	exec := NewGraphExecutor(graph)
	trueResult := map[string]any{"output": "true"}
	next := exec.GetNextActions(source, trueResult)

	if len(next) != 3 {
		t.Fatalf("expected 3 actions (true + 2 always), got %d", len(next))
	}

	// Verify order: A (SortOrder 1), B (SortOrder 2), C (SortOrder 3).
	if next[0].ID != a {
		t.Errorf("expected true_action at index 0, got %s", next[0].Name)
	}
	if next[1].ID != b {
		t.Errorf("expected always_1 at index 1, got %s", next[1].Name)
	}
	if next[2].ID != c {
		t.Errorf("expected always_2 at index 2, got %s", next[2].Name)
	}
}

func TestGetNextActions_AlwaysWithSequence(t *testing.T) {
	t.Parallel()

	// Source -> A (sequence) + B (always)
	source := deterministicUUID("always-seq-source")
	a := deterministicUUID("always-seq-a")
	b := deterministicUUID("always-seq-b")

	srcRef := source
	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: source, Name: "source", ActionType: "set_field"},
			{ID: a, Name: "seq_action", ActionType: "send_email"},
			{ID: b, Name: "always_action", ActionType: "create_alert"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("as-start"), SourceActionID: nil, TargetActionID: source, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("as-seq"), SourceActionID: &srcRef, TargetActionID: a, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("as-always"), SourceActionID: &srcRef, TargetActionID: b, EdgeType: EdgeTypeAlways, SortOrder: 2},
		},
	}

	exec := NewGraphExecutor(graph)
	next := exec.GetNextActions(source, nil)

	if len(next) != 2 {
		t.Fatalf("expected 2 actions (sequence + always), got %d", len(next))
	}
	if next[0].ID != a {
		t.Errorf("expected seq_action at index 0, got %s", next[0].Name)
	}
	if next[1].ID != b {
		t.Errorf("expected always_action at index 1, got %s", next[1].Name)
	}
}

// =============================================================================
// Chained Conditions
// =============================================================================

func TestGetNextActions_ConditionChain(t *testing.T) {
	t.Parallel()

	// Condition1 -> Condition2 (true_branch) -> Action (true_branch)
	cond1 := deterministicUUID("chain-cond1")
	cond2 := deterministicUUID("chain-cond2")
	action := deterministicUUID("chain-action")

	cond1Ref := cond1
	cond2Ref := cond2

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: cond1, Name: "condition1", ActionType: "evaluate_condition"},
			{ID: cond2, Name: "condition2", ActionType: "evaluate_condition"},
			{ID: action, Name: "final_action", ActionType: "send_email"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("cc-start"), SourceActionID: nil, TargetActionID: cond1, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("cc-c1-c2"), SourceActionID: &cond1Ref, TargetActionID: cond2, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("true"), SortOrder: 1},
			{ID: deterministicUUID("cc-c2-act"), SourceActionID: &cond2Ref, TargetActionID: action, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("true"), SortOrder: 2},
		},
	}

	exec := NewGraphExecutor(graph)

	// First condition: true -> goes to condition2.
	trueResult := map[string]any{"output": "true"}
	next1 := exec.GetNextActions(cond1, trueResult)
	if len(next1) != 1 || next1[0].ID != cond2 {
		t.Fatalf("expected condition2 after cond1 true, got %v", next1)
	}

	// Second condition: true -> goes to action.
	trueResult2 := map[string]any{"output": "true"}
	next2 := exec.GetNextActions(cond2, trueResult2)
	if len(next2) != 1 || next2[0].ID != action {
		t.Fatalf("expected final_action after cond2 true, got %v", next2)
	}
}

func TestGetNextActions_ConditionToCondition_FalseTrue(t *testing.T) {
	t.Parallel()

	// Condition1 --(false output)--> Condition2
	// Condition2 --(true output)--> Action
	cond1 := deterministicUUID("ft-cond1")
	cond2 := deterministicUUID("ft-cond2")
	action := deterministicUUID("ft-action")

	cond1Ref := cond1
	cond2Ref := cond2

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: cond1, Name: "condition1", ActionType: "evaluate_condition"},
			{ID: cond2, Name: "condition2", ActionType: "evaluate_condition"},
			{ID: action, Name: "final_action", ActionType: "send_email"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("ft-start"), SourceActionID: nil, TargetActionID: cond1, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("ft-c1-c2"), SourceActionID: &cond1Ref, TargetActionID: cond2, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("false"), SortOrder: 1},
			{ID: deterministicUUID("ft-c2-act"), SourceActionID: &cond2Ref, TargetActionID: action, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("true"), SortOrder: 2},
		},
	}

	exec := NewGraphExecutor(graph)

	// First condition: false -> goes to condition2.
	falseResult := map[string]any{"output": "false"}
	next1 := exec.GetNextActions(cond1, falseResult)
	if len(next1) != 1 || next1[0].ID != cond2 {
		t.Fatalf("expected condition2 after cond1 false, got %v", next1)
	}

	// Second condition: true -> goes to action.
	trueResult := map[string]any{"output": "true"}
	next2 := exec.GetNextActions(cond2, trueResult)
	if len(next2) != 1 || next2[0].ID != action {
		t.Fatalf("expected final_action after cond2 true, got %v", next2)
	}
}

// =============================================================================
// Mixed Edge Types on Diamond
// =============================================================================

func TestGetNextActions_ConditionDiamond(t *testing.T) {
	t.Parallel()

	// Condition -> A (true_branch), B (false_branch)
	// A -> Merge (sequence)
	// B -> Merge (sequence)
	condition := deterministicUUID("cd-condition")
	a := deterministicUUID("cd-true")
	b := deterministicUUID("cd-false")
	merge := deterministicUUID("cd-merge")

	condRef := condition
	aRef := a
	bRef := b

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: condition, Name: "condition", ActionType: "evaluate_condition"},
			{ID: a, Name: "true_path", ActionType: "send_email"},
			{ID: b, Name: "false_path", ActionType: "create_alert"},
			{ID: merge, Name: "merge", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("cd-start"), SourceActionID: nil, TargetActionID: condition, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("cd-true-edge"), SourceActionID: &condRef, TargetActionID: a, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("true"), SortOrder: 1},
			{ID: deterministicUUID("cd-false-edge"), SourceActionID: &condRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("false"), SortOrder: 2},
			{ID: deterministicUUID("cd-a-merge"), SourceActionID: &aRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 3},
			{ID: deterministicUUID("cd-b-merge"), SourceActionID: &bRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 4},
		},
	}

	exec := NewGraphExecutor(graph)

	// True branch: condition -> A -> merge.
	trueResult := map[string]any{"output": "true"}
	next := exec.GetNextActions(condition, trueResult)
	if len(next) != 1 || next[0].ID != a {
		t.Fatalf("expected true_path, got %v", next)
	}
	afterTrue := exec.GetNextActions(a, nil)
	if len(afterTrue) != 1 || afterTrue[0].ID != merge {
		t.Fatalf("expected merge after true_path, got %v", afterTrue)
	}

	// False branch: condition -> B -> merge.
	falseResult := map[string]any{"output": "false"}
	next = exec.GetNextActions(condition, falseResult)
	if len(next) != 1 || next[0].ID != b {
		t.Fatalf("expected false_path, got %v", next)
	}
	afterFalse := exec.GetNextActions(b, nil)
	if len(afterFalse) != 1 || afterFalse[0].ID != merge {
		t.Fatalf("expected merge after false_path, got %v", afterFalse)
	}

	// Verify merge has multiple incoming.
	if !exec.HasMultipleIncoming(merge) {
		t.Error("merge should have multiple incoming edges")
	}
}

// =============================================================================
// Result Map Edge Cases (table-driven)
// =============================================================================

func TestGetNextActions_ResultMapEdgeCases(t *testing.T) {
	t.Parallel()

	conditionID := deterministicUUID("rme-condition")
	actionA := deterministicUUID("rme-action-a")
	actionB := deterministicUUID("rme-action-b")
	actionC := deterministicUUID("rme-action-c")

	condRef := conditionID
	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: conditionID, Name: "condition", ActionType: "evaluate_condition"},
			{ID: actionA, Name: "true_action", ActionType: "send_email"},
			{ID: actionB, Name: "false_action", ActionType: "create_alert"},
			{ID: actionC, Name: "always_action", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("rme-start"), SourceActionID: nil, TargetActionID: conditionID, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("rme-true"), SourceActionID: &condRef, TargetActionID: actionA, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("true"), SortOrder: 1},
			{ID: deterministicUUID("rme-false"), SourceActionID: &condRef, TargetActionID: actionB, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("false"), SortOrder: 2},
			{ID: deterministicUUID("rme-always"), SourceActionID: &condRef, TargetActionID: actionC, EdgeType: EdgeTypeAlways, SortOrder: 3},
		},
	}

	tests := []struct {
		name      string
		result    map[string]any
		expectIDs []uuid.UUID
	}{
		{"nil result", nil, []uuid.UUID{actionC}},
		{"empty map", map[string]any{}, []uuid.UUID{actionC}},
		{"empty output", map[string]any{"output": ""}, []uuid.UUID{actionC}},
		{"missing output key", map[string]any{"other": "val"}, []uuid.UUID{actionC}},
		{"unknown output value", map[string]any{"output": "unknown"}, []uuid.UUID{actionC}},
		{"wrong type int", map[string]any{"output": 123}, []uuid.UUID{actionC}},
		{"wrong type bool", map[string]any{"output": true}, []uuid.UUID{actionC}},
		{"valid true output", map[string]any{"output": "true"}, []uuid.UUID{actionA, actionC}},
		{"valid false output", map[string]any{"output": "false"}, []uuid.UUID{actionB, actionC}},
	}

	exec := NewGraphExecutor(graph)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := exec.GetNextActions(conditionID, tt.result)
			gotIDs := extractIDs(next)

			if len(gotIDs) != len(tt.expectIDs) {
				t.Fatalf("expected %d actions, got %d: %v", len(tt.expectIDs), len(gotIDs), next)
			}
			for i := range tt.expectIDs {
				if gotIDs[i] != tt.expectIDs[i] {
					t.Errorf("index %d: expected %s, got %s", i, tt.expectIDs[i], gotIDs[i])
				}
			}
		})
	}
}

// =============================================================================
// Always-Only Edges (no condition branches)
// =============================================================================

func TestGetNextActions_OnlyAlwaysEdges(t *testing.T) {
	t.Parallel()

	// Source has only always edges - should always be followed regardless of result.
	source := deterministicUUID("only-always-src")
	a := deterministicUUID("only-always-a")
	b := deterministicUUID("only-always-b")

	srcRef := source
	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: source, Name: "source", ActionType: "set_field"},
			{ID: a, Name: "always_a", ActionType: "send_email"},
			{ID: b, Name: "always_b", ActionType: "create_alert"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("oa-start"), SourceActionID: nil, TargetActionID: source, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("oa-a"), SourceActionID: &srcRef, TargetActionID: a, EdgeType: EdgeTypeAlways, SortOrder: 1},
			{ID: deterministicUUID("oa-b"), SourceActionID: &srcRef, TargetActionID: b, EdgeType: EdgeTypeAlways, SortOrder: 2},
		},
	}

	exec := NewGraphExecutor(graph)

	// With nil result.
	next := exec.GetNextActions(source, nil)
	if len(next) != 2 {
		t.Fatalf("expected 2 always actions with nil result, got %d", len(next))
	}

	// With arbitrary result map.
	next = exec.GetNextActions(source, map[string]any{"foo": "bar"})
	if len(next) != 2 {
		t.Fatalf("expected 2 always actions with arbitrary result, got %d", len(next))
	}
}

// =============================================================================
// Condition with Always + No Matching Branch
// =============================================================================

func TestGetNextActions_ConditionNoMatchButAlways(t *testing.T) {
	t.Parallel()

	// Condition -> A (true_branch), B (always)
	// Result is nil -> only always edge followed.
	condition := deterministicUUID("no-match-cond")
	a := deterministicUUID("no-match-a")
	b := deterministicUUID("no-match-b")

	condRef := condition
	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: condition, Name: "condition", ActionType: "evaluate_condition"},
			{ID: a, Name: "true_action", ActionType: "send_email"},
			{ID: b, Name: "always_action", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("nm-start"), SourceActionID: nil, TargetActionID: condition, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("nm-true"), SourceActionID: &condRef, TargetActionID: a, EdgeType: EdgeTypeSequence, SourceOutput: strPtr("true"), SortOrder: 1},
			{ID: deterministicUUID("nm-always"), SourceActionID: &condRef, TargetActionID: b, EdgeType: EdgeTypeAlways, SortOrder: 2},
		},
	}

	exec := NewGraphExecutor(graph)

	// Nil result -> no branch match, only always.
	next := exec.GetNextActions(condition, nil)
	if len(next) != 1 {
		t.Fatalf("expected 1 action (always only), got %d", len(next))
	}
	if next[0].ID != b {
		t.Errorf("expected always_action, got %s", next[0].Name)
	}
}
