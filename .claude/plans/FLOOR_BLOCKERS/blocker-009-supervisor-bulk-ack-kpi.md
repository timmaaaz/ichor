# Blocker 009: Supervisor — No Bulk Acknowledge + No KPI Aggregation Endpoint

> **STATUS: PARTIAL** (verified 2026-03-16)
> **Part A (Bulk Ack): FIXED** — `POST /v1/workflow/alerts/acknowledge-selected` and
> `POST /v1/workflow/alerts/acknowledge-all` exist at `alertapi/route.go:46-47`.
> Business methods: `AcknowledgeSelected()` at `alertbus.go:255`, `AcknowledgeAll()` at `alertbus.go:285`.
> Request models support notes via `BulkSelectedRequest.Notes` and `BulkAllRequest.Notes`.
> **Part B (KPI): STILL MISSING** — no aggregation endpoint exists.

**Severity:** LOW-MEDIUM — ~~supervisor dashboard works but is friction-heavy at scale~~ bulk ack resolved; KPI aggregation still missing
**Domain:** `workflow/alerts` (bulk ack — DONE) + new `inventory/supervisorkpi` or query extension (TODO)
**Backend repo:** `../../../ichor/` relative to Vue frontend

---

## Problem

### Part A: Bulk Acknowledge — FIXED

~~Supervisors acknowledge alerts one at a time.~~ Backend now supports:
- `POST /v1/workflow/alerts/acknowledge-selected` — bulk ack by ID list
- `POST /v1/workflow/alerts/acknowledge-all` — ack all matching alerts

**Frontend** `useSupervisorDashboard.ts` may still need updating to call these endpoints.

### Part B: KPI Aggregation

The supervisor KPI cards (pending approvals count, daily receiving throughput, open exceptions)
are derived from client-side `.length` on list responses. This means:
1. Paginated responses only show KPIs for the current page — if page=1 shows 25 items and
   there are 100, the KPI shows "25 pending" not "100 pending"
2. No historical throughput data (items received today, picks completed today)

---

## What Should Be Built

### Part A: Bulk Acknowledge Endpoint

`POST /v1/alerts/acknowledge-bulk`

Request body:
```json
{
  "alert_ids": ["uuid1", "uuid2", "uuid3"]
}
```

Response: `{ "acknowledged": 3, "failed": 0 }`

Implementation: loop over IDs in a single transaction, acknowledge each.
Partial success is acceptable — return counts of succeeded/failed.

**Files:**
- `business/domain/workflow/` — find the alerts bus, look for `Acknowledge(ctx, alertID)` method
- `api/domain/http/workflow/` — add bulk endpoint to alert routes
- Check `docs/arch/workflow-alerts.md` in the **Go backend** (`../../../ichor/docs/arch/`)
  for the alerts domain structure before touching anything

### Part B: KPI Endpoint

`GET /v1/inventory/supervisor/kpis`

Response:
```json
{
  "pending_approvals": 12,
  "open_adjustments": 5,
  "pending_transfers": 3,
  "picks_completed_today": 47,
  "items_received_today": 203,
  "open_inspections": 8
}
```

Implementation options:
- **Dedicated endpoint** in a new `supervisorkpiapp` — queries multiple buses, returns aggregates
- **Extend existing COUNT endpoints** — add `?date=today` filter to each domain, frontend assembles

Dedicated endpoint is better for the supervisor dashboard performance. Use
`COUNT(*)` queries (not full result sets) for each metric.

---

## Key Files to Read Before Implementing

- `docs/arch/workflow-alerts.md` (Go backend, at `../../../ichor/docs/arch/workflow-alerts.md`) — alerts domain
- `business/domain/workflow/` — current alert business packages
- `api/domain/http/workflow/` — existing alert API routes
- `business/domain/inventory/inventoryadjustmentbus/` — for pending adjustment count
- `business/domain/inventory/transferorderbus/` — for pending transfer count

---

## Acceptance Criteria

**Bulk Ack:**
- [ ] `POST /v1/alerts/acknowledge-bulk` with 10 IDs → all 10 acknowledged in one round-trip
- [ ] IDs that don't exist or are already acknowledged → skipped gracefully, counted in `failed`
- [ ] Empty `alert_ids` array → 400

**KPIs:**
- [ ] `GET /v1/inventory/supervisor/kpis` returns correct counts (not page-limited)
- [ ] `picks_completed_today` reflects picks with `completed_at >= today midnight`
- [ ] Response time < 200ms (use COUNT queries, not full result fetches)
- [ ] `go build ./...` passes
