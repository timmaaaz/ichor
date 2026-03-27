# Floor Worker Backend — Remaining Work

**Generated:** 2026-03-25
**Source:** Gap analysis at `vue/ichor/docs/guides/features/floor-worker-backend-gaps.md`
**Verified against:** Current codebase state as of 2026-03-25

This document lists only the items that are still missing or incomplete. Items marked DONE in the original gap analysis have been removed.

---

## Legend

| Tag | Meaning |
|-----|---------|
| MISSING | Not implemented — needs to be built |
| PARTIAL | Exists but incomplete — needs additional work |
| DESIGN DECISION | May not need traditional implementation — current approach may suffice |

---

## P0 — Blocks Frontend Phases

### 1. Standalone Add-Stock Endpoint (Phase 3 — Put-Away)

**Status:** MISSING | **Impact:** Medium | **Blocks:** Phase 3 (partially)

No `POST /v1/inventory/inventory-items/{id}/add-stock` endpoint exists. The put-away task domain's `complete` flow calls `UpsertQuantity` internally, which covers the primary use case.

**Decision needed:** Is the put-away task completion flow sufficient, or do we need a standalone endpoint for ad-hoc stock additions outside the put-away workflow (e.g., found stock, manual corrections)?

**If needed, implement:**
- `POST /v1/inventory/inventory-items/{id}/add-stock`
- Body: `{ quantity: string, location_id: string, reference_number: string, notes: string }`
- Must atomically: increment `inventory_items.quantity`, create `inventory_transaction` with `transaction_type = "receive"`

---

### 2. Transfer Order Statuses Reference Table (Phase 6 — Transfers)

**Status:** DESIGN DECISION | **Impact:** Low

Statuses are hardcoded Go constants. No `transfer_order_statuses` DB table or seed data exists.

**Current state:** Transfer order lifecycle works — `execute` endpoint sets status to `completed`. Values are validated in the business layer.

**Decision needed:** Is a DB reference table required for runtime configurability, or are hardcoded constants acceptable? If the frontend only needs a known list of statuses, the current approach works and the API could expose a `GET /v1/inventory/transfer-orders/statuses` endpoint returning the hardcoded list.

---

### 3. Adjustment Reason Codes Reference Table (Phase 5 — Adjustments)

**Status:** DESIGN DECISION | **Impact:** Low

Seven reason codes (`damaged`, `theft`, `data_entry_error`, `receiving_error`, `picking_error`, `found_stock`, `other`) are hardcoded in `inventoryadjustmentbus.go:36-55` with a `ValidReasonCodes` map.

**Current state:** Validation works. Codes are enforced.

**Decision needed:** Same as #2 — DB table vs. hardcoded constants. If the list is fixed, expose a `GET /v1/inventory/inventory-adjustments/reason-codes` endpoint returning the hardcoded list so the frontend doesn't duplicate them.

---

## P1 — Correctness & UX Improvements

### 4. serial_id on Inventory Transactions (Phase 7 — Lot & Serial Traceability)

**Status:** PARTIAL | **Impact:** Medium

`lot_id` is present on `inventory_transactions` (model + migration). `serial_id` is absent.

**What to do:**
- Add `serial_id UUID NULL` column to `inventory.inventory_transactions` with FK to `inventory.serial_numbers`
- Add `SerialID *uuid.UUID` to `InventoryTransaction` business model
- Wire through DB model, app model, and filter layers
- Update transaction-creating code (pick, receive, transfer) to pass serial_id when applicable

---

### 5. lot_id on Inventory Items (Phase 4 & 7 — Picking FEFO, Traceability)

**Status:** MISSING (by design) | **Impact:** Medium

A separate `inventory.inventory_lots` junction table was created instead of adding `lot_id` directly to `inventory_items`. The FEFO implementation joins through `serial_numbers -> lot_trackings` to get expiration dates.

**Decision needed:** The current junction table approach is architecturally cleaner for many-to-many relationships. Verify that:
1. FEFO queries perform acceptably with the join path
2. The frontend can resolve lot info for a given inventory item via the junction table
3. Traceability queries ("all stock for lot X") work through `inventory_lots`

If all three hold, this item can be **closed as solved differently**.

---

### 6. Notification Persistence (Phase 12 — Notifications)

**Status:** PARTIAL | **Impact:** High

What exists:
- `workflow.notification_deliveries` — delivery log table (tracks send attempts per channel)
- `GET /v1/workflow/notifications/summary` — aggregates active alerts + pending approvals

What's missing:
- **No `workflow.notifications` inbox table** with `is_read`, `read_at` columns
- **No unread count endpoint** (`GET /v1/workflow/notifications/count?is_read=false`)
- **No mark-as-read endpoints** (`POST /v1/workflow/notifications/{id}/read`, `POST /v1/workflow/notifications/read-all`)
- **No paginated notification list** (`GET /v1/workflow/notifications`)

**What to build:**

```
Table: workflow.notifications
  id UUID PK
  user_id UUID NOT NULL FK → core.users
  title TEXT NOT NULL
  message TEXT
  priority VARCHAR(10) CHECK (IN ('low','medium','high','critical'))
  is_read BOOLEAN DEFAULT false
  source_entity_name VARCHAR(100)
  source_entity_id UUID
  action_url TEXT
  created_date TIMESTAMPTZ DEFAULT NOW()
  read_date TIMESTAMPTZ
```

Endpoints:
- `GET /v1/workflow/notifications` — paginated, filterable by `is_read`
- `GET /v1/workflow/notifications/count?is_read=false` — badge count
- `POST /v1/workflow/notifications/{id}/read`
- `POST /v1/workflow/notifications/read-all`

---

### 7. action_url on Alert Struct (Phase 12 — Notifications)

**Status:** MISSING | **Impact:** Medium

The `Alert` struct (`alertbus/model.go:33-48`) and `workflow.alerts` table have no `action_url` or `deep_link` field.

**What to do:**
- Add `action_url TEXT NULL` to `workflow.alerts` table
- Add `ActionURL *string` to Alert business model
- Wire through DB, app, and API layers
- Update `create_alert` workflow action to accept optional `action_url`

---

## P2 — New Domains

### 8. Pick Tasks Domain (Phase 4 — Picking & Order Fulfillment)

**Status:** MISSING | **Impact:** Critical for Phase 4

No `pick_tasks` table, business domain, or API routes exist. Only `sales/pickingapp` exists (order-level picking logic, not floor task management).

**What to build:**

```
Table: inventory.pick_tasks
  id UUID PK
  sales_order_id UUID FK → sales.orders
  sales_order_line_item_id UUID FK → sales.order_line_items
  product_id UUID NOT NULL FK → products.products
  lot_id UUID NULL FK → inventory.lot_trackings
  serial_id UUID NULL FK → inventory.serial_numbers
  location_id UUID NOT NULL FK → inventory.inventory_locations
  quantity_to_pick NUMERIC NOT NULL
  quantity_picked NUMERIC DEFAULT 0
  assigned_to UUID FK → core.users
  status VARCHAR(20) CHECK (IN ('pending','in_progress','completed','short_picked','cancelled'))
  created_date TIMESTAMPTZ DEFAULT NOW()
  completed_date TIMESTAMPTZ
  completed_by UUID FK → core.users
  short_pick_reason TEXT
```

Endpoints: full CRUD + `POST /v1/inventory/pick-tasks/{id}/complete`

`complete` must atomically:
1. Decrement `inventory_items.quantity` and `allocated_quantity` at the source location
2. Create `inventory_transaction` with `transaction_type = "pick"`
3. Update pick task status
4. Update `sales.order_line_items` fulfillment status

---

### 9. Cycle Count Sessions Domain (Phase 5 — Adjustments & Cycle Counting)

**Status:** MISSING | **Impact:** High for Phase 5

No cycle count tables, domain, or API routes exist.

**What to build:**

```
Table: inventory.cycle_count_sessions
  id UUID PK
  name VARCHAR(200) NOT NULL
  status VARCHAR(20) CHECK (IN ('draft','in_progress','completed','cancelled'))
  created_by UUID FK → core.users
  created_date TIMESTAMPTZ DEFAULT NOW()
  completed_date TIMESTAMPTZ

Table: inventory.cycle_count_items
  id UUID PK
  session_id UUID NOT NULL FK → cycle_count_sessions
  product_id UUID NOT NULL FK → products.products
  location_id UUID NOT NULL FK → inventory.inventory_locations
  system_quantity NUMERIC NOT NULL
  counted_quantity NUMERIC
  variance NUMERIC (computed: counted - system)
  counted_by UUID FK → core.users
  counted_date TIMESTAMPTZ
  status VARCHAR(20) CHECK (IN ('pending','counted','variance_approved','variance_rejected'))
```

Endpoints:
- Full CRUD on sessions and items
- `POST /v1/inventory/cycle-count-sessions/{id}/complete` — locks session, generates adjustment records for approved variances

---

## P3 — Nice to Have (Future Sprints)

These items are not blocking any current frontend phase but are tracked for completeness.

| # | Item | Notes |
|---|------|-------|
| 10 | Photo/attachment infrastructure | No file storage layer exists. Needs S3 integration + `attachments` domain. |
| 11 | Inspection checklist templates | Inspections are flat records with `notes`. Structured checklists need a new table. |
| 12 | Multi-product transfer orders | Transfer orders are 1 product each. Multi-SKU pallets need a `transfer_order_line_items` table. |
| 13 | RecievedDate typo in lot_trackings | Typo persists across all layers (`RecievedDate` instead of `ReceivedDate`). Requires DB migration + model rename. Present in `lottrackingsbus/model.go:20`, `order.go:4`, and DB layer. |
| 14 | Throughput KPI metrics | `GET /v1/inventory/supervisor-kpis` returns counts only. Rate-based metrics (items/hour, picks/shift) not implemented. |

---

## Priority Order for Implementation

**Build next (unblocks frontend phases):**
1. Pick tasks domain (#8) — critical for Phase 4
2. Cycle count sessions (#9) — needed for Phase 5
3. Notification inbox (#6) — needed for Phase 12

**Quick wins (< 1 day each):**
4. action_url on alerts (#7) — single field addition
5. serial_id on transactions (#4) — single field addition
6. Expose reason codes + transfer statuses via GET endpoints (#2, #3)

**Decisions needed (may already be solved):**
7. Add-stock endpoint (#1) — possibly covered by put-away tasks
8. lot_id on inventory_items (#5) — possibly solved by junction table
