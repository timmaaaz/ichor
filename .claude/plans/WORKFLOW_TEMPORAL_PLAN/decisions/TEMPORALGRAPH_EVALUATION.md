# Temporalgraph Library Evaluation

**Date**: 2026-02-09
**Evaluator**: Claude (Phase 2 - Workflow Temporal Implementation)
**Library**: https://github.com/Nickqiaoo/temporalgraph/
**Listed on**: [Temporal Code Exchange](https://temporal.io/code-exchange/temporalgraph-graph-based-orchestration)

## Summary

The temporalgraph library is a well-designed Go library for composing Temporal workflows as DAGs with type-safe I/O, branching, and merge semantics. However, it is fundamentally incompatible with our core requirement: **runtime graph definition from database-stored workflow definitions**. The library requires compile-time Go function references and generic type parameters at graph construction time, making it unsuitable for our dynamic, UI-driven workflow engine.

## Decision

**DO NOT ADOPT** - Proceed with custom `GraphExecutor` implementation as outlined in the plan (Phases 3-6).

## Evaluation Results

### 1. Library Overview & Maintenance

| Criterion | Assessment |
|-----------|-----------|
| Repository | https://github.com/Nickqiaoo/temporalgraph/ |
| License | MIT (compatible) |
| Stars | 5 |
| Forks | 0 |
| Contributors | 2 |
| Language | Go 100% |
| Maturity | Early-stage, minimal adoption |
| Documentation | README + examples only, no API docs |
| Dependencies | Temporal Go SDK (which we already use) |

**Assessment**: The library is very early-stage with minimal community adoption. The low star count, zero forks, and only 2 contributors suggest this is not production-hardened. Using it would couple our core workflow engine to a young, potentially unstable dependency with no guarantee of maintenance.

### 2. Edge Type Support

Our required edge types and how temporalgraph handles them:

| Edge Type | Supported? | How? | Limitations |
|-----------|-----------|------|-------------|
| `start` | Partial | Has `START` constant for graph entry point | START is a reserved node, not an edge type. Our model uses start edges with `source_action_id = nil` |
| `sequence` | Yes | `AddEdge(startNode, endNode)` connects nodes | No edge type metadata - just connectivity |
| `true_branch` | No | `AddBranch` with condition function returns next node key | Condition must be a Go function, not a database-stored evaluation. No concept of typed edges |
| `false_branch` | No | Same as above - branching returns single destination | No first-class true/false branch distinction |
| `always` | No | Not applicable | All edges are implicitly "always" unless gated by a branch condition |
| `parallel` | Partial | Fan-out via multiple edges from one node | Works implicitly through graph structure, but no explicit "parallel" edge type |

**Assessment**: The library uses a fundamentally different branching model. Our system uses **typed edges** where the edge type determines flow (e.g., `true_branch` edge is followed when condition evaluates true). Temporalgraph uses **branch nodes** that return destination keys via Go callback functions. These are architecturally incompatible approaches.

Our `ShouldFollowEdge()` logic (from `executor.go:321-342`) dispatches on edge type:
```go
case EdgeTypeAlways, EdgeTypeSequence:
    return true
case EdgeTypeTrueBranch:
    return result.BranchTaken == EdgeTypeTrueBranch
case EdgeTypeFalseBranch:
    return result.BranchTaken == EdgeTypeFalseBranch
```

This semantic edge-based routing has no equivalent in temporalgraph.

### 3. Convergence Detection

| Criterion | Assessment |
|-----------|-----------|
| Wait-for-all semantics | Yes - `channelManager` tracks data/control predecessors |
| Multiple incoming edges | Yes - handled via predecessor maps |
| Diamond pattern | Yes - node waits for all predecessor completions |
| Fire-and-forget | Partial - all nodes must eventually reach END |

**Assessment**: This is temporalgraph's strongest feature for our use case. The `channelManager` maintains `dataPredecessors` and `controlPredecessors` maps, and nodes only execute when all predecessors have completed. This matches our convergence detection requirement.

However, our "fire-and-forget" pattern (parallel branches where some branches don't converge) would require workarounds since temporalgraph expects all execution paths to reach the END node.

### 4. Runtime Graph Definition (CRITICAL BLOCKER)

| Criterion | Assessment |
|-----------|-----------|
| Graph defined from data? | **NO** - requires Go function references |
| Dynamic dispatch? | **NO** - type-safe generics enforce compile-time types |
| Database-stored definitions? | **NOT SUPPORTED** |

**Assessment**: This is the **critical incompatibility**. The `AddNode` signature requires a Go function:

```go
func (g *graph) AddNode(key string, activity any, input any,
    acopt workflow.ActivityOptions, opts ...GraphAddNodeOpt) error
```

The `activity` parameter must be a function with signature `func(ctx workflow.Context, input I) (output O, error)`. This means:

1. **Every node requires a Go function at definition time** - we cannot construct a graph from database rows containing `action_type: "send_email"` and `config: {...}` without having a Go function reference for each possible action type.

2. **Generic type parameters fight dynamic dispatch** - `Graph[I, O]` requires knowing input/output types at compile time. Our actions have heterogeneous configs (`json.RawMessage`) and results (`map[string]interface{}`).

3. **Our workflow model is fundamentally data-driven** - Rules are created via UI, stored in PostgreSQL as `rule_actions` + `action_edges` rows, and interpreted at runtime. Every rule can have a completely different graph shape with different action types. There's no fixed graph structure to pre-register.

**Workaround Analysis**: We could theoretically:
- Use a single generic `executeAction(ctx, ActionActivityInput) (ActionActivityOutput, error)` activity
- Build the graph at runtime by calling `AddNode` with this single activity for each DB action
- Use `map[string]any` as the generic type parameters

But this would negate all of temporalgraph's type-safety benefits (the main value proposition), add complexity for wrapping/unwrapping, and create a fragile adapter layer between our data model and the library's expectations. We'd be fighting the library's design at every turn.

## Rationale

Three factors drive the DO NOT ADOPT decision:

1. **Runtime graph definition is a hard blocker** (Critical): Our entire workflow system is built around dynamic, UI-defined graphs stored in PostgreSQL. The library requires compile-time Go function references. No amount of adaptation can cleanly bridge this gap.

2. **Edge type model mismatch** (Major): Our typed-edge routing (`ShouldFollowEdge` dispatching on `EdgeType`) is architecturally incompatible with temporalgraph's branch-condition-returns-destination model. Adapting would require a non-trivial translation layer.

3. **Library maturity risk** (Moderate): With 5 stars, 0 forks, 2 contributors, and early-stage documentation, coupling our core engine to this library introduces maintenance risk without sufficient community support.

## Trade-offs

### What We Lose by Not Adopting
- Built-in convergence detection via `channelManager`
- Automatic parallel execution scheduling
- Type-safe I/O between workflow steps
- DAG validation at compile time

### What We Gain by Building Custom
- **Direct mapping from our data model** - `ActionNode`/`ActionEdge` from DB map directly to executor without translation layers
- **Semantic edge types as first-class citizens** - our `ShouldFollowEdge` pattern works naturally
- **Full control over execution semantics** - fire-and-forget, conditional branching, convergence all implemented exactly as needed
- **No external dependency risk** - no coupling to early-stage library
- **Simpler code** - no adapter layer, no type gymnastics to work around generics
- **Determinism guarantees we control** - sorted map iteration, no non-deterministic operations, all verified by our own tests

## Impact on Implementation Plan

**No changes to Phases 3-6 are needed.** The plan already includes a custom `GraphExecutor` implementation:

- **Phase 3** (Core Models): `WorkflowInput`, `GraphDefinition`, `MergedContext` models proceed as planned. These map directly from our database model.
- **Phase 4** (Graph Executor): Custom `GraphExecutor` with `GetStartActions`, `GetNextActions`, `FindConvergencePoint`, `HasMultipleIncoming` proceeds as planned. This is essentially what temporalgraph's `channelManager` does, but built for our data model.
- **Phase 5** (Workflow Implementation): `ExecuteGraphWorkflow` with Continue-As-New, versioning, parallel execution proceeds as planned.
- **Phase 6** (Activities & Async): `ActionHandler` interface, `ActionRegistry`, async completion pattern proceeds as planned.

The custom implementation in the plan is well-designed for our specific use case and will be simpler than adapting temporalgraph.

## Alternative: What We'll Build Instead

Our custom `GraphExecutor` (Phase 4) will:

1. **Accept `ActionNode[]` + `ActionEdge[]` directly from database queries** - no translation needed
2. **Index by node ID** for O(1) lookups of outgoing edges and incoming edges
3. **Implement `GetStartActions()`** - find edges with nil source (our `EdgeTypeStart`)
4. **Implement `GetNextActions(nodeID, result)`** - follow edges based on edge type semantics matching `ShouldFollowEdge` logic
5. **Implement `FindConvergencePoint()`** - BFS traversal to find nodes with multiple incoming edges
6. **Implement `HasMultipleIncoming(nodeID)`** - check if node is a convergence point
7. **Use sorted map iteration throughout** - deterministic for Temporal replay safety
8. **Be fully unit-testable** - determinism verified by running same graph 100+ times

This is ~200-300 lines of focused Go code, versus adapting a library that wasn't designed for our use case.

## References

- Our existing executor: `business/sdk/workflow/executor.go` (lines 118-342)
- Our edge type definitions: `business/sdk/workflow/models.go` (lines 388-392)
- Our graph validation: `app/domain/workflow/workflowsaveapp/graph.go`
- Temporalgraph source: `compose/` package (16 files)
- Temporalgraph Code Exchange listing: https://temporal.io/code-exchange/temporalgraph-graph-based-orchestration
