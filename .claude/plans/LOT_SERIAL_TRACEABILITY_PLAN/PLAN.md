# Phase 7: Lot & Serial Traceability — Backend Plan

## Overview

Six phased changes addressing Gaps 1–6 from the Phase 7 requirements doc.
Each phase is sized for a single `/feature-dev:feature-dev` invocation.

Priority order follows the ticket table: P0 filters first, P1 quality/settings, P2-P3 location mapping.

---

## Context

- **Migration current version**: 2.01
- **Lot trackings domain**: `business/domain/inventory/lottrackingsbus/`
- **Serial numbers domain**: `business/domain/inventory/serialnumberbus/`
- **Settings domain (fully implemented)**: `business/domain/config/settingsbus/` — supports any JSON value, key-based routing via `GET /v1/config/settings/{key}`
- **Filter pattern**: Business layer `QueryFilter` → App layer `parseFilter(QueryParams)` → DB layer `applyFilter()`
- **SQL rule**: Modify original CREATE TABLE statements directly — no ALTER TABLE

---

## Phase 1 — Verify & Harden Lot Number + Serial Number Filters (Gap 5, P0)

**Goal**: Confirm that `GET /v1/inventory/lot-trackings?lot_number=X` and `GET /v1/inventory/serial-numbers?serial_number=X` work end-to-end. Fix any missing wiring.

**Status at exploration time**:
- `lottrackingsbus.QueryFilter.LotNumber *string` — exists ✓
- `lottrackingsapp.parseFilter` maps `qp.LotNumber` → `filter.LotNumber` — exists ✓
- `serialnumberbus.QueryFilter.SerialNumber *string` — exists ✓
- `serialnumberapp.parseFilter` — **needs verification** (check `app/domain/inventory/serialnumberapp/filter.go`)
- DB `applyFilter` in both `lottrackingsdb` and `serialnumberdb` — **needs verification**

**Files to read and verify**:
- `app/domain/inventory/serialnumberapp/filter.go`
- `business/domain/inventory/serialnumberbus/stores/serialnumberdb/filter.go`
- `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/filter.go`

**Work to do** (if any gap is found):
1. Add missing query params to the relevant `QueryParams` struct in the api filter file
2. Add missing parseFilter mapping in app layer
3. Add missing SQL condition in db layer `applyFilter`

**Files to potentially modify**:
- `app/domain/inventory/serialnumberapp/filter.go`
- `api/domain/http/inventory/serialnumberapi/filter.go`
- `business/domain/inventory/serialnumberbus/stores/serialnumberdb/filter.go`
- `api/domain/http/inventory/lottrackingsapi/filter.go` (verify QueryParams has lot_number)
- `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/filter.go`

**Verification**: `go build ./business/domain/inventory/lottrackingsbus/... ./business/domain/inventory/serialnumberbus/... ./app/domain/inventory/lottrackingsapp/... ./app/domain/inventory/serialnumberapp/... ./api/domain/http/inventory/lottrackingsapi/... ./api/domain/http/inventory/serialnumberapi/...`

---

## Phase 2 — Expiry Range Filters on Lot Trackings (Gap 6, P0)

**Goal**: Add `expiry_before` and `expiry_after` query params to `GET /v1/inventory/lot-trackings` for bucketing lots into expiry windows.

**Current state**: Only a single `ExpirationDate` exact-match filter exists. No range support.

**Work to do**:

1. **`business/domain/inventory/lottrackingsbus/filter.go`** — Add two new fields:
   ```go
   ExpirationDateBefore *time.Time
   ExpirationDateAfter  *time.Time
   ```

2. **`business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/filter.go`** — Add two SQL conditions in `applyFilter`:
   ```go
   // ExpirationDateBefore → expiration_date < :expiration_date_before
   // ExpirationDateAfter  → expiration_date > :expiration_date_after
   ```
   Use named parameters consistent with the existing pattern in that file.

3. **`app/domain/inventory/lottrackingsapp/filter.go`** — Add `ExpiryBefore` and `ExpiryAfter` to `QueryParams` struct and `parseFilter` function, parsing as `timeutil.FORMAT`.

4. **`api/domain/http/inventory/lottrackingsapi/filter.go`** — Add `ExpiryBefore` and `ExpiryAfter` to the `QueryParams` struct (if separate from app layer — check the pattern used by this domain).

**Naming convention**: Follow the existing pattern in lottrackingsdb filter.go for date field naming (check whether it's `expiration_date` or `expiration_date_before`/`expiration_date_after` for the SQL named params).

**Verification**: `go build ./business/domain/inventory/lottrackingsbus/... ./app/domain/inventory/lottrackingsapp/... ./api/domain/http/inventory/lottrackingsapi/...`

---

## Phase 3 — Quality Status Enum Enforcement + Status List Endpoint (Gap 2, P1)

**Goal**: Lock down `lot_trackings.quality_status` to known values and expose them via API.

**Canonical enum**: `good | on_hold | quarantined | released | expired`

**Work to do**:

1. **`business/sdk/migrate/sql/migrate.sql`** — Directly modify the `inventory.lot_trackings` CREATE TABLE (version 1.50) to add a CHECK constraint:
   ```sql
   quality_status varchar(20) NOT NULL CHECK (quality_status IN ('good', 'on_hold', 'quarantined', 'released', 'expired')),
   ```

2. **`api/domain/http/inventory/lottrackingsapi/lottrackingsapi.go`** — Add a new handler `queryQualityStatuses` that returns the hardcoded list:
   ```go
   func (api *api) queryQualityStatuses(ctx context.Context, r *http.Request) web.Encoder {
       return web.NewResponse([]string{"good", "on_hold", "quarantined", "released", "expired"})
   }
   ```

3. **`api/domain/http/inventory/lottrackingsapi/routes.go`** — Register the new route **before** the `/{lot_id}` route to avoid path conflicts:
   ```go
   app.HandlerFunc(http.MethodGet, version, "/inventory/lot-trackings/quality-statuses", api.queryQualityStatuses, authen,
       mid.Authorize(..., permissionsbus.Actions.Read, auth.RuleAny))
   ```

**Important**: The `/quality-statuses` static segment route must be registered before `/{lot_id}` in routes.go or the router will match the static path as a lot_id parameter.

**Verification**: `go build ./api/domain/http/inventory/lottrackingsapi/...`

---

## Phase 4 — Settings Seeds for FEFO/FIFO & Quarantine Access + RBAC Enforcement (Gaps 3 & 4, P1)

**Goal**: Add two new company-level settings, and enforce the quarantine access setting on the lot tracking update endpoint.

### Part A — Seed new settings in migration

**`business/sdk/migrate/sql/migrate.sql`** — Add a new migration version (2.02):
```sql
-- Version: 2.02
-- Description: Add inventory lot rotation method and quarantine access control settings.
INSERT INTO config.settings (key, value, description, created_date, updated_date) VALUES
    ('inventory.lot_rotation_method', '"fefo"',          'Lot rotation method for picking: fefo | fifo', NOW(), NOW()),
    ('inventory.quarantine_access',   '"supervisor_only"', 'Who can quarantine lots: floor_worker | supervisor_only', NOW(), NOW());
```

The settings domain endpoints are already fully implemented. These values are read/updated via:
- `GET  /v1/config/settings/inventory.lot_rotation_method`
- `PUT  /v1/config/settings/inventory.lot_rotation_method`
- `GET  /v1/config/settings/inventory.quarantine_access`
- `PUT  /v1/config/settings/inventory.quarantine_access`

### Part B — RBAC enforcement on PUT lot-trackings

Add a quarantine access check to the lot tracking update handler.

**`api/domain/http/inventory/lottrackingsapi/routes.go`** — The config in `Config` struct needs access to `settingsbus.Business` (or the app can be passed `settingsBus`).

**`api/domain/http/inventory/lottrackingsapi/lottrackingsapi.go`** — In the `update` handler:
1. Parse the incoming request body
2. If `quality_status` is being changed to `"quarantined"` or `"on_hold"`:
   a. Read `inventory.quarantine_access` from settingsBus
   b. If value is `"supervisor_only"`: check that the authenticated user has a supervisor-level role (use the permissionsBus or authclient to check)
   c. Return HTTP 403 if the user lacks the required role

**`api/domain/http/inventory/lottrackingsapi/routes.go`** — Add `SettingsBus *settingsbus.Business` to `Config` struct and pass it when creating the API handler.

**`api/cmd/services/ichor/build/all/all.go`** — Wire `settingsBus` into the lottrackingsapi Config.

**Note on role check**: Look at how other handlers check roles (e.g., permissionsbus or authclient patterns used in approval-related handlers). Follow the same approach — likely `mid.Authorize` with a supervisor rule or an in-handler check via `auth.GetClaims(ctx)` and role inspection.

**Verification**: `go build ./business/domain/config/settingsbus/... ./api/domain/http/inventory/lottrackingsapi/... ./api/cmd/services/ichor/build/all/...`

---

## Phase 5 — Lot-to-Location via Serial Aggregation (Gap 1, Option C, P2)

**Goal**: Expose `GET /v1/inventory/lot-trackings/{lot_id}/locations` that aggregates location data from `inventory.serial_numbers` for serialized items.

**Response model** (new `LotLocation` type):
```go
type LotLocation struct {
    LocationID   uuid.UUID `json:"location_id"`
    LocationCode string    `json:"location_code"`
    Aisle        string    `json:"aisle"`
    Rack         string    `json:"rack"`
    Shelf        string    `json:"shelf"`
    Bin          string    `json:"bin"`
    Quantity     int       `json:"quantity"`
}
```

**Work to do**:

1. **`business/domain/inventory/lottrackingsbus/model.go`** — Add `LotLocation` struct.

2. **`business/domain/inventory/lottrackingsbus/lottrackingsbus.go`** — Add method to Business:
   ```go
   func (b *Business) QueryLocationsByLotID(ctx context.Context, lotID uuid.UUID) ([]LotLocation, error)
   ```

3. **`business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/lottrackingsdb.go`** — Add `QueryLocationsByLotID` to the Storer interface and the DB implementation. SQL:
   ```sql
   SELECT
       sn.location_id,
       il.location_code,
       il.aisle,
       il.rack,
       il.shelf,
       il.bin,
       COUNT(*)::int AS quantity
   FROM inventory.serial_numbers sn
   JOIN inventory.inventory_locations il ON il.id = sn.location_id
   WHERE sn.lot_id = :lot_id
   GROUP BY sn.location_id, il.location_code, il.aisle, il.rack, il.shelf, il.bin
   ```
   Use `sqldb.NamedQuerySlice` consistent with other methods in the store. Define a DB-layer struct for scanning.

4. **`app/domain/inventory/lottrackingsapp/lottrackingsapp.go`** — Add `QueryLocationsByLotID(ctx, id) ([]LotLocation, error)` passthrough.

5. **`app/domain/inventory/lottrackingsapp/model.go`** — Add `LotLocation` app model and `toAppLotLocation` conversion.

6. **`api/domain/http/inventory/lottrackingsapi/lottrackingsapi.go`** — Add `queryLocationsByLotID` handler that reads `{lot_id}` from path, calls app method, returns array.

7. **`api/domain/http/inventory/lottrackingsapi/routes.go`** — Register new route:
   ```go
   app.HandlerFunc(http.MethodGet, version, "/inventory/lot-trackings/{lot_id}/locations", api.queryLocationsByLotID, authen,
       mid.Authorize(..., permissionsbus.Actions.Read, auth.RuleAny))
   ```

**Verification**: `go build ./business/domain/inventory/lottrackingsbus/... ./app/domain/inventory/lottrackingsapp/... ./api/domain/http/inventory/lottrackingsapi/...`

---

## Phase 6 — Lot-Location Junction Table (Gap 1, Option B, P3)

**Goal**: Create `inventory.lot_locations` as a proper junction table with quantity tracking, and a full domain implementation. The Phase 5 endpoint becomes backed by this table for non-serialized lots.

### Migration

**`business/sdk/migrate/sql/migrate.sql`** — Add new version (2.03) with a new CREATE TABLE:
```sql
-- Version: 2.03
-- Description: Add inventory.lot_locations junction table for tracking lot quantity per storage location.
CREATE TABLE inventory.lot_locations (
    id           UUID        NOT NULL,
    lot_id       UUID        NOT NULL,
    location_id  UUID        NOT NULL,
    quantity     numeric     NOT NULL DEFAULT 0,
    created_date TIMESTAMP   NOT NULL,
    updated_date TIMESTAMP   NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (lot_id) REFERENCES inventory.lot_trackings(id),
    FOREIGN KEY (location_id) REFERENCES inventory.inventory_locations(id),
    UNIQUE (lot_id, location_id)
);
```

### New Domain: lotlocationbus

**`business/domain/inventory/lotlocationbus/`** — Full domain following the standard pattern:

- `model.go`: `LotLocation`, `NewLotLocation`, `UpdateLotLocation`
  - Fields: `ID uuid`, `LotID uuid`, `LocationID uuid`, `Quantity decimal.Decimal` (use shopspring/decimal), `CreatedDate time.Time`, `UpdatedDate time.Time`
- `filter.go`: `QueryFilter` with `LotID *uuid`, `LocationID *uuid`
- `order.go`: ordering by `lot_id`, `location_id`, `quantity`, `created_date`, `updated_date`
- `lotlocationbus.go`: `Business` struct with `Create`, `Update`, `Delete`, `Query`, `Count`, `QueryByID`
- `testutil.go`: `TestNewLotLocation` seed helper
- `stores/lotlocationdb/`: DB store implementation

### App Layer

**`app/domain/inventory/lotlocationapp/`** — Standard app layer:
- `model.go`: App-facing models + `toBus*()` / `toApp*()` conversion
- `filter.go`: `QueryParams` + `parseFilter`
- `lotlocationapp.go`: `App` struct delegating to `lotlocationbus.Business`

### API Layer

**`api/domain/http/inventory/lotlocationapi/`** — Standard CRUD routes:
```
GET    /v1/inventory/lot-locations
GET    /v1/inventory/lot-locations/{lot_location_id}
POST   /v1/inventory/lot-locations
PUT    /v1/inventory/lot-locations/{lot_location_id}
DELETE /v1/inventory/lot-locations/{lot_location_id}
```

### Wiring

**`api/cmd/services/ichor/build/all/all.go`**:
```go
lotLocationBus := lotlocationbus.NewBusiness(cfg.Log, delegate, lotlocationdb.NewStore(cfg.Log, cfg.DB))
delegateHandler.RegisterDomain(delegate, lotlocationbus.DomainName, lotlocationbus.EntityName)
lotlocationapi.Routes(app, lotlocationapi.Config{...})
```

**`api/cmd/services/ichor/build/crud/crud.go`** — Mirror the wiring from all.go.

**`business/sdk/dbtest/seedFrontend.go`** — Add `inventory.lot_locations` table entry to seed data.

### Integration tests

**`api/cmd/services/ichor/tests/inventory/lotlocationapi/`** — Standard CRUD test file following the `apitest.Table` pattern.

**Verification**: `go build ./business/domain/inventory/lotlocationbus/... ./app/domain/inventory/lotlocationapp/... ./api/domain/http/inventory/lotlocationapi/... ./api/cmd/services/ichor/build/all/... ./api/cmd/services/ichor/build/crud/...`

---

## Summary Table

| Phase | Gap   | Priority | Scope                                                      | New files | Modified files |
|-------|-------|----------|------------------------------------------------------------|-----------|----------------|
| 1     | Gap 5 | P0       | Verify + fix lot_number/serial_number filters end-to-end  | 0         | 0–4 (if gaps)  |
| 2     | Gap 6 | P0       | expiry_before / expiry_after range filters on lot list     | 0         | 3–4            |
| 3     | Gap 2 | P1       | Quality status CHECK constraint + quality-statuses endpoint| 0         | 2              |
| 4     | Gaps 3&4 | P1    | Settings seeds (FEFO + quarantine) + RBAC on PUT endpoint  | 0         | 3–4            |
| 5     | Gap 1 | P2       | GET /{lot_id}/locations via serial aggregation             | 0         | 4–5            |
| 6     | Gap 1 | P3       | inventory.lot_locations full domain + CRUD endpoints        | ~12       | 3              |
