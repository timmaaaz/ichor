# evaluate_condition Action

Evaluates conditions against entity data and determines branch direction for workflow execution. This action enables conditional branching in workflow graphs, allowing different actions to execute based on field values.

## Overview

The `evaluate_condition` action is a **control flow** action that:
- Evaluates one or more field conditions against entity data
- Returns a branch direction (`true_branch` or `false_branch`)
- Works with the edge system to route execution through different action paths
- Supports both AND and OR logic for combining multiple conditions

**Important**: This action does NOT support manual execution - it only makes sense within an automated workflow context where branching is needed.

## Configuration Schema

```json
{
  "conditions": [
    {
      "field_name": "status",
      "operator": "equals",
      "value": "approved"
    }
  ],
  "logic_type": "and"
}
```

**Source**: `business/sdk/workflow/workflowactions/control/condition.go:24-27`

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `conditions` | []FieldCondition | **Yes** | - | Array of conditions to evaluate |
| `logic_type` | string | No | `and` | How to combine conditions: `and` or `or` |

### FieldCondition Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `field_name` | string | **Yes** | Entity field to evaluate |
| `operator` | string | **Yes** | Comparison operator |
| `value` | any | Conditional | Value to compare (required for most operators) |
| `previous_value` | any | Conditional | Previous value for `changed_from` operator |

## Operators

| Operator | Description | Requires Value | Works With |
|----------|-------------|----------------|------------|
| `equals` | Exact match | Yes | All event types |
| `not_equals` | Not equal | Yes | All event types |
| `greater_than` | Numeric/string greater than | Yes | All event types |
| `less_than` | Numeric/string less than | Yes | All event types |
| `contains` | Substring match (strings only) | Yes | All event types |
| `in` | Value in array | Yes (array) | All event types |
| `is_null` | Field is null | No | All event types |
| `is_not_null` | Field is not null | No | All event types |
| `changed_from` | Previous value matched | Yes (`previous_value`) | `on_update` only |
| `changed_to` | New value matches AND differs from previous | Yes | `on_update` only |

**Source**: `business/sdk/workflow/workflowactions/control/condition.go:77-88`

### Operator Details

#### equals / not_equals
Compares current field value against the specified value using string representation.

```json
{
  "field_name": "status",
  "operator": "equals",
  "value": "active"
}
```

#### greater_than / less_than
For numeric values, performs numeric comparison. Falls back to string comparison for non-numeric values.

```json
{
  "field_name": "quantity",
  "operator": "greater_than",
  "value": 100
}
```

#### contains
Checks if the field value (string) contains the specified substring.

```json
{
  "field_name": "description",
  "operator": "contains",
  "value": "urgent"
}
```

#### in
Checks if the field value exists in the provided array.

```json
{
  "field_name": "priority",
  "operator": "in",
  "value": ["high", "critical"]
}
```

#### is_null / is_not_null
Checks whether the field value is null or not. No value parameter needed.

```json
{
  "field_name": "assigned_to",
  "operator": "is_null"
}
```

#### changed_from (on_update only)
Checks if the field's previous value matched the specified value. Only works with `on_update` events where field change history is available.

```json
{
  "field_name": "status",
  "operator": "changed_from",
  "previous_value": "pending"
}
```

#### changed_to (on_update only)
Checks if the field changed TO the specified value (current value matches AND previous value differs). Only works with `on_update` events.

```json
{
  "field_name": "status",
  "operator": "changed_to",
  "value": "approved"
}
```

## Logic Types

### AND Logic (Default)
All conditions must be true for the overall result to be true.

```json
{
  "conditions": [
    {"field_name": "status", "operator": "equals", "value": "approved"},
    {"field_name": "amount", "operator": "greater_than", "value": 1000}
  ],
  "logic_type": "and"
}
```
Result: `true` only if status is "approved" AND amount > 1000

### OR Logic
Any condition being true makes the overall result true.

```json
{
  "conditions": [
    {"field_name": "priority", "operator": "equals", "value": "critical"},
    {"field_name": "amount", "operator": "greater_than", "value": 10000}
  ],
  "logic_type": "or"
}
```
Result: `true` if priority is "critical" OR amount > 10000

## Return Value

The action returns a `ConditionResult`:

```go
type ConditionResult struct {
    Evaluated   bool   `json:"evaluated"`    // Always true when executed
    Result      bool   `json:"result"`       // The boolean evaluation result
    BranchTaken string `json:"branch_taken"` // "true_branch" or "false_branch"
}
```

**Source**: `business/sdk/workflow/models.go:416-422`

The `BranchTaken` field determines which edges are followed in the workflow graph:
- `"true_branch"` - Follow edges of type `true_branch`
- `"false_branch"` - Follow edges of type `false_branch`

## Integration with Edges

The `evaluate_condition` action works with the edge system to enable workflow branching:

```
                    ┌─────────────────┐
                    │ Previous Action │
                    └────────┬────────┘
                             │ (sequence)
                    ┌────────▼────────┐
                    │    Condition    │
                    │  evaluate_cond  │
                    └────────┬────────┘
                    ┌────────┴────────┐
        true_branch │                 │ false_branch
           ┌────────▼────────┐  ┌─────▼──────────┐
           │ Approve Action  │  │ Reject Action  │
           └─────────────────┘  └────────────────┘
```

To create this flow, you need:
1. A `sequence` edge from the previous action to the condition
2. A `true_branch` edge from the condition to the approve action
3. A `false_branch` edge from the condition to the reject action

See [branching.md](../branching.md) for detailed edge configuration.

## Examples

### Example 1: Simple Status Check

Check if an order status is "approved":

```json
{
  "conditions": [
    {
      "field_name": "status",
      "operator": "equals",
      "value": "approved"
    }
  ]
}
```

### Example 2: Threshold-Based Approval

Route high-value orders to manager approval:

```json
{
  "conditions": [
    {
      "field_name": "total_amount",
      "operator": "greater_than",
      "value": 5000
    }
  ]
}
```

### Example 3: Multiple Conditions (AND)

Check if order is approved AND from a priority customer:

```json
{
  "conditions": [
    {
      "field_name": "status",
      "operator": "equals",
      "value": "approved"
    },
    {
      "field_name": "customer_tier",
      "operator": "in",
      "value": ["gold", "platinum"]
    }
  ],
  "logic_type": "and"
}
```

### Example 4: Multiple Conditions (OR)

Trigger escalation if critical OR overdue:

```json
{
  "conditions": [
    {
      "field_name": "priority",
      "operator": "equals",
      "value": "critical"
    },
    {
      "field_name": "days_open",
      "operator": "greater_than",
      "value": 7
    }
  ],
  "logic_type": "or"
}
```

### Example 5: Status Transition Detection

Detect when status changes to "shipped":

```json
{
  "conditions": [
    {
      "field_name": "status",
      "operator": "changed_to",
      "value": "shipped"
    }
  ]
}
```

### Example 6: Null Checks

Check if assignee is not set:

```json
{
  "conditions": [
    {
      "field_name": "assigned_to_id",
      "operator": "is_null"
    }
  ]
}
```

## Execution Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                     Execute Condition                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  1. Parse configuration                                          │
│     └─ Extract conditions array and logic_type                   │
│                                                                   │
│  2. For each condition:                                          │
│     ├─ Get current value from RawData or FieldChanges            │
│     ├─ Get previous value (for on_update events)                 │
│     └─ Apply operator comparison                                 │
│                                                                   │
│  3. Combine results                                              │
│     ├─ AND: all must be true                                     │
│     └─ OR: any must be true                                      │
│                                                                   │
│  4. Return ConditionResult                                       │
│     ├─ result: true/false                                        │
│     └─ branch_taken: "true_branch" or "false_branch"            │
│                                                                   │
│  5. Graph executor follows appropriate edges                     │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

## Validation Rules

The action validates:

1. **At least one condition** - Empty conditions array is rejected
2. **Valid logic_type** - Must be `and` or `or` (empty defaults to `and`)
3. **Field name required** - Each condition must have `field_name`
4. **Operator required** - Each condition must have `operator`
5. **Valid operator** - Must be one of the 10 supported operators

## Handler Properties

| Property | Value |
|----------|-------|
| Action Type | `evaluate_condition` |
| Supports Manual Execution | No |
| Is Async | No (synchronous) |

**Source**: `business/sdk/workflow/workflowactions/control/condition.go:50-58`

## Use Cases

1. **Approval Routing** - Route orders above a threshold to manager approval
2. **Priority Handling** - Execute different actions based on priority levels
3. **Status Transitions** - Trigger specific actions when status changes
4. **Validation Gates** - Only proceed if certain conditions are met
5. **Customer Segmentation** - Different workflows for different customer tiers

## Related Documentation

- [branching.md](../branching.md) - Graph-based execution and edge configuration
- [configuration/triggers.md](../configuration/triggers.md) - Trigger conditions (same operators)
- [actions/overview.md](overview.md) - Action handler interface
