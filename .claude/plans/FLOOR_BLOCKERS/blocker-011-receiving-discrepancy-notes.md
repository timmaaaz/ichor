# Blocker 011: Receiving — Discrepancy Notes Silently Dropped

> **STATUS: CONFIRMED** (verified 2026-03-16)
> Receiving uses `POST /v1/procurement/purchase-order-line-items/{id}/receive-quantity`.
> `ReceiveQuantityRequest` in `purchaseorderlineitemapp/model.go` only has `Quantity` and `ReceivedBy`.
> The business model (`purchaseorderlineitembus/model.go:28`) HAS a `Notes` field that is persisted (lines 45, 61).
> **Root cause:** `ReceiveQuantityRequest` struct is missing a `Notes`/`DiscrepancyNotes` field.
> **Fix:** Add one field to the request struct + pass it through to the business layer. Trivial.

**Severity:** LOW — functional regression; audit trail is incomplete for discrepancy events
**Domain:** `procurement/purchaseorderlineitem` (not a separate receiving domain)
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

When a floor worker flags a discrepancy during receiving (`useReceiving.ts`), they can enter
free-text notes describing the issue (e.g., "3 units damaged in transit, outer box punctured").

These notes are sent to the backend in the receiving submission payload but the backend endpoint
does not store them — they are silently ignored. This means:
1. Discrepancy notes never appear in any audit view
2. Supervisors have no context when reviewing discrepancy records
3. Data is written to the network but goes nowhere

---

## Investigation Steps

Before implementing, confirm the exact gap:

1. Find the receiving submission endpoint — search for `POST /v1/procurement/receiving` or similar
   in `api/domain/http/procurement/` or `api/domain/http/inventory/`
2. Check the `NewReceiving` or `ReceivingLine` model for a `notes` or `discrepancy_notes` field
3. Check the DB schema for a `notes` column on the relevant table
4. Find where `useReceiving.ts` sends the notes field in its payload

The gap is likely one of:
- **Missing DB column:** `notes VARCHAR` not in the table, migration needed
- **Missing model field:** `Notes string` not in the business/app model, silently ignored during bind
- **Missing store write:** Field in model but not included in the SQL INSERT

---

## Likely Fix

### If DB column missing:
```sql
-- Version: X.XX
-- Description: Add notes column to receiving lines for discrepancy context
ALTER TABLE procurement.receiving_lines  -- (verify exact table name)
    ADD COLUMN discrepancy_notes TEXT;
```

### If model field missing:
Add `DiscrepancyNotes *string` to `NewReceivingLine` / `UpdateReceivingLine` model and
include it in the SQL INSERT/UPDATE statement.

### If store write missing:
Locate the INSERT in `stores/receivingdb/receivingdb.go` and add the missing column binding.

---

## Key Files to Read Before Implementing

- `business/sdk/migrate/sql/migrate.sql` — find the receiving table schema
- `business/domain/procurement/` or `business/domain/inventory/` — find receiving bus
- `app/domain/` — find the receiving app model with `NewReceivingLine` struct
- `src/composables/floor/useReceiving.ts` — what field name it sends
- `src/utils/floorApi.ts` — check the payload shape for receiving submission

---

## Acceptance Criteria

- [ ] Submitting a receiving line with discrepancy notes persists the notes in the database
- [ ] `GET /v1/.../receiving/{id}` returns `discrepancy_notes` in the response
- [ ] Supervisor discrepancy review view shows notes (frontend may also need updating)
- [ ] Empty/null notes are accepted (discrepancy notes are optional)
- [ ] `go build ./...` passes
- [ ] Migration applies cleanly on fresh DB
