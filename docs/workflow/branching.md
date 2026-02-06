# Graph-Based Workflow Execution (Branching)

This document explains how workflows can use graph-based execution with conditional branching instead of simple linear action sequences.

## Overview

Workflow rules execute actions using graph-based traversal with action edges. This enables:

- **Conditional branching** - Execute different actions based on condition results
- **Parallel entry points** - Start multiple action paths simultaneously
- **Convergence** - Multiple branches can lead to a common action
- **Complex flows** - Nested conditions, diamond patterns, and more

## Core Concepts

### Action Edges

Edges define how execution flows between actions. Each edge connects a source action to a target action with a specific type.

```
┌─────────────┐         ┌─────────────┐
│   Action A  │──edge──▶│   Action B  │
└─────────────┘         └─────────────┘
```

**Source**: `business/sdk/workflow/models.go:395-414`

### Edge Types

| Type | Constant | Description |
|------|----------|-------------|
| `start` | `EdgeTypeStart` | Entry point into the graph (source is nil) |
| `sequence` | `EdgeTypeSequence` | Linear progression, always followed |
| `always` | `EdgeTypeAlways` | Unconditional edge, always followed |
| `true_branch` | `EdgeTypeTrueBranch` | Only followed when condition evaluates to `true` |
| `false_branch` | `EdgeTypeFalseBranch` | Only followed when condition evaluates to `false` |

**Source**: `business/sdk/workflow/models.go:386-393`

### ActionEdge Model

```go
type ActionEdge struct {
    ID             uuid.UUID
    RuleID         uuid.UUID
    SourceActionID *uuid.UUID // nil for start edges
    TargetActionID uuid.UUID
    EdgeType       string     // start, sequence, true_branch, false_branch, always
    EdgeOrder      int        // For deterministic traversal
    CreatedDate    time.Time
}
```

## How Graph Execution Works

### Algorithm Overview

The executor uses **Breadth-First Search (BFS)** traversal:

1. Load all edges for the rule
2. If no edges exist, return an error (edges are required)
3. Build adjacency list: `source_action_id → [edges]`
4. Identify start edges (where `source_action_id` is nil)
5. Sort start edges by `edge_order`
6. Execute actions using BFS queue
7. After each action, determine which outgoing edges to follow based on `ShouldFollowEdge()`
8. Track executed actions to prevent cycles

**Source**: `business/sdk/workflow/executor.go:211-402`

### Edge Following Logic

```go
func ShouldFollowEdge(edge ActionEdge, result ActionResult) bool {
    switch edge.EdgeType {
    case EdgeTypeAlways, EdgeTypeSequence:
        return true  // Always follow
    case EdgeTypeTrueBranch:
        return result.BranchTaken == EdgeTypeTrueBranch
    case EdgeTypeFalseBranch:
        return result.BranchTaken == EdgeTypeFalseBranch
    case EdgeTypeStart:
        return false  // Start edges handled separately
    default:
        return false  // Unknown type
    }
}
```

**Source**: `business/sdk/workflow/executor.go:404-427`

### Edge Requirement

All rules with actions must have edges defining execution flow. The validation
layer enforces this at save time, and the executor returns an error at runtime
if edges are missing:

```go
if len(edges) == 0 {
    return BatchExecutionResult{
        RuleID: ruleID, Status: "failed",
        ErrorMessage: "rule has no edges - all rules with actions require edges",
    }, fmt.Errorf("executeRuleActionsGraph: rule %s has no edges", ruleID)
}
```

## Workflow Patterns

### Pattern 1: Linear Sequence

The simplest graph - actions execute in order.

```
[Start] ──start──▶ [Action A] ──sequence──▶ [Action B] ──sequence──▶ [Action C]
```

**Edges required**:
```json
[
  {"source_action_id": null, "target_action_id": "A", "edge_type": "start", "edge_order": 1},
  {"source_action_id": "A", "target_action_id": "B", "edge_type": "sequence", "edge_order": 1},
  {"source_action_id": "B", "target_action_id": "C", "edge_type": "sequence", "edge_order": 1}
]
```

### Pattern 2: Simple Branch (If/Else)

A condition determines which path to take.

```
                    ┌─────────────────┐
[Start] ──start──▶  │    Condition    │
                    │ (amount > 1000) │
                    └────────┬────────┘
               ┌─────────────┴─────────────┐
    true_branch│                           │false_branch
      ┌────────▼────────┐         ┌────────▼────────┐
      │ Manager Approve │         │   Auto Approve  │
      └─────────────────┘         └─────────────────┘
```

**Edges required**:
```json
[
  {"source_action_id": null, "target_action_id": "Condition", "edge_type": "start", "edge_order": 1},
  {"source_action_id": "Condition", "target_action_id": "ManagerApprove", "edge_type": "true_branch", "edge_order": 1},
  {"source_action_id": "Condition", "target_action_id": "AutoApprove", "edge_type": "false_branch", "edge_order": 2}
]
```

**Condition configuration** (evaluate_condition action):
```json
{
  "conditions": [
    {"field_name": "amount", "operator": "greater_than", "value": 1000}
  ],
  "logic_type": "and"
}
```

### Pattern 3: Diamond (Converging Branches)

Branches that converge to a common action.

```
                    ┌─────────────────┐
[Start] ──start──▶  │    Condition    │
                    │ (priority=high) │
                    └────────┬────────┘
               ┌─────────────┴─────────────┐
    true_branch│                           │false_branch
      ┌────────▼────────┐         ┌────────▼────────┐
      │   Fast Track    │         │  Normal Queue   │
      └────────┬────────┘         └────────┬────────┘
               │sequence                   │sequence
               └─────────────┬─────────────┘
                    ┌────────▼────────┐
                    │   Send Email    │
                    └─────────────────┘
```

Both branches lead to the same "Send Email" action. The convergence action executes only once (cycle prevention).

**Edges required**:
```json
[
  {"source_action_id": null, "target_action_id": "Condition", "edge_type": "start", "edge_order": 1},
  {"source_action_id": "Condition", "target_action_id": "FastTrack", "edge_type": "true_branch", "edge_order": 1},
  {"source_action_id": "Condition", "target_action_id": "NormalQueue", "edge_type": "false_branch", "edge_order": 2},
  {"source_action_id": "FastTrack", "target_action_id": "SendEmail", "edge_type": "sequence", "edge_order": 1},
  {"source_action_id": "NormalQueue", "target_action_id": "SendEmail", "edge_type": "sequence", "edge_order": 1}
]
```

### Pattern 4: Nested Conditions

Conditions within conditions for complex logic.

```
                         ┌─────────────┐
[Start] ──start──▶       │  Is Urgent? │
                         └──────┬──────┘
                    ┌───────────┴───────────┐
         true_branch│                       │false_branch
            ┌───────▼───────┐        ┌──────▼──────┐
            │ Priority > 5? │        │Queue Later  │
            └───────┬───────┘        └─────────────┘
       ┌────────────┴────────────┐
true   │                         │ false
  ┌────▼────┐             ┌──────▼──────┐
  │Escalate │             │Standard Proc│
  └─────────┘             └─────────────┘
```

**Edges required**:
```json
[
  {"source_action_id": null, "target_action_id": "IsUrgent", "edge_type": "start", "edge_order": 1},
  {"source_action_id": "IsUrgent", "target_action_id": "PriorityCheck", "edge_type": "true_branch", "edge_order": 1},
  {"source_action_id": "IsUrgent", "target_action_id": "QueueLater", "edge_type": "false_branch", "edge_order": 2},
  {"source_action_id": "PriorityCheck", "target_action_id": "Escalate", "edge_type": "true_branch", "edge_order": 1},
  {"source_action_id": "PriorityCheck", "target_action_id": "StandardProc", "edge_type": "false_branch", "edge_order": 2}
]
```

### Pattern 5: Multiple Entry Points

Start multiple independent action paths simultaneously.

```
         ┌──start (order 1)──▶ [Notify Manager]
[Start]──┤
         └──start (order 2)──▶ [Create Audit Log]
```

Both actions execute (in edge_order sequence).

**Edges required**:
```json
[
  {"source_action_id": null, "target_action_id": "NotifyManager", "edge_type": "start", "edge_order": 1},
  {"source_action_id": null, "target_action_id": "CreateAuditLog", "edge_type": "start", "edge_order": 2}
]
```

### Pattern 6: Conditional Skip (Gate)

Only execute an action if a condition is met.

```
                    ┌─────────────────┐
[Start] ──start──▶  │    Condition    │
                    │  (is_premium?)  │
                    └────────┬────────┘
                             │true_branch
                    ┌────────▼────────┐
                    │  Premium Perks  │
                    └─────────────────┘
```

If the condition is false, no `false_branch` edge exists, so execution stops.

## Edge Ordering

When multiple edges share the same source, `edge_order` determines processing order:

```go
// Edges are sorted by EdgeOrder before processing
sortEdgesByOrder(nextEdges)

for _, edge := range nextEdges {
    if shouldFollow && !executed[edge.TargetActionID] {
        queue = append(queue, edge.TargetActionID)
    }
}
```

Lower `edge_order` values execute first. This ensures deterministic behavior.

## Cycle Prevention

The executor tracks executed actions and prevents re-execution:

```go
executed := make(map[uuid.UUID]bool)

for len(queue) > 0 {
    actionID := queue[0]
    queue = queue[1:]

    if executed[actionID] {
        continue  // Skip already-executed actions
    }
    executed[actionID] = true

    // Execute action...
}
```

This prevents infinite loops even if edges create cycles.

## Creating Edges via API

### Create an Edge

```bash
POST /v1/workflow/rules/{ruleID}/edges
Content-Type: application/json

{
  "source_action_id": "uuid-of-source-action",  // null for start edges
  "target_action_id": "uuid-of-target-action",
  "edge_type": "sequence",
  "edge_order": 1
}
```

### List Edges for a Rule

```bash
GET /v1/workflow/rules/{ruleID}/edges
```

### Delete an Edge

```bash
DELETE /v1/workflow/rules/{ruleID}/edges/{edgeID}
```

### Delete All Edges

```bash
DELETE /v1/workflow/rules/{ruleID}/edges-all
```

See [api-reference.md](api-reference.md) for complete API documentation.

## Complete Example: Order Approval Workflow

This example implements a tiered approval system:

**Business Logic**:
- Orders under $1,000: Auto-approve
- Orders $1,000-$10,000: Manager approval
- Orders over $10,000: VP approval

### Actions

| Index | Name | Type | Config |
|-------|------|------|--------|
| 0 | Check Amount | evaluate_condition | `{"conditions": [{"field_name": "total", "operator": "greater_than", "value": 1000}]}` |
| 1 | Auto Approve | update_field | Sets status to "approved" |
| 2 | Check High Value | evaluate_condition | `{"conditions": [{"field_name": "total", "operator": "greater_than", "value": 10000}]}` |
| 3 | Manager Approval | send_email | Notify manager |
| 4 | VP Approval | send_email | Notify VP |
| 5 | Notify Customer | send_email | Always runs after approval path |

### Edges

```json
[
  {"source_action_id": null, "target_action_id": "0", "edge_type": "start", "edge_order": 1},
  {"source_action_id": "0", "target_action_id": "2", "edge_type": "true_branch", "edge_order": 1},
  {"source_action_id": "0", "target_action_id": "1", "edge_type": "false_branch", "edge_order": 2},
  {"source_action_id": "2", "target_action_id": "4", "edge_type": "true_branch", "edge_order": 1},
  {"source_action_id": "2", "target_action_id": "3", "edge_type": "false_branch", "edge_order": 2},
  {"source_action_id": "1", "target_action_id": "5", "edge_type": "sequence", "edge_order": 1},
  {"source_action_id": "3", "target_action_id": "5", "edge_type": "sequence", "edge_order": 1},
  {"source_action_id": "4", "target_action_id": "5", "edge_type": "sequence", "edge_order": 1}
]
```

### Visual Flow

```
                         ┌──────────────┐
[Start] ──start──▶       │ Check Amount │
                         │  (> $1000?)  │
                         └──────┬───────┘
                    ┌───────────┴───────────┐
                    │false                  │true
           ┌────────▼────────┐     ┌────────▼────────┐
           │   Auto Approve  │     │Check High Value │
           │                 │     │   (> $10000?)   │
           └────────┬────────┘     └────────┬────────┘
                    │                  ┌────┴────┐
                    │           false  │         │ true
                    │         ┌────────▼───┐  ┌──▼──────────┐
                    │         │  Manager   │  │     VP      │
                    │         │  Approval  │  │   Approval  │
                    │         └────────┬───┘  └──┬──────────┘
                    │                  │         │
                    └──────────────────┼─────────┘
                                       │
                              ┌────────▼────────┐
                              │ Notify Customer │
                              └─────────────────┘
```

## Best Practices

### Design Guidelines

1. **Always include start edges** - Every rule with actions must have at least one start edge

2. **Use meaningful edge_order values** - Space them out (1, 10, 20) to allow insertions

3. **Handle both branches** - If a condition can be false, either:
   - Add a `false_branch` edge to an action
   - Accept that execution may stop at the condition

4. **Test edge cases** - Test with data that triggers both branches

5. **Keep graphs simple** - Complex nested conditions are hard to debug. Consider multiple simpler rules instead.

### Common Mistakes

| Mistake | Problem | Solution |
|---------|---------|----------|
| Missing start edge | Execution fails (no entry point) | Add a start edge with `source_action_id: null` |
| No false_branch | Execution stops when condition is false | Add false_branch edge or accept early termination |
| Circular edges | Could cause infinite loop | Executor prevents this, but avoid for clarity |
| Wrong edge_order | Non-deterministic execution order | Use explicit ordering (1, 2, 3...) |

## Testing Graph Workflows

The test file `business/sdk/workflow/executor_graph_test.go` demonstrates all patterns:

- `TestGraphExec_SingleStartEdge` / `MultipleStartEdges` - Entry points
- `TestGraphExec_TrueBranch_WhenConditionTrue` - Branch following
- `TestGraphExec_DiamondPattern` - Converging branches
- `TestGraphExec_NestedConditions` - Multi-level conditions
- `TestGraphExec_NoCycleInfiniteLoop` - Cycle prevention
- `TestGraphExec_EdgeOrderRespected` - Ordering guarantees

Run tests with:
```bash
go test -v ./business/sdk/workflow/... -run TestGraphExec
```

## Related Documentation

- [actions/evaluate-condition.md](actions/evaluate-condition.md) - Condition action for branching decisions
- [configuration/rules.md](configuration/rules.md) - Rule and action configuration
- [api-reference.md](api-reference.md) - Edge API endpoints
- [database-schema.md](database-schema.md) - `action_edges` table schema
