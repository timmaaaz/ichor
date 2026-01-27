# Database Schema

All workflow tables are in the `workflow` schema.

## Core Tables

### workflow.trigger_types

Defines types of triggers that can fire automation rules.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `name` | TEXT | NO | - | Trigger type name (unique) |
| `description` | TEXT | YES | - | Human-readable description |
| `is_active` | BOOLEAN | NO | `true` | Whether active |
| `deactivated_by` | UUID | YES | - | FK to core.users |

**Standard values:**
- `on_create` - Entity creation
- `on_update` - Entity update
- `on_delete` - Entity deletion
- `scheduled` - Time-based trigger

### workflow.entity_types

Categories of entities that can be monitored.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `name` | TEXT | NO | - | Entity type name (unique) |
| `description` | TEXT | YES | - | Human-readable description |
| `is_active` | BOOLEAN | NO | `true` | Whether active |
| `deactivated_by` | UUID | YES | - | FK to core.users |

**Standard values:**
- `table` - Database table
- `view` - Database view

### workflow.entities

Registered entities (tables/views) that can be monitored.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `name` | TEXT | NO | - | Entity name (unique, table name only) |
| `entity_type_id` | UUID | NO | - | FK to entity_types |
| `schema_name` | TEXT | YES | `'public'` | Database schema |
| `is_active` | BOOLEAN | YES | `true` | Whether active |
| `created_date` | TIMESTAMPTZ | NO | `now()` | Creation timestamp |
| `deactivated_by` | UUID | YES | - | FK to core.users |

**Note**: Entity names are stored WITHOUT schema prefix (e.g., `orders` not `sales.orders`).

### workflow.automation_rules

Automation rule definitions.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `name` | TEXT | NO | - | Rule name |
| `description` | TEXT | YES | - | Description |
| `entity_id` | UUID | NO | - | FK to entities |
| `entity_type_id` | UUID | NO | - | FK to entity_types |
| `trigger_type_id` | UUID | NO | - | FK to trigger_types |
| `trigger_conditions` | JSONB | YES | - | Conditions (see below) |
| `is_active` | BOOLEAN | NO | `true` | Whether active |
| `created_date` | TIMESTAMPTZ | NO | `now()` | Creation timestamp |
| `updated_date` | TIMESTAMPTZ | NO | `now()` | Last update |
| `created_by` | UUID | NO | - | FK to core.users |
| `updated_by` | UUID | NO | - | FK to core.users |
| `deactivated_by` | UUID | YES | - | FK to core.users |

**trigger_conditions JSONB structure:**
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

### workflow.action_templates

Reusable action configuration templates.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `name` | TEXT | NO | - | Template name (unique) |
| `description` | TEXT | YES | - | Description |
| `action_type` | TEXT | NO | - | Action type identifier |
| `default_config` | JSONB | NO | - | Default configuration |
| `created_date` | TIMESTAMPTZ | NO | `now()` | Creation timestamp |
| `created_by` | UUID | NO | - | FK to core.users |
| `is_active` | BOOLEAN | YES | `true` | Whether active |
| `deactivated_by` | UUID | YES | - | FK to core.users |

### workflow.rule_actions

Actions attached to automation rules.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `automation_rules_id` | UUID | NO | - | FK to automation_rules |
| `name` | TEXT | NO | - | Action name |
| `description` | TEXT | YES | - | Description |
| `action_config` | JSONB | NO | - | Action configuration |
| `execution_order` | INT | NO | - | Execution order (≥1) |
| `is_active` | BOOLEAN | YES | `true` | Whether active |
| `template_id` | UUID | YES | - | FK to action_templates |
| `deactivated_by` | UUID | YES | - | FK to core.users |

### workflow.rule_dependencies

Dependencies between automation rules.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `parent_rule_id` | UUID | NO | - | FK to automation_rules |
| `child_rule_id` | UUID | NO | - | FK to automation_rules |

**Constraints:**
- Parent and child must be different rules
- No circular dependencies allowed

## Execution Tables

### workflow.automation_executions

Records of workflow executions.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `automation_rules_id` | UUID | NO | - | FK to automation_rules |
| `entity_type` | VARCHAR(50) | NO | - | Entity type that triggered |
| `trigger_data` | JSONB | YES | - | Trigger event data |
| `actions_executed` | JSONB | YES | - | Record of executed actions |
| `status` | VARCHAR(20) | NO | - | Execution status |
| `error_message` | TEXT | YES | - | Error if failed |
| `execution_time_ms` | INTEGER | YES | - | Execution duration in ms |
| `executed_at` | TIMESTAMP | NO | `now()` | Execution timestamp |

**Status values:**
- `success` - Successfully finished
- `failed` - Execution failed
- `partial` - Partially completed

### workflow.notification_deliveries

Tracks notification delivery status.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `notification_id` | UUID | NO | - | References NotificationPayload in queue |
| `automation_execution_id` | UUID | YES | - | FK to automation_executions |
| `rule_id` | UUID | YES | - | FK to automation_rules |
| `action_id` | UUID | YES | - | FK to rule_actions |
| `recipient_id` | UUID | NO | - | User ID or email |
| `channel` | VARCHAR(50) | NO | - | Delivery channel (email, sms, push, in_app) |
| `status` | VARCHAR(20) | NO | - | Delivery status |
| `attempts` | INTEGER | YES | `1` | Number of delivery attempts |
| `sent_at` | TIMESTAMP | YES | - | When sent |
| `delivered_at` | TIMESTAMP | YES | - | When delivered |
| `failed_at` | TIMESTAMP | YES | - | When failed |
| `error_message` | TEXT | YES | - | Error if failed |
| `provider_response` | JSONB | YES | - | Provider-specific response data |
| `created_date` | TIMESTAMP | NO | `now()` | Creation timestamp |
| `updated_date` | TIMESTAMP | NO | `now()` | Last update timestamp |

**Status values:**
- `pending` - Awaiting delivery
- `sent` - Sent to provider
- `delivered` - Confirmed delivered
- `failed` - Delivery failed
- `bounced` - Bounced back
- `retrying` - Retry in progress

### workflow.allocation_results

Stores idempotent inventory allocation results.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | - | Primary key |
| `idempotency_key` | VARCHAR(255) | NO | - | Unique key for deduplication |
| `allocation_data` | JSONB | NO | - | Allocation result data |
| `created_date` | TIMESTAMP | NO | - | Creation timestamp |

**Indexes:**
- `idx_allocation_idempotency` on `idempotency_key`

## Alert Tables

### workflow.alerts

Alert records created by `create_alert` action.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `alert_type` | TEXT | NO | - | Category |
| `severity` | TEXT | NO | `'medium'` | Severity level |
| `title` | TEXT | NO | - | Alert title |
| `message` | TEXT | NO | - | Alert message |
| `context` | JSONB | YES | `'{}'` | Additional data |
| `source_entity_name` | TEXT | YES | - | Triggering entity |
| `source_entity_id` | UUID | YES | - | Triggering entity ID |
| `source_rule_id` | UUID | YES | - | FK to automation_rules |
| `status` | TEXT | NO | `'active'` | Alert status |
| `expires_date` | TIMESTAMPTZ | YES | - | Expiration |
| `created_date` | TIMESTAMPTZ | NO | `now()` | Creation |
| `updated_date` | TIMESTAMPTZ | NO | `now()` | Last update |

**Severity values:**
- `low`
- `medium`
- `high`
- `critical`

**Status values:**
- `active` - New, actionable
- `acknowledged` - User acknowledged
- `dismissed` - User dismissed
- `resolved` - Auto-resolved by system

### workflow.alert_recipients

Maps alerts to their recipients.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `alert_id` | UUID | NO | - | FK to alerts (CASCADE) |
| `recipient_type` | TEXT | NO | - | "user" or "role" |
| `recipient_id` | UUID | NO | - | User or Role UUID |
| `created_date` | TIMESTAMPTZ | NO | `now()` | Creation |

### workflow.alert_acknowledgments

Tracks alert acknowledgments.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | NO | `gen_random_uuid()` | Primary key |
| `alert_id` | UUID | NO | - | FK to alerts (CASCADE) |
| `acknowledged_by` | UUID | NO | - | FK to core.users |
| `acknowledged_date` | TIMESTAMP | NO | - | When acknowledged |
| `notes` | TEXT | YES | - | Optional notes |

## Indexes

Key indexes for performance:

```sql
-- Rules by entity and trigger
CREATE INDEX idx_automation_rules_entity ON workflow.automation_rules(entity_id);
CREATE INDEX idx_automation_rules_trigger ON workflow.automation_rules(trigger_type_id);
CREATE INDEX idx_automation_rules_active ON workflow.automation_rules(is_active);

-- Actions by rule
CREATE INDEX idx_rule_actions_rule ON workflow.rule_actions(automation_rule_id);
CREATE INDEX idx_rule_actions_order ON workflow.rule_actions(execution_order);

-- Executions by rule and status
CREATE INDEX idx_automation_executions_rule ON workflow.automation_executions(rule_id);
CREATE INDEX idx_automation_executions_status ON workflow.automation_executions(status);
CREATE INDEX idx_automation_executions_date ON workflow.automation_executions(executed_at);

-- Alerts by status and recipient
CREATE INDEX idx_alerts_status ON workflow.alerts(status);
CREATE INDEX idx_alerts_created ON workflow.alerts(created_date);
CREATE INDEX idx_alert_recipients_alert ON workflow.alert_recipients(alert_id);
CREATE INDEX idx_alert_recipients_recipient ON workflow.alert_recipients(recipient_type, recipient_id);
```

## Foreign Keys

All foreign key relationships:

| Table | Column | References |
|-------|--------|------------|
| entities | entity_type_id | entity_types(id) |
| entities | deactivated_by | core.users(id) |
| automation_rules | entity_id | entities(id) |
| automation_rules | entity_type_id | entity_types(id) |
| automation_rules | trigger_type_id | trigger_types(id) |
| automation_rules | created_by | core.users(id) |
| automation_rules | updated_by | core.users(id) |
| action_templates | created_by | core.users(id) |
| rule_actions | automation_rules_id | automation_rules(id) |
| rule_actions | template_id | action_templates(id) |
| rule_dependencies | parent_rule_id | automation_rules(id) |
| rule_dependencies | child_rule_id | automation_rules(id) |
| automation_executions | automation_rules_id | automation_rules(id) |
| alerts | source_rule_id | automation_rules(id) |
| alert_recipients | alert_id | alerts(id) CASCADE |
| alert_acknowledgments | alert_id | alerts(id) CASCADE |
| alert_acknowledgments | acknowledged_by | core.users(id) |

## Migration Reference

Workflow tables are created in migrations versions 1.64-1.96 in `business/sdk/migrate/sql/migrate.sql`:

- **1.64-1.65**: trigger_types, entity_types
- **1.66-1.67**: automation_rules, automation_executions
- **1.68-1.69**: action_templates, rule_actions
- **1.70-1.71**: rule_dependencies, entities
- **1.72-1.73**: notification_deliveries, allocation_results
- **1.94-1.96**: alerts, alert_recipients, alert_acknowledgments
