# Default Status Management: Form Configuration + Workflow Engine

## Conclusion

**Use a layered approach**:
1. **Form-level configuration** sets initial default statuses (synchronous, deterministic)
2. **Workflow engine** handles status transitions based on business logic (async, conditional)

This separates concerns cleanly:
- Template/Form layer: Deterministic defaults, user input
- Workflow layer: Business logic, conditional transitions, async operations

---

## Architecture Pattern

```
Order Creation via FormData
    │
    ├─► Form Config Default Resolution:
    │   - fulfillment_status_id default_value: "Pending" → resolved to UUID
    │   - line_item_fulfillment_statuses_id default_value: "Pending" → resolved to UUID
    │   - created_by = {{$me}}, created_date = {{$now}} (template vars unchanged)
    │
    ├─► Database: Order + Line Items committed with resolved default status UUIDs
    │
    └─► Workflow Engine (async, post-commit)
        │
        ├─► Rule: "On Order Create" → allocate_inventory action
        │
        ├─► Allocation Success → update_field: Line items → ALLOCATED
        │
        └─► Allocation Failure → create_alert: Notify role/user
```

### Key Design Decision: Form Configuration Default Value Resolution

**Approach: Form Config Handles FK Default Resolution**
- Default values in form configuration use human-readable names (e.g., `"Pending"`)
- The formdata package resolves names to UUIDs during form processing
- No changes needed to template variable system (`{{$me}}`, `{{$now}}` stay separate)

**Why Form Config Resolution:**
- Keeps form configurations readable and maintainable
- Different environments may have different status UUIDs (dev vs prod)
- Status names are stable; UUIDs are not
- Admin can configure defaults without knowing UUIDs
- Resolution happens at formdata processing time

**Form Config Default Value Pattern:**
```json
{
  "name": "fulfillment_status_id",
  "type": "smart-combobox",
  "entity": "sales.order_fulfillment_statuses",
  "default_value": "Pending",
  "default_mode": "create"
}
```

The formdata package:
1. Sees `default_value: "Pending"` for an FK field with `entity: "sales.order_fulfillment_statuses"`
2. Queries the entity table for a record where `name = "Pending"` (or configurable lookup field)
3. Returns the UUID of that record as the resolved default value

---

## Current State

### Template Magic Values (`{{$me}}`, `{{$now}}`)
- Location: `business/sdk/workflow/template.go`
- Synchronous, resolved during formdata processing
- Used for: created_by, updated_by, created_date, updated_date
- Cannot do: database lookups, FK resolution, conditional logic

### Workflow Engine
- Location: `business/sdk/workflow/`
- Async, event-driven, triggers after commit
- Actions: `allocate_inventory`, `update_field`, `create_alert`, `seek_approval`
- Can do: FK resolution, conditional transitions, complex business logic

### Current Gap
- Status fields in `forms.go:284-292` and `tableforms.go:888` are `Required: true` but have no default
- Users must manually select "PENDING" every time

---

## PHASE 1: Form Configuration FK Default Resolution

### Objective
Enable form fields to specify default values by name (e.g., `"Pending"`) for FK fields. The formdata package resolves these names to UUIDs during form processing by querying the referenced entity table.

### Target Configuration
```json
{
  "name": "fulfillment_status_id",
  "type": "smart-combobox",
  "entity": "sales.order_fulfillment_statuses",
  "default_value": "Pending",
  "default_mode": "create"
}
```

When processed, the formdata package:
1. Detects this is an FK field (has `entity` specified)
2. Queries `sales.order_fulfillment_statuses` for `name = "Pending"`
3. Resolves `default_value` to the matching UUID
4. Applies the UUID as the field's default

### Files to Modify

1. **`app/domain/formdata/formdataapp/model.go`**
   - `FieldDefaultConfig` already has `DefaultValue`, `DefaultValueCreate`, `DefaultValueUpdate`
   - May need to add `DefaultLookupField` to specify which column to match (default: `name`)

2. **`app/domain/formdata/formdataapp/formdataapp.go`**
   - Update `mergeFieldDefaults()` to detect FK fields with string default values
   - Add FK resolution logic: query entity table, find record by name, return UUID
   - Handle resolution failures gracefully (log warning, skip default)

3. **`business/sdk/dbtest/seedmodels/forms.go`**
   - `GetFullSalesOrderFormFields()` lines 284-292: Add `DefaultValueCreate: "Pending"`

4. **`business/sdk/dbtest/seedmodels/tableforms.go`**
   - `GetSalesOrderFormFields()` line 888: Add `DefaultValueCreate: "Pending"`
   - `GetSalesOrderLineItemFormFields()` line 903: Add `DefaultValueCreate: "Pending"`

### Implementation Steps

1. **Add FK default resolution to formdataapp** - Query entity table by name when default_value is not a UUID
2. **Wire database access** - formdataapp needs ability to query arbitrary entity tables
3. **Update form seed data** - Set `DefaultValueCreate: "Pending"` on status FK fields
4. **Test** - Orders/line items get Pending status without user input

### Validation
- Form config with `default_value: "Pending"` resolves to correct UUID
- Orders created via formdata have correct fulfillment_status_id (Pending UUID)
- Line items have correct line_item_fulfillment_statuses_id (Pending UUID)
- Invalid status names produce clear validation errors
- Users can override default if field is visible

---

## PHASE 2: Workflow Integration for Status Transitions

### Objective
Wire automation rules to trigger allocation on order creation and update statuses based on results.

### Dependencies
- Phase 1 complete
- Entity types seeded: orders, order_line_items, allocation_results

### Files to Modify

1. **`business/sdk/workflow/workflowactions/inventory/allocate.go`**
   - After `ProcessAllocation()` line 393, fire workflow event for allocation_results

2. **`business/sdk/dbtest/seedFrontend.go`**
   - Add automation rules (see below)

3. **`business/sdk/workflow/workflowactions/data/updatefield.go`**
   - Verify batch update works for multiple line items

### Automation Rules to Seed

**Rule 1: Order Created → Allocate Inventory**
```json
{
  "rule_name": "Sales Order Created - Allocate Inventory",
  "entity_name": "orders",
  "trigger_type": "on_create",
  "is_active": true,
  "actions": [{
    "action_type": "allocate_inventory",
    "config": {
      "allocation_mode": "reserve",
      "allocation_strategy": "fifo",
      "allow_partial": false,
      "priority": "medium",
      "reference_id": "{{entity_id}}",
      "reference_type": "order"
    }
  }]
}
```

**Rule 2: Allocation Success → Update Status**
```json
{
  "rule_name": "Allocation Success - Update Line Items",
  "entity_name": "allocation_results",
  "trigger_type": "on_create",
  "conditions": [{"field": "status", "operator": "equals", "value": "success"}],
  "actions": [{
    "action_type": "update_field",
    "config": {
      "target_entity": "order_line_items",
      "target_field": "line_item_fulfillment_statuses_id",
      "new_value": "ALLOCATED",
      "field_type": "foreign_key",
      "foreign_key_config": {
        "reference_table": "sales.line_item_fulfillment_statuses",
        "lookup_field": "name"
      },
      "conditions": [{"field": "order_id", "operator": "equals", "value": "{{reference_id}}"}]
    }
  }]
}
```

**Rule 3: Allocation Failure → Alert**
```json
{
  "rule_name": "Allocation Failed - Alert Ops",
  "entity_name": "allocation_results",
  "trigger_type": "on_create",
  "conditions": [{"field": "status", "operator": "equals", "value": "failed"}],
  "actions": [{
    "action_type": "create_alert",
    "config": {
      "alert_type": "inventory_allocation_failed",
      "severity": "high",
      "message": "Allocation failed for order: {{failure_reason}}"
    }
  }]
}
```

### Missing Code: Fire Event After Allocation

Add to `allocate.go` after line 393:
```go
triggerEvent := workflow.TriggerEvent{
    EventType:  "on_create",
    EntityName: "allocation_results",
    EntityID:   result.AllocationID,
    RawData: map[string]interface{}{
        "status":           result.Status,
        "reference_id":     request.Config.ReferenceID,
        "reference_type":   request.Config.ReferenceType,
        "total_allocated":  result.TotalAllocated,
        "total_requested":  result.TotalRequested,
        "failed_items":     result.FailedItems,
    },
}
// Need workflow engine access - inject via handler constructor
```

### Validation
- Order creation triggers allocation workflow
- Success updates line items to ALLOCATED
- Failure keeps PENDING and creates alert

---

## PHASE 3: Alert System Enhancement

### Objective
Extend alert action to support role-based recipients and proper delivery.

### Dependencies
- Phase 2 complete
- Decision on alert delivery model

### Files to Explore
- `business/sdk/workflow/workflowactions/communication/alert.go`

### Tables to Create
- `workflow.alerts` - Alert records
- `workflow.alert_recipients` - User/role routing
- `workflow.alert_acknowledgments` - Acknowledgment tracking

### Target Configuration
```json
{
  "action_type": "create_alert",
  "config": {
    "alert_type": "inventory_allocation_failed",
    "severity": "high",
    "recipients": {
      "type": "role",
      "value": "inventory_manager"
    },
    "message": "Allocation failed for order {{order_number}}: {{failure_reason}}",
    "context": {
      "order_id": "{{entity_id}}",
      "failed_items": "{{failed_items}}"
    }
  }
}
```

### Implementation Steps
1. Define alert routing model (user vs role recipients)
2. Create database tables
3. Implement delivery (in-app, email, both)
4. Add UI for viewing/acknowledging alerts

### Open Questions
- In-app only or also email?
- Default roles for inventory alerts?
- SLA timers for acknowledgment?

---

## Technical Notes

### Already Working
- `allocate.go` queues to RabbitMQ async - order creation doesn't block
- `updatefield.go:254-325` handles FK resolution via `ForeignKeyConfig`
- Template variables like `{{entity_id}}` resolve from `ActionExecutionContext`
- Table whitelist includes `order_fulfillment_statuses`, `line_item_fulfillment_statuses`

### File References

**Core Files**
- `business/sdk/workflow/template.go` - Magic value processing
- `app/domain/formdata/formdataapp/formdataapp.go` - Form data processing
- `business/sdk/workflow/workflowactions/inventory/allocate.go` - Allocation action
- `business/sdk/workflow/workflowactions/data/updatefield.go` - Field update action

**Form Definitions**
- `business/sdk/dbtest/seedmodels/forms.go:224-330` - Full sales order form
- `business/sdk/dbtest/seedmodels/tableforms.go:883-894` - Order form fields
- `business/sdk/dbtest/seedmodels/tableforms.go:897-910` - Line item form fields

**Status Business Layer**
- `business/domain/sales/orderfulfillmentstatusbus/`
- `business/domain/sales/lineitemfulfillmentstatusbus/`
