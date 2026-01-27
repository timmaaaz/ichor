# Template Variables

Template variables allow dynamic value substitution in action configurations.

## Syntax

Variables use double curly braces: `{{variable_name}}`

```json
{
  "title": "Order {{number}} Created",
  "message": "Customer {{customer_name}} placed order totaling {{total | currency:USD}}"
}
```

## Variable Pattern

```regex
\{\{([^}]+)\}\}
```

Variable names must match:
```regex
^(\$[a-zA-Z][a-zA-Z0-9_]*|[a-zA-Z][a-zA-Z0-9._]*)$
```

**Source**: `business/sdk/workflow/template.go:574-607`

## Built-in Variables

Special variables available in all contexts:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `$me` | Current user's ID | `5cf37266-3473-4006-984f-9325122678b7` |
| `$now` | Current timestamp (RFC3339) | `2024-01-15T14:30:00Z` |

## Context Variables

Variables from the trigger event execution context:

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

## Entity Data Variables

Any field from the entity's data snapshot is available by field name:

```json
// If order data is:
{
  "id": "order-uuid",
  "number": "ORD-001",
  "customer_name": "Acme Corp",
  "total": 1500.00,
  "status": "pending"
}

// Then these variables are available:
// {{id}}, {{number}}, {{customer_name}}, {{total}}, {{status}}
```

## Field Change Variables

For `on_update` events, field changes are accessible through the `FieldChanges` map in the execution context.

> **Note**: The `old_{field_name}` and `new_{field_name}` prefix patterns are **not automatically created** as separate variables. Instead, field changes are available via the `FieldChanges` map structure.
>
> **Current limitation**: In Phase 2 of the workflow infrastructure, `FieldChanges` is `nil` for delegate-sourced events. Full change tracking (passing old vs new values through the delegate) is planned for Phase 3.

When `FieldChanges` is populated (e.g., when explicitly passed to `PublishUpdateEvent`), the changed field's current value is available directly by field name:

```json
{
  "message": "Status is now {{status}}"
}
```

## Filters

Variables support filters with pipe syntax: `{{variable | filter:arg}}`

### String Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `uppercase` | Convert to uppercase | `{{name \| uppercase}}` → `JOHN` |
| `lowercase` | Convert to lowercase | `{{name \| lowercase}}` → `john` |
| `capitalize` | Capitalize first letter | `{{name \| capitalize}}` → `John` |
| `trim` | Remove whitespace | `{{name \| trim}}` |
| `truncate:N` | Truncate to N characters | `{{description \| truncate:50}}` |
| `default:value` | Default if empty | `{{notes \| default:No notes}}` |

### Number Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `round:N` | Round to N decimal places | `{{total \| round:2}}` → `1500.00` |
| `currency:CODE` | Format as currency | `{{total \| currency:USD}}` → `$1,500.00` |

**Supported currency codes**: USD, EUR, GBP, JPY (no decimals), CAD, AUD, CHF, CNY, INR, MXN.

**Source**: `business/sdk/workflow/template.go:687-709`

### Date Filters

| Filter | Description | Example Output |
|--------|-------------|----------------|
| `formatDate:short` | Short date | `Jan 15, 2024` |
| `formatDate:long` | Long date with weekday | `Monday, January 15, 2024` |
| `formatDate:time` | Time only (24-hour) | `14:30:05` |
| `formatDate:datetime` | Date and time (24-hour) | `2024-01-15 14:30:05` |
| `formatDate:FORMAT` | Custom Go format | See Go time format |

Custom format example:
```
{{created_date | formatDate:2006-01-02}}  →  2024-01-15
```

**Source**: `business/sdk/workflow/template.go:765-777`

### Array Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `join:separator` | Join array elements | `{{tags \| join:, }}` → `urgent, high-priority` |
| `first` | Get first element | `{{items \| first}}` |
| `last` | Get last element | `{{items \| last}}` |

**Source**: `business/sdk/workflow/template.go:643-832`

## Filter Chaining

Filters can be chained:

```
{{name | trim | uppercase}}
{{amount | round:2 | currency:USD}}
```

## Examples

### Email Notification

```json
{
  "recipients": ["{{customer_email}}"],
  "subject": "Order {{number}} - {{status | capitalize}}",
  "body": "Dear {{customer_name | capitalize}},\n\nYour order {{number}} totaling {{total | currency:USD}} has been {{status}}.\n\nThank you for your business!"
}
```

### Alert Message

```json
{
  "title": "Low Stock Alert: {{product_name}}",
  "message": "Product {{product_name}} (SKU: {{sku}}) has only {{quantity}} units remaining. Minimum threshold is {{min_threshold}}.",
  "severity": "high"
}
```

### Status Change Notification

```json
{
  "title": "Status Update: {{number}}",
  "message": "Order {{number}} status changed from {{old_status | uppercase}} to {{new_status | uppercase}} by user {{user_id}}."
}
```

### With Defaults

```json
{
  "message": "Notes: {{notes | default:No notes provided}}\nPriority: {{priority | default:normal | uppercase}}"
}
```

## Validation

Template variables are validated for:

1. Non-empty variable name
2. No invalid characters
3. Valid variable name pattern
4. Valid filter name (must be registered)

Invalid templates will fail validation when the action is created.

**Source**: `business/sdk/workflow/template.go:574-607`

## Processing

The TemplateProcessor:

1. Parses the template string
2. Finds all `{{...}}` patterns
3. For each variable:
   - Looks up value in context
   - Applies filters in order
   - Replaces with result
4. Returns processed string

Missing variables are replaced with empty string (no error).

## Performance

- Templates are parsed once
- Variable lookup is O(1) from map
- Filter application is sequential
- No external I/O during processing
