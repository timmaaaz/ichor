# Triggers and Conditions

This document covers trigger types, trigger conditions, and field condition operators.

## Trigger Types

Trigger types define when automation rules fire.

| Type | Description | Use Case |
|------|-------------|----------|
| `on_create` | Entity creation | New order notifications |
| `on_update` | Entity update | Status change workflows |
| `on_delete` | Entity deletion | Cleanup actions |
| `scheduled` | Time-based | Daily reports, maintenance |

### TriggerType Model

**Database**: `workflow.trigger_types`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Primary key |
| `name` | string | Yes | Trigger type name (unique) |
| `description` | string | No | Human-readable description |
| `is_active` | bool | Yes | Whether trigger type is active |
| `deactivated_by` | UUID | No | User who deactivated |

**Source**: `business/sdk/workflow/models.go:152-159`

## Trigger Conditions

Trigger conditions are stored as JSONB in `automation_rules.trigger_conditions`. They define field-level criteria that must match for a rule to fire.

### TriggerConditions Structure

```json
{
  "field_conditions": [
    {
      "field_name": "status",
      "operator": "equals",
      "value": "shipped"
    }
  ]
}
```

**Source**: `business/sdk/workflow/trigger.go:22-25`

**Evaluation behavior:**
- If empty or null, rule matches all events of the specified trigger type
- Multiple conditions are evaluated with **AND logic** (all must match)

## Field Conditions

Individual conditions for field evaluation.

### FieldCondition Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `field_name` | string | Yes | Field to evaluate |
| `operator` | string | Yes | Comparison operator |
| `value` | interface{} | Conditional | Comparison value |
| `previous_value` | interface{} | Conditional | For `changed_from` operator |

**Source**: `business/sdk/workflow/trigger.go:15-20`

### Operators

| Operator | Description | Value Required | Example |
|----------|-------------|----------------|---------|
| `equals` | Exact match | Yes | `{"field_name": "status", "operator": "equals", "value": "active"}` |
| `not_equals` | Not equal | Yes | `{"field_name": "status", "operator": "not_equals", "value": "deleted"}` |
| `changed_from` | Previous value match | `previous_value` | `{"field_name": "status", "operator": "changed_from", "previous_value": "pending"}` |
| `changed_to` | New value with change detection | Yes | `{"field_name": "status", "operator": "changed_to", "value": "shipped"}` |
| `greater_than` | Numeric/string comparison | Yes | `{"field_name": "quantity", "operator": "greater_than", "value": 100}` |
| `less_than` | Numeric/string comparison | Yes | `{"field_name": "quantity", "operator": "less_than", "value": 10}` |
| `contains` | Substring match (strings) | Yes | `{"field_name": "name", "operator": "contains", "value": "VIP"}` |
| `in` | Value in array | Yes (array) | `{"field_name": "status", "operator": "in", "value": ["active", "pending"]}` |

**Source**: `business/sdk/workflow/trigger.go:329-371`

### Operator Details

#### equals / not_equals

Simple equality comparison. Works with strings, numbers, booleans, and UUIDs.

```json
{
  "field_name": "is_priority",
  "operator": "equals",
  "value": true
}
```

#### changed_from

Matches when a field's **previous** value matches. Only useful for `on_update` events.

```json
{
  "field_name": "status",
  "operator": "changed_from",
  "previous_value": "draft"
}
```

This fires when status **was** "draft" and has changed to any other value.

#### changed_to

Matches when a field changed **to** a specific value. The field must have actually changed (previous != new).

```json
{
  "field_name": "status",
  "operator": "changed_to",
  "value": "approved"
}
```

This only fires when status changes TO "approved", not when it already was "approved".

#### greater_than / less_than

Numeric comparison. Also works with strings (lexicographic).

```json
{
  "field_name": "total_amount",
  "operator": "greater_than",
  "value": 1000
}
```

#### contains

Substring match for string fields. Case-sensitive.

```json
{
  "field_name": "description",
  "operator": "contains",
  "value": "urgent"
}
```

#### in

Checks if field value is in a list of allowed values.

```json
{
  "field_name": "category",
  "operator": "in",
  "value": ["electronics", "furniture", "clothing"]
}
```

## Examples

### Fire on Any Order Creation

Rule with no conditions - fires for every order created:

```json
{
  "entity_name": "orders",
  "trigger_type": "on_create",
  "trigger_conditions": null
}
```

### Fire When Order Ships

```json
{
  "entity_name": "orders",
  "trigger_type": "on_update",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "status",
        "operator": "changed_to",
        "value": "shipped"
      }
    ]
  }
}
```

### Fire for High-Value Orders

```json
{
  "entity_name": "orders",
  "trigger_type": "on_create",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "total",
        "operator": "greater_than",
        "value": 10000
      }
    ]
  }
}
```

### Fire When Priority Order Status Changes from Pending

Multiple conditions (AND logic):

```json
{
  "entity_name": "orders",
  "trigger_type": "on_update",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "is_priority",
        "operator": "equals",
        "value": true
      },
      {
        "field_name": "status",
        "operator": "changed_from",
        "previous_value": "pending"
      }
    ]
  }
}
```

## Validation

The TriggerProcessor validates events before processing:

1. Event type is required and must be supported
2. Entity name is required
3. Timestamp must be non-zero
4. Warns if no rules exist for entity
5. Warns if update event has no field changes

**Source**: `business/sdk/workflow/trigger.go:153-190`

## Field Value Access

When evaluating conditions, field values are extracted from:

1. **RawData**: Current entity data snapshot (`event.RawData[field_name]`)
2. **FieldChanges**: For update events, previous and new values

For `changed_from` and `changed_to` operators, the processor uses `FieldChanges` to detect actual changes.

## Related Documentation

- [Rules](rules.md) - Automation rule configuration
- [Templates](templates.md) - Template variables for action configs
- [Architecture](../architecture.md) - TriggerProcessor implementation
