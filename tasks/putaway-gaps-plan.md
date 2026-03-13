# Put-Away Gaps Plan
## Bridge Missing Test Coverage + UPC Update Fix

**Goal**: Two focused gaps to close. No new features — only missing coverage.

---

## Gap 1: `productapp.Update()` missing `ErrUniqueEntry` → 409

**Scope**: 3 files, ~20 lines total.

### 1a — `app/domain/products/productapp/productapp.go`
Add ErrUniqueEntry check to `Update()` (mirrors the existing `Create()` check):
```go
if errors.Is(err, productbus.ErrUniqueEntry) {
    return Product{}, errs.New(errs.AlreadyExists, productbus.ErrUniqueEntry)
}
```
Place it before the `ErrForeignKeyViolation` check.

### 1b — `api/cmd/services/ichor/tests/products/productapi/update_test.go`
Add `update409_duplicate_upc` table entry:
- Create two products in seed, attempt to set second's UPC to first's UPC
- Expect `http.StatusConflict` + `errs.AlreadyExists`

### 1c — `api/cmd/services/ichor/tests/products/productapi/create_test.go`
Add `create409_duplicate_upc` table entry:
- Attempt to create a product with a UPC that already exists in seed
- Expect `http.StatusConflict` + `errs.AlreadyExists`

**Test command**: `go test ./api/cmd/services/ichor/tests/products/productapi/...`

---

## Gap 2: Integration tests for `putawaytaskapi`

**Scope**: 6 new files + 1 modification to `dbtest`.

### 2a — `business/sdk/dbtest/dbtest.go`
Add `PutAwayTask *putawaytaskbus.Business` field to `BusDomain` struct (inventory section).
Add instantiation in `newBusDomains()`:
```go
PutAwayTask: putawaytaskbus.NewBusiness(log, delegate, putawaytaskdb.NewStore(log, db)),
```

### 2b — `api/cmd/services/ichor/tests/inventory/putawaytaskapi/seed_test.go`
Seed chain:
1. 2 users (admin + floor worker)
2. 1 warehouse → 1 zone → 2 locations
3. 2 products
4. 4 put_away_tasks via `putawaytaskbus.TestSeedPutAwayTasks`:
   - 1 `pending`
   - 1 `in_progress` (assigned to floor worker)
   - 1 `completed` (completed_by admin)
   - 1 `cancelled`

### 2c — `api/cmd/services/ichor/tests/inventory/putawaytaskapi/create_test.go`
| Case | Expect |
|------|--------|
| `create200` | Valid task, all fields present |
| `create400_missing_product` | Missing product_id → 400 |
| `create400_missing_location` | Missing location_id → 400 |
| `create400_bad_quantity` | quantity=0 → 400 |
| `create409_bad_fk` | Non-existent product_id UUID → 409 |
| `create401` | No token, bad sig, bad permission → 401 |

### 2d — `api/cmd/services/ichor/tests/inventory/putawaytaskapi/query_test.go`
| Case | Expect |
|------|--------|
| `query200` | Returns all 4 seeded tasks |
| `queryByID200` | Returns specific task by ID |
| `query200_by_status` | `?status=pending` returns 1 result |
| `query200_by_assigned_to` | `?assigned_to={userID}` returns in_progress task |
| `queryByID404` | Non-existent ID → 404 |
| `query401` | Auth failure |

### 2e — `api/cmd/services/ichor/tests/inventory/putawaytaskapi/update_test.go`
| Case | Expect |
|------|--------|
| `update200_claim` | `pending → in_progress` (sets assigned_to, assigned_at) |
| `update200_abandon` | `in_progress → pending` (clears assigned_to) |
| `update200_complete` | `in_progress → completed` — verify: task has completed_by/at, inventory_transaction created (PUT_AWAY), inventory_item quantity updated |
| `update200_cancel` | `pending → cancelled` |
| `update400_invalid_transition` | `pending → completed` directly → 400 (FailedPrecondition) |
| `update400_from_terminal` | `completed → in_progress` → 400 |
| `update404` | Non-existent ID → 404 |
| `update401` | Auth failure |

> **Note on `update200_complete`**: After completing the task, query `inventorytransactionbus` and `inventoryitembus` via `db.BusDomain` to assert the side-effects without an extra HTTP round-trip.

### 2f — `api/cmd/services/ichor/tests/inventory/putawaytaskapi/delete_test.go`
| Case | Expect |
|------|--------|
| `delete200` | Deletes a cancelled task |
| `delete404` | Non-existent ID → 404 |
| `delete401` | Auth failure |

**Test command**: `go test ./api/cmd/services/ichor/tests/inventory/putawaytaskapi/...`

---

## Reference Implementations

| Need | Look at |
|------|---------|
| Full test suite shape | `api/cmd/services/ichor/tests/inventory/inventorytransactionapi/` |
| Status transition test (400) | `api/cmd/services/ichor/tests/inventory/inventoryadjustmentapi/update_test.go` |
| 3-way tx side-effect verification | `api/cmd/services/ichor/tests/inventory/putawaytaskapi/update_test.go` (new — use `db.BusDomain.InventoryTransaction.QueryByID`) |
| BusDomain addition | `business/sdk/dbtest/dbtest.go` inventory section |
| UPC 409 pattern | `api/cmd/services/ichor/tests/products/productapi/create_test.go` (FK violation pattern) |

---

## Execution Order

1. Gap 1 first (small, isolated, no deps)
2. Gap 2a (dbtest BusDomain addition — required before tests compile)
3. Gap 2b–2f (test files, any order)
4. Run both test commands to verify

---

## Out of Scope

- `putawaycomplete.go` workflow action — **not needed**, inventory side-effects are handled synchronously in `putawaytaskapp.complete()`
- Phase 5 from the original plan — superseded by the app-layer approach
