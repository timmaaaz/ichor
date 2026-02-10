# Phase 10: Graph Executor Unit Tests

**Category**: Testing
**Status**: Pending
**Dependencies**: Phase 4 (Graph Executor)

---

## Overview

Comprehensive determinism and consistency testing for the graph executor. Phase 4 delivered `graph_executor.go` with 38 passing tests covering core functionality (edge types, branching, convergence, PathLeadsTo, determinism × 100 iterations). Phase 10 fills identified gaps: stress-testing determinism at 1000+ iterations on complex graphs, covering missing graph patterns (cycles, deep chains, mixed edge types, orphaned nodes), and hardening convergence detection (tie-breaking, multiple convergence points, pathological structures).

### Existing Test Coverage (38 tests in `graph_executor_test.go`)

Already covered — **do not duplicate**:

| Area | Count | What's Tested |
|------|-------|---------------|
| GetStartActions | 3 | Linear, empty, multiple (SortOrder) |
| GetNextActions (edge types) | 11 | sequence, true_branch, false_branch, always (×2), end-of-path, parallel, no-edges, missing-target, start-filtered, no-matching-branch |
| FindConvergencePoint | 5 | Diamond, fire-and-forget, single branch, asymmetric depth, empty |
| HasMultipleIncoming | 3 | Convergence node, normal node, start node |
| PathLeadsTo | 4 | Direct, no-path, same-node, diamond cross |
| GetAction | 2 | Exists, not-exists |
| Graph() | 1 | Returns definition |
| calculateMinDepth | 4 | Direct, two-hop, unreachable, same-node |
| Determinism | 3 | Start actions × 100, next actions × 100, convergence × 100 |
| SortOrder | 1 | Edge ordering |

## Goals

1. **Stress-test determinism** — Run same graph 1000+ iterations on complex parallel structures; verify identical output every time (critical for Temporal replay safety)
2. **Cover missing graph patterns** — Cycles, deep chains (100 and 1000 nodes), mixed edge type combinations, orphaned nodes, multiple starts feeding divergent branches
3. **Harden convergence detection** — Tie-breaking at equal depths, multiple convergence points in series, condition-based diamond patterns, pathological graph shapes

**Runtime expectation**: Full Phase 10 test suite should complete in <60s. Individual tests should complete in <5s except stress tests (which use `testing.Short()` to reduce iterations in CI).

## Prerequisites

- Phase 4 complete (graph_executor.go + graph_executor_test.go)
- Models from Phase 3 (ActionNode, ActionEdge, GraphDefinition, EdgeType* constants)
- No external dependencies (pure unit tests — no Temporal SDK, no database, no network)
- Go 1.23+ (project minimum version)

### Cycle Handling Behavior (Verified from Source)

All three BFS methods in `graph_executor.go` use `visited` maps that prevent revisiting nodes, so cycles terminate safely:

- **`findReachableNodes`** (line 207): BFS with `visited` map. A cycle like A→B→C→A terminates because A is already visited when the cycle edge arrives. All cycle members ARE included in the reachable set.
- **`calculateMinDepth`** (line 234): BFS with `visited` map. Returns the shortest path depth; cycle back-edges are skipped. If target is only reachable through a cycle, returns -1 (unreachable after visited pruning).
- **`PathLeadsTo`** (line 271): Delegates to `findReachableNodes`, so inherits cycle safety. Returns `true` for any node in the cycle ring.

**Expected cycle test behavior**:
| Method | Cycle Input | Expected Result |
|--------|-------------|-----------------|
| `findReachableNodes(A)` where A→B→C→A | A | `{A:true, B:true, C:true}` |
| `PathLeadsTo(A, C)` where A→B→C→A | A→C | `true` |
| `PathLeadsTo(A, D)` where A→B→C→A, D unconnected | A→D | `false` |
| `calculateMinDepth(A, C)` where A→B→C→A | A→C | `2` (A→B→C) |
| `FindConvergencePoint([B,D])` where B is in cycle, D→E | B,D | `nil` (no shared node) or shared node if exists |

---

## Task Breakdown

### Task 1: Determinism Stress Tests

**Status**: Pending

**Description**: Extend the existing 100-iteration determinism tests to 1000 iterations on significantly more complex graph structures. The existing tests use simple 3-4 node graphs. These tests use realistic parallel/diamond structures with 10+ nodes to verify that sorted map iteration holds under pressure.

**Files**:
- `business/sdk/workflow/temporal/graph_executor_determinism_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "testing"

    "github.com/google/uuid"
)

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
// start -> condition -> N branches (each 2-3 nodes deep) -> convergence -> end
// Uses deterministicUUID for reproducibility.
func buildComplexParallelGraph(branchCount int) GraphDefinition {
    // Implementation builds start, condition, N*2-3 branch nodes, convergence, end
}

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

func TestDeterminism_FindConvergencePoint_LargeGraph(t *testing.T) {
    t.Parallel()
    // 10 parallel branches, some with sub-branches
    // FindConvergencePoint must return same node every time
    // Run 1000x (100x in short mode)
}

func TestDeterminism_GetNextActions_AllEdgeTypes(t *testing.T) {
    t.Parallel()
    // Graph with all 5 edge types in play
    // Run GetNextActions from each node 1000x
}

func TestDeterminism_ManyActions_SameSort(t *testing.T) {
    t.Parallel()
    // Multiple actions with identical SortOrder (all SortOrder=0)
    // UUID string sort must still be deterministic
}
```

**Key Testing Concerns**:
- Run with `-count=10` to exercise Go's map randomization across test invocations
- Use `t.Parallel()` where possible for faster execution
- Baseline comparison pattern: capture first result, compare all subsequent iterations
- Use `testing.Short()` to reduce iterations to 100 for quick feedback: `go test -short`
- Failure messages must include iteration number and both baseline/current values for debugging

---

### Task 2: Edge Type Coverage Tests

**Status**: Pending

**Description**: Cover edge type combinations and patterns NOT tested in Phase 4. Focus on multi-edge scenarios, chained conditions, and mixed edge types on the same source node.

**Files**:
- `business/sdk/workflow/temporal/graph_executor_edges_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "testing"

    "github.com/google/uuid"
)

// --- Multiple Always Edges ---

func TestGetNextActions_MultipleAlwaysEdges(t *testing.T) {
    // Source -> A (true_branch) + B (always) + C (always)
    // With result["branch_taken"] == "true_branch":
    //   expect [A, B, C] in SortOrder
}

func TestGetNextActions_AlwaysWithSequence(t *testing.T) {
    // Source -> A (sequence) + B (always)
    // expect both followed
}

// --- Chained Conditions ---

func TestGetNextActions_ConditionChain(t *testing.T) {
    // Condition1 -> Condition2 (true_branch) -> Action (true_branch)
    // Verify each condition independently dispatches correct branch
}

func TestGetNextActions_ConditionToCondition_FalseTrue(t *testing.T) {
    // Condition1 -> Condition2 (false_branch)
    // Condition2 -> Action (true_branch)
    // Verify traversal with different results at each step
}

// --- Mixed Edge Types on Diamond ---

func TestGetNextActions_ConditionDiamond(t *testing.T) {
    // Condition -> A (true_branch), B (false_branch)
    // A -> Merge (sequence)
    // B -> Merge (sequence)
    // Verify convergence works with condition-based entry
}

// --- Result Map Edge Cases (table-driven) ---

func TestGetNextActions_ResultMapEdgeCases(t *testing.T) {
    // Build graph: condition -> A (true_branch), B (false_branch), C (always)
    conditionID := deterministicUUID("condition")
    actionA := deterministicUUID("action-a")
    actionB := deterministicUUID("action-b")
    actionC := deterministicUUID("action-c")
    // ... build graph with these nodes and edges ...

    tests := []struct {
        name      string
        result    map[string]any
        expectIDs []uuid.UUID // expected action IDs in order
    }{
        {"nil result", nil, []uuid.UUID{actionC}},                                      // only always
        {"empty map", map[string]any{}, []uuid.UUID{actionC}},                          // only always
        {"empty branch_taken", map[string]any{"branch_taken": ""}, []uuid.UUID{actionC}},
        {"missing branch_taken key", map[string]any{"other": "val"}, []uuid.UUID{actionC}},
        {"unknown branch value", map[string]any{"branch_taken": "unknown"}, []uuid.UUID{actionC}},
        {"wrong type int", map[string]any{"branch_taken": 123}, []uuid.UUID{actionC}},        // type assertion fails
        {"wrong type bool", map[string]any{"branch_taken": true}, []uuid.UUID{actionC}},       // type assertion fails
        {"valid true_branch", map[string]any{"branch_taken": "true_branch"}, []uuid.UUID{actionA, actionC}},
        {"valid false_branch", map[string]any{"branch_taken": "false_branch"}, []uuid.UUID{actionB, actionC}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            exec := NewGraphExecutor(graph)
            next := exec.GetNextActions(conditionID, tt.result)
            gotIDs := extractIDs(next)
            // assert gotIDs matches tt.expectIDs
        })
    }
}
```

**Key Testing Concerns**:
- Each test should be self-contained with its own graph definition
- Table-driven tests for result map variations — covers nil, empty, missing key, wrong type, unknown value
- Wrong-type tests (int, bool) verify the `.(string)` type assertion in GetNextActions fails safely
- Verify SortOrder is respected when multiple edges are followed

---

### Task 3: Convergence Point Tests

**Status**: Pending

**Description**: Test advanced convergence detection patterns not covered by Phase 4's 5 convergence tests. Focus on tie-breaking, multiple convergence layers, pathological shapes, and large parallel structures.

**Files**:
- `business/sdk/workflow/temporal/graph_executor_convergence_test.go`

**Implementation Guide**:

```go
package temporal

import (
    "testing"

    "github.com/google/uuid"
)

// --- Tie-Breaking ---

func TestFindConvergencePoint_TieBreaking(t *testing.T) {
    // Two candidate convergence nodes at SAME depth
    // Verify consistent selection via UUID string sort
    // Run 100x to confirm determinism
}

// --- Multiple Convergence Points in Series ---

func TestFindConvergencePoint_SeriesConvergence(t *testing.T) {
    //   A
    //  / \
    // B   C
    //  \ /
    //   D    <- first convergence
    //  / \
    // E   F
    //  \ /
    //   G    <- second convergence
    // FindConvergencePoint([B,C]) should return D (closest)
    // FindConvergencePoint([E,F]) should return G
}

// --- Cycle Handling ---
// All BFS methods use visited maps — cycles terminate safely.
// See Prerequisites section for detailed expected behavior.

func TestFindReachableNodes_CycleInGraph(t *testing.T) {
    // Graph: A -> B -> C -> A (cycle), plus D unconnected
    // findReachableNodes(A) should return {A, B, C} (all cycle members)
    // findReachableNodes(D) should return {D} (not part of cycle)
    // Must complete within 1 second (timeout protection)
}

func TestPathLeadsTo_CycleInGraph(t *testing.T) {
    // Graph: A -> B -> C -> A (cycle)
    // PathLeadsTo(A, C) = true (C is reachable via A->B->C)
    // PathLeadsTo(A, A) = true (A is in its own reachable set - BFS starts at A)
    // PathLeadsTo(C, B) = true (C->A->B)
}

func TestCalculateMinDepth_CycleInGraph(t *testing.T) {
    // Graph: A -> B -> C -> A (cycle)
    // calculateMinDepth(A, C) = 2 (A->B->C, shortest path)
    // calculateMinDepth(A, A) = 0 (same node short-circuit)
}

func TestFindConvergencePoint_CycleWithBranches(t *testing.T) {
    // Graph: start -> [B, C]
    //        B -> D -> B (cycle on B's branch)
    //        C -> E
    //        D -> F, E -> F (convergence)
    // FindConvergencePoint([B, C]) should return F
    // Cycle on B's branch must not prevent convergence detection
}

// --- Deep Graphs (concrete sizes) ---

func TestFindConvergencePoint_DeepDiamond_100(t *testing.T) {
    // 50 nodes deep on EACH branch (100 total) before convergence node
    // buildDeepDiamond(50) helper generates this
    // Verify correct convergence point and BFS handles depth
}

func TestCalculateMinDepth_Chain100(t *testing.T) {
    // Linear chain: node0 -> node1 -> ... -> node99 (100 nodes)
    // calculateMinDepth(node0, node99) should return 99
}

func TestCalculateMinDepth_Chain1000(t *testing.T) {
    // Linear chain: node0 -> node1 -> ... -> node999 (1000 nodes)
    // calculateMinDepth(node0, node999) should return 999
    // Verifies iterative BFS handles large depth without stack overflow
}

// --- Orphaned Nodes ---

func TestFindConvergencePoint_OrphanedActions(t *testing.T) {
    // Graph has actions not connected by any edges
    // Should NOT affect convergence detection
}

func TestGetStartActions_OrphanedActions(t *testing.T) {
    // Actions exist in Actions[] but no start edges point to them
    // GetStartActions should only return start-edge targets
}

// --- Multiple Starts ---

func TestFindConvergencePoint_MultipleStartBranches(t *testing.T) {
    // Two independent start edges -> two chains -> shared convergence
    // FindConvergencePoint should find the shared node
}

// --- Pathological Shapes ---

func TestFindConvergencePoint_WideParallel(t *testing.T) {
    // 10+ parallel branches all converging to single node
    // Performance should remain acceptable
}

func TestFindConvergencePoint_AsymmetricFanOut(t *testing.T) {
    // One branch has 1 node, another has 10 nodes, both reach convergence
    // Closest convergence point should be correct
}

func TestFindConvergencePoint_DiamondWithExtraEdges(t *testing.T) {
    //   A
    //  / \
    // B   C
    // |   |
    // D   E   (D and E also connect to F which is NOT the convergence)
    //  \ /
    //   G     <- true convergence
    // Verify G selected, not F (which only partial branches reach)
}
```

**Key Testing Concerns**:
- Cycle tests are critical — BFS `visited` set should prevent infinite loops
- Deep graph tests verify iterative BFS (not recursive) handles depth
- Tie-breaking tests must run multiple iterations to confirm determinism
- Orphaned node tests verify structural robustness

---

## Validation Criteria

- [ ] All determinism tests pass (1000 iterations per test, 100 in short mode)
- [ ] All determinism tests pass with `-count=10` (randomized map seeds)
- [ ] All 5 edge types tested in isolation AND combination
- [ ] Result map edge cases covered (nil, empty, missing key, unknown value, wrong type int/bool)
- [ ] Convergence detection correct for: diamond, series, tie-breaking, wide parallel, asymmetric, deep
- [ ] Cycle handling verified: terminates correctly for findReachableNodes, calculateMinDepth, PathLeadsTo, FindConvergencePoint
- [ ] Deep chain tests pass at 100 AND 1000 nodes without stack overflow
- [ ] Orphaned node tests verify no interference
- [ ] `go test -timeout=60s ./business/sdk/workflow/temporal/... -count=1` passes (all new + existing tests)
- [ ] `go test -race ./business/sdk/workflow/temporal/...` clean
- [ ] `go vet ./business/sdk/workflow/temporal/...` clean
- [ ] No test flakiness across 5 consecutive runs
- [ ] Test coverage on `graph_executor.go` >=90% (check with `go test -coverprofile`)
- [ ] All test failures include actionable debugging info (iteration number, graph structure, expected vs actual)

---

## Deliverables

- `business/sdk/workflow/temporal/graph_executor_determinism_test.go`
- `business/sdk/workflow/temporal/graph_executor_edges_test.go`
- `business/sdk/workflow/temporal/graph_executor_convergence_test.go`

---

## Gotchas & Tips

### Common Pitfalls

- **UUID generation in tests**: Use `uuid.NewSHA1(uuid.NameSpaceDNS, []byte("action-N"))` (wrapped as `deterministicUUID("action-N")` helper) for reproducible UUIDs in determinism tests. Random UUIDs make debugging failures harder.
- **SortOrder ties**: When two actions have the same SortOrder, the executor uses `TargetActionID.String()` as a secondary sort key for deterministic tie-breaking. This was a bug fix applied during Phase 10 plan review — the original `sort.Slice` had no secondary key, which was a Temporal replay safety risk since `sort.Slice` is unstable. Tests (especially `TestDeterminism_ManyActions_SameSort`) verify this behavior.
- **Cycle handling is safe** (verified): All BFS methods use `visited` maps. Cycles like A→B→C→A terminate because the back-edge to A is skipped (already visited). All cycle members are included in reachable sets. See Prerequisites for full behavior table.
- **Test file organization**: Phase 4 tests are in `graph_executor_test.go`. Phase 10 tests go in NEW files (not appended to existing). This keeps Phase 4's test surface unchanged.
- **Don't duplicate existing tests**: All 38 Phase 4 tests are comprehensive for their scenarios. Phase 10 adds NEW patterns, not re-tests of existing ones.
- **Timeout protection**: Always run with `go test -timeout=60s` to catch potential infinite loops in cycle or deep graph tests. Individual cycle tests should complete in <1s.
- **branch_taken type assertion**: `GetNextActions` uses `result["branch_taken"].(string)` — if the value is not a string (int, bool, nil), the type assertion fails and the condition branch is NOT followed. Only `always` and `sequence` edges proceed. Tests must cover this.

### Test Naming Conventions

Use consistent naming patterns across all three test files:
- `TestDeterminism_{Feature}_{Iterations}` — e.g., `TestDeterminism_ComplexParallel_1000`
- `TestGetNextActions_{EdgePattern}` — e.g., `TestGetNextActions_MultipleAlwaysEdges`
- `TestFindConvergencePoint_{GraphShape}` — e.g., `TestFindConvergencePoint_DeepDiamond_100`
- `Test{Method}_{Scenario}` — e.g., `TestCalculateMinDepth_Chain1000`

### Tips

- Use helper functions to build complex graphs — reduces test boilerplate and makes patterns clearer (`buildComplexParallelGraph`, `buildDeepDiamond`, `buildLinearChain`)
- Table-driven tests for result map variations (Task 2) — see `TestGetNextActions_ResultMapEdgeCases` example
- `t.Parallel()` is safe here since GraphExecutor is read-only after construction
- `testing.Short()` to reduce stress test iterations for quick feedback: `go test -short`
- Run `go test -race ./business/sdk/workflow/temporal/...` to verify no data races in concurrent determinism tests
- Benchmark tests (optional): `BenchmarkFindConvergencePoint_WideParallel` to establish performance baselines

### Expected Failure Modes

These scenarios should return gracefully, NOT panic:
| Scenario | Expected Behavior |
|----------|-------------------|
| Graph with no start actions | `GetStartActions()` returns empty slice |
| Graph with no edges | `GetNextActions()` returns nil |
| Graph with cycle | BFS terminates, cycle members in reachable set |
| Graph with orphaned actions | Ignored by traversal methods |
| Result map with wrong types | Type assertion fails, condition branches not followed |
| 1000-node chain | BFS completes (iterative, not recursive) |

---

## Testing Strategy

### Unit Tests

All Phase 10 work is unit testing. No Temporal SDK, no database, no network — pure Go struct + method tests.

**Test organization**:
- `graph_executor_determinism_test.go` — Stress tests (high iteration count, complex graphs)
- `graph_executor_edges_test.go` — Edge type combinations and result map edge cases
- `graph_executor_convergence_test.go` — Advanced convergence patterns (cycles, depth, tie-breaking)

**Test execution**:
```bash
# Run all temporal tests (existing + new Phase 10) with timeout
go test -timeout=60s -v ./business/sdk/workflow/temporal/...

# Run only Phase 10 determinism tests
go test -timeout=60s -v -run "TestDeterminism_" ./business/sdk/workflow/temporal/...

# Quick mode (100 iterations instead of 1000)
go test -timeout=30s -short -v ./business/sdk/workflow/temporal/...

# Stress test with multiple seeds
go test -timeout=120s -count=10 -run "TestDeterminism_" ./business/sdk/workflow/temporal/...

# Race detection
go test -timeout=60s -race ./business/sdk/workflow/temporal/...

# Coverage report
go test -coverprofile=coverage.out ./business/sdk/workflow/temporal/...
go tool cover -func=coverage.out | grep graph_executor.go
```

### Integration Tests

Not applicable — Phase 10 is purely unit tests.

---

## Scope Boundary

### In Scope (Phase 10)
- GraphExecutor method testing with new graph patterns
- Determinism verification at 1000+ iterations
- Edge type combination coverage
- Convergence detection hardening
- Cycle handling verification

### Out of Scope (Other Phases)
- Workflow execution tests (Phase 11)
- Activity tests (Phase 12)
- Temporal SDK integration tests (Phase 11)
- Database integration tests (Phase 8, already complete)
- Performance optimization (future, if benchmarks reveal issues)

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 10

# Review plan before implementing
/workflow-temporal-plan-review 10

# Review code after implementing
/workflow-temporal-review 10
```
