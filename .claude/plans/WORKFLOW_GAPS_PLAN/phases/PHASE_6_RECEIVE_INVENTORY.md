# Phase 7: Add receive_inventory Action

**Category**: Backend
**Status**: Pending
**Dependencies**: None (but complements Phase 6)
**Effort**: Medium

---

## Overview

When a purchase order is marked as received (status transitions to "received"), there is no automated way to:
1. Increase the `inventory.inventory_items.quantity` for the received items
2. Create an `inventory.inventory_transactions` record (inbound transaction)

Currently you could use `update_field` (raw SQL `quantity = quantity + N`) but this:
- Bypasses business layer validation
- Does not create a transaction record
- Does not trigger the `inventory_item updated` event correctly

This phase adds a `receive_inventory` action handler that processes PO receipt properly.

---

## Goals

1. New `receive_inventory` handler that increases available inventory
2. Creates proper inbound `inventory_transaction` record
3. Supports `source_from_po: true` mode to extract fields from a PO line item update event

---

## Task Breakdown

### Task 1: Create ReceiveInventoryHandler

**New file**: `business/sdk/workflow/workflowactions/inventory/receive.go`

```go
package inventory

type ReceiveInventoryHandler struct {
    log                 *logger.Logger
    db                  *sqlx.DB
    inventoryItemBus    *inventoryitembus.Business
    inventoryTxBus      *inventorytransactionbus.Business
}
```

**Config struct**:
```go
type ReceiveInventoryConfig struct {
    ProductID          string  `json:"product_id"`
    Quantity           float64 `json:"quantity"`
    WarehouseID        string  `json:"warehouse_id"`
    LocationID         string  `json:"location_id"`
    SourceFromPO       bool    `json:"source_from_po,omitempty"` // Extract from PO line item event
    POLineItemID       string  `json:"po_line_item_id,omitempty"`
    LotNumber          string  `json:"lot_number,omitempty"`
    Notes              string  `json:"notes,omitempty"`
}
```

**Execute logic**:

1. Parse config. If `source_from_po: true`, extract `product_id`, `quantity`, `warehouse_id` from `execCtx.RawData` (fields available on a `procurement.purchase_order_line_items` event).

2. Query `inventoryitembus` to find the correct inventory item for this `(product_id, warehouse_id, location_id)` combination:
   ```go
   filter := inventoryitembus.QueryFilter{
       ProductID:   &productID,
       WarehouseID: &warehouseID,
       LocationID:  &locationID,
   }
   items, err := h.inventoryItemBus.Query(ctx, filter, ...)
   ```

3. If no inventory item found: return `failure` port with error message (can't receive without knowing where to put it).

4. Update inventory item — increase `quantity` by received amount:
   ```go
   update := inventoryitembus.UpdateInventoryItem{
       Quantity: ptrFloat64(item.Quantity + cfg.Quantity),
   }
   _, err = h.inventoryItemBus.Update(ctx, item, update)
   ```

5. Create inbound transaction record:
   ```go
   tx := inventorytransactionbus.NewInventoryTransaction{
       InventoryItemID:    item.InventoryItemID,
       TransactionType:    "inbound",
       QuantityChange:     cfg.Quantity,
       ReferenceID:        cfg.POLineItemID, // optional, link back to PO line item
       ReferenceType:      "purchase_order_line_item",
       Notes:              cfg.Notes,
       CreatedDate:        time.Now(),
   }
   _, err = h.inventoryTxBus.Create(ctx, tx)
   ```

6. Return result with updated quantities.

**Output ports**:
- `received` (default) — inventory updated and transaction created
- `item_not_found` — no inventory item found for the product/location
- `failure` — unexpected error

**EntityModifier**:
```go
func (h *ReceiveInventoryHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
    return []workflow.EntityModification{
        {EntityName: "inventory.inventory_items", EventType: "on_update", Fields: []string{"quantity"}},
        {EntityName: "inventory.inventory_transactions", EventType: "on_create"},
    }
}
```

### Task 2: Register in register.go

**File**: `business/sdk/workflow/workflowactions/register.go`

Add to `RegisterGranularInventoryActions`:
```go
registry.Register(inventory.NewReceiveInventoryHandler(
    config.Log,
    config.DB,
    config.Buses.InventoryItem,
    config.Buses.InventoryTransaction,
))
```

This handler uses existing `BusDependencies.InventoryItem` and `BusDependencies.InventoryTransaction` — no new dependencies needed.

---

## Example Workflow

**Trigger**: `procurement.purchase_order_statuses on_update` where status name changes to "received"

**Actions**:
1. `lookup_entity` — find all line items for the PO (filter by `purchase_order_id`)
2. For each line item: `receive_inventory` (source_from_po: false, explicit product_id/quantity)
3. `transition_status` — update PO status to "completed"
4. `create_alert` — notify warehouse team

Or more simply:

**Trigger**: `procurement.purchase_order_line_items on_update` (when a line item is marked received)

**Actions**:
1. `receive_inventory` with `source_from_po: true` — reads product_id/quantity from the line item event
2. `check_reorder_point` — verify stock is now sufficient
3. (conditional) `create_alert` if still low

---

## Validation

```bash
go build ./...

# Verify inventorytransactionbus has the needed fields
grep -A 20 "type NewInventoryTransaction" business/domain/inventory/inventorytransactionbus/model.go

# Verify inventoryitembus has UpdateInventoryItem with Quantity field
grep -A 15 "type UpdateInventoryItem" business/domain/inventory/inventoryitembus/model.go

# Integration test
go test ./api/cmd/services/ichor/tests/...
```

---

## Gotchas

- **`inventoryitembus` Update takes the current entity** — like other Ardan Labs bus Update methods, `inventoryitembus.Update(ctx, currentItem, update)` requires the current state. Use the result of the Query as the first arg.
- **Inventory item uniqueness**: A product may have multiple inventory items across different locations. The `receive_inventory` handler must either target a specific location (`location_id` required) or use a strategy similar to `allocate_inventory` (FIFO, nearest, etc.). For initial implementation, require `location_id` in config (or infer from PO delivery location).
- **Transaction record fields**: Check `inventorytransactionbus/model.go` for the exact `NewInventoryTransaction` fields. The field names may differ from what's shown here.
- **`source_from_po: true` field extraction**: The PO line item event's `RawData` has snake_case JSON keys. Check the `purchaseorderlineitembus/model.go` JSON tags to know the exact field names.
- **Decimal vs float64**: Inventory quantities may use decimal types. Check `inventoryitembus.InventoryItem.Quantity` type. If it's `decimal.Decimal`, adjust accordingly.
