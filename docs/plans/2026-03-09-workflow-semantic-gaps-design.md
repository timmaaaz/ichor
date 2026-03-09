# Workflow Semantic Gaps — Design Document

**Date**: 2026-03-09
**Status**: Approved
**Replaces**: `.claude/plans/WORKFLOW_SEMANTIC_GAPS_PLAN/` (superseded — plan files were inconsistent with revised scope)

---

## Problem Statement

The Ichor workflow engine has two categories of gaps that degrade the experience of building ERP workflows:

### Category 1: Missing Action Handlers

Terminal state transitions exist in the business layer and fire delegate events, but have no corresponding workflow action handler. Workflow authors cannot invoke them as steps in a DAG — only react to them via `on_update` triggers.

### Category 2: Undiscoverable Trigger Conditions

When configuring `field_conditions` on `on_update` triggers, users must know exact internal string values for status/enum fields. No API exposes this metadata.

---

## Design Principles

### Handler Invocation Threshold

A dedicated action handler is justified when the bus method provides something generic actions (`transition_status`, `update_field`) structurally cannot:

| Bus method characteristic | Justifies handler? |
|---|---|
| FK UUID status field (user can't know deployment UUID) | Yes |
| Multi-field atomic write (>1 field in one operation) | Yes |
| User identity injection (`execCtx.UserID` as approver) | Yes |
| Audit trail fields (ApprovedBy, ApprovalReason, timestamps) | Yes |
| String status flip only, no user identity needed | No — `transition_status` covers it |

### Symmetric Approve/Reject Pairs

Every approval operation gets both `approve_*` and `reject_*` handlers. Asymmetric coverage leaves rejection paths as dead ends or forces generic nodes that bypass audit trail capture. Rejection reason is required on reject handlers (audit trail demands *why*), optional on approve handlers.

### No Deferral

If a gap is visible, close it now. Deferral relies on remembering. All four approval domains (purchase orders, inventory adjustments, transfer orders, approval requests) are addressed in this plan.

### Bus Methods First

Handlers are built on top of fleshed-out bus methods. Bus methods must include full audit trail fields before handlers are written. Building handlers on thin bus methods means rewriting them when the bus methods grow.

---

## Phase 1 — Putaway Completion (Full Feature)

### Problem

`putawaytaskbus` has no `Complete()` method. There is no completion HTTP endpoint. The 3-way atomic write (task status → `completed`, inventory item quantity increment, inventory transaction record creation) is completely untested and the endpoint does not exist.

### Design

**Business layer**: Add `putawaytaskbus.Complete(ctx, task, completedBy)` implementing the 3-way atomic write in a single DB transaction via `NewWithTx`. Rollback must revert all three writes on any failure.

**App layer**: Add `putawaytaskapp.complete()` handler that calls the bus method and returns the app model.

**API layer**: `PUT /v1/inventory/put-away-tasks/{id}/complete` — no request body required (task ID from path, user from JWT context).

### Tests

- `putawaytaskbus_test.go` — `complete` subtest: verify all 3 writes, verify rollback on inventory update failure
- `putawaytaskapi/complete_test.go` — `apitest.Table`: happy path (200), already-completed (409 or 200 idempotent — decide during impl), not-found (404), wrong status/cancelled (422)
- `putawaytaskapi/seed_test.go` — add task in `in_progress` state with a real inventory item ready for completion

---

## Phase 2 — Approval Handlers (6 handlers across 3 domains)

### Problem

Purchase orders, inventory adjustments, and transfer orders all have approval operations but no workflow action handlers. Some bus methods lack audit trail fields. No reject handlers exist for POs or transfer orders.

### Design Principle: Bus First

Before writing any handler, verify each bus method has:
- `ApprovedBy uuid.UUID` / `RejectedBy uuid.UUID`
- `ApprovalReason string` (optional) / `RejectionReason string` (required)
- Proper timestamps set in the bus method body
- `delegate.Call()` fired after the write

If missing, add to the bus method first, then build the handler.

### Handlers

**`approve_purchase_order`** / **`reject_purchase_order`**
File: `business/sdk/workflow/workflowactions/procurement/`
Config: `purchase_order_id` (required), `reason` (optional/required)
Bus: `purchaseorderbus.Approve()` already exists — verify audit trail fields. Add `purchaseorderbus.Reject()` (new).
Output ports: `approved`/`rejected`, `not_found`, `already_approved`/`already_rejected`, `failure`

**`approve_inventory_adjustment`** / **`reject_inventory_adjustment`**
File: `business/sdk/workflow/workflowactions/inventory/`
Config: `adjustment_id` (required), `reason` (optional/required)
Bus: Verify `inventoryadjustmentbus.Approve()` and `Reject()` have audit trail fields.
Output ports: `approved`/`rejected`, `not_found`, `already_approved`/`already_rejected`, `failure`

**`approve_transfer_order`** / **`reject_transfer_order`**
File: `business/sdk/workflow/workflowactions/inventory/`
Config: `transfer_order_id` (required), `reason` (optional/required)
Bus: Verify `transferorderbus.Approve()` has audit trail fields. Add `transferorderbus.Reject()` (new).
Output ports: `approved`/`rejected`, `not_found`, `already_approved`/`already_rejected`, `failure`

### Registration

Add all 6 to `register.go`. Add any missing bus fields to `BusDependencies`. Wire in `all.go`.

### Tests

Per handler: `unitest.Table` file with **Validate suite** (missing ID, invalid UUID, missing required reason) and **Execute suite** (happy path, not_found, already resolved, failure). Bus unit tests for any new `Reject()` methods added.

---

## Phase 3 — Resolve Approval (Bug Fix + Handler)

### Part A: Delegate Event Bug Fix

`approvalrequestbus.Business.Resolve()` is the only method across all 51 bus domains that changes entity state without calling `delegate.Call()`. Reactive workflows cannot trigger on approval resolution. Audit systems cannot observe it.

**Fix**: Add `delegate.Call(ctx, ActionUpdatedData(updatedRequest))` at the end of `Resolve()`, following the identical pattern used in every other bus `Update()` method.

### Part B: `resolve_approval_request` Handler

Enables cross-workflow orchestration: Workflow B can programmatically resolve an approval request that Workflow A is blocked on (e.g., a manager escalation workflow closes a purchase approval).

**Config**: `approval_request_id` (required), `resolution` (`"approved"` | `"rejected"`, required), `reason` (optional — audit trail)
**Output ports**: `resolved_approved`, `resolved_rejected`, `not_found`, `already_resolved`, `failure`
**IsAsync**: `true` — approval resolution triggers downstream workflow continuation signals; async avoids nested activity deadlock with `seek_approval`.

**Registration**: `RegisterAll()` and `RegisterCoreActions()` (nil-guarded in core).

### Tests

- `approvalrequestbus_test.go` — verify `Resolve()` now fires delegate event (delegate capture pattern)
- `approval/resolve_test.go` — `unitest.Table`: Validate suite (missing ID, invalid resolution value), Execute suite (resolved_approved, resolved_rejected, not_found, already_resolved)

---

## Phase 4 — Entity Field Schema Discovery API

### Problem

Workflow authors configuring `field_conditions` on `on_update` triggers must know exact internal string values (e.g., `"in_progress"`, `"timed_out"`). No API exposes this.

### Design

**Step 1 — Route discovery**: Locate where existing workflow discovery endpoints (`/trigger-types`, `/entities`, `/action-types`) live. The `ruleapi/route.go` file only has rule CRUD routes — the discovery endpoints are elsewhere. Find them before adding the new route.

**Static registry** (Option A, preferred over DB introspection):
`business/sdk/workflow/fieldschema/registry.go` — a Go map from entity name to known enum field schemas. Explicit, testable, immune to DB schema quirks. The set of business-significant enums is small and stable.

**7 enums to register**:

| Entity | Field | Values |
|--------|-------|--------|
| `inventory.put_away_tasks` | `status` | pending, in_progress, completed, cancelled |
| `inventory.inventory_adjustments` | `approval_status` | pending, approved, rejected |
| `inventory.lot_trackings` | `quality_status` | good, on_hold, quarantined, released, expired |
| `workflow.alerts` | `status` | active, acknowledged, dismissed, resolved |
| `workflow.alerts` | `severity` | low, medium, high, critical |
| `workflow.approval_requests` | `status` | pending, approved, rejected, timed_out, expired |
| `workflow.approval_requests` | `approval_type` | any, all, majority |

**Endpoint**: `GET /v1/workflow/entities/{entity}/fields`
Response: `{"entity": "...", "fields": [{"name": "status", "type": "enum", "values": [...], "description": "..."}]}`
- 200 + fields array — entity in registry
- 200 + empty fields — entity exists in DB catalog but no registered enums (not 404)
- 404 — entity not found in DB catalog at all

**Maintenance link**: Add `// NOTE: update business/sdk/workflow/fieldschema/registry.go when adding new values` comment to each status source file.

### Tests

`workflow/ruleapi/fields_test.go` — `apitest.Table`:
1. Known enum entity → 200 with correct values
2. Entity with no registered enums → 200 with empty fields array
3. Nonexistent entity → 404

---

## Summary

| Phase | Deliverable | Net new handlers | Tests |
|-------|-------------|-----------------|-------|
| 1 | Putaway completion: bus method + endpoint | 0 | bus unit + API integration |
| 2 | 6 approval handlers + bus fleshing | 6 | 6× handler unit + bus unit for new Reject() |
| 3 | Delegate event fix + resolve handler | 1 | bus unit (delegate capture) + handler unit |
| 4 | Field schema discovery API | 0 | API integration (3 cases) |

**Total**: 7 new workflow action handlers, 1 new HTTP endpoint, 1 bus bug fix, 1 discovery API.
