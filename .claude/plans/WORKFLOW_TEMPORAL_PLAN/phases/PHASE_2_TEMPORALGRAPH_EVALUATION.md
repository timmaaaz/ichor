# Phase 2: Temporalgraph Evaluation

**Category**: research
**Status**: Pending
**Dependencies**: None (can run in parallel with Phase 1 - research only, no code)

---

## Overview

Evaluate the [temporalgraph](https://github.com/saltosystems/temporalgraph) library for compatibility with our visual workflow graph requirements. The core question: can we use this library to interpret database-stored graph definitions (nodes + edges) at runtime, or do we need to build a custom graph executor?

This evaluation directly informs Phase 3+ architecture. If temporalgraph fits, we adopt it and build on its abstractions. If not, we proceed with a custom `GraphExecutor` implementation as outlined in the plan.

## Goals

1. **Evaluate temporalgraph library's compatibility with our visual graph edge types** - Determine if the library supports or can be extended to handle our required edge types: `start`, `sequence`, `true_branch`, `false_branch`, `always`, `parallel`
2. **Determine if runtime graph definition (from DB) is supported vs compile-time only** - Our workflows are defined at runtime through a UI and stored in PostgreSQL. The library must interpret database-stored graph definitions, not require compile-time workflow registration
3. **Document a clear recommendation: adopt temporalgraph OR build custom graph executor** - Produce a decision document with rationale, trade-offs, and any required adaptations

## Prerequisites

- Internet access to review the temporalgraph GitHub repository
- Understanding of our current workflow graph model (edges with `source_action_id`, `target_action_id`, `edge_type`)
- Understanding of our required edge types from `workflowsaveapp/model.go`

---

## Task Breakdown

### Task 1: Review temporalgraph Library Documentation and Source

**Status**: Pending

**Description**: Thoroughly review the temporalgraph library's README, documentation, source code, and examples to understand its architecture, API surface, and intended use cases.

**Notes**:
- Repository: https://github.com/saltosystems/temporalgraph
- Focus on: API design, graph definition mechanism, workflow registration pattern
- Determine if it's a thin wrapper around Temporal SDK or a full graph engine
- Check activity and maintenance status (last commit, open issues, release history)

**Files**: (none - research only)

**Evaluation Checklist**:

- [ ] README and documentation reviewed
- [ ] Source code structure understood
- [ ] API surface cataloged (key types, functions, interfaces)
- [ ] Examples reviewed and understood
- [ ] Maintenance status assessed (last commit, issues, releases)
- [ ] License compatibility confirmed (must be compatible with our project)
- [ ] Dependency footprint assessed (what does it pull in beyond Temporal SDK?)

---

### Task 2: Evaluate Edge Type Support

**Status**: Pending

**Description**: Determine if temporalgraph supports (or can be extended to support) all of our required edge types for visual workflow graphs.

**Notes**:
- Required edge types from our workflow model:
  - `start` - Initial edge from trigger to first action(s)
  - `sequence` - Linear flow from one action to the next
  - `true_branch` - Conditional branch when condition evaluates true
  - `false_branch` - Conditional branch when condition evaluates false
  - `always` - Unconditional execution (regardless of condition result)
  - `parallel` - Fork execution into concurrent branches
- Check if the library has a concept of edge types/labels
- Check if conditional branching is supported
- Check if parallel execution with fan-out is supported

**Files**: (none - research only)

**Evaluation Matrix**:

| Edge Type | Supported? | How? | Limitations |
|-----------|-----------|------|-------------|
| `start` | ? | | |
| `sequence` | ? | | |
| `true_branch` | ? | | |
| `false_branch` | ? | | |
| `always` | ? | | |
| `parallel` | ? | | |

**Key Questions**:
- Does the library support labeled/typed edges, or just connectivity?
- Can we add custom edge type semantics on top of the library?
- How does the library handle conditional branching (if at all)?

---

### Task 3: Evaluate Convergence Detection Capability

**Status**: Pending

**Description**: Determine if temporalgraph can detect convergence points - nodes where multiple parallel branches must rejoin before execution continues. This is critical for our parallel execution pattern.

**Notes**:
- Must detect nodes with multiple incoming edges (convergence points)
- Must support "wait for all branches" semantics before continuing
- Our custom implementation uses BFS traversal + `HasMultipleIncoming` check
- The library needs to either provide this or allow us to implement it on top

**Files**: (none - research only)

**Key Questions**:
- Does the library have a concept of "join" or "barrier" nodes?
- Can we query a node's incoming edge count?
- Does the library support "wait for all" semantics at convergence points?
- If not built-in, can we implement convergence detection using the library's graph data structures?

**Test Patterns to Evaluate**:

```
Diamond Pattern:
    A
   / \
  B   C     (parallel branches)
   \ /
    D       (convergence point - must wait for B and C)

Multi-Convergence:
    A
   /|\
  B C D     (3 parallel branches)
   \|/
    E       (convergence point - must wait for B, C, and D)

Fire-and-Forget:
    A
   / \
  B   C     (parallel branches)
  |         (B continues, C has no convergence)
  D
```

---

### Task 4: Evaluate Runtime Graph Definition Support

**Status**: Pending

**Description**: Determine if temporalgraph supports defining workflow graphs at runtime from database-stored definitions, or if it requires compile-time workflow registration.

**Notes**:
- Our graphs are defined at runtime via a UI and stored in PostgreSQL
- Graphs consist of `rule_actions` (nodes) and `action_edges` (edges) tables
- The library must interpret these definitions dynamically - we cannot pre-register every possible graph shape
- This is likely the most critical evaluation criterion

**Files**: (none - research only)

**Key Questions**:
- How does the library define graphs? Code-based registration or data-driven?
- Can we construct a graph at runtime from arbitrary node/edge data?
- Does the library require Go function references at graph definition time?
- Can we map our `ActionNode` + `ActionEdge` database models to the library's graph format?
- Is there a dynamic dispatch mechanism for activities based on action type?

**Our Data Model for Reference**:
```go
// From our database
type ActionNode struct {
    ID         uuid.UUID
    ActionType string    // e.g., "send_email", "set_field", "allocate_inventory"
    Config     map[string]interface{}
}

type ActionEdge struct {
    SourceActionID *uuid.UUID  // nil for start edges
    TargetActionID uuid.UUID
    EdgeType       string      // "start", "sequence", "true_branch", etc.
}
```

**Critical Test**: Can we take the above data model and construct a valid temporalgraph workflow from it at runtime?

---

### Task 5: Document Decision and Rationale

**Status**: Pending

**Description**: Write a comprehensive decision document capturing the evaluation results, recommendation, and rationale.

**Notes**:
- Document should be clear enough that someone not involved in the evaluation can understand the decision
- Include specific code examples if adopting the library
- Include migration path considerations

**Files**:
- `.claude/plans/WORKFLOW_TEMPORAL_PLAN/decisions/TEMPORALGRAPH_EVALUATION.md`

**Decision Document Structure**:

```markdown
# Temporalgraph Library Evaluation

## Summary
[1-2 sentence recommendation]

## Decision
[ADOPT / DO NOT ADOPT / ADOPT WITH MODIFICATIONS]

## Evaluation Results

### Edge Type Support
[Results from Task 2]

### Convergence Detection
[Results from Task 3]

### Runtime Graph Definition
[Results from Task 4]

### Maintenance & Quality
[Results from Task 1]

## Rationale
[Why this decision was made]

## Trade-offs
[What we gain/lose with this decision]

## Impact on Implementation Plan
[How this decision affects Phases 3-6]

## Alternative Considered
[If not adopting: what we'll build instead and why it's better for our use case]
```

---

## Validation Criteria

- [ ] Decision document created with clear recommendation (ADOPT / DO NOT ADOPT / ADOPT WITH MODIFICATIONS)
- [ ] All evaluation criteria addressed:
  - [ ] Edge type support assessed for all 6 edge types
  - [ ] Convergence detection capability evaluated
  - [ ] Runtime graph definition support evaluated
  - [ ] Library maintenance and quality assessed
- [ ] Rationale is well-documented and defensible
- [ ] Impact on Phases 3-6 is explicitly described
- [ ] If not adopting: custom implementation approach is outlined

---

## Deliverables

- `.claude/plans/WORKFLOW_TEMPORAL_PLAN/decisions/TEMPORALGRAPH_EVALUATION.md` - Decision document with clear recommendation

---

## Gotchas & Tips

### Common Pitfalls

- **Don't evaluate in a vacuum**: Our specific use case (runtime graph definition from DB) is unusual. Most Temporal workflow libraries assume compile-time workflow registration. A library that looks great in general might fail this specific requirement.
- **Don't confuse "graph library" with "workflow engine"**: temporalgraph might be a graph data structure library, not a full workflow execution engine. We need both: graph traversal AND Temporal workflow orchestration.
- **API stability matters**: If the library is young/unstable, we're coupling our core engine to something that might change. Check version, release history, and breaking change patterns.
- **Don't over-invest in making it fit**: If the library requires significant adaptation to handle our edge types and runtime definitions, building custom might be simpler and more maintainable.

### Tips

- Start with the runtime graph definition evaluation (Task 4) - this is the most likely blocker. If the library can't define graphs from data at runtime, the other evaluations don't matter.
- Review our existing `workflowsaveapp/graph.go` to understand the current graph validation patterns
- Review `business/sdk/workflow/workflow.go` to understand the current execution model
- The custom `GraphExecutor` in the plan (Phase 4) is already well-designed. The bar for adopting temporalgraph is that it must be clearly better than what we'd build ourselves.
- Check if the library handles Temporal's determinism requirements (sorted map iteration, no random, no time.Now, etc.)

---

## Testing Strategy

### Research Validation

This is a research phase - there are no unit or integration tests to write. Validation is through:

1. **Document completeness**: All evaluation criteria addressed
2. **Code examples**: Where relevant, include code snippets showing how (or why not) the library handles our requirements
3. **Peer review**: The decision document should be reviewable by the team

### Spike (Optional)

If the evaluation is inconclusive from documentation alone, consider writing a small spike:

```go
// spike_test.go - NOT committed to main codebase
func TestTemporalgraphSpike(t *testing.T) {
    // Try to construct a graph from our data model
    // Try to execute it with Temporal test suite
    // Document results
}
```

This spike should be throwaway code, not production quality.

---

## Reference: Our Current Workflow Graph Model

For context, here's how workflows are currently stored:

**Database Tables**:
- `workflow.automation_rules` - Rule definitions (entity, trigger type, conditions)
- `workflow.rule_actions` - Action nodes (action_type, config, rule_id)
- `workflow.action_edges` - Edges between actions (source_action_id, target_action_id, edge_type)

**Edge Types** (from `workflowsaveapp/model.go`):
```go
const (
    EdgeTypeStart      = "start"
    EdgeTypeSequence   = "sequence"
    EdgeTypeTrueBranch = "true_branch"
    EdgeTypeFalseBranch = "false_branch"
    EdgeTypeAlways     = "always"
    EdgeTypeParallel   = "parallel"
)
```

**Current Graph Validation** (from `workflowsaveapp/graph.go`):
- Validates DAG structure (no cycles)
- Validates all actions are reachable from start edges
- Validates edge types are valid
- Validates start edges have nil source_action_id

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 2

# Review plan before implementing
/workflow-temporal-plan-review 2

# Review code after implementing
/workflow-temporal-review 2
```
