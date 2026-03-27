# Blocker 010: PO Visibility + Lot Expiry — Client-Side Date Filtering

> **STATUS: FIXED (backend)** (verified 2026-03-16)
> **PO Date Filters:** `purchaseorderbus/filter.go:21-24` has `StartExpectedDelivery` and
> `EndExpectedDelivery`. SQL at `purchaseorderdb/filter.go:58-66`. API params at `purchaseorderapi/filter.go:25-26`.
> **Lot Expiry Filters:** `lottrackingsbus/filter.go:15-16` has `ExpirationDateBefore` and
> `ExpirationDateAfter`. SQL at `lottrackingsdb/filter.go:28-36`.
> **Remaining:** Frontend composables need to pass these query params instead of filtering client-side.

**Severity:** ~~LOW~~ → **Frontend-only fix** — backend filters fully implemented
**Domains:** `procurement/purchaseorder`, `inventory/lottrackings`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem (backend portion resolved)

~~Two operations filter by date entirely in-memory after fetching all records.~~ Backend now supports
server-side date filtering for both domains. Frontend needs to use the query params:

1. **PO Visibility** (`usePOVisibility.ts`): Pass `StartExpectedDelivery` / `EndExpectedDelivery` query params.
2. **Lot Expiry Dashboard** (`useLotLookup.ts`): Pass `ExpirationDateBefore` query param.

---

## What Should Be Added

### Part A: Date window filter on purchase orders

**Backend change:** Add `expected_delivery_before` and `expected_delivery_after` query params
to `GET /v1/procurement/purchase-orders`.

`purchaseordersbus` filter:
```go
type QueryFilter struct {
    // existing fields ...
    ExpectedDeliveryBefore *time.Time
    ExpectedDeliveryAfter  *time.Time
}
```

SQL WHERE clause addition:
```sql
AND (:expected_delivery_before IS NULL OR expected_delivery_date <= :expected_delivery_before)
AND (:expected_delivery_after  IS NULL OR expected_delivery_date >= :expected_delivery_after)
```

**Frontend change:** `usePOVisibility.ts` — instead of filtering client-side, pass
`expected_delivery_after=<today>` and `expected_delivery_before=<today+30days>` as query params.

### Part B: Expiry-within-N-days filter on lot trackings

**Backend change:** Add `expires_before` query param to `GET /v1/inventory/lot-trackings`.

```go
type QueryFilter struct {
    // existing fields ...
    ExpiresBefore *time.Time
}
```

SQL WHERE clause:
```sql
AND (:expires_before IS NULL OR expiration_date <= :expires_before)
```

**Frontend change:** `useLotLookup.ts` expiry dashboard — pass
`expires_before=<today + N days>` as a query param; remove client-side filter.

---

## Key Files to Read Before Implementing

- `business/domain/procurement/` — find the PO business package and its QueryFilter
- `business/domain/inventory/lottrackingsbus/lottrackingsbus.go` — QueryFilter for lots
- `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/lottrackingsdb.go` — SQL query to update
- `app/domain/inventory/lottrackingsapp/` — QueryParams parsing (add `expires_before` param)
- `src/composables/floor/usePOVisibility.ts` — client-side filter to replace
- `src/composables/floor/useLotLookup.ts` — expiry filter to replace

---

## Acceptance Criteria

- [ ] `GET /v1/procurement/purchase-orders?expected_delivery_before=2026-04-01` returns only POs due by that date
- [ ] `GET /v1/inventory/lot-trackings?expires_before=2026-04-01` returns only lots expiring by that date
- [ ] Frontend PO Visibility passes date params to backend — no client-side filtering remains
- [ ] Frontend Lot Expiry Dashboard passes `expires_before` — no client-side filtering remains
- [ ] Pagination + date filter returns correct `.total` count (not full count)
- [ ] `go build ./...` passes
- [ ] `npm run type-check` passes
