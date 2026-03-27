# Progress Summary: picking.md

## Overview
Architecture for order fulfillment via picking, packing, and shipping. Implements FEFO (First-Expiry-First-Out) inventory allocation with multi-step order state transitions.

## Picking State Machine

```
PICKING → PACKING        auto: triggered by PickQuantity/ShortPick when all items terminal
PACKING → READY_TO_SHIP  manual: CompletePacking

terminal states: PICKED, PARTIALLY_PICKED, BACKORDERED, CANCELLED
```

**Auto-transition:** Order advances from PICKING → PACKING only when ALL line items reach terminal state.

## PickingApp [app] — `app/domain/sales/pickingapp/pickingapp.go`

**Responsibility:** Multi-bus orchestration for order fulfillment operations.

### Struct
```go
type App struct {
    log                        *logger.Logger
    db                         *sqlx.DB
    ordersBus                  *ordersbus.Business
    orderLineItemsBus          *orderlineitemsbus.Business
    inventoryItemBus           *inventoryitembus.Business
    inventoryTransactionBus    *inventorytransactionbus.Business
    orderFulfillmentStatusBus  *orderfulfillmentstatusbus.Business
    lineItemFulfillmentStatusBus *lineitemfulfillmentstatusbus.Business
}
```

### Methods
- `NewApp(log, db, ordersBus, orderLineItemsBus, inventoryItemBus, inventoryTransactionBus, orderFulfillmentStatusBus, lineItemFulfillmentStatusBus) *App`
- `PickQuantity(ctx, lineItemID uuid.UUID, req PickQuantityRequest) (orderlineitemsapp.OrderLineItem, error)`
- `ShortPick(ctx, lineItemID uuid.UUID, req ShortPickRequest) (orderlineitemsapp.OrderLineItem, error)`
- `CompletePacking(ctx, orderID uuid.UUID, req CompletePackingRequest) (ordersapp.Order, error)`

### Key Facts
- **Multi-bus orchestration** — coordinates 7 different business domains
- **PickQuantity and ShortPick** run inside ReadCommitted transactions
- **CompletePacking** runs without transaction (read-only order state check)

## PickQuantity Operation

**TX isolation: sql.LevelReadCommitted**

Atomic steps:
1. **Lock inventory via QueryAvailableForAllocation()** with FOR UPDATE (FEFO ordering)
2. **Decrement Quantity + AllocatedQuantity** on inventory_item row
3. **CREATE inventory_transaction record** (type="pick", negative quantity)
4. **UPDATE line item** → status: * → PICKED
5. **Check if all line items terminal** (PICKED / PARTIALLY_PICKED / BACKORDERED / CANCELLED)
   - If yes: **UPDATE order FulfillmentStatusID** → PICKING → PACKING

### Inventory Operation
- **FEFO allocation** with FOR UPDATE pessimistic lock
- Ensures earliest-expiry stock picked first (compliance for perishables, etc.)

## ShortPick Operation

**TX isolation: sql.LevelReadCommitted**

Four pick types with different outcomes:

| Type       | Inventory Touch | Line Item Result    | New Line Created |
|------------|-----------------|---------------------|------------------|
| partial    | yes (pickedQty) | PARTIALLY_PICKED    | no               |
| backorder  | no              | BACKORDERED         | no               |
| substitute | yes (pickedQty) | original→BACKORDERED | yes (new→PICKED) |
| skip       | no              | PENDING_REVIEW      | no               |

### Key Facts
- **Substitute type:** Creates new line item for substitute product; original line → BACKORDERED
- **Order advancement:** Order advances PICKING → PACKING only when ALL line items reach terminal state
- **Fail-safe:** Each type handles inventory touch independently

## CompletePacking Operation

**TX: none** (read-only state check)

Steps:
1. Validate order is in PACKING state
2. **UPDATE order FulfillmentStatusID** → PACKING → READY_TO_SHIP

## Change Patterns

### ⚠ Adding a New Pick Type (ShortPick Variant)
Affects 4 areas:
1. `app/domain/sales/pickingapp/model.go` — add to ShortPickRequest type enum
2. `app/domain/sales/pickingapp/pickingapp.go` — add case in ShortPick switch
3. `business/domain/inventory/inventorytransactionbus/` — new transaction type constant if needed
4. `api/cmd/services/ichor/tests/sales/pickingapi/` — integration test for new path

### ⚠ Changing Order Status Transitions
Affects 4 areas:
1. `app/domain/sales/pickingapp/pickingapp.go` — transition logic in PickQuantity + ShortPick
2. `business/domain/sales/orderfulfillmentstatusbus/` — status constants must exist before reference
3. `business/sdk/migrate/sql/migrate.sql` — if new status value added to enum/table
4. `api/cmd/services/ichor/tests/sales/` — update transition assertions

### ⚠ Changing the FEFO Allocation Query
Affects 2 areas:
1. `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go` — QueryAvailableForAllocation (FEFO ORDER BY + FOR UPDATE clause)
2. `app/domain/sales/pickingapp/pickingapp.go` — consumer of query result

## Critical Points
- **FEFO ordering** is mandatory for compliance (earliest expiry picked first)
- **FOR UPDATE pessimistic lock** prevents concurrent allocation races
- **Order advancement is deterministic** — ALL line items must be terminal before state change
- **Substitute creates new line item** — increases order complexity (must be tracked separately)
- **Atomicity per operation** — PickQuantity and ShortPick are atomic; CompletePacking is not

## Notes for Future Development
Picking implements sophisticated inventory allocation with FEFO compliance and multi-type short-pick handling. Most changes will be:
- Adding new pick types (moderate complexity)
- Changing order status transitions (requires careful coordination)
- Tuning FEFO algorithm (risky, requires testing with inventory data)

The FOR UPDATE pessimistic lock is critical — removing it would allow race conditions in concurrent picking.
