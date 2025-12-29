# Default Status Management Implementation Plan

**Project**: Ichor ERP System
**Goal**: Implement default status management using form configuration + workflow engine for automatic status assignment and transitions
**Timeline**: 3 phases
**Status**: Planning Complete - Ready for Phase Execution

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Pattern](#architecture-pattern)
3. [Current State](#current-state)
4. [Phase Overview](#phase-overview)
5. [Technology Stack](#technology-stack)
6. [Success Criteria](#success-criteria)
7. [Quick Reference Guide](#quick-reference-guide)

---

## Executive Summary

### The Goal

Enable automatic default status assignment for orders and line items, eliminating manual status selection while maintaining configurable behavior. Status transitions (Pending → Allocated) are handled by the workflow engine based on business events.

### The Approach

**Use a layered approach**:
1. **Form-level configuration** sets initial default statuses (synchronous, deterministic)
2. **Workflow engine** handles status transitions based on business logic (async, conditional)

This separates concerns cleanly:
- Form layer: Deterministic defaults, user input, FK name-to-UUID resolution
- Workflow layer: Business logic, conditional transitions, async operations

### Why This Matters

- **User Experience**: No more manually selecting "Pending" on every order
- **Data Consistency**: Ensures all orders start with correct initial status
- **Automation**: Status transitions happen automatically based on business events
- **Maintainability**: Defaults configured by name, not hardcoded UUIDs

---

## Architecture Pattern

```
Order Creation via FormData
    │
    ├─► Form Config Default Resolution:
    │   - fulfillment_status_id default_value: "Pending" → resolved to UUID
    │   - line_item_fulfillment_statuses_id default_value: "Pending" → resolved to UUID
    │   - created_by = {{$me}}, created_date = {{$now}} (template vars unchanged)
    │
    ├─► Database: Order + Line Items committed with resolved default status UUIDs
    │
    └─► Workflow Engine (async, post-commit)
        │
        ├─► Rule: "On Order Create" → allocate_inventory action
        │
        ├─► Allocation Success → update_field: Line items → ALLOCATED
        │
        └─► Allocation Failure → create_alert: Notify role/user
```

### Key Design Decision: Form Configuration Default Value Resolution

**Approach: Form Config Handles FK Default Resolution**
- Default values in form configuration use human-readable names (e.g., `"Pending"`)
- The formdata package resolves names to UUIDs during form processing
- No changes needed to template variable system (`{{$me}}`, `{{$now}}` stay separate)

**Why Form Config Resolution:**
- Keeps form configurations readable and maintainable
- Different environments may have different status UUIDs (dev vs prod)
- Status names are stable; UUIDs are not
- Admin can configure defaults without knowing UUIDs
- Resolution happens at formdata processing time

**Form Config Default Value Pattern:**
```json
{
  "name": "fulfillment_status_id",
  "type": "smart-combobox",
  "entity": "sales.order_fulfillment_statuses",
  "default_value": "Pending",
  "default_mode": "create"
}
```

The formdata package:
1. Sees `default_value: "Pending"` for an FK field with `entity: "sales.order_fulfillment_statuses"`
2. Queries the entity table for a record where `name = "Pending"` (or configurable lookup field)
3. Returns the UUID of that record as the resolved default value

---

## Current State

### Template Magic Values (`{{$me}}`, `{{$now}}`)
- Location: `business/sdk/workflow/template.go`
- Synchronous, resolved during formdata processing
- Used for: created_by, updated_by, created_date, updated_date
- Cannot do: database lookups, FK resolution, conditional logic

### Workflow Engine
- Location: `business/sdk/workflow/`
- Async, event-driven, triggers after commit
- Actions: `allocate_inventory`, `update_field`, `create_alert`, `seek_approval`
- Can do: FK resolution, conditional transitions, complex business logic

### Current Gap
- Status fields in `forms.go:284-292` and `tableforms.go:888` are `Required: true` but have no default
- Users must manually select "PENDING" every time

---

## Phase Overview

| Phase | Name | Category | Description | Key Deliverables |
|-------|------|----------|-------------|------------------|
| 1 | Form Configuration FK Default Resolution | backend | Enable form fields to specify default values by name for FK fields | FK resolution in formdataapp, form seed updates |
| 2 | Workflow Integration for Status Transitions | backend | Wire automation rules for allocation and status updates | Automation rules, workflow event firing |
| 3 | Alert System Enhancement | fullstack | Extend alert action with role-based recipients | Alert tables, routing, UI |

---

## Technology Stack

### Backend
- **Language/Runtime**: Go 1.23
- **Framework**: Ardan Labs Service Architecture
- **Database**: PostgreSQL 16.4

### Key Components
- **Form Processing**: `app/domain/formdata/formdataapp/`
- **Workflow Engine**: `business/sdk/workflow/`
- **Status Entities**: `business/domain/sales/*fulfillmentstatusbus/`

---

## Success Criteria

### Functional Requirements
- Orders/line items get Pending status by default (via form config resolution)
- Workflow triggers allocation on order create
- Status transitions to ALLOCATED on allocation success
- Alerts created on allocation failure
- Integration tests validate the complete flow

### Non-Functional Requirements
- Form configs use human-readable names, not hardcoded UUIDs
- Default resolution is environment-agnostic
- Invalid status names produce clear validation errors

### Quality Metrics
- **Code Quality**: All changes pass `make lint`
- **Testing**: Integration tests cover the full workflow
- **Performance**: No noticeable impact on order creation time

---

## Quick Reference Guide

### Commands

Execute phases:
```bash
/default-statuses-status      # View current progress
/default-statuses-next        # Execute next pending phase
/default-statuses-phase N     # Jump to specific phase
/default-statuses-validate    # Run validation checks
```

Code review:
```bash
/default-statuses-review N    # Manual code review for phase N
```

Planning:
```bash
/default-statuses-build-phase # Generate next phase documentation
/default-statuses-summary     # Generate executive summary
/default-statuses-dependencies # Show cross-plan dependencies
```

### PROGRESS.yaml

Track progress in [PROGRESS.yaml](./PROGRESS.yaml):
- Real-time phase and task status
- Context (current focus, next task)
- Blockers and decisions
- Deliverables tracking

---

## Phase Documentation

1. [Phase 1: Form Configuration FK Default Resolution](./phases/PHASE_1_FORM_FK_DEFAULTS.md)
2. [Phase 2: Workflow Integration for Status Transitions](./phases/PHASE_2_WORKFLOW_INTEGRATION.md)
3. [Phase 3: Alert System Enhancement](./phases/PHASE_3_ALERT_SYSTEM.md)

---

## Technical Notes

### Already Working
- `allocate.go` queues to RabbitMQ async - order creation doesn't block
- `updatefield.go:254-325` handles FK resolution via `ForeignKeyConfig`
- Template variables like `{{entity_id}}` resolve from `ActionExecutionContext`
- Table whitelist includes `order_fulfillment_statuses`, `line_item_fulfillment_statuses`

### File References

**Core Files**
- `business/sdk/workflow/template.go` - Magic value processing
- `app/domain/formdata/formdataapp/formdataapp.go` - Form data processing
- `business/sdk/workflow/workflowactions/inventory/allocate.go` - Allocation action
- `business/sdk/workflow/workflowactions/data/updatefield.go` - Field update action

**Form Definitions**
- `business/sdk/dbtest/seedmodels/forms.go:224-330` - Full sales order form
- `business/sdk/dbtest/seedmodels/tableforms.go:883-894` - Order form fields
- `business/sdk/dbtest/seedmodels/tableforms.go:897-910` - Line item form fields

**Status Business Layer**
- `business/domain/sales/orderfulfillmentstatusbus/`
- `business/domain/sales/lineitemfulfillmentstatusbus/`

---

**Last Updated**: 2025-12-29
**Created By**: Plan Creator
**Status**: Planning
