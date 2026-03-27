# Blocker 008: Inspections — No Atomic Fail + Quarantine Endpoint

> **STATUS: CONFIRMED** (verified 2026-03-16)
> No `Fail()` method in `inspectionbus.go` — only CRUD (Create, Update, Delete, Query, Count, QueryByID).
> `inspectionapi/routes.go:27-47` — only standard CRUD routes, no `/fail` composite endpoint.
> `inspectionapp.go` does not import or reference `lottrackingsbus` — cannot quarantine.
> `migrate.sql:749` — inspection `status VARCHAR(20) NOT NULL` with no CHECK constraint.
> Note: lot `quality_status` at `migrate.sql:659` DOES have a CHECK constraint.

**Severity:** MEDIUM — fail and quarantine are two separate API calls with no transaction guarantee
**Domain:** `inventory/inspection`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

When a floor worker fails an inspection AND the item is lot-tracked, the frontend
(`src/composables/floor/useInspection.ts`) must:
1. `PATCH /v1/inventory/inspections/{id}` — set `status = failed`
2. `PATCH /v1/inventory/lot-trackings/{lotId}` — set `quality_status = quarantine`

These are two separate HTTP calls with no transaction. If the second call fails (network error,
race condition, server crash), inventory goes in a partially-failed state:
- Inspection marked failed ✓
- Lot still shows active ✗ — lot continues to be picked

Additionally, inspection `status` is stored as a free-text `VARCHAR` with no CHECK constraint —
the backend accepts any string. Frontend hardcodes `passed | failed | pending`.

---

## What Should Be Built

### 1. Add a composite `fail` endpoint to inspectionapi

`POST /v1/inventory/inspections/{id}/fail`

Request body:
```json
{
  "notes": "Surface damage on pallets 2 and 3",
  "quarantine_lot": true  // only relevant if lot exists on the inspection
}
```

This endpoint runs an atomic transaction:
1. UPDATE `inventory.inspections` → `status = failed`, `notes = req.Notes`
2. IF `quarantine_lot == true` AND inspection has a `lot_id`:
   UPDATE `inventory.lot_trackings` → `quality_status = quarantine`
3. Optionally: INSERT `inventory.inventory_transactions` type = `inspection_fail` (audit trail)

TX isolation: `sql.LevelReadCommitted`

### 2. Create an orchestrating app for inspections

Current `inventoradjustmentapp` only wraps `inspectionbus`. For the composite operation,
create or update `app/domain/inventory/inspectionapp/inspectionapp.go` to hold both
`inspectionbus` and `lottrackingsbus`, plus `db` for transaction management.

```go
type App struct {
    inspectionbus   *inspectionbus.Business
    lotTrackingsBus *lottrackingsbus.Business
    db              *sqlx.DB
    auth            *auth.Auth
}
```

### 3. Add CHECK constraint on inspection status

Migration:
```sql
ALTER TABLE inventory.inspections
    ADD CONSTRAINT chk_inspection_status CHECK (
        status IN ('pending', 'passed', 'failed', 'in_progress')
    );
```

Verify exact current values in `business/domain/inventory/inspectionbus/inspectionbus.go`
before applying — don't invalidate existing data.

### 4. Schedule next inspection

The frontend also schedules the next inspection (7/30/90 days). Verify whether this is
a separate call or bundled. If separate, it's acceptable — scheduling failure is recoverable
unlike the fail+quarantine split.

---

## Key Files to Read Before Implementing

- `app/domain/inventory/putawaytaskapp/putawaytaskapp.go` — atomic write pattern reference
- `business/domain/inventory/inspectionbus/inspectionbus.go` — current inspection business logic
- `business/domain/inventory/lottrackingsbus/lottrackingsbus.go` — quarantine method
- `api/domain/http/inventory/inspectionapi/` — existing routes to extend
- `business/sdk/migrate/sql/migrate.sql` — migration format

---

## Acceptance Criteria

- [ ] `POST /v1/inventory/inspections/{id}/fail` sets inspection `failed` and (if requested) quarantines the lot atomically
- [ ] If quarantine update fails, the inspection status change also rolls back
- [ ] `POST /v1/inventory/inspections/{id}/fail` with `quarantine_lot: false` only fails the inspection (lot untouched)
- [ ] Inspection with no lot: `quarantine_lot: true` is a no-op for lot (not an error)
- [ ] Status column has CHECK constraint — `status = "oops"` returns 400/500
- [ ] Integration test: fail inspection + quarantine, verify both rows updated
- [ ] `go build ./...` passes
