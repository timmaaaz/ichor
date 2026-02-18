# Phase 1: Add Missing Tables to Whitelist

**Category**: Backend
**Status**: Pending
**Dependencies**: None
**Effort**: Low (~10 lines added to one file)

---

## Overview

The data action handlers (`update_field`, `create_entity`, `lookup_entity`, `transition_status`) all validate target table names against a whitelist in `business/sdk/workflow/workflowactions/data/tables.go`. This prevents workflow actions from targeting arbitrary or sensitive tables.

Currently the whitelist is missing entire schemas that are registered for workflow triggers. Any workflow that triggers on a procurement, HR, or assets entity can currently receive the event — but then cannot take any data action against those same entities. The handlers will reject the target table with a validation error.

**This is a pure addition** — no existing entries are modified, no interfaces change.

---

## Goals

1. Add all `procurement.*` tables that are registered for workflow events
2. Add missing `hr.*` tables (only `hr.offices` is currently whitelisted)
3. Add missing `assets.*` tables (only `assets.assets` and `assets.valid_assets` are currently whitelisted)

---

## Task Breakdown

### Task 1: Add Procurement Tables

**File**: `business/sdk/workflow/workflowactions/data/tables.go`

Add the following entries to `validTables`:

```go
// procurement schema
"procurement.purchase_orders":                    true,
"procurement.purchase_order_line_items":          true,
"procurement.purchase_order_statuses":            true,
"procurement.purchase_order_line_item_statuses":  true,
```

**Verify table names match schema**: Check `business/sdk/migrate/sql/migrate.sql` for exact table names before adding. The bus packages are named `purchaseorderbus`, `purchaseorderstatusbus`, etc. but the DB table names may differ.

Current whitelist already has:
```go
"procurement.suppliers":         true,
"procurement.supplier_products": true,
```

### Task 2: Add HR Tables

**File**: `business/sdk/workflow/workflowactions/data/tables.go`

Add the following entries:

```go
// hr schema (currently only hr.offices is whitelisted)
"hr.titles":                   true,
"hr.reports_to":               true,
"hr.homes":                    true,
"hr.user_approval_statuses":   true,
"hr.user_approval_comments":   true,
```

**Verify**: Check migrate.sql for exact table names. The `hr.user_approval_statuses` table name is confirmed in `business/domain/hr/userapprovalstatusbus/`.

### Task 3: Add Assets Tables

**File**: `business/sdk/workflow/workflowactions/data/tables.go`

Add the following entries:

```go
// assets schema (currently only assets.assets and assets.valid_assets are whitelisted)
"assets.user_assets":       true,
"assets.asset_conditions":  true,
"assets.asset_types":       true,
"assets.asset_tags":        true,
```

---

## Implementation

Open `business/sdk/workflow/workflowactions/data/tables.go`. The file currently ends at line ~62 (the `workflow` schema section). Add the new entries in the correct schema section for each.

**Final result** should look like:

```go
// procurement schema
"procurement.suppliers":                          true,
"procurement.supplier_products":                  true,
"procurement.purchase_orders":                    true,
"procurement.purchase_order_line_items":          true,
"procurement.purchase_order_statuses":            true,
"procurement.purchase_order_line_item_statuses":  true,
// hr schema
"hr.offices":                true,
"hr.titles":                 true,
"hr.reports_to":             true,
"hr.homes":                  true,
"hr.user_approval_statuses": true,
"hr.user_approval_comments": true,
// assets schema
"assets.assets":             true,
"assets.valid_assets":       true,
"assets.user_assets":        true,
"assets.asset_conditions":   true,
"assets.asset_types":        true,
"assets.asset_tags":         true,
```

---

## Validation

```bash
# Compilation
go build ./business/sdk/workflow/workflowactions/...

# Confirm new entries are present
grep -c "true" business/sdk/workflow/workflowactions/data/tables.go
# Should be higher than the original 39

# Check no typos vs actual migrate.sql table names
grep -E "^CREATE TABLE (procurement|hr|assets)\." business/sdk/migrate/sql/migrate.sql
```

---

## Gotchas

- **Table names must exactly match the SQL schema** — the whitelist is checked with a string equality map lookup. A typo like `purchase_order` instead of `purchase_orders` will silently fail at runtime.
- **Do not add `workflow.allocation_results`** — that table is accessed directly by inventory handlers, not through the generic data action path. Adding it here would expose it to raw `update_field` mutations which could corrupt idempotency tracking.
- **Do not add `core.permissions` or auth tables** — the whitelist exists to prevent this. Keep it targeted to ERP domain tables.
