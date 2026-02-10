# Phase 4: Graph Executor

**Category**: backend
**Status**: Pending
**Dependencies**: Phase 3 (Core Models & Context) - requires `ActionNode`, `ActionEdge`, `GraphDefinition` types

---

## Overview

Implement the `GraphExecutor` - the core graph traversal engine that interprets workflow DAGs stored in PostgreSQL. The executor indexes actions and edges for fast lookup, follows edges based on type and action result (mirroring the existing `ShouldFollowEdge` logic in `business/sdk/workflow/executor.go`), and detects convergence points where parallel branches rejoin.

**Critical requirement**: All operations must be deterministic for Temporal replay safety. Go maps have random iteration order, so every map iteration in this file must sort keys before processing.

## Goals

1. **Implement GraphExecutor with deterministic graph traversal that mirrors the existing `ShouldFollowEdge` logic for all 5 edge types** - `GetStartActions`, `GetNextActions`, `GetAction`, `Graph`, and `HasMultipleIncoming` with edge-type-aware dispatch (`sequence`/`always` always follow, `true_branch`/`false_branch` conditional, `start` skip)
2. **Build convergence point detection using BFS reachability analysis with sorted map iteration for Temporal replay safety** - `FindConvergencePoint` identifies where parallel branches rejoin, `findReachableNodes` performs BFS, `calculateMinDepth` selects the closest convergence point, `PathLeadsTo` checks reachability
3. **Validate determinism with comprehensive tests** - Same graph produces identical traversal results across 100+ iterations, all edge types tested, convergence detection verified for diamond, multi-branch, and fire-and-forget patterns

## Prerequisites

- Phase 3 complete: `models.go` with `ActionNode`, `ActionEdge`, `GraphDefinition` types available
- Understanding of existing `ShouldFollowEdge` logic (`business/sdk/workflow/executor.go:321-342`)
- Understanding of existing `sortEdgesByOrder` pattern (`business/sdk/workflow/executor.go:344-353`)

---

## Go Package Structure

```
business/sdk/workflow/temporal/
    models.go              <- Phase 3 (COMPLETED)
    models_test.go         <- Phase 3 (COMPLETED)
    graph_executor.go      <- THIS PHASE
    graph_executor_test.go <- THIS PHASE
    workflow.go            <- Phase 5
    activities.go          <- Phase 6
```

---

## Task Breakdown

### Task 1: Implement graph_executor.go

**Status**: Pending

**Description**: Create the `GraphExecutor` struct with indexed lookups and all traversal methods. This is the deterministic heart of the Temporal workflow engine - it decides which actions to execute next based on the graph structure and action results.

**Notes**:
- `GraphExecutor` struct with pre-built indexes: `actionsByID`, `edgesBySource`, `incomingCount`
- `NewGraphExecutor` builds indexes and pre-sorts edges by `SortOrder` for determinism
- `GetStartActions` returns actions targeted by start edges (SourceActionID == nil)
- `GetNextActions` dispatches on edge type, matching existing `ShouldFollowEdge` logic exactly
- `FindConvergencePoint` with **SORTED** map iteration (uuid string sort) - critical for Temporal replay
- `findReachableNodes` BFS traversal (recursive with visited set)
- `calculateMinDepth` BFS to find shortest path between two nodes
- `PathLeadsTo` convenience wrapper over `findReachableNodes`
- `HasMultipleIncoming` checks `incomingCount > 1`
- `GetAction` and `Graph` accessor methods

**Files**:
- `business/sdk/workflow/temporal/graph_executor.go`

**Implementation Guide**:

```go
package temporal

import (
    "sort"

    "github.com/google/uuid"
)

// Edge type constants matching business/sdk/workflow/models.go.
// Defined locally to avoid import cycle (workflow package will import
// temporal in Phase 7+).
const (
    edgeTypeStart       = "start"
    edgeTypeSequence    = "sequence"
    edgeTypeTrueBranch  = "true_branch"
    edgeTypeFalseBranch = "false_branch"
    edgeTypeAlways      = "always"
)

// GraphExecutor traverses the workflow graph respecting edge types.
// All operations are deterministic for Temporal replay safety.
//
// Edge traversal rules (matching business/sdk/workflow/executor.go ShouldFollowEdge):
//   - edgeTypeSequence:    Always follow
//   - edgeTypeAlways:      Always follow
//   - edgeTypeTrueBranch:  Follow if result["branch_taken"] == "true_branch"
//   - edgeTypeFalseBranch: Follow if result["branch_taken"] == "false_branch"
//   - edgeTypeStart:       Skip (only used for initial dispatch via GetStartActions)
type GraphExecutor struct {
    graph         GraphDefinition
    actionsByID   map[uuid.UUID]ActionNode
    edgesBySource map[uuid.UUID][]ActionEdge // source_action_id -> outgoing edges (sorted by SortOrder)
    incomingCount map[uuid.UUID]int          // target_action_id -> number of incoming edges
}

// NewGraphExecutor creates an executor from a graph definition.
// Builds indexes for O(1) lookups and pre-sorts edges for deterministic traversal.
func NewGraphExecutor(graph GraphDefinition) *GraphExecutor {
    e := &GraphExecutor{
        graph:         graph,
        actionsByID:   make(map[uuid.UUID]ActionNode),
        edgesBySource: make(map[uuid.UUID][]ActionEdge),
        incomingCount: make(map[uuid.UUID]int),
    }

    // Index actions by ID
    for _, action := range graph.Actions {
        e.actionsByID[action.ID] = action
    }

    // Index edges by source and count incoming edges
    for _, edge := range graph.Edges {
        if edge.SourceActionID != nil {
            e.edgesBySource[*edge.SourceActionID] = append(
                e.edgesBySource[*edge.SourceActionID], edge)
        } else {
            // Start edges (source_action_id = nil) stored under uuid.Nil
            e.edgesBySource[uuid.Nil] = append(e.edgesBySource[uuid.Nil], edge)
        }
        e.incomingCount[edge.TargetActionID]++
    }

    // Sort edges by SortOrder for deterministic execution
    // This matches sortEdgesByOrder in business/sdk/workflow/executor.go
    for sourceID := range e.edgesBySource {
        edges := e.edgesBySource[sourceID]
        sort.Slice(edges, func(i, j int) bool {
            return edges[i].SortOrder < edges[j].SortOrder
        })
        e.edgesBySource[sourceID] = edges
    }

    return e
}

// GetStartActions returns all actions targeted by start edges (source_action_id = nil).
// Returns actions in SortOrder of their start edges.
func (e *GraphExecutor) GetStartActions() []ActionNode {
    startEdges := e.edgesBySource[uuid.Nil]
    actions := make([]ActionNode, 0, len(startEdges))

    for _, edge := range startEdges {
        if edge.EdgeType == edgeTypeStart {
            if action, ok := e.actionsByID[edge.TargetActionID]; ok {
                actions = append(actions, action)
            }
        }
    }

    return actions
}

// GetNextActions returns the next actions to execute based on edge types and action result.
//
// Edge type behavior (mirrors ShouldFollowEdge in business/sdk/workflow/executor.go):
//   - edgeTypeSequence:    Always follow
//   - edgeTypeAlways:      Always follow (runs regardless of condition result)
//   - edgeTypeTrueBranch:  Follow if result["branch_taken"] == "true_branch"
//   - edgeTypeFalseBranch: Follow if result["branch_taken"] == "false_branch"
//   - edgeTypeStart:       Skip (start edges are only for initial dispatch)
//
// Returns nil if no outgoing edges exist (end of path) or if edges exist
// but none match (e.g., condition with no matching branch). Callers should
// treat nil as "nothing more to execute on this path."
// Multiple returned actions indicate parallel execution.
func (e *GraphExecutor) GetNextActions(sourceActionID uuid.UUID, result map[string]any) []ActionNode {
    edges := e.edgesBySource[sourceActionID]
    if len(edges) == 0 {
        return nil
    }

    var nextActions []ActionNode

    for _, edge := range edges {
        shouldFollow := false

        switch edge.EdgeType {
        case edgeTypeSequence:
            shouldFollow = true

        case edgeTypeAlways:
            shouldFollow = true

        case edgeTypeTrueBranch:
            if branch, ok := result["branch_taken"].(string); ok && branch == edgeTypeTrueBranch {
                shouldFollow = true
            }

        case edgeTypeFalseBranch:
            if branch, ok := result["branch_taken"].(string); ok && branch == edgeTypeFalseBranch {
                shouldFollow = true
            }

        case edgeTypeStart:
            // Start edges are only used for initial dispatch via GetStartActions
            continue
        }

        if shouldFollow {
            if action, ok := e.actionsByID[edge.TargetActionID]; ok {
                nextActions = append(nextActions, action)
            }
        }
    }

    return nextActions
}

// FindConvergencePoint detects if multiple branches lead to the same node.
// Returns the closest convergence node reachable by ALL branches, or nil
// if branches are fire-and-forget (no common downstream node).
//
// IMPORTANT: This function must be deterministic for Temporal replay.
// All map iterations sort keys by UUID string before processing.
func (e *GraphExecutor) FindConvergencePoint(branches []ActionNode) *ActionNode {
    if len(branches) <= 1 {
        return nil
    }

    // Track which nodes each branch can reach
    reachable := make(map[uuid.UUID]int) // node_id -> count of branches that reach it

    for _, branch := range branches {
        visited := e.findReachableNodes(branch.ID)
        for nodeID := range visited {
            reachable[nodeID]++
        }
    }

    // DETERMINISM: Sort node IDs before iteration
    nodeIDs := make([]uuid.UUID, 0, len(reachable))
    for nodeID := range reachable {
        nodeIDs = append(nodeIDs, nodeID)
    }
    sort.Slice(nodeIDs, func(i, j int) bool {
        return nodeIDs[i].String() < nodeIDs[j].String()
    })

    // Find the closest node reachable by ALL branches.
    // "Closest" means the smallest maximum depth across all branches -
    // we need every branch to reach the convergence point, so the
    // limiting factor is the longest path from any branch.
    var convergencePoint *ActionNode
    bestMaxDepth := -1

    for _, nodeID := range nodeIDs {
        count := reachable[nodeID]
        if count == len(branches) {
            // Calculate max depth across ALL branches to this candidate
            maxDepth := 0
            for _, branch := range branches {
                depth := e.calculateMinDepth(branch.ID, nodeID)
                if depth > maxDepth {
                    maxDepth = depth
                }
            }
            if bestMaxDepth == -1 || maxDepth < bestMaxDepth {
                bestMaxDepth = maxDepth
                action := e.actionsByID[nodeID]
                convergencePoint = &action
            }
        }
    }

    return convergencePoint
}

// findReachableNodes performs iterative BFS to find all nodes reachable from startID.
// Uses iterative approach (not recursive) to prevent stack overflow on deep graphs.
// Follows ALL outgoing edges regardless of type (condition results don't
// matter for reachability analysis - both branches are structurally reachable).
func (e *GraphExecutor) findReachableNodes(startID uuid.UUID) map[uuid.UUID]bool {
    visited := make(map[uuid.UUID]bool)
    queue := []uuid.UUID{startID}

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        if visited[current] {
            continue
        }
        visited[current] = true

        for _, edge := range e.edgesBySource[current] {
            if !visited[edge.TargetActionID] {
                queue = append(queue, edge.TargetActionID)
            }
        }
    }

    return visited
}

// calculateMinDepth calculates minimum edge count from source to target using BFS.
// Returns -1 if target is not reachable from source.
func (e *GraphExecutor) calculateMinDepth(sourceID, targetID uuid.UUID) int {
    if sourceID == targetID {
        return 0
    }

    visited := make(map[uuid.UUID]bool)
    queue := []struct {
        id    uuid.UUID
        depth int
    }{{sourceID, 0}}

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        if visited[current.id] {
            continue
        }
        visited[current.id] = true

        for _, edge := range e.edgesBySource[current.id] {
            if edge.TargetActionID == targetID {
                return current.depth + 1
            }
            queue = append(queue, struct {
                id    uuid.UUID
                depth int
            }{edge.TargetActionID, current.depth + 1})
        }
    }

    return -1
}

// PathLeadsTo checks if starting from actionID eventually reaches targetID.
func (e *GraphExecutor) PathLeadsTo(actionID, targetID uuid.UUID) bool {
    reachable := e.findReachableNodes(actionID)
    return reachable[targetID]
}

// Graph returns the underlying graph definition.
func (e *GraphExecutor) Graph() GraphDefinition {
    return e.graph
}

// GetAction retrieves an action by ID.
func (e *GraphExecutor) GetAction(id uuid.UUID) (ActionNode, bool) {
    action, ok := e.actionsByID[id]
    return action, ok
}

// HasMultipleIncoming returns true if the node has more than one incoming edge.
// This indicates a potential convergence point where parallel branches rejoin.
func (e *GraphExecutor) HasMultipleIncoming(actionID uuid.UUID) bool {
    return e.incomingCount[actionID] > 1
}
```

**Existing Code Mapping**:

| GraphExecutor Method | Existing Equivalent | Notes |
|---------------------|---------------------|-------|
| `GetNextActions` edge dispatch | `ShouldFollowEdge` (executor.go:321-342) | Same logic, different shape (returns actions vs bool) |
| Edge sort in `NewGraphExecutor` | `sortEdgesByOrder` (executor.go:344-353) | Bubble sort -> `sort.Slice` (same result, cleaner) |
| `GetStartActions` | Start edge handling in `ExecuteRuleActionsGraph` | Extracted into dedicated method |
| `FindConvergencePoint` | No equivalent (new capability) | Enables parallel branch execution |

---

### Task 2: Write Comprehensive Unit Tests

**Status**: Pending

**Description**: Write thorough unit tests covering all edge types, convergence detection, and determinism. The determinism test is especially critical - it runs the same graph 100 times and verifies identical results each time, catching any non-deterministic map iteration.

**Notes**:
- Test all 5 edge types in isolation (`start`, `sequence`, `true_branch`, `false_branch`, `always`)
- Test convergence detection for diamond pattern, asymmetric diamond, and fire-and-forget
- Test determinism: same graph produces same `GetStartActions`, `GetNextActions`, and `FindConvergencePoint` results 100 times
- Test `HasMultipleIncoming` for convergence candidates vs non-convergence nodes
- Test `PathLeadsTo` for reachability (forward and reverse)
- Test `calculateMinDepth` for shortest path
- Test empty graph edge cases (`GetStartActions`, `FindConvergencePoint` with nil/empty)
- Test `GetNextActions` with unknown source ID (no outgoing edges)
- Test start edges filtered from `GetNextActions` output
- Test asymmetric depth convergence (different path lengths to same node)

**Files**:
- `business/sdk/workflow/temporal/graph_executor_test.go`

**Implementation Guide**:

```go
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
//     A (condition)
//    / \
//   B   C
//  (T) (F)
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
//       A
//      / \
//     B   C   (parallel branches via sequence edges)
//      \ /
//       D     (convergence point)
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
//       A
//      / \
//     B   C   (B continues, C ends - no convergence)
//     |
//     D
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
//     A (condition)
//    /|\
//   B C D
//  (T)(F)(always)
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

    // With true result: should follow true_branch AND always
    trueResult := map[string]any{"branch_taken": "true_branch"}
    next := exec.GetNextActions(a, trueResult)

    if len(next) != 2 {
        t.Fatalf("expected 2 actions (true + always), got %d", len(next))
    }

    // Verify always action is included
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

func TestGetNextActions_EndOfPath(t *testing.T) {
    graph := buildLinearGraph()
    exec := NewGraphExecutor(graph)

    // Get to the last action
    starts := exec.GetStartActions()
    second := exec.GetNextActions(starts[0].ID, nil)
    third := exec.GetNextActions(second[0].ID, nil)

    // Last action should have no next actions
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

// =============================================================================
// FindConvergencePoint Tests
// =============================================================================

func TestFindConvergencePoint_Diamond(t *testing.T) {
    graph, a, _, _, d := buildDiamondGraph()
    exec := NewGraphExecutor(graph)

    // Get the two branches from A
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

    // No common downstream node -> fire-and-forget
    if convergence != nil {
        t.Errorf("expected nil convergence for fire-and-forget, got %s", convergence.Name)
    }
}

func TestFindConvergencePoint_SingleBranch(t *testing.T) {
    graph := buildLinearGraph()
    exec := NewGraphExecutor(graph)

    starts := exec.GetStartActions()
    convergence := exec.FindConvergencePoint(starts)

    // Single branch -> no convergence needed
    if convergence != nil {
        t.Error("expected nil convergence for single branch")
    }
}

// =============================================================================
// HasMultipleIncoming Tests
// =============================================================================

func TestHasMultipleIncoming_ConvergenceNode(t *testing.T) {
    graph, _, _, _, d := buildDiamondGraph()
    exec := NewGraphExecutor(graph)

    if !exec.HasMultipleIncoming(d) {
        t.Error("convergence node D should have multiple incoming edges")
    }
}

func TestHasMultipleIncoming_NormalNode(t *testing.T) {
    graph := buildLinearGraph()
    exec := NewGraphExecutor(graph)

    // Middle node in linear chain has exactly 1 incoming edge
    starts := exec.GetStartActions()
    next := exec.GetNextActions(starts[0].ID, nil)
    if exec.HasMultipleIncoming(next[0].ID) {
        t.Error("middle node in linear chain should NOT have multiple incoming")
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

    // Reverse direction should not have path
    if exec.PathLeadsTo(last[0].ID, starts[0].ID) {
        t.Error("C should NOT lead to A (reverse direction)")
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
// Determinism Test - CRITICAL for Temporal Replay
// =============================================================================

func TestDeterminism_GetStartActions(t *testing.T) {
    graph := buildDiamondGraph // Use function that generates random UUIDs
    // Build graph ONCE, then test it 100 times
    g, _, _, _, _ := graph()
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
// Additional Edge Case Tests
// =============================================================================

// buildAsymmetricDiamondGraph creates:
//       A
//      / \
//     B   C    (parallel)
//     |
//     D        (B -> D -> E, C -> E)
//     |
//     E        (convergence point, asymmetric depths: B=2 hops, C=1 hop)
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

func TestGetNextActions_NoEdges(t *testing.T) {
    graph := buildLinearGraph()
    exec := NewGraphExecutor(graph)

    // Use an ID that doesn't exist as a source in any edge
    result := exec.GetNextActions(uuid.New(), nil)
    if result != nil {
        t.Errorf("expected nil for action with no outgoing edges, got %d actions", len(result))
    }
}

func TestFindConvergencePoint_EmptyGraph(t *testing.T) {
    exec := NewGraphExecutor(GraphDefinition{})

    // No branches at all
    convergence := exec.FindConvergencePoint(nil)
    if convergence != nil {
        t.Error("expected nil convergence for nil branches")
    }

    // Empty branch list
    convergence = exec.FindConvergencePoint([]ActionNode{})
    if convergence != nil {
        t.Error("expected nil convergence for empty branches")
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
            // A has a sequence edge to B and a (pathological) start edge to C
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
```

---

## Validation Criteria

- [ ] `go build ./business/sdk/workflow/temporal/...` passes
- [ ] `go test ./business/sdk/workflow/temporal/...` passes - all unit tests green
- [ ] Determinism test: `GetStartActions` returns identical results 100 times
- [ ] Determinism test: `GetNextActions` returns identical results 100 times
- [ ] Determinism test: `FindConvergencePoint` returns identical results 100 times
- [ ] All 5 edge types tested: `start`, `sequence`, `true_branch`, `false_branch`, `always`
- [ ] Convergence detection correct for diamond pattern (returns merge node)
- [ ] Convergence detection correct for fire-and-forget (returns nil)
- [ ] Convergence detection correct for single branch (returns nil)
- [ ] `HasMultipleIncoming` returns true for convergence nodes, false for linear nodes
- [ ] `PathLeadsTo` returns true for connected nodes, false for unreachable nodes
- [ ] `GetNextActions` returns nil at end of path
- [ ] `GetAction` returns (action, true) for existing IDs, (zero, false) for missing
- [ ] Edge type constants defined locally (matching `business/sdk/workflow/models.go` values exactly)
- [ ] `findReachableNodes` uses iterative BFS (not recursive DFS)
- [ ] `FindConvergencePoint` selects convergence using max depth across ALL branches
- [ ] Convergence detection correct for asymmetric diamond (different path lengths)
- [ ] `GetNextActions` returns nil for unknown source IDs (no outgoing edges)
- [ ] `GetNextActions` filters start edges (never returned as next actions)
- [ ] `FindConvergencePoint` handles nil and empty branch inputs
- [ ] `go vet ./business/sdk/workflow/temporal/...` passes

---

## Deliverables

- `business/sdk/workflow/temporal/graph_executor.go` - Graph traversal engine
- `business/sdk/workflow/temporal/graph_executor_test.go` - Comprehensive unit tests

---

## Gotchas & Tips

### Common Pitfalls

- **Non-deterministic map iteration**: This is the #1 risk. Go maps iterate in random order. Every place you iterate over a map in graph_executor.go MUST sort keys first. The `FindConvergencePoint` method sorts `nodeIDs` before checking reachability counts. If you add any new map iteration, sort first.
- **`uuid.Nil` for start edges**: Start edges have `SourceActionID == nil`, but they're stored under `uuid.Nil` (all zeros) in `edgesBySource`. Don't confuse `nil` with `uuid.Nil` - they're different types.
- **Edge type constants**: Local `edgeType*` constants are defined in `graph_executor.go` to avoid importing `business/sdk/workflow` (which would create an import cycle in Phase 7+). These MUST match the `EdgeType*` constants in `business/sdk/workflow/models.go` exactly. If the upstream constants change, these must be updated too.
- **Convergence detection follows ALL edges**: `findReachableNodes` follows all outgoing edges regardless of type. This is intentional - structurally, both true_branch and false_branch paths exist even if only one is taken at runtime.
- **Convergence depth uses max across branches**: `FindConvergencePoint` picks the closest convergence node where "closest" is defined as the smallest _maximum_ depth from any branch. Using depth from only one branch would give wrong results for asymmetric graphs.
- **`SortOrder` ties**: If two edges have the same `SortOrder`, their order is undefined but stable within a single `sort.Slice` call. The database enforces unique edge orders per source, so ties shouldn't occur in practice.
- **Graph validation is NOT this phase's responsibility**: The `GraphExecutor` assumes it receives a valid graph (no dangling references, no orphan nodes). Validation happens at load time in the Phase 8 Edge Store adapter. If an edge references a non-existent action ID, the executor silently skips it (the `if action, ok := e.actionsByID[...]` check).
- **Dual graph executors during development**: Until Phase 9 wiring, the existing `workflow.ActionExecutor.ExecuteRuleActionsGraph` and the new `temporal.GraphExecutor` coexist. Do NOT modify the existing executor during Phases 4-8. Phase 9 wires the Temporal path; Phase 13 validates end-to-end.
- **Performance**: `NewGraphExecutor` builds indexes on every workflow task. For typical graphs (5-20 nodes), this is negligible. For graphs approaching 100+ nodes: indexing is O(N+E), edge sorting is O(E log E) per source. If graph loading becomes a bottleneck (>10ms in profiling), consider caching the executor in workflow-local state.
- **Graph size limits**: Temporal has a 50K history event limit; each action generates ~2-3 events. Combined with Continue-As-New at 10K events, practical max is ~3-5K actions per workflow execution. Phase 8 edge store should enforce reasonable limits (e.g., max 1000 actions per rule).

### Tips

- The reference implementation in `.claude/plans/workflow-temporal-implementation.md` (lines 303-555) has the complete code. Use it as the primary source.
- Compare `GetNextActions` edge dispatch with `ShouldFollowEdge` in `business/sdk/workflow/executor.go:321-342` to verify logic parity.
- The test helpers (`buildLinearGraph`, `buildDiamondGraph`, etc.) generate fresh UUIDs each call. This is good for testing - it ensures the executor doesn't depend on specific UUID values.
- Run determinism tests with `-count=10` to stress them: `go test -count=10 -run TestDeterminism ./business/sdk/workflow/temporal/...`
- The `GraphExecutor` has no Temporal SDK dependency - it's pure Go graph logic. This makes it easy to test without any infrastructure.

---

## Testing Strategy

### Unit Tests

Tests in `graph_executor_test.go` cover:

1. **Graph builders**: Reusable helpers that create common graph patterns (linear, condition, diamond, asymmetric diamond, fire-and-forget, always-edge)
2. **GetStartActions**: Linear graph (1 start), empty graph (0 starts)
3. **GetNextActions per edge type**: Sequence (always follow), true_branch (conditional), false_branch (conditional), always (unconditional), end-of-path (nil return), multiple parallel (returns >1), no outgoing edges (nil), start edges filtered
4. **FindConvergencePoint**: Diamond (returns merge node), asymmetric diamond (correct despite different depths), fire-and-forget (nil), single branch (nil), empty/nil input (nil)
5. **HasMultipleIncoming**: Convergence node (true), linear node (false)
6. **PathLeadsTo**: Forward path (true), reverse direction (false)
7. **GetAction**: Existing ID (found), missing ID (not found)
8. **Determinism**: GetStartActions, GetNextActions, FindConvergencePoint all produce identical results across 100 iterations. Tests run within a single process - cross-process replay determinism validated in Phase 11.

### Temporal-Specific Patterns

The `GraphExecutor` itself doesn't use the Temporal SDK, but it's designed for Temporal determinism:
- All map iterations sorted
- No randomization
- No time-dependent logic
- No I/O or side effects
- Pure function of input graph + action result

Phase 10 adds stress tests (1000 iterations, complex parallel structures) and edge type combination tests.

---

## Existing Code Reference

### ShouldFollowEdge (`business/sdk/workflow/executor.go:321-342`)
```go
func (ae *ActionExecutor) ShouldFollowEdge(edge ActionEdge, result ActionResult) bool {
    switch edge.EdgeType {
    case EdgeTypeAlways, EdgeTypeSequence:
        return true
    case EdgeTypeTrueBranch:
        return result.BranchTaken == EdgeTypeTrueBranch
    case EdgeTypeFalseBranch:
        return result.BranchTaken == EdgeTypeFalseBranch
    case EdgeTypeStart:
        return false
    default:
        ae.log.Warn(context.Background(), "Unknown edge type", ...)
        return false
    }
}
```

### sortEdgesByOrder (`business/sdk/workflow/executor.go:344-353`)
```go
func sortEdgesByOrder(edges []ActionEdge) {
    for i := 0; i < len(edges)-1; i++ {
        for j := i + 1; j < len(edges); j++ {
            if edges[i].EdgeOrder > edges[j].EdgeOrder {
                edges[i], edges[j] = edges[j], edges[i]
            }
        }
    }
}
```

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 4

# Review plan before implementing
/workflow-temporal-plan-review 4

# Review code after implementing
/workflow-temporal-review 4
```
