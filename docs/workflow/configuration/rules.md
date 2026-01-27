# Automation Rules

This document covers automation rules, rule actions, action templates, and rule dependencies.

## AutomationRule

An automation rule defines when and what actions should execute.

### Model

**Database**: `workflow.automation_rules`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Primary key |
| `name` | string | Yes | Rule name (max 100 chars) |
| `description` | string | No | Human-readable description |
| `entity_id` | UUID | Yes | FK to `workflow.entities` |
| `entity_type_id` | UUID | Yes | FK to `workflow.entity_types` |
| `trigger_type_id` | UUID | Yes | FK to `workflow.trigger_types` |
| `trigger_conditions` | JSONB | No | See [Triggers](triggers.md) |
| `is_active` | bool | Yes | Whether rule is active |
| `created_date` | timestamp | Auto | Creation timestamp |
| `updated_date` | timestamp | Auto | Last update timestamp |
| `created_by` | UUID | Yes | User who created |
| `updated_by` | UUID | Yes | User who last updated |
| `deactivated_by` | UUID | No | User who deactivated |

**Source**: `business/sdk/workflow/models.go:238-252`

### Example

```json
{
  "name": "Order Shipped Notification",
  "description": "Send email when order ships",
  "entity_id": "uuid-of-orders-entity",
  "entity_type_id": "uuid-of-table-type",
  "trigger_type_id": "uuid-of-on-update-trigger",
  "trigger_conditions": {
    "field_conditions": [
      {
        "field_name": "status",
        "operator": "changed_to",
        "value": "shipped"
      }
    ]
  },
  "is_active": true
}
```

## Entity

Represents a database entity (table/view) that can be monitored.

### Model

**Database**: `workflow.entities`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Primary key |
| `name` | string | Yes | Entity name (unique, max 100 chars) |
| `entity_type_id` | UUID | Yes | FK to `workflow.entity_types` |
| `schema_name` | string | No | Database schema (default 'public') |
| `is_active` | bool | No | Whether entity is active (default true) |
| `created_date` | timestamp | Auto | Creation timestamp |
| `deactivated_by` | UUID | No | User who deactivated |

**Source**: `business/sdk/workflow/models.go:207-216`

**Note**: Entity names are stored WITHOUT schema prefix (e.g., `orders` not `sales.orders`).

## EntityType

Categories of entities that can be monitored.

### Model

**Database**: `workflow.entity_types`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Primary key |
| `name` | string | Yes | Entity type name (unique, max 50 chars) |
| `description` | string | No | Human-readable description |
| `is_active` | bool | Yes | Whether entity type is active |
| `deactivated_by` | UUID | No | User who deactivated |

**Source**: `business/sdk/workflow/models.go:180-187`

**Standard entity types:**
- `table` - Database table
- `view` - Database view

## RuleAction

Individual actions attached to automation rules.

### Model

**Database**: `workflow.rule_actions`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Primary key |
| `automation_rule_id` | UUID | Yes | FK to `workflow.automation_rules` |
| `name` | string | Yes | Action name (max 100 chars) |
| `description` | string | No | Human-readable description |
| `action_config` | JSONB | Yes | Action configuration (varies by type) |
| `execution_order` | int | Yes | Order of execution (>= 1) |
| `is_active` | bool | No | Whether action is active (default true) |
| `template_id` | UUID | No | FK to `workflow.action_templates` |
| `deactivated_by` | UUID | No | User who deactivated |

**Source**: `business/sdk/workflow/models.go:315-325`

### Execution Order

Actions with the **same** `execution_order` run in **parallel**.
Actions with **different** orders run **sequentially**.

```
Order 1: [send_email, create_alert]  ← parallel
Order 2: [update_field]               ← waits for 1
Order 3: [seek_approval]              ← waits for 2
```

### Action Config

The `action_config` field is polymorphic - its structure depends on the action type.

See [Actions](../actions/) for configuration schemas for each action type:
- [create_alert](../actions/create-alert.md)
- [update_field](../actions/update-field.md)
- [send_email](../actions/send-email.md)
- [send_notification](../actions/send-notification.md)
- [seek_approval](../actions/seek-approval.md)
- [allocate_inventory](../actions/allocate-inventory.md)

### Example

```json
{
  "automation_rule_id": "rule-uuid",
  "name": "Send Shipping Email",
  "action_config": {
    "recipients": ["customer@example.com"],
    "subject": "Your Order {{number}} Has Shipped",
    "body": "Dear {{customer_name}}, your order is on its way!"
  },
  "execution_order": 1,
  "is_active": true
}
```

## ActionTemplate

Reusable action configurations that can be referenced by multiple rule actions.

### Model

**Database**: `workflow.action_templates`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Primary key |
| `name` | string | Yes | Template name (unique, max 100 chars) |
| `description` | string | No | Human-readable description |
| `action_type` | string | Yes | Action type (must be registered) |
| `default_config` | JSONB | Yes | Default action configuration |
| `created_date` | timestamp | Auto | Creation timestamp |
| `created_by` | UUID | Yes | User who created |
| `is_active` | bool | No | Whether template is active (default true) |
| `deactivated_by` | UUID | No | User who deactivated |

**Source**: `business/sdk/workflow/models.go:282-292`

### Using Templates

When a `RuleAction` references a template via `template_id`:
1. The template's `default_config` is loaded
2. Values from `action_config` override template defaults
3. Final merged config is used for execution

### Example

Template:
```json
{
  "name": "Standard Shipping Email",
  "action_type": "send_email",
  "default_config": {
    "subject": "Order {{number}} Shipped",
    "body": "Your order has shipped. Tracking: {{tracking_number}}"
  }
}
```

Rule action using template:
```json
{
  "template_id": "template-uuid",
  "action_config": {
    "recipients": ["customer@example.com"]
  }
}
```

Merged result:
```json
{
  "recipients": ["customer@example.com"],
  "subject": "Order {{number}} Shipped",
  "body": "Your order has shipped. Tracking: {{tracking_number}}"
}
```

## RuleDependency

Defines dependencies between automation rules.

### Model

**Database**: `workflow.rule_dependencies`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Auto | Primary key |
| `parent_rule_id` | UUID | Yes | Parent rule (must complete first) |
| `child_rule_id` | UUID | Yes | Child rule (runs after parent) |

**Source**: `business/sdk/workflow/models.go:352-356`

### Validation Rules

1. Cannot create circular dependencies
2. Parent and child must be different rules
3. Both rules should be active

### Example

Rule B depends on Rule A completing first:
```json
{
  "parent_rule_id": "rule-a-uuid",
  "child_rule_id": "rule-b-uuid"
}
```

## Best Practices

### Rule Naming

Use descriptive names that indicate:
- The entity being monitored
- The trigger condition
- The primary action

Good: `Order Shipped - Send Customer Email`
Bad: `Rule 1`

### Execution Order

1. Put independent actions at the same order (parallel)
2. Put dependent actions at higher orders (sequential)
3. Put validation/approval actions last

### Active State

- Use `is_active` to disable rules temporarily
- Prefer deactivation over deletion for audit trail
- Set `deactivated_by` when deactivating

### Conditions

- Start simple, add conditions as needed
- Test with broad conditions first
- Use `changed_to` instead of `equals` for status workflows
