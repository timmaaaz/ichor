# Workflow Configuration Validator - Complete Field Reference

> **Note:** This document should be copied to `business/sdk/workflow/docs/workflow-config-validator-spec.md` for permanent storage within the package.

This document provides a comprehensive mapping of all fields in the workflow configuration system, intended to serve as the foundation for building a robust configuration validator.

---

## Implementation Plan

### Phase 1: Copy Documentation
1. Create `business/sdk/workflow/docs/` directory
2. Copy this file to `business/sdk/workflow/docs/workflow-config-validator-spec.md`

### Phase 2: Build Validator (Optional Follow-up)
Based on this specification, implement enhanced validation in a new `validation.go` file

---

## Table of Contents

1. [TriggerType](#1-triggertype)
2. [EntityType](#2-entitytype)
3. [Entity](#3-entity)
4. [AutomationRule](#4-automationrule)
5. [TriggerConditions (JSONB)](#5-triggerconditions-jsonb)
6. [FieldCondition](#6-fieldcondition)
7. [ActionTemplate](#7-actiontemplate)
8. [RuleAction](#8-ruleaction)
9. [RuleDependency](#9-ruledependency)
10. [Action Configurations (Polymorphic)](#10-action-configurations-polymorphic)
    - [create_alert](#101-create_alert)
    - [update_field](#102-update_field)
    - [send_email](#103-send_email)
    - [send_notification](#104-send_notification)
    - [seek_approval](#105-seek_approval)
    - [allocate_inventory](#106-allocate_inventory)
11. [Template Variables](#11-template-variables)
12. [TriggerEvent (Runtime)](#12-triggerevent-runtime)
13. [Alerts Subsystem](#13-alerts-subsystem)
14. [Allowed Values (Whitelists)](#14-allowed-values-whitelists)
15. [Existing Validation Rules](#15-existing-validation-rules)
16. [Proposed Validator Enhancements](#16-proposed-validator-enhancements)
17. [CLI Command: validate-workflows](#17-cli-command-validate-workflows)

---

## 1. TriggerType

Represents types of triggers for automation rules.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `name` | `string` | **Yes** | Trigger type name | Non-empty, unique, max 50 chars |
| `description` | `string` | No | Human-readable description | - |
| `is_active` | `bool` | **Yes** | Whether trigger type is active | - |
| `deactivated_by` | `uuid.UUID` | No | User who deactivated | Valid user UUID if provided |

**Source:** [models.go:152-159](business/sdk/workflow/models.go#L152-L159)

**Database:** `workflow.trigger_types`

---

## 2. EntityType

Represents types/categories of entities that can be monitored.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `name` | `string` | **Yes** | Entity type name | Non-empty, unique, max 50 chars |
| `description` | `string` | No | Human-readable description | - |
| `is_active` | `bool` | **Yes** | Whether entity type is active | - |
| `deactivated_by` | `uuid.UUID` | No | User who deactivated | Valid user UUID if provided |

**Source:** [models.go:180-187](business/sdk/workflow/models.go#L180-L187)

**Database:** `workflow.entity_types`

---

## 3. Entity

Represents a specific monitored database entity (table/view).

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `name` | `string` | **Yes** | Entity name | Non-empty, unique, max 100 chars |
| `entity_type_id` | `uuid.UUID` | **Yes** | FK to entity_types | Must exist in entity_types |
| `schema_name` | `string` | No | Database schema | Valid schema name, default 'public' |
| `is_active` | `bool` | No | Whether entity is active | Default true |
| `created_date` | `time.Time` | Auto | Creation timestamp | Auto-generated |
| `deactivated_by` | `uuid.UUID` | No | User who deactivated | Valid user UUID if provided |

**Source:** [models.go:207-216](business/sdk/workflow/models.go#L207-L216)

**Database:** `workflow.entities`

---

## 4. AutomationRule

Represents a workflow automation rule.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `name` | `string` | **Yes** | Rule name | Non-empty, max 100 chars |
| `description` | `string` | No | Human-readable description | - |
| `entity_id` | `uuid.UUID` | **Yes** | FK to entities | Must exist in entities |
| `entity_type_id` | `uuid.UUID` | **Yes** | FK to entity_types | Must exist in entity_types |
| `trigger_type_id` | `uuid.UUID` | **Yes** | FK to trigger_types | Must exist in trigger_types |
| `trigger_conditions` | `*json.RawMessage` | No | JSONB conditions | See TriggerConditions |
| `is_active` | `bool` | **Yes** | Whether rule is active | Default true |
| `created_date` | `time.Time` | Auto | Creation timestamp | Auto-generated |
| `updated_date` | `time.Time` | Auto | Last update timestamp | Auto-updated |
| `created_by` | `uuid.UUID` | **Yes** | User who created | Must exist in users |
| `updated_by` | `uuid.UUID` | **Yes** | User who last updated | Must exist in users |
| `deactivated_by` | `uuid.UUID` | No | User who deactivated | Valid user UUID if provided |

**Source:** [models.go:238-252](business/sdk/workflow/models.go#L238-L252)

**Database:** `workflow.automation_rules`

---

## 5. TriggerConditions (JSONB)

Stored in `automation_rules.trigger_conditions` as JSONB.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `field_conditions` | `[]FieldCondition` | No | Array of field conditions | All conditions must be valid |

**Source:** [trigger.go:22-25](business/sdk/workflow/trigger.go#L22-L25)

**Notes:**
- If empty or null, rule matches all events of the specified trigger type
- Multiple conditions are evaluated with AND logic (all must match)

---

## 6. FieldCondition

Individual condition for field evaluation.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `field_name` | `string` | **Yes** | Field to evaluate | Non-empty, valid identifier |
| `operator` | `string` | **Yes** | Comparison operator | See allowed operators |
| `value` | `interface{}` | Cond. | Comparison value | Required for most operators |
| `previous_value` | `interface{}` | Cond. | For change detection | Required for `changed_from` |

**Source:** [trigger.go:15-20](business/sdk/workflow/trigger.go#L15-L20)

**Allowed Operators (Trigger Evaluation):**
- `equals` - exact match
- `not_equals` - not equal
- `changed_from` - previous value match (update events only)
- `changed_to` - new value match with change detection
- `greater_than` - numeric/string comparison
- `less_than` - numeric/string comparison
- `contains` - substring match (strings only)
- `in` - value in array

---

## 7. ActionTemplate

Reusable action configuration template.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `name` | `string` | **Yes** | Template name | Non-empty, unique, max 100 chars |
| `description` | `string` | No | Human-readable description | - |
| `action_type` | `string` | **Yes** | Action type | Must be registered action type |
| `default_config` | `json.RawMessage` | **Yes** | Default action config | Must validate against action type |
| `created_date` | `time.Time` | Auto | Creation timestamp | Auto-generated |
| `created_by` | `uuid.UUID` | **Yes** | User who created | Must exist in users |
| `is_active` | `bool` | No | Whether template is active | Default true |
| `deactivated_by` | `uuid.UUID` | No | User who deactivated | Valid user UUID if provided |

**Source:** [models.go:282-292](business/sdk/workflow/models.go#L282-L292)

**Database:** `workflow.action_templates`

---

## 8. RuleAction

Individual action within an automation rule.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `automation_rule_id` | `uuid.UUID` | **Yes** | FK to automation_rules | Must exist in automation_rules |
| `name` | `string` | **Yes** | Action name | Non-empty, max 100 chars |
| `description` | `string` | No | Human-readable description | - |
| `action_config` | `json.RawMessage` | **Yes** | Action configuration | Must validate for action type |
| `execution_order` | `int` | **Yes** | Order of execution | >= 1 |
| `is_active` | `bool` | No | Whether action is active | Default true |
| `template_id` | `*uuid.UUID` | No | FK to action_templates | Must exist if provided |
| `deactivated_by` | `uuid.UUID` | No | User who deactivated | Valid user UUID if provided |

**Source:** [models.go:315-325](business/sdk/workflow/models.go#L315-L325)

**Database:** `workflow.rule_actions`

**Notes:**
- Action type is inferred from template or must be specified in action_config
- Each action_config has polymorphic validation based on action type

---

## 9. RuleDependency

Represents a dependency between two automation rules.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `parent_rule_id` | `uuid.UUID` | **Yes** | Parent rule | Must exist in automation_rules |
| `child_rule_id` | `uuid.UUID` | **Yes** | Child rule | Must exist in automation_rules |

**Source:** [models.go:352-356](business/sdk/workflow/models.go#L352-L356)

**Database:** `workflow.rule_dependencies`

**Validation:**
- Cannot create circular dependencies
- Parent and child must be different rules
- Both rules should be active

---

## 10. Action Configurations (Polymorphic)

Each action type has its own configuration schema stored in `rule_actions.action_config` or `action_templates.default_config`.

### 10.1. create_alert

Creates workflow alerts for recipients.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `alert_type` | `string` | No | Alert category | - |
| `severity` | `string` | No | Alert severity | `low`, `medium`, `high`, `critical`; default `medium` |
| `title` | `string` | No | Alert title | Supports `{{template_vars}}` |
| `message` | `string` | **Yes** | Alert message | Non-empty, supports `{{template_vars}}` |
| `recipients.users` | `[]string` | Cond. | User UUIDs | Valid UUIDs |
| `recipients.roles` | `[]string` | Cond. | Role UUIDs | Valid UUIDs |
| `context` | `json.RawMessage` | No | Additional context | Valid JSON |
| `resolve_prior` | `bool` | No | Auto-resolve prior alerts | - |

**Source:** [workflowactions/communication/alert.go:24-35](business/sdk/workflow/workflowactions/communication/alert.go#L24-L35)

**Validation:**
- `message` is required and must be non-empty
- At least one recipient (user or role) is required
- All UUIDs must be valid format

---

### 10.2. update_field

Updates a field in a database entity.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `target_entity` | `string` | **Yes** | Target table | Must be in whitelist |
| `target_field` | `string` | **Yes** | Column to update | Valid identifier |
| `new_value` | `any` | **Yes** | New value | Supports `{{template_vars}}` |
| `field_type` | `string` | No | Value type | `string`, `number`, `date`, `uuid`, `foreign_key` |
| `foreign_key_config` | `ForeignKeyConfig` | Cond. | FK resolution | Required when field_type is `foreign_key` |
| `conditions` | `[]FieldCondition` | No | WHERE conditions | All conditions must be valid |
| `timeout_ms` | `int` | No | Operation timeout | >= 0 |

**Source:** [workflowactions/data/updatefield.go:19-28](business/sdk/workflow/workflowactions/data/updatefield.go#L19-L28)

**ForeignKeyConfig:**

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `reference_table` | `string` | **Yes** | FK target table | Must be in whitelist |
| `lookup_field` | `string` | **Yes** | Field to match | Valid identifier |
| `id_field` | `string` | No | ID field name | Default `id` |
| `allow_create` | `bool` | No | Create if not found | - |
| `create_config` | `map[string]any` | Cond. | Creation defaults | Required if allow_create |

**Whitelist (Valid Tables):**
```go
// sales schema
"sales.customers", "sales.orders", "sales.order_line_items",
"sales.order_fulfillment_statuses", "sales.line_item_fulfillment_statuses",
// products schema
"products.products", "products.brands", "products.product_categories",
"products.physical_attributes", "products.product_costs", "products.cost_history",
"products.quality_metrics",
// inventory schema
"inventory.inventory_items", "inventory.inventory_locations", "inventory.inventory_transactions",
"inventory.warehouses", "inventory.zones", "inventory.lot_trackings",
"inventory.serial_numbers", "inventory.inspections", "inventory.inventory_adjustments",
"inventory.transfer_orders",
// procurement schema
"procurement.suppliers", "procurement.supplier_products",
// core schema
"core.users", "core.roles", "core.user_roles", "core.contact_infos", "core.table_access",
// hr schema
"hr.offices",
// geography schema
"geography.countries", "geography.regions", "geography.cities", "geography.streets",
// assets schema
"assets.assets", "assets.valid_assets",
// config schema
"config.table_configs",
// workflow schema
"workflow.automation_rules", "workflow.rule_actions", "workflow.action_templates",
"workflow.rule_dependencies", "workflow.trigger_types", "workflow.entity_types",
"workflow.entities", "workflow.automation_executions", "workflow.notification_deliveries",
```

**Source:** [workflowactions/data/updatefield.go:373-403](business/sdk/workflow/workflowactions/data/updatefield.go#L373-L403)

**Condition Operators (update_field):**
- `equals`, `not_equals`, `greater_than`, `less_than`
- `contains`, `is_null`, `is_not_null`, `in`, `not_in`

---

### 10.3. send_email

Sends email notifications.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `recipients` | `[]string` | **Yes** | Email addresses | At least one required |
| `subject` | `string` | **Yes** | Email subject | Non-empty |
| `body` | `string` | No | Email body | Supports `{{template_vars}}` |
| `simulate_failure` | `bool` | No | Testing flag | For testing only |
| `failure_message` | `string` | No | Test failure msg | Used with simulate_failure |

**Source:** [workflowactions/communication/email.go:32-38](business/sdk/workflow/workflowactions/communication/email.go#L32-L38)

**Validation:**
- Recipients list is required and must not be empty
- Subject is required

---

### 10.4. send_notification

Sends multi-channel notifications.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `recipients` | `[]string` | **Yes** | Recipient IDs | At least one required |
| `channels` | `[]Channel` | **Yes** | Delivery channels | At least one required |
| `priority` | `string` | **Yes** | Priority level | `low`, `medium`, `high`, `critical` |

**Source:** [workflowactions/communication/notification.go:32-39](business/sdk/workflow/workflowactions/communication/notification.go#L32-L39)

**Channel Types:** `email`, `sms`, `push`, `in_app`

---

### 10.5. seek_approval

Initiates an approval workflow.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `approvers` | `[]string` | **Yes** | Approver UUIDs | At least one required |
| `approval_type` | `string` | **Yes** | Approval mode | `any`, `all`, `majority` |

**Source:** [workflowactions/approval/seek.go:32-36](business/sdk/workflow/workflowactions/approval/seek.go#L32-L36)

---

### 10.6. allocate_inventory

Allocates or reserves inventory items.

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `inventory_items` | `[]AllocationItem` | Cond. | Items to allocate | Required unless source_from_line_item |
| `source_from_line_item` | `bool` | No | Extract from trigger | - |
| `allocation_mode` | `string` | **Yes** | Mode | `reserve`, `allocate` |
| `allocation_strategy` | `string` | **Yes** | Strategy | `fifo`, `lifo`, `nearest_expiry`, `lowest_cost` |
| `allow_partial` | `bool` | No | Allow partial allocation | - |
| `reservation_duration_hours` | `int` | Cond. | Reservation TTL | Required if mode is `reserve`, default 24 |
| `priority` | `string` | **Yes** | Priority level | `low`, `medium`, `high`, `critical` |
| `timeout_ms` | `int` | No | Operation timeout | >= 0 |
| `reference_id` | `string` | No | Reference ID | Order ID, etc. |
| `reference_type` | `string` | No | Reference type | `order`, `transfer`, etc. |

**Source:** [workflowactions/inventory/allocate.go:30-41](business/sdk/workflow/workflowactions/inventory/allocate.go#L30-L41)

**AllocationItem:**

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `product_id` | `uuid.UUID` | **Yes** | Product to allocate | Non-nil UUID |
| `quantity` | `int` | **Yes** | Quantity to allocate | > 0 |
| `warehouse_id` | `*uuid.UUID` | No | Specific warehouse | Valid UUID if provided |
| `location_id` | `*uuid.UUID` | No | Specific location | Valid UUID if provided |

---

## 11. Template Variables

Template variables use `{{variable_name}}` syntax for dynamic value substitution.

### Variable Pattern
```regex
\{\{([^}]+)\}\}
```

### Built-in Variables
| Variable | Description |
|----------|-------------|
| `$me` | Current user's ID |
| `$now` | Current timestamp (RFC3339) |

### Context Variables
Variables available from the trigger event execution context:

| Variable | Source | Description |
|----------|--------|-------------|
| `entity_id` | TriggerEvent | The entity's UUID |
| `entity_name` | TriggerEvent | Table/view name |
| `event_type` | TriggerEvent | `on_create`, `on_update`, `on_delete` |
| `timestamp` | TriggerEvent | Event timestamp |
| `user_id` | TriggerEvent | User who triggered |
| `rule_id` | ExecutionContext | Current rule UUID |
| `rule_name` | ExecutionContext | Current rule name |
| `execution_id` | ExecutionContext | Workflow execution UUID |
| `{field_name}` | RawData | Any field from entity data |
| `old_{field_name}` | FieldChanges | Previous value (updates) |
| `new_{field_name}` | FieldChanges | New value (updates) |

### Filters
Variables support filters with pipe syntax: `{{variable | filter:arg}}`

**Built-in Filters:**
- `uppercase` - Convert to uppercase
- `lowercase` - Convert to lowercase
- `capitalize` - Capitalize first letter
- `trim` - Remove whitespace
- `truncate:N` - Truncate to N characters
- `currency:CODE` - Format as currency (USD, EUR, GBP, JPY, etc.)
- `round:N` - Round to N decimal places
- `formatDate:format` - Format date (short, long, time, datetime, or Go format)
- `join:separator` - Join array elements
- `first` - Get first array element
- `last` - Get last array element
- `default:value` - Default if empty

**Source:** [template.go:643-832](business/sdk/workflow/template.go#L643-L832)

### Variable Validation
```regex
^(\$[a-zA-Z][a-zA-Z0-9_]*|[a-zA-Z][a-zA-Z0-9._]*)$
```

**Source:** [template.go:574-607](business/sdk/workflow/template.go#L574-L607)

---

## 12. TriggerEvent (Runtime)

Represents an event that triggers workflow execution (not persisted, runtime only).

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `event_type` | `string` | **Yes** | Event type | `on_create`, `on_update`, `on_delete`, `scheduled` |
| `entity_name` | `string` | **Yes** | Entity name | Non-empty |
| `entity_id` | `uuid.UUID` | No | Entity UUID | - |
| `field_changes` | `map[string]FieldChange` | Cond. | Changed fields | Required for `on_update` |
| `timestamp` | `time.Time` | **Yes** | Event timestamp | Non-zero |
| `raw_data` | `map[string]any` | No | Entity data snapshot | - |
| `user_id` | `uuid.UUID` | No | Triggering user | - |

**Source:** [models.go:14-22](business/sdk/workflow/models.go#L14-L22)

---

## 13. Alerts Subsystem

### Alert

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `alert_type` | `string` | **Yes** | Alert category | Max 100 chars |
| `severity` | `string` | **Yes** | Priority level | Enum: `low`, `medium`, `high`, `critical` |
| `title` | `string` | **Yes** | Alert title | Non-empty |
| `message` | `string` | **Yes** | Alert message | Non-empty |
| `context` | `json.RawMessage` | No | Additional context | Default `{}` |
| `source_entity_name` | `string` | No | Source entity | Max 100 chars |
| `source_entity_id` | `uuid.UUID` | No | Source entity ID | - |
| `source_rule_id` | `uuid.UUID` | No | Source rule | FK to automation_rules |
| `status` | `string` | **Yes** | Alert status | Enum: `active`, `acknowledged`, `dismissed`, `resolved` |
| `expires_date` | `*time.Time` | No | Expiration timestamp | - |
| `created_date` | `time.Time` | **Yes** | Creation timestamp | - |
| `updated_date` | `time.Time` | **Yes** | Last update timestamp | - |

**Source:** [alertbus/model.go:33-47](business/domain/workflow/alertbus/model.go#L33-L47)

**Database:** `workflow.alerts`

### AlertRecipient

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `alert_id` | `uuid.UUID` | **Yes** | FK to alerts | Must exist in alerts |
| `recipient_type` | `string` | **Yes** | Recipient type | `user` or `role` |
| `recipient_id` | `uuid.UUID` | **Yes** | User/Role UUID | Valid UUID |
| `created_date` | `time.Time` | **Yes** | Creation timestamp | - |

**Source:** [alertbus/model.go:50-56](business/domain/workflow/alertbus/model.go#L50-L56)

**Database:** `workflow.alert_recipients`

### AlertAcknowledgment

| Field | Type | Required | Description | Validation Rules |
|-------|------|----------|-------------|------------------|
| `id` | `uuid.UUID` | Auto | Primary key | Auto-generated |
| `alert_id` | `uuid.UUID` | **Yes** | FK to alerts | Must exist in alerts |
| `acknowledged_by` | `uuid.UUID` | **Yes** | User who acked | Must exist in users |
| `acknowledged_date` | `time.Time` | **Yes** | Ack timestamp | - |
| `notes` | `string` | No | Acknowledgment notes | - |

**Source:** [alertbus/model.go:59-65](business/domain/workflow/alertbus/model.go#L59-L65)

**Database:** `workflow.alert_acknowledgments`

---

## 14. Allowed Values (Whitelists)

### Trigger Event Types
```go
"on_create"   // Entity creation
"on_update"   // Entity update
"on_delete"   // Entity deletion
"scheduled"   // Time-based trigger
```
**Source:** [trigger.go:434-441](business/sdk/workflow/trigger.go#L434-L441)

### Execution Status
```go
"pending"     // Not yet started
"running"     // Currently executing
"completed"   // Successfully finished
"failed"      // Execution failed
"cancelled"   // Execution cancelled
```
**Source:** [models.go:67-73](business/sdk/workflow/models.go#L67-L73)

### Delivery Status
```go
"pending"     // Awaiting delivery
"sent"        // Sent to provider
"delivered"   // Confirmed delivered
"failed"      // Delivery failed
"bounced"     // Bounced back
"retrying"    // Retry in progress
```
**Source:** [models.go:424-431](business/sdk/workflow/models.go#L424-L431)

### Alert Severity
```go
"low"
"medium"
"high"
"critical"
```
**Source:** [alertbus/model.go:25-29](business/domain/workflow/alertbus/model.go#L25-L29)

### Alert Status
```go
"active"       // New, actionable
"acknowledged" // User acknowledged
"dismissed"    // User dismissed
"resolved"     // Auto-resolved by system
```
**Source:** [alertbus/model.go:17-22](business/domain/workflow/alertbus/model.go#L17-L22)

### Recipient Type
```go
"user"  // Individual user
"role"  // Role-based (all users with role)
```

### Notification Channels
```go
"email"
"sms"
"push"
"in_app"
```

### Approval Types
```go
"any"       // Any one approver
"all"       // All approvers
"majority"  // Majority of approvers
```

### Allocation Modes
```go
"reserve"   // Reserve inventory (temporary hold)
"allocate"  // Allocate inventory (permanent)
```

### Allocation Strategies
```go
"fifo"           // First in, first out
"lifo"           // Last in, first out
"nearest_expiry" // Closest to expiration
"lowest_cost"    // Cheapest warehouse
```

### Priority Levels
```go
"low"      // Priority 1
"medium"   // Priority 5
"high"     // Priority 10
"critical" // Priority 20
```

### Condition Operators (Trigger Evaluation)
```go
"equals"
"not_equals"
"changed_from"
"changed_to"
"greater_than"
"less_than"
"contains"
"in"
```
**Source:** [trigger.go:329-371](business/sdk/workflow/trigger.go#L329-L371)

### Condition Operators (update_field)
```go
"equals"
"not_equals"
"greater_than"
"less_than"
"contains"
"is_null"
"is_not_null"
"in"
"not_in"
```
**Source:** [workflowactions/data/updatefield.go:414-425](business/sdk/workflow/workflowactions/data/updatefield.go#L414-L425)

---

## 15. Existing Validation Rules

### TriggerProcessor.validateEvent()
1. Event type is required and must be supported
2. Entity name is required
3. Timestamp must be non-zero
4. Warns if no rules exist for entity
5. Warns if update event has no field changes

**Source:** [trigger.go:153-190](business/sdk/workflow/trigger.go#L153-L190)

### CreateAlertHandler.Validate()
1. Message is required
2. At least one recipient (user or role)
3. Severity must be valid if provided

**Source:** [workflowactions/communication/alert.go:59-81](business/sdk/workflow/workflowactions/communication/alert.go#L59-L81)

### UpdateFieldHandler.Validate()
1. target_entity is required and in whitelist
2. target_field is required
3. new_value is required
4. foreign_key_config required when field_type is foreign_key
5. All conditions must have valid operators

**Source:** [workflowactions/data/updatefield.go:61-114](business/sdk/workflow/workflowactions/data/updatefield.go#L61-L114)

### SendEmailHandler.Validate()
1. Recipients required (unless simulate_failure)
2. Subject required (unless simulate_failure)

**Source:** [workflowactions/communication/email.go:32-58](business/sdk/workflow/workflowactions/communication/email.go#L32-L58)

### SendNotificationHandler.Validate()
1. Recipients list required
2. At least one channel required
3. Priority must be valid

**Source:** [workflowactions/communication/notification.go:32-61](business/sdk/workflow/workflowactions/communication/notification.go#L32-L61)

### SeekApprovalHandler.Validate()
1. Approvers list required
2. approval_type must be: any, all, or majority

**Source:** [workflowactions/approval/seek.go:32-52](business/sdk/workflow/workflowactions/approval/seek.go#L32-L52)

### AllocateInventoryHandler.Validate()
1. inventory_items required unless source_from_line_item
2. allocation_strategy must be valid
3. allocation_mode must be valid
4. priority must be valid
5. Each item must have product_id and quantity > 0

**Source:** [workflowactions/inventory/allocate.go:179-222](business/sdk/workflow/workflowactions/inventory/allocate.go#L179-L222)

### TemplateProcessor.validateVariable()
1. Variable name must be non-empty
2. No invalid characters
3. Valid variable name pattern
4. Valid filter name pattern

**Source:** [template.go:574-607](business/sdk/workflow/template.go#L574-L607)

---

## 16. Proposed Validator Enhancements

### High Priority

1. **AutomationRule validation**
   - entity_id must reference valid entity
   - entity_type_id must reference valid entity_type
   - trigger_type_id must reference valid trigger_type
   - All foreign key references must exist

2. **RuleAction validation**
   - automation_rule_id must reference valid rule
   - template_id must reference valid template if provided
   - action_config must validate against inferred action type
   - execution_order uniqueness within rule

3. **RuleDependency validation**
   - No circular dependencies (graph cycle detection)
   - Parent and child rules must exist
   - Parent and child must be different

4. **Template variable validation**
   - Validate variable syntax in all action configs
   - Check filter names against built-in filters
   - Warn about potentially missing context variables

### Medium Priority

1. **ActionTemplate validation**
   - action_type must be registered
   - default_config must validate for action_type

2. **Cross-reference validation**
   - User UUIDs in recipients must exist in users
   - Role UUIDs in recipients must exist in roles
   - Entity references map to real database tables

3. **Alert configuration validation**
   - At least one active rule should create alerts for each alert_type
   - Recipients should have notification preferences set

4. **Consistency validation**
   - Entity in rule should match entity_type
   - Active rules should have at least one active action

### Low Priority (Schema Validation)

1. **Database schema validation**
   - Verify target_entity tables exist
   - Verify target_field columns exist
   - Validate foreign key relationships against actual DB

2. **Workflow completeness checks**
   - All entities should have at least one rule
   - Alert types should have handling rules
   - All action types should be registered

### Security Validations

1. **SQL injection prevention** - Already in place via table whitelist
2. **Template variable sanitization** - Validate against context
3. **Permission validation** - Users must have workflow management permissions

---

## Implementation Notes

### Validator Function Signatures
```go
// Rule-level validation
func (r *AutomationRule) Validate() error
func (r *AutomationRule) ValidateWithDependencies(ctx context.Context, bus *Business) error

// Action config validation (polymorphic)
func ValidateActionConfig(actionType string, config json.RawMessage) error

// Template validation
func ValidateTemplateVariables(template string, availableContext []string) error

// Dependency graph validation
func ValidateDependencyGraph(dependencies []RuleDependency) error
```

### Error Types
Consider creating workflow-specific errors:
- `ErrInvalidRule` - Rule configuration error
- `ErrInvalidAction` - Action configuration error
- `ErrInvalidCondition` - Condition configuration error
- `ErrCircularDependency` - Dependency cycle detected
- `ErrMissingReference` - Foreign key doesn't exist
- `ErrInvalidTemplate` - Template syntax error
- `ErrUnknownActionType` - Action type not registered

### Validation Approach
1. Validate structure (required fields, types)
2. Validate values (whitelists, formats)
3. Validate references (FKs exist, entities valid)
4. Validate templates (syntax, available variables)
5. Validate dependencies (no cycles, proper ordering)
6. Optional: Validate against database schema

---

## 17. CLI Command: validate-workflows

Similar to the `validate-configs` command for table configurations, a CLI command can validate workflow configurations.

### Command Location

**File:** `api/cmd/tooling/admin/commands/validateworkflows.go`

### Command Registration

**File:** `api/cmd/tooling/admin/main.go`

Add to the switch statement:
```go
case "validate-workflows":
    if err := commands.ValidateWorkflows(); err != nil {
        return fmt.Errorf("validating workflows: %w", err)
    }
```

Add to help output:
```go
fmt.Println("validate-workflows: validate workflow rules and action configurations")
```

### Implementation Pattern

Following the pattern from `validateconfigs.go`:

```go
// api/cmd/tooling/admin/commands/validateworkflows.go
package commands

import (
    "fmt"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
)

// workflowEntry pairs a workflow config with its name for validation reporting
type workflowEntry struct {
    name        string
    actionType  string
    config      []byte
}

// ValidateWorkflows validates all seed workflow configurations.
// Returns nil if all configs are valid, otherwise returns validation errors.
// This command does not require a database connection.
func ValidateWorkflows() error {
    // Create action registry with all handlers
    registry := workflow.NewActionRegistry()
    // Note: Handlers need to be registered without DB dependencies for validation-only mode

    // Collect all action configs to validate from testutil.go seed data
    configs := collectWorkflowConfigs()

    var (
        hasErrors    bool
        validCount   int
        invalidCount int
        warnCount    int
    )

    fmt.Println("Validating workflow configurations...")
    fmt.Println()

    for _, entry := range configs {
        handler, exists := registry.Get(entry.actionType)
        if !exists {
            invalidCount++
            fmt.Printf("❌ %s: unknown action type %s\n", entry.name, entry.actionType)
            hasErrors = true
            continue
        }

        if err := handler.Validate(entry.config); err != nil {
            hasErrors = true
            invalidCount++
            fmt.Printf("❌ %s:\n", entry.name)
            fmt.Printf("   • %s\n", err.Error())
        } else {
            validCount++
            fmt.Printf("✓ %s\n", entry.name)
        }
    }

    fmt.Println()
    fmt.Printf("Summary: %d valid, %d invalid, %d warnings\n", validCount, invalidCount, warnCount)

    if hasErrors {
        return fmt.Errorf("validation failed: %d workflow config(s) have errors", invalidCount)
    }

    fmt.Println("\nAll workflow configurations valid!")
    return nil
}

func collectWorkflowConfigs() []workflowEntry {
    // Collect from testutil.go or seed data
    // This would include sample action configs for each action type
    return []workflowEntry{
        // Add entries from seed data or test configurations
    }
}
```

### Usage

```bash
# Run workflow validation
make admin validate-workflows

# Or directly
go run api/cmd/tooling/admin/main.go validate-workflows
```

### Output Format

```
Validating workflow configurations...

✓ LowInventoryAlert
✓ OrderCreatedNotification
❌ InvalidEmailAction:
   • email recipients list is required and must not be empty
✓ StatusUpdateAction
   ⚠ target_entity: consider using fully qualified schema.table name

Summary: 3 valid, 1 invalid, 1 warnings

validation failed: 1 workflow config(s) have errors
```

### Validation Categories

The CLI command should validate:

1. **Structural validation** - Required fields present, correct types
2. **Value validation** - Enum values in whitelists, UUIDs are valid format
3. **Reference validation** (optional, requires DB) - FKs exist
4. **Template validation** - Variable syntax is valid
5. **Dependency validation** - No circular dependencies

### Makefile Integration

Add to Makefile:
```makefile
validate-workflows:
	go run api/cmd/tooling/admin/main.go validate-workflows
```

---

## Critical Files Reference

| Purpose | File Path |
|---------|-----------|
| Core models | `business/sdk/workflow/models.go` |
| Trigger processing | `business/sdk/workflow/trigger.go` |
| Template processing | `business/sdk/workflow/template.go` |
| Action interfaces | `business/sdk/workflow/interfaces.go` |
| Action registration | `business/sdk/workflow/workflowactions/register.go` |
| create_alert action | `business/sdk/workflow/workflowactions/communication/alert.go` |
| update_field action | `business/sdk/workflow/workflowactions/data/updatefield.go` |
| send_email action | `business/sdk/workflow/workflowactions/communication/email.go` |
| send_notification action | `business/sdk/workflow/workflowactions/communication/notification.go` |
| seek_approval action | `business/sdk/workflow/workflowactions/approval/seek.go` |
| allocate_inventory action | `business/sdk/workflow/workflowactions/inventory/allocate.go` |
| Alert models | `business/domain/workflow/alertbus/model.go` |
| Database migrations | `business/sdk/migrate/sql/migrate.sql` (lines 938-1844) |
