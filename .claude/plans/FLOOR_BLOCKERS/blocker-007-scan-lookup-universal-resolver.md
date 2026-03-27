# Blocker 007: Scan Lookup — Backend Fan-Out Exists But No Location Barcode Column

> **STATUS: DOWNGRADED / NOT A BLOCKER** (verified 2026-03-16)
> `location_code` column exists via migration at `migrate.sql:2188-2189`.
> `scanapp.go:82` queries locations via `LocationCodeExact` filter — works correctly.
> This is a design choice (barcodes encode the location code string), not a missing feature.
> A separate `barcode` column would be a nice-to-have for warehouses that use different
> barcode values, but current implementation is functional.

**Severity:** ~~MEDIUM~~ → **NICE-TO-HAVE** — current location_code lookup works if barcodes encode the code
**Domain:** `inventory/scan`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

The `scanapp` (`app/domain/inventory/scanapp/scanapp.go`) already implements a parallel
fan-out resolver across 4 domains: product (UPC), lot, serial, and location. This is
architecturally solid.

However, **location lookup uses `QueryByCode`** — it matches the scanned barcode against
`location_code` (e.g., `A-01-02-03`). This works when the barcode IS the code, but real
warehouses print barcodes that encode a separate barcode field — not the human-readable code.

**Exact fan-out (from `docs/arch/inventory-ops.md`):**
```
goroutine 4 → locationBus.QueryByCode(barcode)
```

There is no `barcode` column on `inventory.inventory_locations`. The table has only
`location_code` (human-readable). In a real deployment, the scanner will read a barcode
that doesn't match the code string.

---

## What Needs to Be Added

### 1. Add `barcode` column to `inventory_locations`

Migration:
```sql
-- Version: X.XX
-- Description: Add barcode column to inventory_locations for physical label scanning
ALTER TABLE inventory.inventory_locations
    ADD COLUMN barcode VARCHAR(128) UNIQUE;

CREATE INDEX idx_inventory_locations_barcode
    ON inventory.inventory_locations(barcode)
    WHERE barcode IS NOT NULL;
```

Column is nullable — locations without a physical barcode label fall back to `location_code`.

### 2. Update `inventorylocationbus` storer and model

- `business/domain/inventory/inventorylocationbus/model.go` — add `Barcode *string`
- `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/inventorylocationdb.go` — add `barcode` to SELECT, INSERT, UPDATE queries
- Add `QueryByBarcode(ctx, barcode string) (InventoryLocation, error)` method

### 3. Update `scanapp` to prefer barcode over code

`app/domain/inventory/scanapp/scanapp.go`:

```go
// goroutine 4: try barcode first, fall back to code
go func() {
    defer wg.Done()
    loc, err := s.locationBus.QueryByBarcode(ctx, barcode)
    if err != nil || loc is zero {
        loc, err = s.locationBus.QueryByCode(ctx, barcode)
    }
    if err == nil {
        mu.Lock()
        results = append(results, ...)
        mu.Unlock()
    }
}()
```

### 4. Expose barcode field in admin UI

The `inventorylocationapp` and `inventorylocationapi` should expose the new `barcode` field
in Create and Update so warehouse admins can set barcodes when labeling locations.

---

## Key Files to Read Before Implementing

- `docs/arch/inventory-ops.md` — ScanApp fan-out pattern, `## ⚠ Adding a new scan result type`
- `app/domain/inventory/scanapp/scanapp.go` — current fan-out implementation
- `business/domain/inventory/inventorylocationbus/model.go` — current location model
- `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/inventorylocationdb.go` — SQL queries to update
- `business/sdk/migrate/sql/migrate.sql` — migration format

---

## Acceptance Criteria

- [ ] `inventory_locations` table has nullable `barcode` column with unique index
- [ ] `GET /v1/inventory/inventory-locations/{id}` includes `barcode` field
- [ ] `PATCH /v1/inventory/inventory-locations/{id}` accepts `barcode` field
- [ ] `POST /v1/inventory/scan` with a barcode value matches location by `barcode` column (not code)
- [ ] Locations without a barcode still resolve via `location_code` match
- [ ] `go build ./...` passes
- [ ] Migration applies cleanly
