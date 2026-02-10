# Workflow Engine Documentation

The Ichor workflow engine provides event-driven automation for business processes. When entities are created, updated, or deleted, the engine evaluates automation rules and executes configured actions.

## Quick Links

| Document | Description |
|----------|-------------|
| [Architecture](architecture.md) | System overview, event flow, components |
| [Configuration](configuration/) | Triggers, rules, and template variables |
| [Actions](actions/) | All 7 action types and their configuration |
| [Branching](branching.md) | Graph-based execution and conditional workflows |
| [Cascade Visualization](cascade-visualization.md) | Downstream workflow detection |
| [Database Schema](database-schema.md) | Workflow tables and relationships |
| [API Reference](api-reference.md) | REST endpoints for alerts |
| [Event Infrastructure](event-infrastructure.md) | Delegate pattern and workflow dispatch |
| [Temporal Integration](temporal.md) | Temporal architecture, GraphExecutor, activities |
| [Worker Deployment](worker-deployment.md) | Workflow-worker service operations |
| [Migration from RabbitMQ](migration-from-rabbitmq.md) | What changed and why |
| [Testing](testing.md) | Testing patterns and examples |
| [Adding Domains](adding-domains.md) | How to add workflow events to new domains |

## Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WORKFLOW ENGINE                                    │
│                                                                             │
│  DelegateHandler ──► WorkflowTrigger ──► Temporal ──► Worker ──► Activities │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Core Concepts

| Concept | Description |
|---------|-------------|
| **TriggerEvent** | An event fired when an entity changes (create/update/delete) |
| **AutomationRule** | Defines conditions that must match for actions to execute |
| **RuleAction** | A configured action to execute when a rule matches |
| **ActionHandler** | Executes a specific action type (email, alert, etc.) |

### Supported Actions

| Action Type | Description |
|-------------|-------------|
| `create_alert` | Creates in-app alerts for users/roles |
| `update_field` | Updates database fields dynamically |
| `send_email` | Sends email notifications |
| `send_notification` | Multi-channel notifications (email, SMS, push, in-app) |
| `seek_approval` | Initiates approval workflows |
| `allocate_inventory` | Reserves or allocates inventory |
| `evaluate_condition` | Evaluates conditions for branching workflows |

## Getting Started

### 1. Understanding the Event Flow

1. An entity is created/updated/deleted via API or formdata
2. The business layer fires a delegate event
3. `TemporalDelegateHandler` converts it to a `TriggerEvent`
4. `WorkflowTrigger` evaluates matching automation rules
5. For each matched rule, loads graph definition from DB
6. Dispatches workflow to Temporal (`workflow-worker` picks up)
7. `GraphExecutor` traverses action graph, executing each action as an activity

### 2. Creating an Automation Rule

Automation rules are stored in the `workflow.automation_rules` table. A rule requires:

- **Entity**: Which table/entity to monitor (e.g., `orders`)
- **Trigger Type**: When to fire (`on_create`, `on_update`, `on_delete`, `scheduled`)
- **Conditions** (optional): Field conditions that must match
- **Actions**: What to do when the rule fires

### 3. Adding Actions to a Rule

Actions are stored in `workflow.rule_actions`. Each action has:

- **Action Config**: JSON configuration specific to the action type
- **Action Edges**: Directed edges defining execution flow (see [Branching](branching.md))
- **Template Variables**: Dynamic values like `{{entity_id}}`, `{{customer_name}}`

See [Actions](actions/) for configuration details for each action type.

### 4. Action Execution

Rules with actions execute using graph-based BFS traversal via `action_edges`. Edges define the execution order and flow:

| Edge Type | Description |
|-----------|-------------|
| **start** | Entry point — connects to the first action (no source) |
| **sequence** | Unconditional flow from one action to the next |
| **true_branch** | Conditional path taken when `evaluate_condition` returns true |
| **false_branch** | Conditional path taken when `evaluate_condition` returns false |
| **always** | Always-execute path from a condition (runs regardless of branch) |

**All rules with actions must have edges.** Rules without actions (trigger-only rules) are valid and saved as inactive drafts.

See [Branching](branching.md) for complete documentation on conditional workflows.

## Configuration Reference

### Trigger Types

| Type | Description |
|------|-------------|
| `on_create` | Fires when entity is created |
| `on_update` | Fires when entity is updated |
| `on_delete` | Fires when entity is deleted |
| `scheduled` | Fires on a schedule (cron-based) |

### Condition Operators

| Operator | Description |
|----------|-------------|
| `equals` | Exact match |
| `not_equals` | Not equal |
| `changed_from` | Previous value match (updates only) |
| `changed_to` | New value match with change detection |
| `greater_than` | Numeric/string comparison |
| `less_than` | Numeric/string comparison |
| `contains` | Substring match |
| `in` | Value in array |
| `is_null` | Field is null/empty |
| `is_not_null` | Field has a value |

### Template Variables

Template variables use `{{variable}}` syntax:

```json
{
  "title": "Order {{number}} Created",
  "message": "Customer {{customer_name}} placed order {{number}}"
}
```

See [Templates](configuration/templates.md) for all available variables and filters.

## Key Files

| Purpose | Location |
|---------|----------|
| Core models | `business/sdk/workflow/models.go` |
| Trigger processing | `business/sdk/workflow/trigger.go` |
| Template processing | `business/sdk/workflow/template.go` |
| Temporal delegate handler | `business/sdk/workflow/temporal/delegatehandler.go` |
| Workflow trigger | `business/sdk/workflow/temporal/trigger.go` |
| Graph executor | `business/sdk/workflow/temporal/graph_executor.go` |
| Workflow implementation | `business/sdk/workflow/temporal/workflow.go` |
| Activities | `business/sdk/workflow/temporal/activities.go` |
| Edge store adapter | `business/sdk/workflow/temporal/stores/edgedb/edgedb.go` |
| Action handlers | `business/sdk/workflow/workflowactions/` |
| Worker service | `api/cmd/services/workflow-worker/main.go` |
| Alert business layer | `business/domain/workflow/alertbus/` |
| Alert API | `api/domain/http/workflow/alertapi/` |

## Database Tables

All workflow tables are in the `workflow` schema:

| Table | Purpose |
|-------|---------|
| `trigger_types` | Types of triggers (on_create, on_update, etc.) |
| `entity_types` | Categories of entities (table, view, etc.) |
| `entities` | Registered entities that can be monitored |
| `automation_rules` | Rule definitions |
| `rule_actions` | Actions attached to rules |
| `action_templates` | Reusable action configurations |
| `action_edges` | Directed edges for graph-based execution (branching) |
| `rule_dependencies` | Dependencies between rules |
| `automation_executions` | Execution history |
| `notification_deliveries` | Notification delivery tracking |
| `alerts` | Alert records |
| `alert_recipients` | Alert recipients (users/roles) |
| `alert_acknowledgments` | Alert acknowledgment tracking |

See [Database Schema](database-schema.md) for complete field definitions.
