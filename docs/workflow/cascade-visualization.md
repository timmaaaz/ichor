# Cascade Visualization

This document explains the cascade visualization feature, which shows how workflow rules trigger downstream workflows through entity modifications.

## Overview

When a workflow action modifies an entity (e.g., `update_field` changes `sales.orders.status`), it may trigger other workflows that are listening for changes on that entity. The cascade visualization API shows these downstream dependencies, helping you:

- **Understand side effects** - See what other workflows will run when your rule executes
- **Debug cascade chains** - Trace multi-rule execution paths
- **Prevent infinite loops** - Identify self-triggering patterns
- **Plan workflow changes** - Understand impact before modifying rules

## How It Works

### The EntityModifier Interface

Action handlers that modify entities implement the `EntityModifier` interface to declare what they change:

```go
type EntityModifier interface {
    // GetEntityModifications returns what entities/fields this action modifies
    GetEntityModifications(config json.RawMessage) []EntityModification
}
```

**Source**: `business/sdk/workflow/interfaces.go:140-145`

### EntityModification Struct

```go
type EntityModification struct {
    // EntityName is the fully-qualified table name (e.g., "sales.orders")
    EntityName string `json:"entity_name"`

    // EventType indicates what event this modification triggers
    // Valid values: "on_create", "on_update", "on_delete"
    EventType string `json:"event_type"`

    // Fields lists which fields are modified (for on_update events)
    Fields []string `json:"fields,omitempty"`
}
```

**Source**: `business/sdk/workflow/interfaces.go:147-161`

### Handlers That Implement EntityModifier

Currently, the following handlers implement `EntityModifier`:

| Handler | Entity Modified | Event Type |
|---------|-----------------|------------|
| `update_field` | Configured `target_entity` | `on_update` |

Other handlers (`send_email`, `send_notification`, `create_alert`, `allocate_inventory`, `evaluate_condition`) do NOT modify entities and therefore don't appear in cascade detection.

**Example Implementation** (`update_field`):

```go
func (h *UpdateFieldHandler) GetEntityModifications(config json.RawMessage) []EntityModification {
    var cfg UpdateFieldConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil
    }

    return []workflow.EntityModification{{
        EntityName: cfg.TargetEntity,  // e.g., "sales.orders"
        EventType:  "on_update",
        Fields:     []string{cfg.TargetField},  // e.g., ["status"]
    }}
}
```

**Source**: `business/sdk/workflow/workflowactions/data/updatefield.go:447-461`

## API Reference

### Get Cascade Map

Returns all downstream workflows that could be triggered by a rule's actions.

**Endpoint**: `GET /v1/workflow/rules/{id}/cascade-map`

**Authentication**: Required (Bearer token)

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Rule ID to analyze |

### Response Structure

```json
{
  "rule_id": "uuid",
  "rule_name": "Order Status Update",
  "actions": [
    {
      "action_id": "uuid",
      "action_name": "Update Status",
      "action_type": "update_field",
      "modifies_entity": "sales.orders",
      "triggers_event": "on_update",
      "modified_fields": ["status"],
      "downstream_workflows": [
        {
          "rule_id": "uuid",
          "rule_name": "Order Shipped Notification",
          "trigger_conditions": {...},
          "will_trigger_if": "any sales.orders update (with conditions)"
        }
      ]
    },
    {
      "action_id": "uuid",
      "action_name": "Send Email",
      "action_type": "send_email",
      "downstream_workflows": []
    }
  ]
}
```

### Response Types

**CascadeResponse**:
| Field | Type | Description |
|-------|------|-------------|
| `rule_id` | string | ID of the analyzed rule |
| `rule_name` | string | Name of the analyzed rule |
| `actions` | []ActionCascadeInfo | Actions and their downstream effects |

**Source**: `api/domain/http/workflow/ruleapi/cascade.go:22-26`

**ActionCascadeInfo**:
| Field | Type | Description |
|-------|------|-------------|
| `action_id` | string | Action ID |
| `action_name` | string | Action name |
| `action_type` | string | Type (e.g., "update_field") |
| `modifies_entity` | string | Entity being modified (if any) |
| `triggers_event` | string | Event type triggered (if any) |
| `modified_fields` | []string | Fields being changed (if any) |
| `downstream_workflows` | []DownstreamWorkflowInfo | Workflows that may be triggered |

**Source**: `api/domain/http/workflow/ruleapi/cascade.go:35-43`

**DownstreamWorkflowInfo**:
| Field | Type | Description |
|-------|------|-------------|
| `rule_id` | string | Downstream rule ID |
| `rule_name` | string | Downstream rule name |
| `trigger_conditions` | json.RawMessage | Raw trigger conditions (nullable) |
| `will_trigger_if` | string | Human-readable trigger description |

**Source**: `api/domain/http/workflow/ruleapi/cascade.go:46-51`

### Error Responses

| Status | Error Code | Description |
|--------|------------|-------------|
| 400 | InvalidArgument | Invalid rule ID format |
| 401 | Unauthenticated | Missing or invalid token |
| 404 | NotFound | Rule does not exist |
| 500 | Internal | Action registry not configured |

## Example Usage

### cURL Request

```bash
curl -X GET "http://localhost:3000/v1/workflow/rules/550e8400-e29b-41d4-a716-446655440000/cascade-map" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

### Example Response

```json
{
  "rule_id": "550e8400-e29b-41d4-a716-446655440000",
  "rule_name": "Auto-approve Low Value Orders",
  "actions": [
    {
      "action_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "action_name": "Set Order Status",
      "action_type": "update_field",
      "modifies_entity": "sales.orders",
      "triggers_event": "on_update",
      "modified_fields": ["status"],
      "downstream_workflows": [
        {
          "rule_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
          "rule_name": "Order Approved - Notify Warehouse",
          "trigger_conditions": {"status": {"operator": "equals", "value": "approved"}},
          "will_trigger_if": "any sales.orders update (with conditions)"
        },
        {
          "rule_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
          "rule_name": "Order Status Changed - Update Dashboard",
          "will_trigger_if": "any sales.orders update"
        }
      ]
    },
    {
      "action_id": "a3bb189e-8bf9-3888-9912-ace4e6543002",
      "action_name": "Send Confirmation Email",
      "action_type": "send_email",
      "downstream_workflows": []
    }
  ]
}
```

## How Downstream Rules Are Found

The cascade detection algorithm:

1. **For each action** in the rule:
   - Check if the handler implements `EntityModifier`
   - If yes, call `GetEntityModifications(config)` to get modifications

2. **For each modification**:
   - Parse entity name (handle `schema.table` format)
   - Find the entity in `workflow.entities` table
   - Query all rules that monitor this entity

3. **Filter rules**:
   - Exclude the current rule (self-trigger prevention)
   - Exclude inactive rules (`is_active = false`)
   - Match trigger type (`on_create`, `on_update`, `on_delete`)

4. **Build response**:
   - Include human-readable `will_trigger_if` description
   - Include raw `trigger_conditions` for advanced analysis

**Source**: `api/domain/http/workflow/ruleapi/cascade.go:122-187`

### Self-Trigger Exclusion

The cascade map explicitly excludes the queried rule from its own downstream workflows. This prevents confusing output when a rule modifies an entity it also listens to.

```go
// Skip the current rule (don't show self-triggers)
if rule.ID == excludeRuleID {
    continue
}
```

### Human-Readable Descriptions

The `will_trigger_if` field provides a friendly description:

| Event Type | Without Conditions | With Conditions |
|------------|-------------------|-----------------|
| `on_create` | "any sales.orders creation" | "any sales.orders creation (with conditions)" |
| `on_update` | "any sales.orders update" | "any sales.orders update (with conditions)" |
| `on_delete` | "any sales.orders deletion" | "any sales.orders deletion (with conditions)" |

**Source**: `api/domain/http/workflow/ruleapi/cascade.go:189-209`

## Use Cases

### 1. Understanding Workflow Dependencies

Before modifying a rule, check what downstream rules depend on it:

```bash
GET /v1/workflow/rules/{rule_id}/cascade-map
```

Review the response to understand:
- Which rules will trigger when this rule runs
- What entities and fields each action modifies
- Whether any workflows depend on specific field changes

### 2. Debugging Cascade Chains

If workflows are triggering unexpectedly:

1. Get the cascade map for the original rule
2. For each downstream workflow, get its cascade map
3. Build a dependency tree to trace execution paths

### 3. Impact Analysis Before Changes

Before changing an `update_field` action's target entity or field:

1. Get current cascade map
2. Note all downstream workflows
3. After making changes, verify the new cascade map matches expectations
4. Test that downstream workflows still function correctly

### 4. Identifying Missing Downstream Workflows

If an expected workflow isn't triggering:

1. Get cascade map for the upstream rule
2. Check if the expected downstream appears
3. If not, verify:
   - Downstream rule is active (`is_active = true`)
   - Downstream rule listens to the same entity
   - Downstream rule has the correct trigger type
   - Entity modification event matches trigger type

## Limitations

### Not All Actions Implement EntityModifier

Only actions that modify database entities participate in cascade detection. The following do NOT:

| Action Type | Reason |
|-------------|--------|
| `send_email` | Sends external email, no DB change |
| `send_notification` | Creates notification, doesn't modify tracked entities |
| `create_alert` | Creates alert record, not a tracked entity cascade |
| `allocate_inventory` | Currently doesn't implement EntityModifier (future enhancement) |
| `evaluate_condition` | Decision node, doesn't modify anything |
| `seek_approval` | Waits for human input, no automatic modification |

### Static Analysis Only

The cascade map shows **potential** downstream workflows based on configuration. It cannot account for:

- Runtime conditions that may prevent the action from executing
- Trigger conditions on downstream rules that may not match
- Field-level matching (conditions checking specific field values)

### No Recursive Cascade

The API shows immediate downstream workflows only. It does not recursively analyze what workflows those downstream rules might trigger. To build a full cascade tree, call the API for each downstream rule.

## Testing

Integration tests cover cascade visualization in `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go`:

| Test | Description |
|------|-------------|
| `cascadeMap200WithDownstream` | Verifies downstream workflows are found |
| `cascadeMap200NoDownstream` | Verifies empty response for non-modifying actions |
| `cascadeMap200MultipleActions` | Tests rules with mixed action types |
| `cascadeMapExcludesSelf` | Verifies self-trigger exclusion |
| `cascadeMapOnlyActiveRules` | Verifies inactive rules are filtered |
| `cascadeMapResponseStructure` | Validates response format |
| `cascadeMapEmptyActions` | Tests rules with no actions |
| `cascadeMapRuleNotFound404` | Tests 404 for non-existent rules |
| `cascadeMap401` | Tests authentication requirement |

Run tests with:
```bash
go test -v ./api/cmd/services/ichor/tests/workflow/ruleapi/... -run Test_CascadeAPI
```

## Adding EntityModifier to a New Action Handler

If you create a new action handler that modifies entities:

1. **Implement the interface**:

```go
// Ensure your handler type implements EntityModifier
var _ workflow.EntityModifier = (*MyHandler)(nil)

func (h *MyHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
    var cfg MyConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return nil
    }

    return []workflow.EntityModification{{
        EntityName: cfg.TargetEntity,
        EventType:  "on_update",  // or "on_create", "on_delete"
        Fields:     []string{cfg.ModifiedField},
    }}
}
```

2. **Register normally** - The cascade API automatically detects `EntityModifier` implementations:

```go
handler, exists := registry.Get(actionType)
if exists {
    if modifier, ok := handler.(workflow.EntityModifier); ok {
        mods := modifier.GetEntityModifications(config)
        // Process modifications...
    }
}
```

No changes to the cascade API are needed.

## Related Documentation

- [architecture.md](architecture.md) - System overview including action handlers
- [actions/overview.md](actions/overview.md) - Action handler interface
- [actions/update-field.md](actions/update-field.md) - The main entity-modifying action
- [api-reference.md](api-reference.md) - Complete API documentation
- [adding-domains.md](adding-domains.md) - Integrating new domains with workflows

## Key Files

| File | Purpose |
|------|---------|
| `api/domain/http/workflow/ruleapi/cascade.go` | Cascade API handler and response types |
| `business/sdk/workflow/interfaces.go` | `EntityModifier` interface definition |
| `business/sdk/workflow/workflowactions/data/updatefield.go:447-461` | EntityModifier implementation |
| `api/cmd/services/ichor/tests/workflow/ruleapi/cascade_test.go` | Integration tests |
