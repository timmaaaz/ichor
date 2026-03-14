# Workflow Semantic Gaps Plan

## Problem Statement

The Ichor workflow engine has two distinct categories of semantic leakiness that degrade the user experience when building ERP workflows:

### Category 1: Missing Action Handlers (Composability Gaps)

Terminal state transitions exist in the business layer and fire delegate events, but have **no corresponding workflow action handler**. This means users can *react to* these operations via `on_update` triggers, but cannot *invoke them as steps* in a workflow DAG.

Current coverage: **0%** of terminal state transitions have workflow action handlers.

Affected operations (by priority):
1. `complete_putaway` — completes a putaway task (3-way atomic: task + inventory item + transaction)
2. `approve_inventory_adjustment` — approves a pending inventory correction
3. `reject_inventory_adjustment` — rejects a pending inventory correction
4. `approve_transfer_order` — approves an inter-location inventory transfer
5. `approve_purchase_order` — approves a purchase order before goods receipt
6. `resolve_approval_request` — closes an open approval request programmatically (currently no delegate event either)

### Category 2: Undiscoverable Trigger Conditions (Trigger Leakiness)

When a user configures an `on_update` trigger with `field_conditions`, they must know the **exact internal string values** for status/enum fields. There is no API that exposes field metadata or enum values for a given entity. This is the "semantic leak" that the putaway debate surfaced.

Affected hardcoded enums (7 total):
| Entity | Field | Values |
|--------|-------|--------|
| `inventory.put_away_tasks` | `status` | pending, in_progress, completed, cancelled |
| `inventory.inventory_adjustments` | `approval_status` | pending, approved, rejected |
| `inventory.lot_trackings` | `quality_status` | good, on_hold, quarantined, released, expired |
| `workflow.alerts` | `status` | active, acknowledged, dismissed, resolved |
| `workflow.alerts` | `severity` | low, medium, high, critical |
| `workflow.approval_requests` | `status` | pending, approved, rejected, timed_out, expired |
| `workflow.approval_requests` | `approval_type` | any, all, majority |

## Design Principles Applied (revised after critical review)

### Invocation Threshold

A dedicated action handler is justified when the bus method provides something generic actions
(`transition_status`, `update_field`) **structurally cannot**:

| What the bus adds | Justifies handler? |
|---|---|
| FK UUID abstraction (status is a lookup table) | Yes |
| Multi-record atomic write (>1 table) | Yes |
| Physical act confirmation (putaway completion) | No — wrong agent, delegate-reactive is correct |
| Delegate event only, string status field | Marginal — defer until concrete need |

### complete_putaway is a category error

Completing a putaway task reports that a physical act occurred. A workflow engine invoking this
fabricates the confirmation. The correct pattern: worker confirms placement → delegate fires →
workflow reacts. This plan originally included a complete_putaway handler; it has been removed.

### Correctness risk of thin wrappers

Adding a handler for an operation `transition_status` can already cover creates two paths to
the same state change. Authors using `transition_status` directly will silently miss future
changes to the bus method's guards. Only add a dedicated handler when the bus method does
something the generic path structurally cannot.

## Phases (revised)

| Phase | Name | Category | Priority |
|-------|------|----------|----------|
| 1 | Putaway completion integration tests (no handler) | Test gap | Urgent (untested 3-way txn) |
| 2 | `approve_purchase_order` handler | Action gap | High (FK abstraction, 4-field write) |
| 3 | Delegate event fix for `approvalrequestbus.Resolve()` | Bug fix | High (only bus that fires no event) |
| 4 | Entity field schema discovery API | Trigger gap | Medium (removes enum leakiness) |

## Key Files

### Existing Action Handler Pattern
- `business/sdk/workflow/workflowactions/inventory/receive.go` — canonical reference
- `business/sdk/workflow/workflowactions/register.go` — registration and BusDependencies struct
- `business/sdk/workflow/workflowactions/procurement/createpo.go` — procurement pattern

### Business Layer (targets)
- `business/domain/inventory/putawaytaskbus/putawaytaskbus.go` — Complete() method
- `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go` — Approve(), Reject()
- `business/domain/inventory/transferorderbus/transferorderbus.go` — Approve()
- `business/domain/procurement/purchaseorderbus/purchaseorderbus.go` — Approve()
- `business/domain/workflow/approvalrequestbus/approvalrequestbus.go` — Resolve() (missing delegate event)

### Discovery API Target
- `api/domain/http/workflow/ruleapi/` — existing trigger-types, entities, action-types endpoints
- Add: entity field schema endpoint with enum value metadata

## Status

See `PROGRESS.yaml` for current phase status.
