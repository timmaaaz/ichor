# Floor Module Blockers

Consolidated backend and frontend gaps preventing full floor worker operation.
Each file is a self-contained task with enough detail for Superpowers to implement independently.

**Last verified:** 2026-03-16 (agents confirmed each blocker against actual codebase)

**Backend repo:** `../../../ichor/` relative to this Vue frontend
**Architecture:** Ardan Labs 7-layer (business → app → api). All domain changes must address all 7 layers.
**Arch docs:** `../../../ichor/docs/arch/` — read the relevant arch file before touching any domain

---

## Priority Order

### Genuine Blockers (verified 2026-03-16)

| # | File | Severity | Summary |
|---|------|----------|---------|
| 003 | `blocker-003-adjustments-qty-change-not-applied.md` | 🔴 CRITICAL | Approving an adjustment never mutates inventory — also blocks Cycle Count reconciliation |
| 004 | `blocker-004-transfers-execute-flow.md` | 🔴 CRITICAL | Transfers have no in_transit/completed status or atomic stock-move endpoint |
| 002 | `blocker-002-picking-fefo-sql.md` | 🔴 HIGH | FEFO allocation falls back to FIFO silently — lot-tracked picks ignore expiry |
| 008 | `blocker-008-inspections-atomic-fail-quarantine.md` | 🟡 MEDIUM | Fail + quarantine are two non-atomic calls; no composite endpoint; status has no CHECK constraint |
| 006 | `blocker-006-adjustments-reason-code-validation.md` | 🟡 MEDIUM | reason_code is free-text; no validation or lookup table |
| 005 | `blocker-005-putaway-frontend-wiring.md` | 🟠 HIGH | Backend complete — frontend usePutAway not wired to correct status-transition endpoints |
| 009 | `blocker-009-supervisor-bulk-ack-kpi.md` | 🟢 LOW-MED | Bulk ack FIXED; KPI aggregation endpoint still missing |
| 011 | `blocker-011-receiving-discrepancy-notes.md` | 🟢 LOW | `ReceiveQuantityRequest` missing `Notes` field — trivial fix, one struct field |

### Resolved / Not Blockers (verified 2026-03-16)

| # | File | Original Severity | Resolution |
|---|------|-------------------|------------|
| 001 | `blocker-001-picking-http-routes.md` | 🔴 CRITICAL | **FIXED** — routes exist at `/v1/sales/order-line-items/{id}/pick-quantity` via `orderlineitemsapi`. Frontend calls wrong URL path. |
| 010 | `blocker-010-client-side-date-filters.md` | 🟢 LOW | **FIXED** — both PO date filters and lot expiry filters fully implemented in backend. Frontend needs to use them. |
| 007 | `blocker-007-scan-lookup-universal-resolver.md` | 🟡 MEDIUM | **DOWNGRADED** — `location_code` column exists, `scanapp.go:82` queries via `LocationCodeExact`. Works if barcodes encode location code. Design choice, not a blocker. |

---

## Architecture Quick Reference

When implementing any backend change, follow this checklist (from `docs/arch/domain-template.md`):

1. Business layer — `business/domain/{area}/{entity}bus/`
2. DB store — `business/domain/{area}/{entity}bus/stores/{entity}db/`
3. App layer — `app/domain/{area}/{entity}app/`
4. API layer — `api/domain/http/{area}/{entity}api/`
5. Route wiring — `api/cmd/services/ichor/build/all/all.go`
6. Migration — `business/sdk/migrate/sql/migrate.sql` (add new version, never edit existing)
7. Tests — `api/cmd/services/ichor/tests/{area}/{entity}api/`

**NEVER run `go test ./...`** — run only the packages you changed.
**Always run `go build ./...`** before reporting completion.
