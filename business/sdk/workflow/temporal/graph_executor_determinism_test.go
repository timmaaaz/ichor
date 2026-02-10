package temporal

import (
	"testing"

	"github.com/google/uuid"
)

// =============================================================================
// Helpers
// =============================================================================

// deterministicUUID generates a reproducible UUID from a string label.
// Use this instead of uuid.New() for test graphs — makes failures debuggable.
func deterministicUUID(label string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(label))
}

// extractIDs extracts action IDs in order for determinism comparison.
func extractIDs(actions []ActionNode) []uuid.UUID {
	ids := make([]uuid.UUID, len(actions))
	for i, a := range actions {
		ids[i] = a.ID
	}
	return ids
}

// assertIDsEqual fails the test if two ID slices differ (order matters).
func assertIDsEqual(t *testing.T, iteration int, baseline, current []uuid.UUID) {
	t.Helper()
	if len(baseline) != len(current) {
		t.Fatalf("iteration %d: length mismatch: baseline=%d, current=%d",
			iteration, len(baseline), len(current))
	}
	for i := range baseline {
		if baseline[i] != current[i] {
			t.Fatalf("iteration %d: non-deterministic at index %d\nbaseline: %v\ncurrent:  %v",
				iteration, i, baseline, current)
		}
	}
}

// buildComplexParallelGraph creates a graph with N parallel branches + convergence.
//
//	start -> condition -> N branches (each 2 nodes deep) -> convergence -> end
//
// Uses deterministicUUID for reproducibility.
func buildComplexParallelGraph(branchCount int) GraphDefinition {
	start := deterministicUUID("start")
	condition := deterministicUUID("condition")
	convergence := deterministicUUID("convergence")
	end := deterministicUUID("end")

	actions := []ActionNode{
		{ID: start, Name: "start", ActionType: "set_field"},
		{ID: condition, Name: "condition", ActionType: "evaluate_condition"},
		{ID: convergence, Name: "convergence", ActionType: "set_field"},
		{ID: end, Name: "end", ActionType: "create_alert"},
	}

	edges := []ActionEdge{
		{ID: deterministicUUID("edge-start"), SourceActionID: nil, TargetActionID: start, EdgeType: EdgeTypeStart, SortOrder: 0},
		{ID: deterministicUUID("edge-start-cond"), SourceActionID: &start, TargetActionID: condition, EdgeType: EdgeTypeSequence, SortOrder: 1},
	}

	for i := range branchCount {
		branchA := deterministicUUID("branch-a-" + string(rune('0'+i)))
		branchB := deterministicUUID("branch-b-" + string(rune('0'+i)))

		actions = append(actions,
			ActionNode{ID: branchA, Name: "branch_a_" + string(rune('0'+i)), ActionType: "send_email"},
			ActionNode{ID: branchB, Name: "branch_b_" + string(rune('0'+i)), ActionType: "set_field"},
		)

		condRef := condition
		edges = append(edges,
			ActionEdge{ID: deterministicUUID("edge-cond-ba-" + string(rune('0'+i))), SourceActionID: &condRef, TargetActionID: branchA, EdgeType: EdgeTypeSequence, SortOrder: 10 + i*2},
			ActionEdge{ID: deterministicUUID("edge-ba-bb-" + string(rune('0'+i))), SourceActionID: &branchA, TargetActionID: branchB, EdgeType: EdgeTypeSequence, SortOrder: 11 + i*2},
			ActionEdge{ID: deterministicUUID("edge-bb-conv-" + string(rune('0'+i))), SourceActionID: &branchB, TargetActionID: convergence, EdgeType: EdgeTypeSequence, SortOrder: 50 + i},
		)
	}

	convRef := convergence
	edges = append(edges,
		ActionEdge{ID: deterministicUUID("edge-conv-end"), SourceActionID: &convRef, TargetActionID: end, EdgeType: EdgeTypeSequence, SortOrder: 100},
	)

	return GraphDefinition{Actions: actions, Edges: edges}
}

// buildMixedEdgeTypeGraph creates a graph using all 5 edge types.
//
//	start -> condition --(true_branch)--> trueAction --(sequence)--> merge
//	                   --(false_branch)--> falseAction --(sequence)--> merge
//	                   --(always)--> alwaysAction --(sequence)--> merge
func buildMixedEdgeTypeGraph() (GraphDefinition, uuid.UUID) {
	start := deterministicUUID("mixed-start")
	condition := deterministicUUID("mixed-cond")
	trueAction := deterministicUUID("mixed-true")
	falseAction := deterministicUUID("mixed-false")
	alwaysAction := deterministicUUID("mixed-always")
	merge := deterministicUUID("mixed-merge")

	condRef := condition
	trueRef := trueAction
	falseRef := falseAction
	alwaysRef := alwaysAction

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: start, Name: "start", ActionType: "set_field"},
			{ID: condition, Name: "condition", ActionType: "evaluate_condition"},
			{ID: trueAction, Name: "true_action", ActionType: "send_email"},
			{ID: falseAction, Name: "false_action", ActionType: "create_alert"},
			{ID: alwaysAction, Name: "always_action", ActionType: "set_field"},
			{ID: merge, Name: "merge", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("me-start"), SourceActionID: nil, TargetActionID: start, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("me-seq"), SourceActionID: &start, TargetActionID: condition, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("me-true"), SourceActionID: &condRef, TargetActionID: trueAction, EdgeType: EdgeTypeTrueBranch, SortOrder: 2},
			{ID: deterministicUUID("me-false"), SourceActionID: &condRef, TargetActionID: falseAction, EdgeType: EdgeTypeFalseBranch, SortOrder: 3},
			{ID: deterministicUUID("me-always"), SourceActionID: &condRef, TargetActionID: alwaysAction, EdgeType: EdgeTypeAlways, SortOrder: 4},
			{ID: deterministicUUID("me-true-merge"), SourceActionID: &trueRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 5},
			{ID: deterministicUUID("me-false-merge"), SourceActionID: &falseRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 6},
			{ID: deterministicUUID("me-always-merge"), SourceActionID: &alwaysRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 7},
		},
	}

	return graph, condition
}

// =============================================================================
// Determinism Stress Tests
// =============================================================================

func TestDeterminism_ComplexParallel_1000(t *testing.T) {
	t.Parallel()
	iterations := 1000
	if testing.Short() {
		iterations = 100
	}

	graph := buildComplexParallelGraph(5)
	exec := NewGraphExecutor(graph)

	baselineIDs := extractIDs(exec.GetStartActions())
	for i := 0; i < iterations; i++ {
		currentIDs := extractIDs(exec.GetStartActions())
		assertIDsEqual(t, i, baselineIDs, currentIDs)
	}
}

func TestDeterminism_GetNextActions_ComplexParallel_1000(t *testing.T) {
	t.Parallel()
	iterations := 1000
	if testing.Short() {
		iterations = 100
	}

	graph := buildComplexParallelGraph(5)
	exec := NewGraphExecutor(graph)

	// Get condition node (second in the chain after start).
	starts := exec.GetStartActions()
	conditionActions := exec.GetNextActions(starts[0].ID, nil)
	if len(conditionActions) != 1 {
		t.Fatalf("expected 1 condition action, got %d", len(conditionActions))
	}
	conditionID := conditionActions[0].ID

	// Get the parallel branches from condition.
	baselineIDs := extractIDs(exec.GetNextActions(conditionID, nil))
	if len(baselineIDs) != 5 {
		t.Fatalf("expected 5 parallel branches, got %d", len(baselineIDs))
	}

	for i := 0; i < iterations; i++ {
		currentIDs := extractIDs(exec.GetNextActions(conditionID, nil))
		assertIDsEqual(t, i, baselineIDs, currentIDs)
	}
}

func TestDeterminism_FindConvergencePoint_LargeGraph(t *testing.T) {
	t.Parallel()
	iterations := 1000
	if testing.Short() {
		iterations = 100
	}

	graph := buildComplexParallelGraph(10)
	exec := NewGraphExecutor(graph)

	// Navigate to the condition node and get parallel branches.
	starts := exec.GetStartActions()
	condActions := exec.GetNextActions(starts[0].ID, nil)
	condID := condActions[0].ID
	branches := exec.GetNextActions(condID, nil)

	if len(branches) != 10 {
		t.Fatalf("expected 10 parallel branches, got %d", len(branches))
	}

	baseline := exec.FindConvergencePoint(branches)
	if baseline == nil {
		t.Fatal("expected convergence point for 10-branch parallel graph")
	}

	for i := 0; i < iterations; i++ {
		result := exec.FindConvergencePoint(branches)
		if result == nil {
			t.Fatalf("iteration %d: convergence point became nil", i)
		}
		if result.ID != baseline.ID {
			t.Fatalf("iteration %d: non-deterministic convergence\nbaseline: %s\ncurrent:  %s",
				i, baseline.ID, result.ID)
		}
	}
}

func TestDeterminism_GetNextActions_AllEdgeTypes(t *testing.T) {
	t.Parallel()
	iterations := 1000
	if testing.Short() {
		iterations = 100
	}

	graph, conditionID := buildMixedEdgeTypeGraph()
	exec := NewGraphExecutor(graph)

	// Test with true_branch result.
	trueResult := map[string]any{"branch_taken": "true_branch"}
	baselineTrueIDs := extractIDs(exec.GetNextActions(conditionID, trueResult))

	// Test with false_branch result.
	falseResult := map[string]any{"branch_taken": "false_branch"}
	baselineFalseIDs := extractIDs(exec.GetNextActions(conditionID, falseResult))

	for i := 0; i < iterations; i++ {
		currentTrueIDs := extractIDs(exec.GetNextActions(conditionID, trueResult))
		assertIDsEqual(t, i, baselineTrueIDs, currentTrueIDs)

		currentFalseIDs := extractIDs(exec.GetNextActions(conditionID, falseResult))
		assertIDsEqual(t, i, baselineFalseIDs, currentFalseIDs)
	}
}

func TestDeterminism_ManyActions_SameSort(t *testing.T) {
	t.Parallel()
	iterations := 1000
	if testing.Short() {
		iterations = 100
	}

	// Create 8 actions all with SortOrder=0, fed from a single source.
	// UUID string sort must provide deterministic ordering.
	source := deterministicUUID("same-sort-source")
	actions := []ActionNode{
		{ID: source, Name: "source", ActionType: "set_field"},
	}
	edges := []ActionEdge{
		{ID: deterministicUUID("same-sort-start"), SourceActionID: nil, TargetActionID: source, EdgeType: EdgeTypeStart, SortOrder: 0},
	}

	for i := range 8 {
		label := "same-sort-target-" + string(rune('a'+i))
		target := deterministicUUID(label)
		actions = append(actions, ActionNode{ID: target, Name: label, ActionType: "send_email"})

		srcRef := source
		edges = append(edges, ActionEdge{
			ID:             deterministicUUID("same-sort-edge-" + label),
			SourceActionID: &srcRef,
			TargetActionID: target,
			EdgeType:       EdgeTypeSequence,
			SortOrder:      0, // All same SortOrder
		})
	}

	graph := GraphDefinition{Actions: actions, Edges: edges}
	exec := NewGraphExecutor(graph)

	baselineIDs := extractIDs(exec.GetNextActions(source, nil))
	if len(baselineIDs) != 8 {
		t.Fatalf("expected 8 next actions, got %d", len(baselineIDs))
	}

	for i := 0; i < iterations; i++ {
		currentIDs := extractIDs(exec.GetNextActions(source, nil))
		assertIDsEqual(t, i, baselineIDs, currentIDs)
	}
}
