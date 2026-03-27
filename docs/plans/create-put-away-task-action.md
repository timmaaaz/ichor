# Requirements: `create_put_away_task` Workflow Action

## Summary

New workflow action handler that creates a put-away task when triggered by a PO line item receive event. Ships with a seeded default workflow rule so receiving automatically generates put-away work for floor workers.

---

## Motivation

Today, receiving and put-away are completely disconnected. After a floor worker receives items against a PO, someone must manually create a put-away task via `POST /v1/inventory/put-away-tasks`. In a real warehouse, received goods sit on the dock with no system-generated instruction to move them to storage. This action closes that gap.

---

## Action Handler: `create_put_away_task`

### Registration

- **File:** New file `business/sdk/workflow/workflowactions/inventory/createputawaytask.go`
- **Register in:** `register.go` → add `PutAwayTask *putawaytaskbus.Business` to `BusDependencies`, register inside `RegisterGranularInventoryActions` (or new helper) with nil-guard on the bus dependency
- **Interface:** Implements `workflow.ActionHandler`

### Config Schema

```go
type CreatePutAwayTaskConfig struct {
    // When true, resolves product_id from trigger context's supplier_product_id
    // via supplierProductBus.QueryByID. When false, uses ProductID from config.
    SourceFromPO bool `json:"source_from_po"`

    // Static product UUID — used only when source_from_po is false.
    ProductID string `json:"product_id,omitempty"`

    // How to determine the destination location:
    //   "po_delivery" — use the PO's delivery_location_id (requires PO lookup)
    //   "static"      — use the LocationID field below
    LocationStrategy string `json:"location_strategy"`

    // Static location UUID — used when location_strategy is "static".
    LocationID string `json:"location_id,omitempty"`

    // Template string for reference_number. Supports {{variable}} syntax.
    // Example: "PO-RCV-{{purchase_order_id}}"
    ReferenceNumber string `json:"reference_number,omitempty"`
}
```

### Execute Logic

```
1. Unmarshal config
2. Resolve product_id:
   if source_from_po:
     supplier_product_id = execCtx.RawData["supplier_product_id"]
     supplierProduct = supplierProductBus.QueryByID(supplier_product_id)
     product_id = supplierProduct.ProductID
   else:
     product_id = config.ProductID

3. Determine quantity:
   quantity_received = execCtx.RawData["quantity_received"] (current total)
   // The delta (newly received amount) is what matters for the put-away task.
   // Use execCtx.FieldChanges["quantity_received"] to get before/after,
   // then: delta = after - before
   // If no FieldChanges available, fall back to quantity_received as-is.

4. Resolve location_id:
   if location_strategy == "po_delivery":
     purchase_order_id = execCtx.RawData["purchase_order_id"]
     po = purchaseOrderBus.QueryByID(purchase_order_id)
     location_id = po.DeliveryLocationID
     // If DeliveryLocationID is nil/zero, return "no_location" output port
   else if location_strategy == "static":
     location_id = config.LocationID

5. Resolve reference_number:
   reference_number = resolveTemplateVars(config.ReferenceNumber, execCtx.RawData)
   // If empty, default to "PO-{purchase_order_id}"

6. Create the task:
   putAwayTaskBus.Create(ctx, putawaytaskbus.NewPutAwayTask{
     ProductID:       product_id,
     LocationID:      location_id,
     Quantity:        delta,
     ReferenceNumber: reference_number,
     CreatedBy:       execCtx.UserID,
   })

7. Return result with output port
```

### Bus Dependencies

The handler needs injected:
- `putawaytaskbus.Business` — to create the task
- `supplierproductbus.Business` — to resolve supplier_product_id → product_id
- `purchaseorderbus.Business` — to resolve purchase_order_id → delivery_location_id (when `location_strategy == "po_delivery"`)

### Output Ports

| Port | When |
|------|------|
| `created` | Task created successfully |
| `no_location` | `po_delivery` strategy but PO has no `delivery_location_id` set |
| `product_not_found` | `supplier_product_id` lookup failed |
| `failure` | Unexpected error |

### Validation

- If `source_from_po == false`, `product_id` must be a valid UUID
- If `location_strategy == "static"`, `location_id` must be a valid UUID
- `location_strategy` must be one of: `"po_delivery"`, `"static"`

### Metadata

```go
func (h) GetType() string                 { return "create_put_away_task" }
func (h) SupportsManualExecution() bool   { return true }
func (h) IsAsync() bool                   { return false }
func (h) GetDescription() string          { return "Creates a put-away task for received inventory" }
```

---

## Action Template (seed)

Add to `seed_workflow.go` alongside existing templates:

```go
workflow.NewActionTemplate{
    Name:          "Create Put-Away Task",
    Description:   "Creates a put-away task directing floor workers to shelve received goods",
    ActionType:    "create_put_away_task",
    Icon:          "material-symbols:shelves",
    DefaultConfig: json.RawMessage(`{
        "source_from_po": true,
        "location_strategy": "po_delivery",
        "reference_number": "PO-RCV-{{purchase_order_id}}"
    }`),
    CreatedBy:     adminID,
}
```

This makes the action appear in the workflow editor palette.

---

## Seeded Default Workflow Rule

Create a new rule in `seed_workflow.go`:

### Rule Definition

| Field | Value |
|-------|-------|
| Name | "Auto-Create Put-Away on Receive" |
| Description | "When a PO line item receives quantity, create a put-away task at the PO's delivery location" |
| EntityID | `purchase_order_line_items` entity UUID |
| TriggerTypeID | `on_update` trigger type UUID |
| IsActive | `true` |

### Trigger Conditions

**No conditions** — set `TriggerConditions` to `nil`.

The `TriggerProcessor` (`business/sdk/workflow/trigger.go`) does not support a generic `"changed"` operator. Supported operators are: `equals`, `not_equals`, `changed_from`, `changed_to`, `greater_than`, `less_than`, `contains`, `in`. An unknown operator silently returns `Matched: false`, meaning the rule would never fire.

`greater_than > 0` is also incorrect: when `quantity_received` is not in `FieldChanges` for a given update (e.g., a different field changed), the evaluator falls back to `RawData["quantity_received"]` which could be > 0 from a prior receive — causing spurious firings.

The correct approach: no conditions. The rule fires on any `on_update` of `purchase_order_line_items`, and the handler's own delta logic (`FieldChanges["quantity_received"].NewValue - OldValue`) handles filtering — skipping when delta ≤ 0.

### Action

Single action node using the template above:

```json
{
    "source_from_po": true,
    "location_strategy": "po_delivery",
    "reference_number": "PO-RCV-{{purchase_order_id}}"
}
```

### Edge

```go
EdgeType: "start"
SourceActionID: nil
```

---

## Quantity: Delta vs Total

**Important design decision:** The put-away task quantity should be the **delta** (newly received amount), not the cumulative total. Reason: if a PO line item for 100 units is received in two batches (60 then 40), it should create two put-away tasks (one for 60, one for 40), not one for 60 and then update to 100.

The delta is available via:
```go
fc := execCtx.FieldChanges["quantity_received"]
before := fc.OldValue  // e.g., 0   (workflow.FieldChange.OldValue)
after  := fc.NewValue  // e.g., 60  (workflow.FieldChange.NewValue)
delta  := after - before             // 60
```

> `FieldChange` struct (`business/sdk/workflow/models.go`): fields are `OldValue` and `NewValue` — **not** `Before`/`After`.

If `FieldChanges` is not populated for some reason, log a warning and fall back to `quantity_received` from `RawData`.

---

## Edge Cases

| Case | Behavior |
|------|----------|
| PO has no `delivery_location_id` | Return `"no_location"` output port — downstream nodes can alert a supervisor |
| `supplier_product_id` doesn't resolve | Return `"product_not_found"` output port |
| `quantity_received` decreased (correction) | Delta is negative — skip task creation, return `"created"` with `skipped: true` in result |
| Delta is 0 | Skip task creation (update was to another field, condition passed anyway) |
| Multiple line items on same PO | Each fires independently — creates separate put-away tasks per product, which is correct |
| Same line item received twice | Creates two tasks (one per receive event) — correct for partial receives |

---

## Testing

### Standard Action Handler Test Pattern

All action handler tests in this codebase use a **real database** via `dbtest.NewDatabase()` — no mocks for bus dependencies. Follow `business/sdk/workflow/workflowactions/inventory/receive_test.go` as the direct template.

**File:** `business/sdk/workflow/workflowactions/inventory/createputawaytask_test.go`

**Structure:**
```
Test_CreatePutAwayTask(t)               — dbtest.NewDatabase, insertSeedData, Run(validate), Run(execute)
createPutAwayTaskSeedData struct        — embeds unitest.SeedData + Handler + domain entities
insertCreatePutAwayTaskSeedData()       — full dependency chain (see below)
createPutAwayTaskValidateTests()        — returns []unitest.Table
createPutAwayTaskExecuteTests()         — returns []unitest.Table
```

**Seed chain required** (this handler touches both procurement and inventory):
```
users (admin x1)
  → regions (query existing) → cities → streets → timezones → contactInfos
    → brands → productCategories → products
      → warehouses → zones → inventoryLocations
        → suppliers → supplierProducts        (for source_from_po=true path)
          → purchaseOrders x2:
              one WITH delivery_location_id   (for po_delivery happy path)
              one WITHOUT delivery_location_id (for no_location output port)
            → purchaseOrderLineItems           (linking to purchase orders)
```

### Validate Tests

| Test name | Config | Expected |
|---|---|---|
| `missing_location_strategy` | `{}` | error contains "location_strategy" |
| `invalid_location_strategy` | `{"location_strategy":"bad"}` | error |
| `missing_product_id_when_static` | `{"location_strategy":"static","location_id":"<uuid>"}` | error contains "product_id" |
| `invalid_product_id` | bad UUID in product_id | error |
| `missing_location_id_when_static` | `{"location_strategy":"static","product_id":"<uuid>"}` | error |
| `invalid_location_id` | bad UUID in location_id | error |
| `source_from_po_po_delivery_valid` | `{"source_from_po":true,"location_strategy":"po_delivery"}` | nil |

### Execute Tests

| Test name | Setup | Expected `"output"` |
|---|---|---|
| `happy_path_static` | real productID + locationID in config; FieldChanges delta=10 | `"created"` — verify task exists in DB with correct qty |
| `source_from_po_po_delivery` | RawData has `supplier_product_id` + `purchase_order_id` pointing to PO with delivery_location_id | `"created"` |
| `no_location_on_po` | purchase_order_id pointing to PO with nil delivery_location_id | `"no_location"` |
| `product_not_found` | bad `supplier_product_id` UUID in RawData | `"product_not_found"` |
| `zero_delta` | FieldChanges OldValue == NewValue | `"created"` with `skipped: true` in result map |
| `negative_delta` | FieldChanges NewValue < OldValue | `"created"` with `skipped: true` in result map |
| `template_reference_number` | RawData has `purchase_order_id`; verify created task's ReferenceNumber | `"created"` + assert `ReferenceNumber == "PO-RCV-{id}"` |

### E2E Workflow Integration Test (separate, future)

A full end-to-end test at `api/cmd/services/ichor/tests/` requires Temporal infra (see `apitest.InitWorkflowInfra`). This is **distinct** from the handler unit tests above and is optional for the initial implementation. Scope: seed a PO with a line item → trigger `on_update` with quantity_received delta → verify Temporal dispatches → verify put-away task created in DB.

---

## Frontend Changes (minimal)

The workflow editor already renders action nodes dynamically from templates. Once the action template is seeded, `create_put_away_task` will appear in the palette automatically. The only frontend consideration:

- **Property panel:** The generic action property panel should handle the config fields. If `source_from_po` / `location_strategy` need special UI (dropdowns instead of text inputs), that's a follow-up. For now, the JSON config editor in the workflow editor works.

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `business/sdk/workflow/workflowactions/inventory/createputawaytask.go` | **Create** — new action handler |
| `business/sdk/workflow/workflowactions/inventory/createputawaytask_test.go` | **Create** — unit tests |
| `business/sdk/workflow/workflowactions/register.go` | **Modify** — add `PutAwayTask *putawaytaskbus.Business` to `BusDependencies` struct, register handler in `RegisterGranularInventoryActions` with nil-guard |
| `business/sdk/dbtest/seed_workflow.go` | **Modify** — add action template + default rule + action + edge |
| Dependency wiring (wherever `BusDependencies` is constructed) | **Modify** — pass `putawaytaskbus.Business` instance |
