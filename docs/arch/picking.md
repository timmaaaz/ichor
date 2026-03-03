# picking

[bus]=business [app]=application [api]=HTTP [db]=store [sdk]=shared
→=depends on ⊕=writes ⊗=reads ⚡=external [tx]=transaction [cache]=cached

---

## StateMachine

PICKING → PACKING        auto: triggered by PickQuantity/ShortPick when all items terminal
PACKING → READY_TO_SHIP  manual: CompletePacking

terminal states: PICKED, PARTIALLY_PICKED, BACKORDERED, CANCELLED

---

## PickingApp [app]

file: app/domain/sales/pickingapp/pickingapp.go
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

  NewApp(log, db, ordersBus, orderLineItemsBus, inventoryItemBus, inventoryTransactionBus, orderFulfillmentStatusBus, lineItemFulfillmentStatusBus) *App
  PickQuantity(ctx, lineItemID uuid.UUID, req PickQuantityRequest) (orderlineitemsapp.OrderLineItem, error)
  ShortPick(ctx, lineItemID uuid.UUID, req ShortPickRequest) (orderlineitemsapp.OrderLineItem, error)
  CompletePacking(ctx, orderID uuid.UUID, req CompletePackingRequest) (ordersapp.Order, error)

key facts:
  - Multi-bus orchestration layer for order fulfillment
  - Three operations: PickQuantity (normal), ShortPick (exception paths), CompletePacking (order advance)
  - PickQuantity and ShortPick run inside ReadCommitted transactions; CompletePacking does not

---

## PickQuantity

TX isolation: sql.LevelReadCommitted

steps:
  1. Lock inventory via QueryAvailableForAllocation() with FOR UPDATE (FEFO ordering)
  2. Decrement Quantity + AllocatedQuantity on inventory_item row
  3. CREATE inventory_transaction record (type="pick", negative quantity)
  4. UPDATE line item → status: * → PICKED
  5. If all line items terminal (PICKED / PARTIALLY_PICKED / BACKORDERED / CANCELLED):
       UPDATE order FulfillmentStatusID: PICKING → PACKING

inventory operation: FEFO allocation with FOR UPDATE pessimistic lock

---

## ShortPick

TX isolation: sql.LevelReadCommitted

four pick types:

| Type       | Inventory touch | Line item result    | New line created |
|------------|----------------|---------------------|------------------|
| partial    | yes (pickedQty) | PARTIALLY_PICKED    | no               |
| backorder  | no             | BACKORDERED         | no               |
| substitute | yes (pickedQty) | original→BACKORDERED | yes (new→PICKED) |
| skip       | no             | PENDING_REVIEW      | no               |

key facts:
  - substitute: creates new line item for substitute product; original line → BACKORDERED
  - order advances PICKING → PACKING only when ALL line items reach terminal state

---

## CompletePacking

TX: none
steps:
  1. Validate order is in PACKING state
  2. UPDATE order FulfillmentStatusID: PACKING → READY_TO_SHIP

---

## ⚠ Adding a new pick type (ShortPick variant)

  app/domain/sales/pickingapp/model.go              (add to ShortPickRequest type enum)
  app/domain/sales/pickingapp/pickingapp.go         (add case in ShortPick switch)
  business/domain/inventory/inventorytransactionbus/ (new transaction type constant if needed)
  api/cmd/services/ichor/tests/sales/pickingapi/    (integration test for new path)

## ⚠ Changing order status transitions

  app/domain/sales/pickingapp/pickingapp.go         (transition logic in PickQuantity + ShortPick)
  business/domain/sales/orderfulfillmentstatusbus/  (status constants — must exist before reference)
  business/sdk/migrate/sql/migrate.sql              (if new status value added to enum/table)
  api/cmd/services/ichor/tests/sales/               (update transition assertions)

## ⚠ Changing the FEFO allocation query

  business/domain/inventory/inventoryitembus/stores/inventoryitemdb/inventoryitemdb.go
    (QueryAvailableForAllocation — FEFO ORDER BY + FOR UPDATE clause)
  app/domain/sales/pickingapp/pickingapp.go         (consumer of query result)
