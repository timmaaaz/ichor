# Blocker 001: Picking — No HTTP Routes for PickQuantity / ShortPick

> **STATUS: FIXED** (verified 2026-03-16)
> Routes exist — they live in `orderlineitemsapi`, not a separate `pickingapi`.
> The frontend is calling the wrong URL path. See "Resolution" below.

**Severity:** ~~CRITICAL~~ → **Frontend URL fix only**
**Domain:** `sales/picking`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Resolution (verified 2026-03-16)

The picking routes are registered in `orderlineitemsapi`, not a separate `pickingapi` package:

- `api/domain/http/sales/orderlineitemsapi/routes.go:45-48`:
  - `POST /v1/sales/order-line-items/{order_line_items_id}/pick-quantity`
  - `POST /v1/sales/order-line-items/{order_line_items_id}/short-pick`
- `all.go:447` constructs `pickingApp`, wired into `orderlineitemsapi.Config` at line 1096

**Remaining fix:** Frontend `usePicking` calls `/v1/sales/picking/{lineItemId}/pick-quantity` but
the actual route is `/v1/sales/order-line-items/{id}/pick-quantity`. Update the frontend URL.

---

## Original Problem (outdated)

`pickingapp` at `app/domain/sales/pickingapp/pickingapp.go` is fully implemented with
`PickQuantity`, `ShortPick`, and `CompletePacking`. However, **no HTTP layer exists** for
these operations. There is no `pickingapi` package under `api/domain/http/sales/`.

The frontend composable `usePicking` (`src/composables/floor/usePicking.ts`) calls endpoints
like `POST /v1/sales/picking/{lineItemId}/pick-quantity` and
`POST /v1/sales/picking/{lineItemId}/short-pick` which return 404.

---

## What Exists

```
app/domain/sales/pickingapp/pickingapp.go     ← COMPLETE (PickQuantity, ShortPick, CompletePacking)
app/domain/sales/pickingapp/model.go          ← PickQuantityRequest, ShortPickRequest, CompletePackingRequest
```

`pickingapp.PickQuantity` already:
1. Locks inventory via `QueryAvailableForAllocation()` with `FOR UPDATE`
2. Decrements `Quantity` + `AllocatedQuantity` on `inventory_items`
3. Creates an `inventory_transactions` record (type=`pick`, negative qty)
4. Updates line item status → `PICKED`
5. Advances order `PICKING → PACKING` when all items reach terminal state

`pickingapp.ShortPick` already handles all four resolution types:
`partial | backorder | substitute | skip`

---

## What Is Missing

### 1. HTTP API package (new files)
Create `api/domain/http/sales/pickingapi/` with:

- `pickingapi.go` — handlers:
  - `POST /v1/sales/picking/{lineItemId}/pick-quantity` → calls `app.PickQuantity()`
  - `POST /v1/sales/picking/{lineItemId}/short-pick` → calls `app.ShortPick()`
  - `POST /v1/sales/orders/{orderId}/complete-packing` → calls `app.CompletePacking()`
- `route.go` — route registration, middleware chain (`mid.Authenticate`, `mid.Authorize`)
- `model.go` — request/response HTTP models if they differ from app models

### 2. Wire routes in the service
File: `api/cmd/services/ichor/build/all/all.go`

Add:
```go
pickingApp := pickingapp.NewApp(log, db, ordersBus, orderLineItemsBus, inventoryItemBus,
    inventoryTransactionBus, orderFulfillmentStatusBus, lineItemFulfillmentStatusBus)
pickingapi.Routes(app, pickingapi.Config{
    Log:     log,
    App:     pickingApp,
    AuthClient: authClient,
})
```

### 3. Integration tests
Create `api/cmd/services/ichor/tests/sales/pickingapi/`:
- `picking_test.go` — table-driven test covering:
  - Normal pick (full quantity)
  - Short pick: partial
  - Short pick: backorder
  - Short pick: substitute (verify new line item created)
  - Short pick: skip
  - Order advances to PACKING after all items terminal

---

## Key Files to Read Before Implementing

- `docs/arch/picking.md` — state machine, TX isolation, all 4 short-pick types
- `docs/arch/domain-template.md` — 7-layer checklist (all layers required)
- `app/domain/sales/pickingapp/pickingapp.go` — the app layer that already works
- `api/domain/http/sales/ordersapi/` — reference pattern for sales HTTP handlers
- `api/cmd/services/ichor/build/all/all.go` — where to wire new routes

---

## Acceptance Criteria

- [ ] `POST /v1/sales/picking/{lineItemId}/pick-quantity` returns updated line item, decrements inventory
- [ ] `POST /v1/sales/picking/{lineItemId}/short-pick` handles all 4 types correctly
- [ ] `POST /v1/sales/orders/{orderId}/complete-packing` advances order to READY_TO_SHIP
- [ ] 401 returned for unauthenticated requests
- [ ] Integration tests pass: `go test ./api/cmd/services/ichor/tests/sales/pickingapi/...`
- [ ] `go build ./...` passes
