# Floor Worker Phase 10 — Supervisor Dashboard Backend Gap Analysis

**Date:** 2026-02-26
**Source:** Parallel 3-agent audit of backend vs. `docs/usage.md` Phase 10 spec
**Frontend branch:** `feature/floor-worker-ux-phase-10`
**Scope:** All backend work required to fully deliver Phase 10 (Supervisor Dashboard & Approvals)

---

## Phase 10 Overview

Phase 10 is a tablet-optimized supervisor interface with four functional areas:

| Screen | Purpose |
|---|---|
| **Exception Inbox** | Unified queue of all pending approvals — workflow approvals, inventory adjustments, transfer orders — with approve/reject actions |
| **Real-time Activity Feed** | Live transaction stream showing who did what, where, and when; anomaly drill-down |
| **Throughput KPIs** | Today's units received/picked/transferred; open task count; active workers |
| **Location Heatmap** | Future scope — not blocking Phase 10 |

---

## Audit Status Summary

| Domain | Endpoints Required | Status |
|---|---|---|
| Workflow Approvals | 4 routes | ✅ All present and functional |
| Workflow Alerts WebSocket | 1 route | ✅ Functional, browser-compatible |
| Workflow Alerts REST | 10 routes | ✅ All present |
| Inventory Adjustments GET | Filter | ⚠️ `approval_status` filter silently broken |
| Inventory Adjustments PUT | Update | ✅ Supports approve/reject via `approval_status` |
| Inventory Transfer Orders GET | Filter | ✅ `?status=` filter works end-to-end |
| Inventory Transactions GET | Date filter | ⚠️ No date-range — only exact equality |
| Throughput KPIs | Aggregation | ✅ Chart builder handles — DB seed records only |
| Active Workers | Presence | ❌ No backend concept exists |
| Worker Session Tracking | N/A | ❌ No backend concept exists |

---

## Layer Checklist (apply to every gap)

Per CLAUDE.md, all applicable layers must be addressed:
1. **Business layer** — model, filter, storer interface, business methods
2. **DB layer** — db model (null types), SQL queries, filter WHERE clauses, store implementation
3. **App layer** — app model, QueryParams, parseFilter, conversion functions (`toBus*`/`toApp*`)
4. **API layer** — HTTP handlers, route registration, query param binding
5. **Migration** — any schema changes (new column, constraint, index)
6. **Unit tests** — business layer tests for new methods/filters
7. **Integration tests** — `api/cmd/services/ichor/tests/` for new endpoints or params

---

## GAP 1 — `approval_status` Filter Broken on Inventory Adjustments GET

**Priority: P0 — Blocking**
**Effort: XS (~30 min, 3 lines)**
**Type: Bug**

### What's Wrong

`GET /v1/inventory/inventory-adjustments?approval_status=pending` silently ignores the filter and returns all adjustments. The supervisor's primary use case — "show me pending adjustments" — does not work.

### Root Cause

The filter pipeline has a break at the HTTP layer. All downstream logic is correct:

```
HTTP layer:  inventoryadjustmentapi/filter.go:12  ← parseQueryParams() never reads approval_status
App layer:   inventoryadjustmentapp/filter.go      ← ApprovalStatus field exists, never populated
Bus layer:   inventoryadjustmentbus/filter.go      ← correctly applies WHERE clause if set
DB layer:    inventoryadjustmentdb/filter.go:38    ← SQL: WHERE approval_status = :approval_status ✅
```

### Fix

**File:** `api/domain/http/inventory/inventoryadjustmentapi/filter.go`

Inside `parseQueryParams()`, add (mirror pattern of other string filters like `reason_code`):
```go
if v := r.URL.Query().Get("approval_status"); v != "" {
    qp.ApprovalStatus = v
}
```

**Verify:** `app/domain/inventory/inventoryadjustmentapp/filter.go` — confirm `ApprovalStatus` is already passed through to the bus filter (it is, but verify the mapping in `parseFilter`).

### Layer Checklist

- [ ] API layer — `inventoryadjustmentapi/filter.go` — add URL param extraction
- [ ] App layer — `inventoryadjustmentapp/filter.go` — verify `ApprovalStatus` passes to bus filter
- [ ] Integration test — `tests/inventory/inventoryadjustmentapi/` — add test case for `?approval_status=pending`

---

## GAP 2 — No Date-Range Filtering on Inventory Transactions

**Priority: P1 — Needed for Activity Feed**
**Effort: M (2–3 hrs, pattern work across 4 files)**
**Type: Missing Feature**

### What's Wrong

All date filters on `inventory_transactions` are exact equality (`=`). "Show me the last 2 hours of activity" or "today's transactions" is impossible. This cripples the Real-time Activity Feed screen.

### Root Cause

`QueryFilter` has only `TransactionDate *time.Time` (single point). DB filter applies `transaction_date = :transaction_date`. No range fields exist at any layer.

### Fix

**File:** `business/domain/inventory/inventorytransactionbus/filter.go`
```go
// Add to QueryFilter:
TransactionDateFrom *time.Time
TransactionDateTo   *time.Time
CreatedDateFrom     *time.Time
CreatedDateTo       *time.Time
```

**File:** `business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb/filter.go`
```go
if filter.TransactionDateFrom != nil {
    data["transaction_date_from"] = *filter.TransactionDateFrom
    wc = append(wc, "transaction_date >= :transaction_date_from")
}
if filter.TransactionDateTo != nil {
    data["transaction_date_to"] = *filter.TransactionDateTo
    wc = append(wc, "transaction_date < :transaction_date_to")
}
// same pattern for created_date_from / created_date_to
```

**File:** `app/domain/inventory/inventorytransactionapp/filter.go` and `model.go`
- Add `TransactionDateFrom`, `TransactionDateTo`, `CreatedDateFrom`, `CreatedDateTo` to `QueryParams`
- Map through to bus `QueryFilter` in `parseFilter`

**File:** `api/domain/http/inventory/inventorytransactionapi/filter.go`
- Add URL param extraction for `transaction_date_from`, `transaction_date_to`, `created_date_from`, `created_date_to`
- Parse as `time.Time` (ISO 8601 format, same pattern as other date params in the codebase)

### Layer Checklist

- [ ] Business layer — `inventorytransactionbus/filter.go` — add range fields to `QueryFilter`
- [ ] DB layer — `inventorytransactiondb/filter.go` — add range WHERE clauses
- [ ] App layer — `inventorytransactionapp/filter.go` + `model.go` — add to `QueryParams`, pass through
- [ ] API layer — `inventorytransactionapi/filter.go` — parse URL params
- [ ] Unit test — `inventorytransactionbus_test.go` — test date-range filter returns correct rows
- [ ] Integration test — `tests/inventory/inventorytransactionapi/` — test `?transaction_date_from=` + `to=`

---

## GAP 3 — Date-Range Gap on Inventory Adjustments (Same Pattern as GAP 2)

**Priority: P2**
**Effort: S (1 hr, identical pattern)**
**Type: Missing Feature**

### What's Wrong

`inventory_adjustments` has the same exact-equality-only date filter on `adjustment_date`, `created_date`, and `updated_date`. Supervisors cannot query "adjustments submitted in the last shift."

### Fix

Identical pattern to GAP 2, applied to the adjustments domain:

- `business/domain/inventory/inventoryadjustmentbus/filter.go` — add `AdjustmentDateFrom`, `AdjustmentDateTo`, `CreatedDateFrom`, `CreatedDateTo`
- `inventoryadjustmentdb/filter.go` — add `>=` / `<` WHERE clauses
- `inventoryadjustmentapp/filter.go` + `model.go` — pass through
- `inventoryadjustmentapi/filter.go` — parse URL params

### Layer Checklist

- [ ] Business layer — `inventoryadjustmentbus/filter.go`
- [ ] DB layer — `inventoryadjustmentdb/filter.go`
- [ ] App layer — `inventoryadjustmentapp/filter.go` + `model.go`
- [ ] API layer — `inventoryadjustmentapi/filter.go`
- [ ] Integration test — `tests/inventory/inventoryadjustmentapi/`

---

## GAP 4 — `lot_id` Dropped from Transaction API Response

**Priority: P2 — Audit Trail Completeness**
**Effort: XS (~30 min)**
**Type: Data Omission**

### What's Wrong

`inventory_transactions.lot_id` is present in the DB and the business model but `ToAppInventoryTransaction()` in the app layer never maps it. The field is silently discarded — lot traceability is invisible in the activity feed drill-down.

### Root Cause

```go
// app/domain/inventory/inventorytransactionapp/model.go
// ToAppInventoryTransaction() — LotID is never set on the return struct
// The app struct has no LotID field at all
```

The business model (`inventorytransactionbus/model.go`) has `LotID *uuid.UUID`. The DB model reads it. The app struct omits it.

### Fix

**File:** `app/domain/inventory/inventorytransactionapp/model.go`
- Add `LotID string `json:"lot_id"`` to the `InventoryTransaction` app struct
- Map `LotID` in `ToAppInventoryTransaction()` (handle nullable UUID → empty string)

### Layer Checklist

- [ ] App layer — `inventorytransactionapp/model.go` — add field + mapping
- [ ] Integration test — verify `lot_id` appears in response for lot-tracked transactions

---

## GAP 5 — Transfer Order Status Has No Validation

**Priority: P2 — Data Integrity**
**Effort: XS (~30 min)**
**Type: Data Integrity**

### What's Wrong

`inventory.transfer_orders.status` is a free-form `varchar(20)` with no CHECK constraint and no `oneof` app-layer validation. Any string is accepted. Contrast with inventory adjustments which correctly validates `oneof=pending approved rejected`.

Risk: supervisor filter `?status=pending` silently misses rows written with `"Pending"` or `"PENDING"`.

### Fix

**Option A — App layer `oneof` validation (recommended, faster):**

**File:** `app/domain/inventory/transferorderapp/model.go`
```go
// On UpdateTransferOrder.Status:
Status *string `json:"status" validate:"omitempty,oneof=pending approved in_progress completed cancelled"`
```

**Option B — Database constraint (more durable, add alongside Option A):**
```sql
ALTER TABLE inventory.transfer_orders
  ADD CONSTRAINT transfer_orders_status_check
  CHECK (status IN ('pending', 'approved', 'in_progress', 'completed', 'cancelled'));
```

### Layer Checklist

- [ ] App layer — `transferorderapp/model.go` — add `oneof` validation tag
- [ ] Migration (optional) — add DB CHECK constraint
- [ ] Unit test — verify invalid status string returns 400

---

## GAP 6 — No Active Workers / Worker Presence System

**Priority: P2 — KPI Screen**
**Effort: M (Option A: 3–4 hrs) or L (Option B: full day)**
**Type: Missing Feature**

### What's Wrong

The "Active Workers" KPI on the supervisor dashboard has no backend support. The `User` model has no `last_active`, `online`, or session tracking fields. The `AlertHub` WebSocket manager tracks live connections in memory but has no HTTP query endpoint.

### Option A — WebSocket Presence Endpoint (Recommended for Phase 10)

Expose connected user IDs from the existing in-memory `AlertHub` via a new HTTP route. No DB migration.

**File:** `api/domain/http/workflow/alertws/alerthub.go`
- Add `ConnectedUserIDs() []uuid.UUID` method that reads from the hub's connection map

**New package:** `api/domain/http/floor/activeapi/`
- Handler: `GET /v1/floor/active-workers` → returns `{ count: N, user_ids: [...] }`
- Wire into `all.go`

Limitation: only counts users with live WebSocket connections — represents supervisors/workers who currently have the app open.

### Option B — Last-Activity Timestamp on User (Durable, Historical)

Add `last_activity_at *time.Time` to `core.users`. Update via auth middleware (debounced — only write if > 5 min since last update to avoid write amplification).

**Files:**
- New migration: `last_activity_at timestamptz NULL` on `core.users`
- `business/domain/core/userbus/model.go` — add field
- `api/sdk/http/mid/authen.go` — update timestamp post-auth (debounced)
- `core/userbus/filter.go` — add `ActiveSince *time.Time` filter
- `api/domain/http/core/userapi/filter.go` — expose `?active_since=` query param

**Recommendation:** Option A for Phase 10 (no migration, immediate value). Option B in a later phase for historical shift reporting.

### Acceptable Workaround (Zero Backend Changes)

Approximate active workers by counting distinct workers in put-away tasks with `status=in_progress`:
```
GET /v1/inventory/put-away-tasks?status=in_progress&rows=1
```
Read `total` from pagination wrapper. Represents "workers with tasks in-progress right now." Reasonable proxy for Phase 10.

### Layer Checklist (Option A)

- [ ] `alertws/alerthub.go` — add `ConnectedUserIDs()` method
- [ ] New `api/domain/http/floor/activeapi/` package — handler + route
- [ ] `all.go` — wire new route
- [ ] Integration test — connect WebSocket, verify user appears in active list

---

## GAP 7 — Chart Builder KPI Configs Not Seeded

**Priority: P1 — Required for KPI Screen**
**Effort: S (2–3 hrs, SQL seed records only — no Go code)**
**Type: Missing Seed Data**

### What's Wrong

The Throughput KPI screen requires no new Go code — the chart builder engine (`POST /v1/data/chart/name/{name}`) handles aggregation via `MetricConfig` + `GroupByConfig`. The `KPICard.vue` and `KPIGrid.vue` components are wired and ready. However, the four `config.table_configs` records that define the KPI queries don't exist yet.

### KPI Configs to Seed

| Config Name | Source Table | Metric | Static Filters |
|---|---|---|---|
| `floor_kpi_open_tasks` | `inventory.put_away_tasks` | `count(id)` | `status IN (pending, in_progress)` |
| `floor_kpi_received_today` | `inventory.put_away_tasks` | `count(id)` | `status = completed` + dynamic `completed_at >= today` |
| `floor_kpi_transferred_today` | `inventory.transfer_orders` | `count(id)` | `status = completed` + dynamic `updated_date >= today` |
| `floor_kpi_picked_today` | `inventory.inventory_transactions` | `sum(quantity)` | `transaction_type = pick` + dynamic `transaction_date >= today` |

### "Today" Date Boundary

"Today" filters cannot be baked in as static string literals. Two approaches:
- **Dynamic filter injection:** Frontend passes today's ISO date at midnight UTC via `ChartQuery.dynamic` map in the POST body. `useChartBuilder` supports this via `setQuery()`.
- **SQL expression groupby:** Use `GroupByConfig` with `Expression: true` and `DATE_TRUNC('day', column) = CURRENT_DATE` as the expression (avoids frontend date passing entirely).

Recommend the SQL expression approach for robustness — no clock-skew risk.

### Config JSON Structure (reference `SeedKPIOrderCount` in `business/sdk/dbtest/seedmodels/charts.go`)

```json
{
  "title": "Open Tasks",
  "widget_type": "chart",
  "visualization": "kpi",
  "data_source": [{
    "type": "query",
    "schema": "inventory",
    "source": "put_away_tasks",
    "metrics": [{
      "name": "task_count",
      "function": "count",
      "column": "put_away_tasks.id"
    }],
    "filters": [
      { "column": "put_away_tasks.status", "operator": "in", "value": ["pending", "in_progress"] }
    ]
  }],
  "visual_settings": {
    "columns": {
      "_chart": {
        "cell_template": "{\"chartType\":\"kpi\",\"valueColumns\":[\"task_count\"],\"kpi\":{\"label\":\"Open Tasks\",\"format\":\"number\"}}"
      }
    }
  }
}
```

### Layer Checklist

- [ ] Seed SQL — insert 4 `config.table_configs` records
- [ ] Seed SQL — insert corresponding `config.page_content` rows linking to supervisor dashboard page config
- [ ] Verify each config runs correctly via `POST /v1/data/chart/name/{config_name}`
- [ ] Add to `dbtest/seedmodels/` for use in integration tests

---

## GAP 8 — No Dedicated Approve/Reject Endpoints for Adjustments

**Priority: P3 — Ergonomics**
**Effort: S (1–2 hrs)**
**Type: Ergonomics / API Symmetry**

### What's Wrong

Transfer orders have `POST /v1/inventory/transfer-orders/{id}/approve` which is semantically clear. Inventory adjustments require a generic `PUT` with `{ "approval_status": "approved", "approved_by": "<uuid>" }` — asymmetric and less intuitive for the supervisor action pattern.

### Fix (Optional)

Add two new routes:
```
POST /v1/inventory/inventory-adjustments/{id}/approve
POST /v1/inventory/inventory-adjustments/{id}/reject
```
Body: `{ "approved_by": "<uuid>", "reason": "..." }`

Internally delegates to the existing `Update` business method with `ApprovalStatus` set.

### Layer Checklist

- [ ] App layer — new `ApprovementRequest` body struct in `inventoryadjustmentapp/model.go`
- [ ] API layer — new `approve` and `reject` handlers in `inventoryadjustmentapi/inventoryadjustmentapi.go`
- [ ] API layer — register routes in `inventoryadjustmentapi/routes.go`
- [ ] Integration test — `tests/inventory/inventoryadjustmentapi/` — approve + reject happy paths

---

## GAP 9 — No Date-Range on Workflow Alerts

**Priority: P3**
**Effort: S (1 hr, same pattern)**
**Type: Missing Feature**

### What's Wrong

`alertbus/filter.go` `QueryFilter` has no `CreatedAfter`/`CreatedBefore` range fields. Cannot query "alerts in the last hour" — needed for a time-bounded supervisor alert feed.

### Fix

Same date-range pattern as GAP 2, applied to the alert domain:

- `business/domain/workflow/alertbus/filter.go` — add `CreatedDateFrom`, `CreatedDateTo`
- `alertdb/filter.go` — add `>=` / `<` WHERE clauses
- `app/domain/workflow/alertapp/filter.go` + `model.go` — pass through
- `api/domain/http/workflow/alertapi/filter.go` — parse URL params

### Layer Checklist

- [ ] Business layer — `alertbus/filter.go`
- [ ] DB layer — `alertdb/filter.go`
- [ ] App layer — `alertapp/filter.go` + `model.go`
- [ ] API layer — `alertapi/filter.go`
- [ ] Integration test — `tests/workflow/alertapi/`

---

## GAP 10 — WebSocket CORS Hardcoded to `["*"]`

**Priority: P3 — Security/Config**
**Effort: XS (~30 min)**
**Type: Configuration**

### What's Wrong

`all.go:1324` passes `CORSAllowedOrigins: []string{"*"}` to the WebSocket hub with a `// TODO: Configure from environment` comment. Fine for development; needs to respect `ICHOR_WEB_CORSALLOWEDORIGINS` before production.

### Fix

**File:** `api/cmd/services/ichor/build/all/all.go` (line ~1324)

Replace hardcoded `[]string{"*"}` with the `cfg.Web.CORSAllowedOrigins` value that is already used for HTTP CORS.

### Layer Checklist

- [ ] `all.go` — pass `cfg.Web.CORSAllowedOrigins` to WebSocket hub config

---

## Behavioral Notes for Frontend (No Backend Changes Needed)

### Workflow Approvals

- `/mine` returns **all statuses by default** — pass `?status=pending` to show only actionable items
- `/resolve` returns `409 FailedPrecondition` if already resolved — show "already actioned by another user" message
- Resolve body: `{ "resolution": "approved"|"rejected", "reason": "..." }` — `reason` is optional but surfaces in audit trail
- `/{id}` has no ownership check — any authenticated user can fetch any approval by UUID (intentional for deep-link from notifications)
- Temporal integration is conditional — in dev without Temporal, DB updates correctly but workflow is not unblocked

### Inventory Adjustments

- Approve/reject via `PUT`: `{ "approval_status": "approved", "approved_by": "<supervisor_uuid>" }`
- Valid `approval_status` values: `pending`, `approved`, `rejected` (enforced by `oneof` validation)
- New adjustments always start as `pending` — business layer hard-codes this on create
- No row-level security — all users with Read permission see all adjustments across all locations

### Transfer Orders

- Approve via dedicated endpoint: `POST /{id}/approve` with `{ "approved_by": "<uuid>" }`
- Reject via `PUT`: `{ "status": "cancelled" }` — no `"rejected"` status exists in business logic
- `?status=pending` filter works correctly end-to-end

### WebSocket Alert Stream

- URL: `wss://{host}/v1/workflow/alerts/ws?token={jwt}` — JWT in query param (browsers can't set Authorization header on WebSocket upgrade)
- Message envelope: `{ "type": "alert", "payload": { ...Alert... }, "timestamp": "RFC3339" }`
- Alert fields: `id`, `alertType`, `severity` (low/medium/high/critical), `title`, `message`, `context` (arbitrary JSON), `sourceEntityName`, `sourceEntityID`, `sourceRuleID`, `status`, `expiresDate`
- Frontend must implement exponential backoff reconnect — backend has no reconnect management
- CORS uses `OriginPatterns` from env var (defaults to `["*"]` in dev)

---

## Recommended Implementation Order

### Sprint 1 — Unlock Core Functionality (~half day backend)

| # | Gap | Effort | Outcome |
|---|---|---|---|
| 1 | Fix `approval_status` filter bug | XS | Exception Inbox adjustment queue works |
| 2 | Date-range on inventory transactions | M | Activity feed time-scoping works |
| 7 | Seed chart builder KPI configs | S | KPI screen fully functional |

### Sprint 2 — Completeness (~half day backend)

| # | Gap | Effort | Outcome |
|---|---|---|---|
| 3 | Date-range on inventory adjustments | S | Adjustment queue time-scoping |
| 4 | `lot_id` in transaction response | XS | Full lot traceability in audit feed |
| 5 | Transfer order status validation | XS | Data integrity enforcement |
| 9 | Date-range on workflow alerts | S | Time-bounded alert feed |

### Sprint 3 — Enhancement (~1 day backend)

| # | Gap | Effort | Outcome |
|---|---|---|---|
| 6 | Active workers presence endpoint | M | Real-time worker count KPI |
| 8 | Dedicated approve/reject for adjustments | S | API symmetry with transfer orders |
| 10 | WebSocket CORS from env config | XS | Production-ready CORS |

---

## Complete Gap Registry

| # | Gap | Priority | Type | Effort | Key Files |
|---|---|---|---|---|---|
| 1 | `approval_status` filter broken on adjustments GET | P0 | Bug | XS | `inventoryadjustmentapi/filter.go` |
| 2 | No date-range filter on inventory transactions | P1 | Feature | M | `inventorytransactionbus/filter.go`, `inventorytransactiondb/filter.go`, `inventorytransactionapp/filter.go`, `inventorytransactionapi/filter.go` |
| 3 | No date-range filter on inventory adjustments | P2 | Feature | S | Same pattern as #2 for adjustments |
| 4 | `lot_id` dropped from transaction API response | P2 | Data | XS | `inventorytransactionapp/model.go` |
| 5 | Transfer order status has no validation | P2 | Integrity | XS | `transferorderapp/model.go` + optional migration |
| 6 | No active workers / worker presence system | P2 | Feature | M–L | New `floor/activeapi/`, `alertws/alerthub.go` |
| 7 | Chart builder KPI configs not seeded | P1 | Seed data | S | `config.table_configs` DB records only |
| 8 | No dedicated approve/reject for adjustments | P3 | Ergonomics | S | `inventoryadjustmentapi/` new routes |
| 9 | No date-range filter on workflow alerts | P3 | Feature | S | `alertbus/filter.go`, `alertdb/filter.go`, `alertapi/filter.go` |
| 10 | WebSocket CORS hardcoded to `["*"]` | P3 | Security | XS | `all.go:1324` |
