package temporal

import (
	"testing"

	"github.com/google/uuid"
)

// =============================================================================
// Test Graph Builders
// =============================================================================

// buildLinearGraph creates: A -> B -> C (all sequence edges)
func buildLinearGraph() GraphDefinition {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	return GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "action_a", ActionType: "set_field"},
			{ID: b, Name: "action_b", ActionType: "send_email"},
			{ID: c, Name: "action_c", ActionType: "create_alert"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "sequence", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &b, TargetActionID: c, EdgeType: "sequence", SortOrder: 2},
		},
	}
}

// buildConditionGraph creates:
//
//	    A (condition)
//	   / \
//	  B   C
//	 (T) (F)
func buildConditionGraph() (GraphDefinition, uuid.UUID, uuid.UUID, uuid.UUID) {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "check_status", ActionType: "evaluate_condition"},
			{ID: b, Name: "true_action", ActionType: "send_email"},
			{ID: c, Name: "false_action", ActionType: "create_alert"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "true_branch", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: c, EdgeType: "false_branch", SortOrder: 2},
		},
	}

	return graph, a, b, c
}

// buildDiamondGraph creates:
//
//	      A
//	     / \
//	    B   C   (parallel branches via sequence edges)
//	     \ /
//	      D     (convergence point)
func buildDiamondGraph() (GraphDefinition, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "start_action", ActionType: "set_field"},
			{ID: b, Name: "branch_left", ActionType: "send_email"},
			{ID: c, Name: "branch_right", ActionType: "create_alert"},
			{ID: d, Name: "merge_action", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "sequence", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: c, EdgeType: "sequence", SortOrder: 2},
			{ID: uuid.New(), SourceActionID: &b, TargetActionID: d, EdgeType: "sequence", SortOrder: 3},
			{ID: uuid.New(), SourceActionID: &c, TargetActionID: d, EdgeType: "sequence", SortOrder: 4},
		},
	}

	return graph, a, b, c, d
}

// buildFireAndForgetGraph creates:
//
//	      A
//	     / \
//	    B   C   (B continues, C ends - no convergence)
//	    |
//	    D
func buildFireAndForgetGraph() (GraphDefinition, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "start_action", ActionType: "set_field"},
			{ID: b, Name: "main_path", ActionType: "send_email"},
			{ID: c, Name: "fire_forget", ActionType: "create_alert"},
			{ID: d, Name: "continue_main", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "sequence", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: c, EdgeType: "sequence", SortOrder: 2},
			{ID: uuid.New(), SourceActionID: &b, TargetActionID: d, EdgeType: "sequence", SortOrder: 3},
			// c has no outgoing edges - fire and forget
		},
	}

	return graph, a, b, c, d
}

// buildAlwaysEdgeGraph creates:
//
//	    A (condition)
//	   /|\
//	  B C D
//	 (T)(F)(always)
func buildAlwaysEdgeGraph() (GraphDefinition, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "check_status", ActionType: "evaluate_condition"},
			{ID: b, Name: "true_action", ActionType: "send_email"},
			{ID: c, Name: "false_action", ActionType: "create_alert"},
			{ID: d, Name: "always_action", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "true_branch", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: c, EdgeType: "false_branch", SortOrder: 2},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: d, EdgeType: "always", SortOrder: 3},
		},
	}

	return graph, a, b, c, d
}

// buildAsymmetricDiamondGraph creates:
//
//	      A
//	     / \
//	    B   C    (parallel)
//	    |
//	    D        (B -> D -> E, C -> E)
//	    |
//	    E        (convergence point, asymmetric depths: B=2 hops, C=1 hop)
func buildAsymmetricDiamondGraph() (GraphDefinition, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	e := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "start_action", ActionType: "set_field"},
			{ID: b, Name: "long_path", ActionType: "send_email"},
			{ID: c, Name: "short_path", ActionType: "create_alert"},
			{ID: d, Name: "intermediate", ActionType: "set_field"},
			{ID: e, Name: "merge_action", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "sequence", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: c, EdgeType: "sequence", SortOrder: 2},
			{ID: uuid.New(), SourceActionID: &b, TargetActionID: d, EdgeType: "sequence", SortOrder: 3},
			{ID: uuid.New(), SourceActionID: &d, TargetActionID: e, EdgeType: "sequence", SortOrder: 4},
			{ID: uuid.New(), SourceActionID: &c, TargetActionID: e, EdgeType: "sequence", SortOrder: 5},
		},
	}

	return graph, a, b, c, d, e
}

// =============================================================================
// GetStartActions Tests
// =============================================================================

func TestGetStartActions_Linear(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	if len(starts) != 1 {
		t.Fatalf("expected 1 start action, got %d", len(starts))
	}
	if starts[0].Name != "action_a" {
		t.Errorf("expected action_a, got %s", starts[0].Name)
	}
}

func TestGetStartActions_Empty(t *testing.T) {
	exec := NewGraphExecutor(GraphDefinition{})
	starts := exec.GetStartActions()
	if len(starts) != 0 {
		t.Errorf("expected 0 start actions, got %d", len(starts))
	}
}

func TestGetStartActions_MultipleInSortOrder(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "action_a", ActionType: "set_field"},
			{ID: b, Name: "action_b", ActionType: "send_email"},
			{ID: c, Name: "action_c", ActionType: "create_alert"},
		},
		Edges: []ActionEdge{
			// Defined out of SortOrder: 2, 0, 1
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: c, EdgeType: "start", SortOrder: 2},
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: b, EdgeType: "start", SortOrder: 1},
		},
	}

	exec := NewGraphExecutor(graph)
	starts := exec.GetStartActions()

	if len(starts) != 3 {
		t.Fatalf("expected 3 start actions, got %d", len(starts))
	}
	if starts[0].ID != a || starts[1].ID != b || starts[2].ID != c {
		t.Error("start actions not returned in SortOrder")
	}
}

// =============================================================================
// GetNextActions Tests - Edge Types
// =============================================================================

func TestGetNextActions_Sequence(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	next := exec.GetNextActions(starts[0].ID, nil)
	if len(next) != 1 {
		t.Fatalf("expected 1 next action, got %d", len(next))
	}
	if next[0].Name != "action_b" {
		t.Errorf("expected action_b, got %s", next[0].Name)
	}
}

func TestGetNextActions_TrueBranch(t *testing.T) {
	graph, a, b, _ := buildConditionGraph()
	exec := NewGraphExecutor(graph)

	trueResult := map[string]any{"branch_taken": "true_branch"}
	next := exec.GetNextActions(a, trueResult)

	if len(next) != 1 {
		t.Fatalf("expected 1 action for true_branch, got %d", len(next))
	}
	if next[0].ID != b {
		t.Error("expected true_action to be followed")
	}
}

func TestGetNextActions_FalseBranch(t *testing.T) {
	graph, a, _, c := buildConditionGraph()
	exec := NewGraphExecutor(graph)

	falseResult := map[string]any{"branch_taken": "false_branch"}
	next := exec.GetNextActions(a, falseResult)

	if len(next) != 1 {
		t.Fatalf("expected 1 action for false_branch, got %d", len(next))
	}
	if next[0].ID != c {
		t.Error("expected false_action to be followed")
	}
}

func TestGetNextActions_AlwaysEdge(t *testing.T) {
	graph, a, _, _, d := buildAlwaysEdgeGraph()
	exec := NewGraphExecutor(graph)

	// With true result: should follow true_branch AND always.
	trueResult := map[string]any{"branch_taken": "true_branch"}
	next := exec.GetNextActions(a, trueResult)

	if len(next) != 2 {
		t.Fatalf("expected 2 actions (true + always), got %d", len(next))
	}

	// Verify always action is included.
	foundAlways := false
	for _, action := range next {
		if action.ID == d {
			foundAlways = true
		}
	}
	if !foundAlways {
		t.Error("always_action should be in next actions")
	}
}

func TestGetNextActions_AlwaysEdge_WithFalseBranch(t *testing.T) {
	graph, a, _, _, d := buildAlwaysEdgeGraph()
	exec := NewGraphExecutor(graph)

	// With false result: should follow false_branch AND always.
	falseResult := map[string]any{"branch_taken": "false_branch"}
	next := exec.GetNextActions(a, falseResult)

	if len(next) != 2 {
		t.Fatalf("expected 2 actions (false + always), got %d", len(next))
	}

	foundAlways := false
	for _, action := range next {
		if action.ID == d {
			foundAlways = true
		}
	}
	if !foundAlways {
		t.Error("always_action should be in next actions with false branch too")
	}
}

func TestGetNextActions_EndOfPath(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	// Get to the last action.
	starts := exec.GetStartActions()
	second := exec.GetNextActions(starts[0].ID, nil)
	third := exec.GetNextActions(second[0].ID, nil)

	// Last action should have no next actions.
	last := exec.GetNextActions(third[0].ID, nil)
	if len(last) != 0 {
		t.Errorf("expected 0 next actions at end of path, got %d", len(last))
	}
}

func TestGetNextActions_MultipleParallel(t *testing.T) {
	graph, a, _, _, _ := buildDiamondGraph()
	exec := NewGraphExecutor(graph)

	next := exec.GetNextActions(a, nil)
	if len(next) != 2 {
		t.Fatalf("expected 2 parallel actions, got %d", len(next))
	}
}

func TestGetNextActions_NoEdges(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	// Use an ID that doesn't exist as a source in any edge.
	result := exec.GetNextActions(uuid.New(), nil)
	if result != nil {
		t.Errorf("expected nil for action with no outgoing edges, got %d actions", len(result))
	}
}

func TestGetNextActions_MissingTargetAction(t *testing.T) {
	a := uuid.New()
	missing := uuid.New() // Not in Actions list

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "action_a", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: missing, EdgeType: "sequence", SortOrder: 1},
		},
	}

	exec := NewGraphExecutor(graph)
	next := exec.GetNextActions(a, nil)

	if len(next) != 0 {
		t.Errorf("expected 0 next actions (target missing), got %d", len(next))
	}
}

func TestGetNextActions_StartEdgesFiltered(t *testing.T) {
	// Verify start edges are never returned by GetNextActions,
	// even if they appear as outgoing edges from a source action.
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "action_a", ActionType: "set_field"},
			{ID: b, Name: "action_b", ActionType: "send_email"},
			{ID: c, Name: "action_c", ActionType: "create_alert"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			// A has a sequence edge to B and a (pathological) start edge to C.
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "sequence", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: c, EdgeType: "start", SortOrder: 2},
		},
	}

	exec := NewGraphExecutor(graph)
	next := exec.GetNextActions(a, nil)

	if len(next) != 1 {
		t.Fatalf("expected 1 next action (start edge filtered), got %d", len(next))
	}
	if next[0].ID != b {
		t.Error("expected action_b, start edge to action_c should be filtered")
	}
}

func TestGetNextActions_ConditionWithNoBranch(t *testing.T) {
	// Condition with no matching branch result should return nil.
	graph, a, _, _ := buildConditionGraph()
	exec := NewGraphExecutor(graph)

	// Result has no branch_taken key.
	next := exec.GetNextActions(a, map[string]any{})
	if len(next) != 0 {
		t.Errorf("expected 0 actions when no branch matches, got %d", len(next))
	}

	// Result is nil.
	next = exec.GetNextActions(a, nil)
	if len(next) != 0 {
		t.Errorf("expected 0 actions when result is nil for condition, got %d", len(next))
	}
}

// =============================================================================
// FindConvergencePoint Tests
// =============================================================================

func TestFindConvergencePoint_Diamond(t *testing.T) {
	graph, a, _, _, d := buildDiamondGraph()
	exec := NewGraphExecutor(graph)

	// Get the two branches from A.
	branches := exec.GetNextActions(a, nil)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}

	convergence := exec.FindConvergencePoint(branches)
	if convergence == nil {
		t.Fatal("expected convergence point")
	}
	if convergence.ID != d {
		t.Errorf("expected merge_action (D) as convergence, got %s", convergence.Name)
	}
}

func TestFindConvergencePoint_FireAndForget(t *testing.T) {
	graph, a, _, _, _ := buildFireAndForgetGraph()
	exec := NewGraphExecutor(graph)

	branches := exec.GetNextActions(a, nil)
	convergence := exec.FindConvergencePoint(branches)

	// No common downstream node -> fire-and-forget.
	if convergence != nil {
		t.Errorf("expected nil convergence for fire-and-forget, got %s", convergence.Name)
	}
}

func TestFindConvergencePoint_SingleBranch(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	convergence := exec.FindConvergencePoint(starts)

	// Single branch -> no convergence needed.
	if convergence != nil {
		t.Error("expected nil convergence for single branch")
	}
}

func TestFindConvergencePoint_AsymmetricDepth(t *testing.T) {
	graph, a, _, _, _, e := buildAsymmetricDiamondGraph()
	exec := NewGraphExecutor(graph)

	branches := exec.GetNextActions(a, nil)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}

	convergence := exec.FindConvergencePoint(branches)
	if convergence == nil {
		t.Fatal("expected convergence point for asymmetric diamond")
	}
	if convergence.ID != e {
		t.Errorf("expected merge_action (E) as convergence, got %s", convergence.Name)
	}
}

func TestFindConvergencePoint_EmptyGraph(t *testing.T) {
	exec := NewGraphExecutor(GraphDefinition{})

	// No branches at all.
	convergence := exec.FindConvergencePoint(nil)
	if convergence != nil {
		t.Error("expected nil convergence for nil branches")
	}

	// Empty branch list.
	convergence = exec.FindConvergencePoint([]ActionNode{})
	if convergence != nil {
		t.Error("expected nil convergence for empty branches")
	}
}

// =============================================================================
// HasMultipleIncoming Tests
// =============================================================================

func TestHasMultipleIncoming_ConvergenceNode(t *testing.T) {
	_, _, _, _, d := buildDiamondGraph()
	exec := NewGraphExecutor(GraphDefinition{
		Actions: []ActionNode{{ID: d}},
	})
	// Need to use the actual diamond graph for proper edge indexing.
	graph, _, _, _, d2 := buildDiamondGraph()
	exec = NewGraphExecutor(graph)

	if !exec.HasMultipleIncoming(d2) {
		t.Error("convergence node D should have multiple incoming edges")
	}
}

func TestHasMultipleIncoming_NormalNode(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	// Middle node in linear chain has exactly 1 incoming edge.
	starts := exec.GetStartActions()
	next := exec.GetNextActions(starts[0].ID, nil)
	if exec.HasMultipleIncoming(next[0].ID) {
		t.Error("middle node in linear chain should NOT have multiple incoming")
	}
}

func TestHasMultipleIncoming_StartNode(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	// Start node has 1 incoming (the start edge).
	if exec.HasMultipleIncoming(starts[0].ID) {
		t.Error("start node should NOT have multiple incoming edges")
	}
}

// =============================================================================
// PathLeadsTo Tests
// =============================================================================

func TestPathLeadsTo_DirectPath(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	next := exec.GetNextActions(starts[0].ID, nil)
	last := exec.GetNextActions(next[0].ID, nil)

	if !exec.PathLeadsTo(starts[0].ID, last[0].ID) {
		t.Error("A should lead to C in linear graph")
	}
}

func TestPathLeadsTo_NoPath(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	last := exec.GetNextActions(exec.GetNextActions(starts[0].ID, nil)[0].ID, nil)

	// Reverse direction should not have path.
	if exec.PathLeadsTo(last[0].ID, starts[0].ID) {
		t.Error("C should NOT lead to A (reverse direction)")
	}
}

func TestPathLeadsTo_SameNode(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	// A node should always reach itself.
	if !exec.PathLeadsTo(starts[0].ID, starts[0].ID) {
		t.Error("node should be reachable from itself")
	}
}

func TestPathLeadsTo_DiamondCrossPath(t *testing.T) {
	graph, a, _, _, d := buildDiamondGraph()
	exec := NewGraphExecutor(graph)

	// A should reach D through either branch.
	if !exec.PathLeadsTo(a, d) {
		t.Error("A should lead to D through diamond branches")
	}
}

// =============================================================================
// GetAction Tests
// =============================================================================

func TestGetAction_Exists(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	action, ok := exec.GetAction(graph.Actions[0].ID)
	if !ok {
		t.Fatal("expected action to exist")
	}
	if action.Name != graph.Actions[0].Name {
		t.Errorf("expected %s, got %s", graph.Actions[0].Name, action.Name)
	}
}

func TestGetAction_NotExists(t *testing.T) {
	exec := NewGraphExecutor(GraphDefinition{})

	_, ok := exec.GetAction(uuid.New())
	if ok {
		t.Error("expected action to not exist")
	}
}

// =============================================================================
// Graph Tests
// =============================================================================

func TestGraph_ReturnsDefinition(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	returned := exec.Graph()
	if len(returned.Actions) != len(graph.Actions) {
		t.Errorf("expected %d actions, got %d", len(graph.Actions), len(returned.Actions))
	}
	if len(returned.Edges) != len(graph.Edges) {
		t.Errorf("expected %d edges, got %d", len(graph.Edges), len(returned.Edges))
	}
}

// =============================================================================
// calculateMinDepth Tests
// =============================================================================

func TestCalculateMinDepth_DirectNeighbor(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	next := exec.GetNextActions(starts[0].ID, nil)

	depth := exec.calculateMinDepth(starts[0].ID, next[0].ID)
	if depth != 1 {
		t.Errorf("expected depth 1 for direct neighbor, got %d", depth)
	}
}

func TestCalculateMinDepth_TwoHops(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	last := exec.GetNextActions(exec.GetNextActions(starts[0].ID, nil)[0].ID, nil)

	depth := exec.calculateMinDepth(starts[0].ID, last[0].ID)
	if depth != 2 {
		t.Errorf("expected depth 2 for two hops, got %d", depth)
	}
}

func TestCalculateMinDepth_Unreachable(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	last := exec.GetNextActions(exec.GetNextActions(starts[0].ID, nil)[0].ID, nil)

	// Reverse direction: unreachable.
	depth := exec.calculateMinDepth(last[0].ID, starts[0].ID)
	if depth != -1 {
		t.Errorf("expected depth -1 for unreachable, got %d", depth)
	}
}

func TestCalculateMinDepth_SameNode(t *testing.T) {
	graph := buildLinearGraph()
	exec := NewGraphExecutor(graph)

	starts := exec.GetStartActions()
	depth := exec.calculateMinDepth(starts[0].ID, starts[0].ID)
	if depth != 0 {
		t.Errorf("expected depth 0 for same node, got %d", depth)
	}
}

// =============================================================================
// Determinism Test - CRITICAL for Temporal Replay
// =============================================================================

func TestDeterminism_GetStartActions(t *testing.T) {
	// Build graph ONCE, then test it 100 times.
	g, _, _, _, _ := buildDiamondGraph()
	exec := NewGraphExecutor(g)

	firstResult := exec.GetStartActions()
	for i := 0; i < 100; i++ {
		result := exec.GetStartActions()
		if len(result) != len(firstResult) {
			t.Fatalf("iteration %d: different start action count", i)
		}
		for j := range result {
			if result[j].ID != firstResult[j].ID {
				t.Fatalf("iteration %d: different start action order", i)
			}
		}
	}
}

func TestDeterminism_GetNextActions(t *testing.T) {
	g, a, _, _, _ := buildDiamondGraph()
	exec := NewGraphExecutor(g)

	firstResult := exec.GetNextActions(a, nil)
	for i := 0; i < 100; i++ {
		result := exec.GetNextActions(a, nil)
		if len(result) != len(firstResult) {
			t.Fatalf("iteration %d: different next action count", i)
		}
		for j := range result {
			if result[j].ID != firstResult[j].ID {
				t.Fatalf("iteration %d: different next action order at index %d", i, j)
			}
		}
	}
}

func TestDeterminism_FindConvergencePoint(t *testing.T) {
	g, a, _, _, _ := buildDiamondGraph()
	exec := NewGraphExecutor(g)

	branches := exec.GetNextActions(a, nil)
	firstResult := exec.FindConvergencePoint(branches)

	for i := 0; i < 100; i++ {
		result := exec.FindConvergencePoint(branches)
		if (firstResult == nil) != (result == nil) {
			t.Fatalf("iteration %d: convergence nil mismatch", i)
		}
		if firstResult != nil && result.ID != firstResult.ID {
			t.Fatalf("iteration %d: different convergence point", i)
		}
	}
}

// =============================================================================
// SortOrder Tests
// =============================================================================

func TestSortOrder_EdgesReturnedInOrder(t *testing.T) {
	// Create a graph where edges are defined out of SortOrder.
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "start", ActionType: "set_field"},
			{ID: b, Name: "first", ActionType: "send_email"},
			{ID: c, Name: "second", ActionType: "create_alert"},
			{ID: d, Name: "third", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: uuid.New(), SourceActionID: nil, TargetActionID: a, EdgeType: "start", SortOrder: 0},
			// Edges out of order: 3, 1, 2
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: d, EdgeType: "sequence", SortOrder: 3},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: b, EdgeType: "sequence", SortOrder: 1},
			{ID: uuid.New(), SourceActionID: &a, TargetActionID: c, EdgeType: "sequence", SortOrder: 2},
		},
	}

	exec := NewGraphExecutor(graph)
	next := exec.GetNextActions(a, nil)

	if len(next) != 3 {
		t.Fatalf("expected 3 next actions, got %d", len(next))
	}

	// Should be returned in SortOrder: b(1), c(2), d(3).
	if next[0].Name != "first" {
		t.Errorf("expected first at index 0, got %s", next[0].Name)
	}
	if next[1].Name != "second" {
		t.Errorf("expected second at index 1, got %s", next[1].Name)
	}
	if next[2].Name != "third" {
		t.Errorf("expected third at index 2, got %s", next[2].Name)
	}
}
