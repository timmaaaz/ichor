# Blocker 006: Adjustments — reason_code is Free-Text with No Backend Validation

> **STATUS: CONFIRMED** (verified 2026-03-16)
> `migrate.sql:789` — `reason_code varchar(50) NOT NULL` with no CHECK constraint.
> `inventoryadjustmentbus.go:92` — ReasonCode assigned with zero validation.
> No `ValidReasonCodes` map or enum exists in the package.

**Severity:** MEDIUM — data quality issue; invalid reason codes are accepted silently
**Domain:** `inventory/inventoryadjustment`
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

`inventory_adjustments.reason_code` is a free-text `VARCHAR` field. The frontend
(`src/composables/floor/useAdjustment.ts`) hardcodes a list of valid reason options:

```
damaged | theft | data_entry_error | receiving_error | picking_error | found_stock | other
```

But the backend accepts any string — there's no validation, no lookup table, no enum. This means:
1. Typos silently persist in the audit trail
2. Reporting on reasons is unreliable (same reason may appear with different strings)
3. No single source of truth for valid codes

---

## What Should Be Built

### Option A: CHECK constraint (lightweight, no new domain needed)

Add a Postgres CHECK constraint on `inventory.inventory_adjustments.reason_code`:

```sql
-- Version: X.XX
-- Description: Add CHECK constraint on inventory_adjustments.reason_code
ALTER TABLE inventory.inventory_adjustments
    ADD CONSTRAINT chk_reason_code CHECK (
        reason_code IN (
            'damaged', 'theft', 'data_entry_error',
            'receiving_error', 'picking_error', 'found_stock', 'other'
        )
    );
```

Also add validation in the business layer before the DB write:

```go
// business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go
var ValidReasonCodes = map[string]bool{
    "damaged": true, "theft": true, "data_entry_error": true,
    "receiving_error": true, "picking_error": true, "found_stock": true, "other": true,
}
```

### Option B: Reason codes lookup table (richer, enables future extensibility)

Create a new `inventory.adjustment_reason_codes` table:

```sql
CREATE TABLE inventory.adjustment_reason_codes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(64) UNIQUE NOT NULL,
    label       TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    date_created TIMESTAMP NOT NULL DEFAULT NOW()
);
```

Seed with the 7 current values. Add `GET /v1/inventory/adjustment-reason-codes` endpoint
for the frontend to fetch dynamically instead of hardcoding.

**Recommendation:** Option A is faster. Option B is better long-term. Choose based on roadmap.

---

## Frontend Impact

If Option B is chosen, update `src/composables/floor/useAdjustment.ts` to fetch reason codes
dynamically from `GET /v1/inventory/adjustment-reason-codes` instead of using the hardcoded array.

---

## Key Files to Read Before Implementing

- `business/domain/inventory/inventoryadjustmentbus/model.go` — current ReasonCode field
- `business/sdk/migrate/sql/migrate.sql` — migration format and version numbering
- `docs/arch/domain-template.md` — if creating new domain (Option B)

---

## Acceptance Criteria

- [ ] Submitting an adjustment with `reason_code = "typo_reason"` returns 400
- [ ] All 7 valid reason codes are accepted
- [ ] Migration applies cleanly on fresh DB
- [ ] `go build ./...` passes
- [ ] (Option B only) `GET /v1/inventory/adjustment-reason-codes` returns active codes
