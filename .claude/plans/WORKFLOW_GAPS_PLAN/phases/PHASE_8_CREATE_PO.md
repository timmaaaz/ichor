# Phase 6: Add create_purchase_order Action

**Category**: Backend
**Status**: Pending
**Dependencies**: Phase 1 (procurement tables in whitelist) recommended first
**Effort**: High

---

## Overview

The reorder automation chain has a broken link:

```
inventory_item updated
    → check_reorder_point
        → needs_reorder: ??? (dead end — only create_alert or send_email possible)
        → ok: done
```

This phase adds a `create_purchase_order` action handler that creates a PO through the proper business layer (not raw SQL), making the full reorder chain possible:

```
inventory_item.quantity updated
    → check_reorder_point
        → needs_reorder:
            → lookup_entity (find product's preferred supplier)
            → create_purchase_order
                → created: create_alert (notify procurement team)
                → no_supplier_found: create_alert (urgent — manual intervention needed)
        → ok: done
```

---

## Goals

1. New `create_purchase_order` action handler in a new `procurement/` subpackage
2. Auto-supplier lookup when `supplier_id` not specified
3. Create PO + line item via business layer (not raw SQL)
4. Wire procurement bus dependencies into ActionConfig and all.go

---

## Task Breakdown

### Task 1: Create procurement/ subpackage and Handler

**New file**: `business/sdk/workflow/workflowactions/procurement/createpo.go`

```go
package procurement

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
    "github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
    "github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/foundation/logger"
)

type CreatePurchaseOrderHandler struct {
    log               *logger.Logger
    purchaseOrderBus  *purchaseorderbus.Business
    lineItemBus       *purchaseorderlineitembus.Business
    supplierProductBus *supplierproductbus.Business
}
```

**Config struct**:
```go
type CreatePOConfig struct {
    ProductID           string  `json:"product_id"`             // UUID string
    Quantity            float64 `json:"quantity"`
    SupplierID          string  `json:"supplier_id,omitempty"`  // Optional: auto-lookup if absent
    DeliveryWarehouseID string  `json:"delivery_warehouse_id"`
    DeliveryLocationID  string  `json:"delivery_location_id,omitempty"`
    Notes               string  `json:"notes,omitempty"`
    SourceFromLineItem  bool    `json:"source_from_line_item,omitempty"` // Extract from trigger event
}
```

**Output ports**:
- `created` (default) — PO successfully created
- `no_supplier_found` — no active supplier found for the product
- `failure` — unexpected error

**Execute logic**:
1. Parse config; if `source_from_line_item: true`, extract `product_id` and `quantity` from `execCtx.RawData`
2. If `supplier_id` absent: call `supplierProductBus.Query()` filtered by `product_id`, pick the one with the lowest unit cost (or first active)
3. Look up the "draft" PO status ID from DB (or require `purchase_order_status_id` in config)
4. Call `purchaseOrderBus.Create()` with proper fields
5. Call `lineItemBus.Create()` with the product line item
6. Return created PO ID and line item ID

**Important**: Use the business layer, not raw SQL. This ensures:
- Proper UUID generation via delegate
- All business validation runs
- Downstream workflow events fire (PO created event will trigger any rules watching procurement.purchase_orders)

### Task 2: Add Procurement Dependencies to ActionConfig

**File**: `business/sdk/workflow/workflowactions/register.go`

```go
type BusDependencies struct {
    // ... existing fields ...

    // Procurement domain (for create_purchase_order action)
    PurchaseOrder         *purchaseorderbus.Business
    PurchaseOrderLineItem *purchaseorderlineitembus.Business
    SupplierProduct       *supplierproductbus.Business
}
```

Register the new handler in `RegisterAll()`:
```go
// Procurement actions
if config.Buses.PurchaseOrder != nil {
    registry.Register(procurement.NewCreatePurchaseOrderHandler(
        config.Log,
        config.Buses.PurchaseOrder,
        config.Buses.PurchaseOrderLineItem,
        config.Buses.SupplierProduct,
    ))
}
```

Note the nil guard — this handler requires real procurement buses. Don't add to `RegisterCoreActions`.

### Task 3: Wire Procurement Buses in all.go

**File**: `api/cmd/services/ichor/build/all/all.go`

The procurement buses are likely already instantiated (they're registered for workflow events). Find their instantiation and add them to `ActionConfig.Buses`:

```go
// Procurement buses (already exist in all.go — just add to ActionConfig)
actionConfig.Buses.PurchaseOrder = purchaseOrderBus
actionConfig.Buses.PurchaseOrderLineItem = purchaseOrderLineItemBus
actionConfig.Buses.SupplierProduct = supplierProductBus
```

### Task 4: Handle PO Status Lookup

POs require a `purchase_order_status_id` FK. The status table has pre-seeded rows (e.g., "draft", "submitted", "approved").

Options:
- **Option A**: Accept `initial_status_name` in config (e.g., `"draft"`) and look it up by name
- **Option B**: Look up the status with the lowest sort order / first status as default

Use Option A for clarity. Add to config:
```go
InitialStatusName string `json:"initial_status_name,omitempty"` // default: "draft"
```

Query `procurement.purchase_order_statuses WHERE name = $1` to get the UUID.

---

## Validation

```bash
go build ./...

# Test: trigger a reorder chain
# 1. Reduce inventory_item quantity below reorder_point
# 2. Verify check_reorder_point → needs_reorder
# 3. Verify create_purchase_order creates PO in DB
# 4. Verify PO has correct supplier and line item
go test ./api/cmd/services/ichor/tests/...
```

---

## Gotchas

- **Circular event fire**: Creating a PO fires `procurement.purchase_orders on_create` which could trigger other rules. This is intentional (the cascade visualization feature exists for this reason) but be aware.
- **`supplierProductBus.Query()` filter**: Check the filter struct for `supplierproductbus` — you need to filter by `product_id`. Look at `supplierproductbus/filter.go` for available filter fields.
- **Currency ID**: POs require a `currency_id`. Fetch the default currency from `core.currencies` or require it in the config. The simplest approach: accept `currency_id` in config with no default (validation error if absent).
- **`DeliveryLocationID` and `DeliveryStreetID`**: PO model requires both warehouse and location. `DeliveryStreetID` is for the supplier's delivery address. Consider making these optional with sane defaults for the initial implementation.
- **Business layer vs raw SQL tradeoff**: The business layer may enforce additional validations (e.g., the PO `order_number` must be unique). The `create_entity` generic action could INSERT directly into `procurement.purchase_orders` (once Phase 1 adds it to the whitelist) but would bypass these validations. This action handler intentionally uses the bus layer to keep validation consistent.
