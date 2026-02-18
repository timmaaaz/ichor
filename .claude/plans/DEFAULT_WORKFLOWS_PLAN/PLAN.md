# Default Workflows & Duplicate Feature Plan

## Overview

Two related features:
1. **`is_default` flag** — mark seed workflows as system defaults; protect them from modification via `PUT /v1/workflow/rules/{id}/full` (returns 403); allow enable/disable and duplicate
2. **Duplicate endpoint** — `POST /v1/workflow/rules/{id}/duplicate`: deep copy (rule + actions + edges), name becomes `{original}-DUPLICATE`, `is_active=false`, `is_default=false`, owned by requester
3. **Expanded default seed data** — existing 5 workflows marked `is_default=true` + 8 new default workflows across HR, sales, procurement, inventory, assets domains

## Architecture Summary

- Migration Version **1.996**: adds `is_default BOOLEAN NOT NULL DEFAULT FALSE` to `workflow.automation_rules`; updates `automation_rules_view` to expose it
- `is_default` is **immutable after creation** (not in `UpdateAutomationRule`, not in `UpdateRule` SQL SET)
- `PATCH /active` (enable/disable) still works on default workflows
- `DuplicateWorkflow` lives in `workflowsaveapp.App` (already owns db + transaction logic)
- Duplicate endpoint at `POST /v1/workflow/rules/{id}/duplicate` in `workflowsaveapi`

## Phases

### Phase 1 — Migration + Data Layer
**Scope**: Database + Go model/store changes only. No business logic changes yet.

Files:
- `business/sdk/migrate/sql/migrate.sql` — add migration 1.996
- `business/sdk/workflow/models.go` — add `IsDefault bool` to `AutomationRule`, `NewAutomationRule`, `AutomationRuleView`
- `business/sdk/workflow/stores/workflowdb/models.go` — add `IsDefault` to `automationRule` and `automationRulesView` DB structs; update `toCoreAutomationRule`, `toDBAutomationRule`, `toCoreAutomationRuleView`
- `business/sdk/workflow/stores/workflowdb/workflowdb.go` — add `is_default` to SELECT in `QueryRuleByID`, `QueryRulesByEntity`, `QueryActiveRules`; add to INSERT in `CreateRule`; NOT in `UpdateRule` SET

Verify: `go build ./...` passes. `go test ./...` passes.

---

### Phase 2 — Business + App Layer (is_default enforcement)
**Scope**: Business layer wires `IsDefault` through `CreateRule`; app layer adds default-workflow protection to `SaveWorkflow`.

Files:
- `business/sdk/workflow/workflowbus.go` — `CreateRule`: copy `IsDefault` from `NewAutomationRule` → `AutomationRule`; add `ErrDefaultWorkflow` sentinel error
- `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` — in `updateRule()`: after fetching the rule, if `rule.IsDefault` → return `errs.Newf(errs.PermissionDenied, "cannot modify default workflow: use POST /v1/workflow/rules/{id}/duplicate to create an editable copy")`

Verify: `go build ./...` passes. `go test ./...` passes.

---

### Phase 3 — Duplicate Endpoint
**Scope**: `DuplicateWorkflow` app method + HTTP handler + route.

Files:
- `app/domain/workflow/workflowsaveapp/workflowsaveapp.go` — add `DuplicateWorkflow(ctx, ruleID, userID uuid.UUID) (SaveWorkflowResponse, error)`:
  1. `QueryRuleByID` → source rule
  2. `QueryActionsByRule` → active actions only
  3. `QueryEdgesByRuleID` → edges
  4. BEGIN TRANSACTION
  5. `CreateRule` (name=`orig+"-DUPLICATE"`, is_active=false, is_default=false, created_by=userID, copies entity/trigger/conditions/canvasLayout)
  6. `CreateRuleAction` for each active action → build `oldID→newID` map
  7. `CreateActionEdge` for each edge, remapping source/target via `oldID→newID` map; skip edges whose source or target action was inactive
  8. COMMIT
  9. Fire `ActionRuleCreated` delegate event
  10. Return `buildResponse(...)`
- `api/domain/http/workflow/workflowsaveapi/workflowsaveapi.go` — add `duplicate(ctx, r)` handler
- `api/domain/http/workflow/workflowsaveapi/route.go` — add `POST /v1/workflow/rules/{id}/duplicate` with Create permission

Verify: `go build ./...` passes. `go test ./...` passes.

---

### Phase 4 — API Response + Seed Data
**Scope**: Expose `is_default` in API response models; update seedFrontend with defaults.

Files:
- `api/domain/http/workflow/ruleapi/model.go` — add `IsDefault bool \`json:"is_default"\`` to `RuleResponse`
- `api/domain/http/workflow/ruleapi/ruleapi.go` — update `toRuleResponse()` to include `IsDefault` from `AutomationRuleView`
- `business/sdk/dbtest/seedFrontend.go`:
  - Add `IsDefault: true` to all 5 existing workflow rules
  - Register 8 new workflow entities (inventory_items, sales_orders, users, suppliers, supplier_products, user_assets)
  - Create 8 new default workflows (see table below), each with `IsDefault: true`, `IsActive: false`

New default workflows:

| Domain | Name | Entity | Trigger | Actions |
|--------|------|--------|---------|---------|
| Inventory | "Low Stock Alert Pipeline" | `inventory_items` | on_update | check_reorder_point → create_alert |
| Inventory | "Item Created - Log Audit Entry" | `inventory_items` | on_create | log_audit_entry |
| Sales | "Order Created - Confirmation Alert" | `sales_orders` | on_create | create_alert |
| Sales | "Order Updated - Notify Operations" | `sales_orders` | on_update | create_alert |
| HR | "New User Created - Welcome Alert" | `users` | on_create | create_alert |
| Procurement | "Supplier Added - Alert Team" | `suppliers` | on_create | create_alert |
| Procurement | "Supplier Product Added - Log Audit" | `supplier_products` | on_create | log_audit_entry |
| Assets | "Asset Assigned - Send Notification" | `user_assets` | on_create | send_notification |

Verify: `go build ./...` passes. `go test ./...` passes.

---

## Key Constraints

- `is_default` is **never** updated via `UpdateRule` or `UpdateAutomationRule` — treat it as immutable system metadata
- `PATCH /active` (toggle enable/disable) **does work** on default workflows
- Duplicate always sets `is_default=false` and `is_active=false`
- Duplicate endpoint requires **Create** permission (same as `POST /v1/workflow/rules/full`)
- All new seed entities use the `"table"` entity type
- Follow existing seedFrontend.go error handling pattern: log errors but don't fail the whole seed

## Migration Note

Before writing migration 1.996, read the existing `automation_rules_view` definition in migrate.sql to reproduce it exactly with `is_default` added to the SELECT. Use `CREATE OR REPLACE VIEW`.
