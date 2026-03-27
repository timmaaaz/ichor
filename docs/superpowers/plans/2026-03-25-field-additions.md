# Field Additions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `serial_id` to inventory transactions, `action_url` to alerts, and fix the `RecievedDate` typo across lot trackings.

**Architecture:** Three independent field-level changes. Each touches a different domain (inventory transactions, workflow alerts, inventory lot trackings). No cross-dependencies. Each task produces a working, buildable codebase.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, Ardan Labs Service architecture (bus/db/app/api layers)

---

## Task 1: Fix RecievedDate Typo in Lot Trackings

**Files:**
- Modify: `business/domain/inventory/lottrackingsbus/model.go`
- Modify: `business/domain/inventory/lottrackingsbus/filter.go`
- Modify: `business/domain/inventory/lottrackingsbus/order.go`
- Modify: `business/domain/inventory/lottrackingsbus/lottrackingsbus.go`
- Modify: `business/domain/inventory/lottrackingsbus/testutil.go`
- Modify: `business/domain/inventory/lottrackingsbus/lottrackingsbus_test.go`
- Modify: `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/model.go`
- Modify: `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/filter.go`
- Modify: `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/order.go`
- Modify: `app/domain/inventory/lottrackingsapp/model.go`
- Modify: `app/domain/inventory/lottrackingsapp/filter.go`
- Modify: `app/domain/inventory/lottrackingsapp/order.go`
- Modify: `api/domain/http/inventory/lottrackingsapi/filter.go`
- Modify: `business/domain/inventory/inventoryitembus/inventoryitembus_test.go`
- Modify: `api/cmd/services/ichor/tests/inventory/lottrackingsapi/create_test.go`
- Modify: `api/cmd/services/ichor/tests/inventory/lottrackingsapi/update_test.go`

**Important notes:**
- The DB column is already `received_date` (correct). No migration needed.
- The app-layer JSON tags are already `received_date` (correct). No API response change.
- The bus-layer JSON tags have the typo (`recieved_date`) — fix those too. These are only used for internal workflow event serialization.
- This is a pure Go identifier rename: `RecievedDate` → `ReceivedDate` and `OrderByRecievedDate` → `OrderByReceivedDate`.

- [ ] **Step 1: Rename in bus model**

In `business/domain/inventory/lottrackingsbus/model.go`, use find-and-replace across the file:
- `RecievedDate` → `ReceivedDate` (all struct fields in `LotTrackings`, `NewLotTrackings`, `UpdateLotTrackings`)
- `recieved_date` → `received_date` (JSON tags on bus model)

- [ ] **Step 2: Rename in bus filter, order, business logic**

In `business/domain/inventory/lottrackingsbus/filter.go`:
- `RecievedDate` → `ReceivedDate`

In `business/domain/inventory/lottrackingsbus/order.go`:
- `OrderByRecievedDate` → `OrderByReceivedDate`

In `business/domain/inventory/lottrackingsbus/lottrackingsbus.go`:
- All references to `RecievedDate` → `ReceivedDate` (Create and Update methods)

- [ ] **Step 3: Rename in DB layer**

In `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/model.go`:
- `RecievedDate` → `ReceivedDate` (struct field + both mapping functions)
- The `db:"received_date"` tag is already correct — do NOT change it.

In `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/filter.go`:
- `filter.RecievedDate` → `filter.ReceivedDate`

In `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/order.go`:
- `lottrackingsbus.OrderByRecievedDate` → `lottrackingsbus.OrderByReceivedDate`

- [ ] **Step 4: Rename in app layer**

In `app/domain/inventory/lottrackingsapp/model.go`:
- All `RecievedDate` → `ReceivedDate` (struct fields in `LotTrackings`, `NewLotTrackings`, `UpdateLotTrackings`, `QueryParams`, and all conversion functions)
- The `json:"received_date"` tags are already correct — do NOT change them.

In `app/domain/inventory/lottrackingsapp/filter.go`:
- `RecievedDate` → `ReceivedDate`

In `app/domain/inventory/lottrackingsapp/order.go`:
- `lottrackingsbus.OrderByRecievedDate` → `lottrackingsbus.OrderByReceivedDate`

- [ ] **Step 5: Rename in API layer**

In `api/domain/http/inventory/lottrackingsapi/filter.go`:
- `RecievedDate:` → `ReceivedDate:`

- [ ] **Step 6: Rename in test files**

In `business/domain/inventory/lottrackingsbus/testutil.go`:
- `RecievedDate` → `ReceivedDate`

In `business/domain/inventory/lottrackingsbus/lottrackingsbus_test.go`:
- All `RecievedDate` → `ReceivedDate` (4 occurrences)

In `business/domain/inventory/inventoryitembus/inventoryitembus_test.go`:
- `RecievedDate` → `ReceivedDate` (1 occurrence)

In `api/cmd/services/ichor/tests/inventory/lottrackingsapi/create_test.go`:
- All `RecievedDate` → `ReceivedDate` (12 occurrences)

In `api/cmd/services/ichor/tests/inventory/lottrackingsapi/update_test.go`:
- All `RecievedDate` → `ReceivedDate` (2 occurrences)

- [ ] **Step 7: Build and verify**

Run: `go build ./business/domain/inventory/lottrackingsbus/... ./app/domain/inventory/lottrackingsapp/... ./api/domain/http/inventory/lottrackingsapi/... ./business/domain/inventory/inventoryitembus/... ./api/cmd/services/ichor/tests/inventory/lottrackingsapi/...`
Expected: Clean build, no errors.

- [ ] **Step 8: Commit**

```bash
git add business/domain/inventory/lottrackingsbus/ app/domain/inventory/lottrackingsapp/ api/domain/http/inventory/lottrackingsapi/ business/domain/inventory/inventoryitembus/inventoryitembus_test.go api/cmd/services/ichor/tests/inventory/lottrackingsapi/
git commit -m "fix: rename RecievedDate → ReceivedDate across lot trackings domain

Fixes typo in Go identifiers across all layers (bus/db/app/api/tests).
DB column was already 'received_date'. App JSON tags were already correct.
Only bus-layer JSON tags and Go field names had the misspelling."
```

---

## Task 2: Add serial_id to Inventory Transactions

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql` (add version 2.18)
- Modify: `business/domain/inventory/inventorytransactionbus/model.go`
- Modify: `business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb/model.go`
- Modify: `business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb/inventorytransactiondb.go`
- Modify: `app/domain/inventory/inventorytransactionapp/model.go`

**Pattern reference:** Follow the exact same pattern as `lot_id` which is already wired through all these files.

- [ ] **Step 1: Add migration**

Append to `business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: 2.18
-- Description: Add serial_id column to inventory_transactions for serial number traceability.
ALTER TABLE inventory.inventory_transactions ADD COLUMN serial_id UUID NULL REFERENCES inventory.serial_numbers(id);
```

- [ ] **Step 2: Add SerialID to bus model**

In `business/domain/inventory/inventorytransactionbus/model.go`:

Add after the `LotID` field in `InventoryTransaction`:
```go
SerialID *uuid.UUID `json:"serial_id,omitempty"`
```

Add after the `LotID` field in `NewInventoryTransaction`:
```go
SerialID *uuid.UUID `json:"serial_id,omitempty"`
```

`UpdateInventoryTransaction` — do NOT add SerialID (matches lot_id pattern: not updatable).

- [ ] **Step 3: Add SerialID to DB model**

In `business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb/model.go`:

Add after the `LotID` field in the `inventoryTransaction` struct:
```go
SerialID uuid.NullUUID `db:"serial_id"`
```

In `toBusInventoryTransaction`, add after the lot_id block:
```go
var serialID *uuid.UUID
if db.SerialID.Valid {
    id := db.SerialID.UUID
    serialID = &id
}
```
And add `SerialID: serialID,` to the return struct.

In `toDBInventoryTransaction`, add after the lot_id block:
```go
var serialID uuid.NullUUID
if bus.SerialID != nil {
    serialID = uuid.NullUUID{UUID: *bus.SerialID, Valid: true}
}
```
And add `SerialID: serialID,` to the return struct.

- [ ] **Step 4: Add serial_id to SQL queries**

In `business/domain/inventory/inventorytransactionbus/stores/inventorytransactiondb/inventorytransactiondb.go`:

**Create** — add `serial_id` to column list and `:serial_id` to values:
```sql
INSERT INTO inventory.inventory_transactions (
    id, product_id, location_id, user_id, lot_id, serial_id, transaction_type, reference_number,
    quantity, transaction_date, created_date, updated_date
) VALUES (
    :id, :product_id, :location_id, :user_id, :lot_id, :serial_id, :transaction_type, :reference_number,
    :quantity, :transaction_date, :created_date, :updated_date
)
```

**Update** — add `serial_id = :serial_id` to SET clause.

**Query** and **QueryByID** — add `serial_id` to SELECT column lists:
```sql
id, product_id, location_id, user_id, lot_id, serial_id, transaction_type, reference_number,
quantity, transaction_date, created_date, updated_date
```

- [ ] **Step 5: Add SerialID to app model**

In `app/domain/inventory/inventorytransactionapp/model.go`:

Add after the `LotID` field in `InventoryTransaction`:
```go
SerialID string `json:"serial_id"`
```

In `ToAppInventoryTransaction`, add after the lot_id block:
```go
serialID := ""
if bus.SerialID != nil {
    serialID = bus.SerialID.String()
}
```
And add `SerialID: serialID,` to the return struct.

`NewInventoryTransaction` and `UpdateInventoryTransaction` — do NOT add SerialID (matches lot_id pattern).

- [ ] **Step 6: Build and verify**

Run: `go build ./business/domain/inventory/inventorytransactionbus/... ./app/domain/inventory/inventorytransactionapp/... ./api/cmd/...`
Expected: Clean build, no errors.

- [ ] **Step 7: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql business/domain/inventory/inventorytransactionbus/ app/domain/inventory/inventorytransactionapp/
git commit -m "feat(inventory): add serial_id to inventory transactions

Adds nullable serial_number FK to inventory_transactions table for
serial-level traceability. Wired through bus, db, and app layers
following the existing lot_id pattern."
```

---

## Task 3: Add action_url to Alerts

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql` (add version 2.19)
- Modify: `business/domain/workflow/alertbus/model.go`
- Modify: `business/domain/workflow/alertbus/stores/alertdb/model.go`
- Modify: `business/domain/workflow/alertbus/stores/alertdb/alertdb.go` (INSERT + all SELECT queries)
- Modify: `api/domain/http/workflow/alertapi/model.go`
- Modify: `business/sdk/workflow/workflowactions/communication/alert.go`

**Note:** The alert domain has no `alertapp` package. The API layer (`alertapi`) contains the app-layer models directly. This is different from most other domains.

- [ ] **Step 1: Add migration**

Append to `business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: 2.19
-- Description: Add action_url column to alerts for deep-linking from notifications.
ALTER TABLE workflow.alerts ADD COLUMN action_url TEXT NULL;
```

- [ ] **Step 2: Add ActionURL to bus model**

In `business/domain/workflow/alertbus/model.go`, add after `SourceRuleName` in the `Alert` struct:
```go
ActionURL string `json:"action_url,omitempty"`
```

- [ ] **Step 3: Add ActionURL to DB model**

In `business/domain/workflow/alertbus/stores/alertdb/model.go`:

Add after `SourceRuleName` in `dbAlert`:
```go
ActionURL sql.NullString `db:"action_url"`
```

In `toDBAlert`, add after the `SourceRuleID` block:
```go
if a.ActionURL != "" {
    db.ActionURL = sql.NullString{String: a.ActionURL, Valid: true}
}
```

In `toBusAlert`, add after the `SourceRuleName` block:
```go
if db.ActionURL.Valid {
    a.ActionURL = db.ActionURL.String
}
```

- [ ] **Step 4: Add action_url to SQL queries**

In `business/domain/workflow/alertbus/stores/alertdb/alertdb.go`:

**Create INSERT** (line ~74): Add `action_url` to column list and `:action_url` to values:
```sql
INSERT INTO workflow.alerts (
    id, alert_type, severity, title, message, context,
    source_entity_name, source_entity_id, source_rule_id,
    action_url, status, expires_date, created_date, updated_date
) VALUES (
    :id, :alert_type, :severity, :title, :message, :context,
    :source_entity_name, :source_entity_id, :source_rule_id,
    :action_url, :status, :expires_date, :created_date, :updated_date
)
```

**All SELECT queries** that select from `workflow.alerts a` — add `a.action_url` to the column list. There are multiple SELECT statements in this file (QueryByID, Query, QueryByUserID, QueryAlertsByUserID, etc.). Each one that selects the alert columns needs `a.action_url` added after `a.source_rule_id`. Use grep to find them all:
```bash
grep -n "a.source_rule_id" business/domain/workflow/alertbus/stores/alertdb/alertdb.go
```
Add `a.action_url,` on the line after `a.source_rule_id,` in each matching SELECT.

**Important:** Some queries use `rule.name AS source_rule_name` — place `a.action_url` BEFORE that line.

- [ ] **Step 5: Add ActionURL to API model**

In `api/domain/http/workflow/alertapi/model.go`:

Add after `SourceRuleName` in `Alert`:
```go
ActionURL string `json:"actionUrl,omitempty"`
```

In `toAppAlert`, add after the `SourceRuleName` block:
```go
if bus.ActionURL != "" {
    app.ActionURL = bus.ActionURL
}
```

- [ ] **Step 6: Add ActionURL to create_alert handler**

In `business/sdk/workflow/workflowactions/communication/alert.go`:

Add to `AlertConfig`:
```go
ActionURL string `json:"action_url"`
```

In `Execute`, add after `alert.SourceRuleID = sourceRuleID`:
```go
// Set action URL with template variable substitution
if cfg.ActionURL != "" {
    alert.ActionURL = resolveTemplateVars(cfg.ActionURL, execCtx.RawData)
}
```

In `publishAlertToWebSocket`, add after the `sourceEntityId` block:
```go
if alert.ActionURL != "" {
    alertData["actionUrl"] = alert.ActionURL
}
```

- [ ] **Step 7: Build and verify**

Run: `go build ./business/domain/workflow/alertbus/... ./api/domain/http/workflow/alertapi/... ./business/sdk/workflow/workflowactions/... ./api/cmd/...`
Expected: Clean build, no errors.

- [ ] **Step 8: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql business/domain/workflow/alertbus/ api/domain/http/workflow/alertapi/model.go business/sdk/workflow/workflowactions/communication/alert.go
git commit -m "feat(workflow): add action_url to alerts for deep-linking

Adds nullable action_url field to alerts table and wires through all
layers. The create_alert workflow action handler accepts action_url
in config with template variable substitution. WebSocket delivery
includes actionUrl for frontend navigation."
```

---

## Final Verification

- [ ] **Step 1: Full service build**

Run: `go build ./api/cmd/...`
Expected: Clean build.

- [ ] **Step 2: Verify migration order**

Run: `grep "^-- Version:" business/sdk/migrate/sql/migrate.sql | tail -5`
Expected: Versions 2.15, 2.16, 2.17, 2.18, 2.19 in order.
