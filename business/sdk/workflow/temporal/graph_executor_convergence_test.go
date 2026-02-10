package temporal

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// =============================================================================
// Graph Builder Helpers
// =============================================================================

// buildDeepDiamond creates a diamond with `depth` nodes on each branch before convergence.
//
//	start -> [branch_left_0 -> ... -> branch_left_{depth-1}] -> convergence
//	      -> [branch_right_0 -> ... -> branch_right_{depth-1}] -> convergence
func buildDeepDiamond(depth int) (GraphDefinition, uuid.UUID, uuid.UUID) {
	start := deterministicUUID("deep-start")
	convergence := deterministicUUID("deep-convergence")

	actions := []ActionNode{
		{ID: start, Name: "start", ActionType: "set_field"},
		{ID: convergence, Name: "convergence", ActionType: "set_field"},
	}

	edges := []ActionEdge{
		{ID: deterministicUUID("deep-start-edge"), SourceActionID: nil, TargetActionID: start, EdgeType: EdgeTypeStart, SortOrder: 0},
	}

	// Build left branch.
	prevLeft := start
	for i := range depth {
		nodeID := deterministicUUID(fmt.Sprintf("deep-left-%d", i))
		actions = append(actions, ActionNode{ID: nodeID, Name: fmt.Sprintf("left_%d", i), ActionType: "set_field"})

		src := prevLeft
		edges = append(edges, ActionEdge{
			ID: deterministicUUID(fmt.Sprintf("deep-edge-left-%d", i)), SourceActionID: &src, TargetActionID: nodeID,
			EdgeType: EdgeTypeSequence, SortOrder: i + 1,
		})
		prevLeft = nodeID
	}
	leftLast := prevLeft
	edges = append(edges, ActionEdge{
		ID: deterministicUUID("deep-left-conv"), SourceActionID: &leftLast, TargetActionID: convergence,
		EdgeType: EdgeTypeSequence, SortOrder: depth + 1,
	})

	// Build right branch.
	prevRight := start
	for i := range depth {
		nodeID := deterministicUUID(fmt.Sprintf("deep-right-%d", i))
		actions = append(actions, ActionNode{ID: nodeID, Name: fmt.Sprintf("right_%d", i), ActionType: "set_field"})

		src := prevRight
		edges = append(edges, ActionEdge{
			ID: deterministicUUID(fmt.Sprintf("deep-edge-right-%d", i)), SourceActionID: &src, TargetActionID: nodeID,
			EdgeType: EdgeTypeSequence, SortOrder: depth + 100 + i,
		})
		prevRight = nodeID
	}
	rightLast := prevRight
	edges = append(edges, ActionEdge{
		ID: deterministicUUID("deep-right-conv"), SourceActionID: &rightLast, TargetActionID: convergence,
		EdgeType: EdgeTypeSequence, SortOrder: depth*2 + 200,
	})

	return GraphDefinition{Actions: actions, Edges: edges}, start, convergence
}

// buildLinearChain creates a chain of n nodes: node0 -> node1 -> ... -> node_{n-1}.
func buildLinearChain(n int) (GraphDefinition, uuid.UUID, uuid.UUID) {
	actions := make([]ActionNode, n)
	edges := make([]ActionEdge, 0, n)

	for i := range n {
		actions[i] = ActionNode{
			ID:         deterministicUUID(fmt.Sprintf("chain-%d", i)),
			Name:       fmt.Sprintf("node_%d", i),
			ActionType: "set_field",
		}
	}

	// Start edge.
	edges = append(edges, ActionEdge{
		ID: deterministicUUID("chain-start"), SourceActionID: nil, TargetActionID: actions[0].ID,
		EdgeType: EdgeTypeStart, SortOrder: 0,
	})

	// Chain edges.
	for i := 0; i < n-1; i++ {
		src := actions[i].ID
		edges = append(edges, ActionEdge{
			ID: deterministicUUID(fmt.Sprintf("chain-edge-%d", i)), SourceActionID: &src, TargetActionID: actions[i+1].ID,
			EdgeType: EdgeTypeSequence, SortOrder: i + 1,
		})
	}

	return GraphDefinition{Actions: actions, Edges: edges}, actions[0].ID, actions[n-1].ID
}

// =============================================================================
// Tie-Breaking
// =============================================================================

func TestFindConvergencePoint_TieBreaking(t *testing.T) {
	t.Parallel()

	// Two candidate convergence nodes at SAME depth from both branches.
	//
	//       start
	//      /     \
	//     B       C
	//    / \     / \
	//   D   E   D   E     (D and E both reachable from both branches, both at depth 1)
	//
	// But we need to build this as explicit edges. Instead:
	//   start -> B, start -> C
	//   B -> D, B -> E
	//   C -> D, C -> E
	// Both D and E are convergence candidates at depth 1 from both branches.
	// UUID string sort determines which one wins.
	start := deterministicUUID("tie-start")
	b := deterministicUUID("tie-b")
	c := deterministicUUID("tie-c")
	d := deterministicUUID("tie-d")
	e := deterministicUUID("tie-e")

	startRef := start
	bRef := b
	cRef := c

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: start, Name: "start", ActionType: "set_field"},
			{ID: b, Name: "branch_b", ActionType: "set_field"},
			{ID: c, Name: "branch_c", ActionType: "set_field"},
			{ID: d, Name: "candidate_d", ActionType: "set_field"},
			{ID: e, Name: "candidate_e", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("tie-start-edge"), SourceActionID: nil, TargetActionID: start, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("tie-s-b"), SourceActionID: &startRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("tie-s-c"), SourceActionID: &startRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("tie-b-d"), SourceActionID: &bRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 3},
			{ID: deterministicUUID("tie-b-e"), SourceActionID: &bRef, TargetActionID: e, EdgeType: EdgeTypeSequence, SortOrder: 4},
			{ID: deterministicUUID("tie-c-d"), SourceActionID: &cRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 5},
			{ID: deterministicUUID("tie-c-e"), SourceActionID: &cRef, TargetActionID: e, EdgeType: EdgeTypeSequence, SortOrder: 6},
		},
	}

	exec := NewGraphExecutor(graph)
	branches := exec.GetNextActions(start, nil)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}

	// Run 100x to verify consistent tie-breaking.
	baseline := exec.FindConvergencePoint(branches)
	if baseline == nil {
		t.Fatal("expected convergence point")
	}

	for i := range 100 {
		result := exec.FindConvergencePoint(branches)
		if result == nil || result.ID != baseline.ID {
			t.Fatalf("iteration %d: non-deterministic tie-breaking\nbaseline: %s\ncurrent:  %v",
				i, baseline.ID, result)
		}
	}
}

// =============================================================================
// Multiple Convergence Points in Series
// =============================================================================

func TestFindConvergencePoint_SeriesConvergence(t *testing.T) {
	t.Parallel()

	//   A (start)
	//  / \
	// B   C
	//  \ /
	//   D    <- first convergence
	//  / \
	// E   F
	//  \ /
	//   G    <- second convergence
	a := deterministicUUID("series-a")
	b := deterministicUUID("series-b")
	c := deterministicUUID("series-c")
	d := deterministicUUID("series-d")
	e := deterministicUUID("series-e")
	f := deterministicUUID("series-f")
	g := deterministicUUID("series-g")

	aRef := a
	bRef := b
	cRef := c
	dRef := d
	eRef := e
	fRef := f

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "a", ActionType: "set_field"},
			{ID: b, Name: "b", ActionType: "set_field"},
			{ID: c, Name: "c", ActionType: "set_field"},
			{ID: d, Name: "d", ActionType: "set_field"},
			{ID: e, Name: "e", ActionType: "set_field"},
			{ID: f, Name: "f", ActionType: "set_field"},
			{ID: g, Name: "g", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("ser-start"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("ser-a-b"), SourceActionID: &aRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("ser-a-c"), SourceActionID: &aRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("ser-b-d"), SourceActionID: &bRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 3},
			{ID: deterministicUUID("ser-c-d"), SourceActionID: &cRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 4},
			{ID: deterministicUUID("ser-d-e"), SourceActionID: &dRef, TargetActionID: e, EdgeType: EdgeTypeSequence, SortOrder: 5},
			{ID: deterministicUUID("ser-d-f"), SourceActionID: &dRef, TargetActionID: f, EdgeType: EdgeTypeSequence, SortOrder: 6},
			{ID: deterministicUUID("ser-e-g"), SourceActionID: &eRef, TargetActionID: g, EdgeType: EdgeTypeSequence, SortOrder: 7},
			{ID: deterministicUUID("ser-f-g"), SourceActionID: &fRef, TargetActionID: g, EdgeType: EdgeTypeSequence, SortOrder: 8},
		},
	}

	exec := NewGraphExecutor(graph)

	// First diamond: branches [B, C] should converge at D (closest).
	branches1 := exec.GetNextActions(a, nil)
	if len(branches1) != 2 {
		t.Fatalf("expected 2 first-level branches, got %d", len(branches1))
	}
	conv1 := exec.FindConvergencePoint(branches1)
	if conv1 == nil || conv1.ID != d {
		t.Fatalf("expected D as first convergence, got %v", conv1)
	}

	// Second diamond: branches [E, F] should converge at G.
	branches2 := exec.GetNextActions(d, nil)
	if len(branches2) != 2 {
		t.Fatalf("expected 2 second-level branches, got %d", len(branches2))
	}
	conv2 := exec.FindConvergencePoint(branches2)
	if conv2 == nil || conv2.ID != g {
		t.Fatalf("expected G as second convergence, got %v", conv2)
	}
}

// =============================================================================
// Cycle Handling
// =============================================================================

func TestFindReachableNodes_CycleInGraph(t *testing.T) {
	t.Parallel()

	// Graph: A -> B -> C -> A (cycle), plus D unconnected.
	a := deterministicUUID("cycle-a")
	b := deterministicUUID("cycle-b")
	c := deterministicUUID("cycle-c")
	d := deterministicUUID("cycle-d")

	aRef := a
	bRef := b
	cRef := c

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "a", ActionType: "set_field"},
			{ID: b, Name: "b", ActionType: "set_field"},
			{ID: c, Name: "c", ActionType: "set_field"},
			{ID: d, Name: "d", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("cyc-start"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("cyc-a-b"), SourceActionID: &aRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("cyc-b-c"), SourceActionID: &bRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("cyc-c-a"), SourceActionID: &cRef, TargetActionID: a, EdgeType: EdgeTypeSequence, SortOrder: 3},
		},
	}

	exec := NewGraphExecutor(graph)

	// All cycle members should be reachable from A.
	reachable := exec.findReachableNodes(a)
	for _, nodeID := range []uuid.UUID{a, b, c} {
		if !reachable[nodeID] {
			t.Errorf("expected %s to be reachable from A in cycle", nodeID)
		}
	}
	if reachable[d] {
		t.Error("expected D to NOT be reachable from A (unconnected)")
	}

	// D only reaches itself.
	reachableD := exec.findReachableNodes(d)
	if len(reachableD) != 1 || !reachableD[d] {
		t.Errorf("expected only D reachable from D, got %v", reachableD)
	}
}

func TestPathLeadsTo_CycleInGraph(t *testing.T) {
	t.Parallel()

	a := deterministicUUID("cyc-pl-a")
	b := deterministicUUID("cyc-pl-b")
	c := deterministicUUID("cyc-pl-c")

	aRef := a
	bRef := b
	cRef := c

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "a", ActionType: "set_field"},
			{ID: b, Name: "b", ActionType: "set_field"},
			{ID: c, Name: "c", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("cyc-pl-start"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("cyc-pl-ab"), SourceActionID: &aRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("cyc-pl-bc"), SourceActionID: &bRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("cyc-pl-ca"), SourceActionID: &cRef, TargetActionID: a, EdgeType: EdgeTypeSequence, SortOrder: 3},
		},
	}

	exec := NewGraphExecutor(graph)

	// All cycle members should be reachable from each other.
	if !exec.PathLeadsTo(a, c) {
		t.Error("A should lead to C via A->B->C")
	}
	if !exec.PathLeadsTo(a, a) {
		t.Error("A should lead to A (itself)")
	}
	if !exec.PathLeadsTo(c, b) {
		t.Error("C should lead to B via C->A->B")
	}
}

func TestCalculateMinDepth_CycleInGraph(t *testing.T) {
	t.Parallel()

	a := deterministicUUID("cyc-md-a")
	b := deterministicUUID("cyc-md-b")
	c := deterministicUUID("cyc-md-c")

	aRef := a
	bRef := b
	cRef := c

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "a", ActionType: "set_field"},
			{ID: b, Name: "b", ActionType: "set_field"},
			{ID: c, Name: "c", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("cyc-md-start"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("cyc-md-ab"), SourceActionID: &aRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("cyc-md-bc"), SourceActionID: &bRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("cyc-md-ca"), SourceActionID: &cRef, TargetActionID: a, EdgeType: EdgeTypeSequence, SortOrder: 3},
		},
	}

	exec := NewGraphExecutor(graph)

	// A -> B -> C: depth 2.
	depth := exec.calculateMinDepth(a, c)
	if depth != 2 {
		t.Errorf("expected depth 2 for A->C, got %d", depth)
	}

	// Same node: depth 0.
	depth = exec.calculateMinDepth(a, a)
	if depth != 0 {
		t.Errorf("expected depth 0 for A->A, got %d", depth)
	}
}

func TestFindConvergencePoint_CycleWithBranches(t *testing.T) {
	t.Parallel()

	// start -> [B, C]
	// B -> D -> B (cycle on B's branch)
	// D -> F
	// C -> E -> F (convergence at F)
	start := deterministicUUID("cyc-conv-start")
	b := deterministicUUID("cyc-conv-b")
	c := deterministicUUID("cyc-conv-c")
	d := deterministicUUID("cyc-conv-d")
	e := deterministicUUID("cyc-conv-e")
	f := deterministicUUID("cyc-conv-f")

	startRef := start
	bRef := b
	cRef := c
	dRef := d
	eRef := e

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: start, Name: "start", ActionType: "set_field"},
			{ID: b, Name: "b", ActionType: "set_field"},
			{ID: c, Name: "c", ActionType: "set_field"},
			{ID: d, Name: "d", ActionType: "set_field"},
			{ID: e, Name: "e", ActionType: "set_field"},
			{ID: f, Name: "f", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("cc-start"), SourceActionID: nil, TargetActionID: start, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("cc-s-b"), SourceActionID: &startRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("cc-s-c"), SourceActionID: &startRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("cc-b-d"), SourceActionID: &bRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 3},
			{ID: deterministicUUID("cc-d-b"), SourceActionID: &dRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 4}, // Cycle
			{ID: deterministicUUID("cc-d-f"), SourceActionID: &dRef, TargetActionID: f, EdgeType: EdgeTypeSequence, SortOrder: 5},
			{ID: deterministicUUID("cc-c-e"), SourceActionID: &cRef, TargetActionID: e, EdgeType: EdgeTypeSequence, SortOrder: 6},
			{ID: deterministicUUID("cc-e-f"), SourceActionID: &eRef, TargetActionID: f, EdgeType: EdgeTypeSequence, SortOrder: 7},
		},
	}

	exec := NewGraphExecutor(graph)
	branches := exec.GetNextActions(start, nil)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}

	conv := exec.FindConvergencePoint(branches)
	if conv == nil {
		t.Fatal("expected convergence point despite cycle")
	}
	if conv.ID != f {
		t.Errorf("expected F as convergence, got %s (%s)", conv.Name, conv.ID)
	}
}

// =============================================================================
// Deep Graphs
// =============================================================================

func TestFindConvergencePoint_DeepDiamond_100(t *testing.T) {
	t.Parallel()

	graph, start, convergence := buildDeepDiamond(50) // 50 nodes per branch = 100 total
	exec := NewGraphExecutor(graph)

	branches := exec.GetNextActions(start, nil)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}

	conv := exec.FindConvergencePoint(branches)
	if conv == nil {
		t.Fatal("expected convergence point for deep diamond")
	}
	if conv.ID != convergence {
		t.Errorf("expected convergence node, got %s", conv.Name)
	}
}

func TestCalculateMinDepth_Chain100(t *testing.T) {
	t.Parallel()

	graph, first, last := buildLinearChain(100)
	exec := NewGraphExecutor(graph)

	depth := exec.calculateMinDepth(first, last)
	if depth != 99 {
		t.Errorf("expected depth 99 for 100-node chain, got %d", depth)
	}
}

func TestCalculateMinDepth_Chain1000(t *testing.T) {
	t.Parallel()

	graph, first, last := buildLinearChain(1000)
	exec := NewGraphExecutor(graph)

	depth := exec.calculateMinDepth(first, last)
	if depth != 999 {
		t.Errorf("expected depth 999 for 1000-node chain, got %d", depth)
	}
}

// =============================================================================
// Orphaned Nodes
// =============================================================================

func TestFindConvergencePoint_OrphanedActions(t *testing.T) {
	t.Parallel()

	// Diamond graph + orphaned node not connected to anything.
	a := deterministicUUID("orph-a")
	b := deterministicUUID("orph-b")
	c := deterministicUUID("orph-c")
	d := deterministicUUID("orph-d")
	orphan := deterministicUUID("orph-orphan")

	aRef := a
	bRef := b
	cRef := c

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "start", ActionType: "set_field"},
			{ID: b, Name: "left", ActionType: "set_field"},
			{ID: c, Name: "right", ActionType: "set_field"},
			{ID: d, Name: "merge", ActionType: "set_field"},
			{ID: orphan, Name: "orphan", ActionType: "set_field"}, // Not connected
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("orph-start"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("orph-a-b"), SourceActionID: &aRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("orph-a-c"), SourceActionID: &aRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("orph-b-d"), SourceActionID: &bRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 3},
			{ID: deterministicUUID("orph-c-d"), SourceActionID: &cRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 4},
		},
	}

	exec := NewGraphExecutor(graph)
	branches := exec.GetNextActions(a, nil)

	conv := exec.FindConvergencePoint(branches)
	if conv == nil || conv.ID != d {
		t.Fatalf("orphaned node should not affect convergence: expected D, got %v", conv)
	}
}

func TestGetStartActions_OrphanedActions(t *testing.T) {
	t.Parallel()

	// Actions exist but no start edges point to them.
	a := deterministicUUID("orph-sa-a")
	b := deterministicUUID("orph-sa-b")
	orphan := deterministicUUID("orph-sa-orphan")

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "a", ActionType: "set_field"},
			{ID: b, Name: "b", ActionType: "set_field"},
			{ID: orphan, Name: "orphan", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("orph-sa-start"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
		},
	}

	exec := NewGraphExecutor(graph)
	starts := exec.GetStartActions()

	if len(starts) != 1 {
		t.Fatalf("expected 1 start action, got %d", len(starts))
	}
	if starts[0].ID != a {
		t.Errorf("expected action a as start, got %s", starts[0].Name)
	}
}

// =============================================================================
// Multiple Starts
// =============================================================================

func TestFindConvergencePoint_MultipleStartBranches(t *testing.T) {
	t.Parallel()

	// Two independent start edges -> two chains -> shared convergence.
	a := deterministicUUID("ms-a")
	b := deterministicUUID("ms-b")
	c := deterministicUUID("ms-c")
	d := deterministicUUID("ms-d")
	merge := deterministicUUID("ms-merge")

	aRef := a
	bRef := b
	cRef := c
	dRef := d

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "start_a", ActionType: "set_field"},
			{ID: b, Name: "start_b", ActionType: "set_field"},
			{ID: c, Name: "chain_a", ActionType: "set_field"},
			{ID: d, Name: "chain_b", ActionType: "set_field"},
			{ID: merge, Name: "merge", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("ms-start-a"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("ms-start-b"), SourceActionID: nil, TargetActionID: b, EdgeType: EdgeTypeStart, SortOrder: 1},
			{ID: deterministicUUID("ms-a-c"), SourceActionID: &aRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("ms-b-d"), SourceActionID: &bRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 3},
			{ID: deterministicUUID("ms-c-merge"), SourceActionID: &cRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 4},
			{ID: deterministicUUID("ms-d-merge"), SourceActionID: &dRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 5},
		},
	}

	exec := NewGraphExecutor(graph)

	// Use the intermediate nodes as branches (they share convergence at merge).
	conv := exec.FindConvergencePoint([]ActionNode{
		{ID: c, Name: "chain_a"},
		{ID: d, Name: "chain_b"},
	})
	if conv == nil || conv.ID != merge {
		t.Fatalf("expected merge as convergence, got %v", conv)
	}
}

// =============================================================================
// Pathological Shapes
// =============================================================================

func TestFindConvergencePoint_WideParallel(t *testing.T) {
	t.Parallel()

	// 10+ parallel branches all converging to single node.
	start := deterministicUUID("wide-start")
	merge := deterministicUUID("wide-merge")

	actions := []ActionNode{
		{ID: start, Name: "start", ActionType: "set_field"},
		{ID: merge, Name: "merge", ActionType: "set_field"},
	}

	startRef := start
	edges := []ActionEdge{
		{ID: deterministicUUID("wide-start-edge"), SourceActionID: nil, TargetActionID: start, EdgeType: EdgeTypeStart, SortOrder: 0},
	}

	branchNodes := make([]ActionNode, 12)
	for i := range 12 {
		nodeID := deterministicUUID(fmt.Sprintf("wide-branch-%d", i))
		branchNodes[i] = ActionNode{ID: nodeID, Name: fmt.Sprintf("branch_%d", i), ActionType: "set_field"}
		actions = append(actions, branchNodes[i])

		edges = append(edges,
			ActionEdge{ID: deterministicUUID(fmt.Sprintf("wide-s-%d", i)), SourceActionID: &startRef, TargetActionID: nodeID, EdgeType: EdgeTypeSequence, SortOrder: i + 1},
		)

		nRef := nodeID
		edges = append(edges,
			ActionEdge{ID: deterministicUUID(fmt.Sprintf("wide-%d-m", i)), SourceActionID: &nRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 100 + i},
		)
	}

	graph := GraphDefinition{Actions: actions, Edges: edges}
	exec := NewGraphExecutor(graph)

	branches := exec.GetNextActions(start, nil)
	if len(branches) != 12 {
		t.Fatalf("expected 12 parallel branches, got %d", len(branches))
	}

	conv := exec.FindConvergencePoint(branches)
	if conv == nil || conv.ID != merge {
		t.Fatalf("expected merge as convergence for 12 parallel branches, got %v", conv)
	}
}

func TestFindConvergencePoint_AsymmetricFanOut(t *testing.T) {
	t.Parallel()

	// One branch has 1 node, another has 5 nodes, both reach convergence.
	start := deterministicUUID("asym-start")
	short := deterministicUUID("asym-short")
	merge := deterministicUUID("asym-merge")

	actions := []ActionNode{
		{ID: start, Name: "start", ActionType: "set_field"},
		{ID: short, Name: "short_path", ActionType: "set_field"},
		{ID: merge, Name: "merge", ActionType: "set_field"},
	}

	startRef := start
	shortRef := short
	edges := []ActionEdge{
		{ID: deterministicUUID("asym-start-edge"), SourceActionID: nil, TargetActionID: start, EdgeType: EdgeTypeStart, SortOrder: 0},
		{ID: deterministicUUID("asym-s-short"), SourceActionID: &startRef, TargetActionID: short, EdgeType: EdgeTypeSequence, SortOrder: 1},
		{ID: deterministicUUID("asym-short-merge"), SourceActionID: &shortRef, TargetActionID: merge, EdgeType: EdgeTypeSequence, SortOrder: 100},
	}

	// Long branch: 5 nodes.
	prev := start
	for i := range 5 {
		nodeID := deterministicUUID(fmt.Sprintf("asym-long-%d", i))
		actions = append(actions, ActionNode{ID: nodeID, Name: fmt.Sprintf("long_%d", i), ActionType: "set_field"})

		pRef := prev
		edges = append(edges, ActionEdge{
			ID: deterministicUUID(fmt.Sprintf("asym-long-edge-%d", i)), SourceActionID: &pRef, TargetActionID: nodeID,
			EdgeType: EdgeTypeSequence, SortOrder: 10 + i,
		})
		prev = nodeID
	}
	prevRef := prev
	edges = append(edges, ActionEdge{
		ID: deterministicUUID("asym-long-merge"), SourceActionID: &prevRef, TargetActionID: merge,
		EdgeType: EdgeTypeSequence, SortOrder: 99,
	})

	graph := GraphDefinition{Actions: actions, Edges: edges}
	exec := NewGraphExecutor(graph)

	branches := exec.GetNextActions(start, nil)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches (short + long), got %d", len(branches))
	}

	conv := exec.FindConvergencePoint(branches)
	if conv == nil || conv.ID != merge {
		t.Fatalf("expected merge as convergence for asymmetric fan-out, got %v", conv)
	}
}

func TestFindConvergencePoint_DiamondWithExtraEdges(t *testing.T) {
	t.Parallel()

	//   A
	//  / \
	// B   C
	// |   |
	// D   E   (D and E also connect to F which is NOT the convergence)
	//  \ /
	//   G     <- true convergence
	//
	// D -> F, E -> F (partial convergence, but both B and C also reach G)
	a := deterministicUUID("extra-a")
	b := deterministicUUID("extra-b")
	c := deterministicUUID("extra-c")
	d := deterministicUUID("extra-d")
	e := deterministicUUID("extra-e")
	f := deterministicUUID("extra-f")
	g := deterministicUUID("extra-g")

	aRef := a
	bRef := b
	cRef := c
	dRef := d
	eRef := e

	graph := GraphDefinition{
		Actions: []ActionNode{
			{ID: a, Name: "a", ActionType: "set_field"},
			{ID: b, Name: "b", ActionType: "set_field"},
			{ID: c, Name: "c", ActionType: "set_field"},
			{ID: d, Name: "d", ActionType: "set_field"},
			{ID: e, Name: "e", ActionType: "set_field"},
			{ID: f, Name: "f", ActionType: "set_field"},
			{ID: g, Name: "g", ActionType: "set_field"},
		},
		Edges: []ActionEdge{
			{ID: deterministicUUID("ex-start"), SourceActionID: nil, TargetActionID: a, EdgeType: EdgeTypeStart, SortOrder: 0},
			{ID: deterministicUUID("ex-a-b"), SourceActionID: &aRef, TargetActionID: b, EdgeType: EdgeTypeSequence, SortOrder: 1},
			{ID: deterministicUUID("ex-a-c"), SourceActionID: &aRef, TargetActionID: c, EdgeType: EdgeTypeSequence, SortOrder: 2},
			{ID: deterministicUUID("ex-b-d"), SourceActionID: &bRef, TargetActionID: d, EdgeType: EdgeTypeSequence, SortOrder: 3},
			{ID: deterministicUUID("ex-c-e"), SourceActionID: &cRef, TargetActionID: e, EdgeType: EdgeTypeSequence, SortOrder: 4},
			{ID: deterministicUUID("ex-d-f"), SourceActionID: &dRef, TargetActionID: f, EdgeType: EdgeTypeSequence, SortOrder: 5},
			{ID: deterministicUUID("ex-e-f"), SourceActionID: &eRef, TargetActionID: f, EdgeType: EdgeTypeSequence, SortOrder: 6},
			{ID: deterministicUUID("ex-d-g"), SourceActionID: &dRef, TargetActionID: g, EdgeType: EdgeTypeSequence, SortOrder: 7},
			{ID: deterministicUUID("ex-e-g"), SourceActionID: &eRef, TargetActionID: g, EdgeType: EdgeTypeSequence, SortOrder: 8},
		},
	}

	exec := NewGraphExecutor(graph)
	branches := exec.GetNextActions(a, nil)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}

	conv := exec.FindConvergencePoint(branches)
	if conv == nil {
		t.Fatal("expected convergence point")
	}

	// Both F and G are reachable by both branches. The closest convergence
	// should be selected (both at depth 2 from B and C respectively).
	// Since both are at equal depth, UUID string sort determines the winner.
	// Either F or G is acceptable as long as it's consistent.
	if conv.ID != f && conv.ID != g {
		t.Errorf("expected F or G as convergence, got %s (%s)", conv.Name, conv.ID)
	}
}
