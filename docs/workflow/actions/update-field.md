# update_field Action

Updates database fields dynamically based on configuration. Supports conditional updates, foreign key resolution, and template variables.

## Configuration Schema

```json
{
  "target_entity": "schema.table",
  "target_field": "column_name",
  "new_value": "value or {{template}}",
  "field_type": "string|number|date|uuid|foreign_key",
  "foreign_key_config": {
    "reference_table": "schema.table",
    "lookup_field": "field_name",
    "id_field": "id",
    "allow_create": false,
    "create_config": {}
  },
  "conditions": [],
  "timeout_ms": 5000
}
```

**Source**: `business/sdk/workflow/workflowactions/data/updatefield.go:19-28`

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `target_entity` | string | **Yes** | - | Table to update (must be in whitelist) |
| `target_field` | string | **Yes** | - | Column to update |
| `new_value` | any | **Yes** | - | New value (template supported) |
| `field_type` | string | No | `string` | Value type for conversion |
| `foreign_key_config` | object | Conditional | - | Required if field_type is `foreign_key` |
| `conditions` | []FieldCondition | No | - | WHERE conditions |
| `timeout_ms` | int | No | 5000 | Operation timeout |

### Field Types

| Type | Description |
|------|-------------|
| `string` | String value (default) |
| `number` | Numeric value |
| `date` | Date/timestamp |
| `uuid` | UUID value |
| `foreign_key` | Foreign key with lookup |

### Foreign Key Config

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `reference_table` | string | **Yes** | - | FK target table |
| `lookup_field` | string | **Yes** | - | Field to match |
| `id_field` | string | No | `id` | ID field name |
| `allow_create` | bool | No | `false` | Create if not found |
| `create_config` | object | Conditional | - | Defaults for creation |

### Condition Operators

| Operator | Description |
|----------|-------------|
| `equals` | Exact match |
| `not_equals` | Not equal |
| `greater_than` | Greater than |
| `less_than` | Less than |
| `contains` | Substring match |
| `is_null` | Field is NULL |
| `is_not_null` | Field is NOT NULL |
| `in` | Value in array |
| `not_in` | Value not in array (⚠️ see note) |

**Source**: `business/sdk/workflow/workflowactions/data/updatefield.go:414-425`

> **Note**: The `not_in` operator is validated but currently falls through to `equals` in query building. Use `not_equals` with multiple conditions as an alternative.

## Table Whitelist

Only these tables can be updated (security feature):

**Sales schema:**
- `sales.customers`, `sales.orders`, `sales.order_line_items`
- `sales.order_fulfillment_statuses`, `sales.line_item_fulfillment_statuses`

**Products schema:**
- `products.products`, `products.brands`, `products.product_categories`
- `products.physical_attributes`, `products.product_costs`
- `products.cost_history`, `products.quality_metrics`

**Inventory schema:**
- `inventory.inventory_items`, `inventory.inventory_locations`
- `inventory.inventory_transactions`, `inventory.warehouses`
- `inventory.zones`, `inventory.lot_trackings`, `inventory.serial_numbers`
- `inventory.inspections`, `inventory.inventory_adjustments`
- `inventory.transfer_orders`

**Procurement schema:**
- `procurement.suppliers`, `procurement.supplier_products`

**Core schema:**
- `core.users`, `core.roles`, `core.user_roles`
- `core.contact_infos`, `core.table_access`

**HR schema:**
- `hr.offices`

**Geography schema:**
- `geography.countries`, `geography.regions`
- `geography.cities`, `geography.streets`

**Assets schema:**
- `assets.assets`, `assets.valid_assets`

**Config schema:**
- `config.table_configs`

**Workflow schema:**
- `workflow.automation_rules`, `workflow.rule_actions`
- `workflow.action_templates`, `workflow.rule_dependencies`
- `workflow.trigger_types`, `workflow.entity_types`, `workflow.entities`
- `workflow.automation_executions`, `workflow.notification_deliveries`

**Source**: `business/sdk/workflow/workflowactions/data/updatefield.go:373-403`

## Validation Rules

1. `target_entity` is required and must be in whitelist
2. `target_field` is required
3. `new_value` is required
4. `foreign_key_config` required when `field_type` is `foreign_key`
5. All conditions must have valid operators

**Source**: `business/sdk/workflow/workflowactions/data/updatefield.go:61-114`

## Example Configurations

### Simple Field Update

```json
{
  "target_entity": "sales.orders",
  "target_field": "status",
  "new_value": "processed"
}
```

### Update with Template

```json
{
  "target_entity": "sales.orders",
  "target_field": "processed_by",
  "new_value": "{{user_id}}",
  "field_type": "uuid"
}
```

### Update with Conditions

```json
{
  "target_entity": "sales.orders",
  "target_field": "status",
  "new_value": "shipped",
  "conditions": [
    {
      "field_name": "id",
      "operator": "equals",
      "value": "{{entity_id}}"
    },
    {
      "field_name": "status",
      "operator": "equals",
      "value": "pending"
    }
  ]
}
```

### Foreign Key Update with Lookup

```json
{
  "target_entity": "sales.orders",
  "target_field": "fulfillment_status_id",
  "new_value": "shipped",
  "field_type": "foreign_key",
  "foreign_key_config": {
    "reference_table": "sales.order_fulfillment_statuses",
    "lookup_field": "name",
    "id_field": "id"
  }
}
```

### Foreign Key with Auto-Create

```json
{
  "target_entity": "products.products",
  "target_field": "category_id",
  "new_value": "{{category_name}}",
  "field_type": "foreign_key",
  "foreign_key_config": {
    "reference_table": "products.product_categories",
    "lookup_field": "name",
    "id_field": "id",
    "allow_create": true,
    "create_config": {
      "description": "Auto-created category"
    }
  }
}
```

## Execution Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        UpdateFieldHandler                                    │
│                                                                             │
│  1. Validation Phase                                                        │
│     ├── Validate configuration structure                                    │
│     ├── Check table names against whitelist                                 │
│     └── Validate operators and conditions                                   │
│                                                                             │
│  2. Template Processing                                                     │
│     ├── Process template variables in new_value                             │
│     └── Build context from execution metadata                               │
│                                                                             │
│  3. Foreign Key Resolution (if applicable)                                  │
│     ├── Look up ID from display value                                       │
│     └── Create new record if allowed and not found                          │
│                                                                             │
│  4. Database Update                                                         │
│     ├── Build SQL UPDATE query with conditions                              │
│     └── Execute update and return affected row count                        │
│                                                                             │
│  5. Result Generation                                                       │
│     └── Return status, records affected, execution time                     │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Security Features

| Feature | Description |
|---------|-------------|
| **Table Whitelist** | Only predefined tables can be updated |
| **Operator Validation** | SQL operators restricted to safe set |
| **Parameterized Queries** | Named parameters prevent SQL injection |
| **Transaction Support** | Can participate in larger transactions |

## Common Use Cases

1. **Update order status** when conditions are met
2. **Set foreign key relationships** using lookup values
3. **Bulk update records** matching specific criteria
4. **Cascading field updates** in workflow automation

## Error Handling

- Validates all inputs before execution
- Returns detailed error messages in result
- Logs operations for debugging
- Handles missing foreign keys gracefully (create or fail)

## Related Documentation

- [Templates](../configuration/templates.md) - Template variables and filters
- [Rules](../configuration/rules.md) - How to create automation rules
- [Architecture](../architecture.md) - How action handlers are executed

## Key Files

| File | Purpose |
|------|---------|
| `business/sdk/workflow/workflowactions/data/updatefield.go` | Handler implementation |
