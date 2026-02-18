# Workflow Action Gap Remediation Plan

**Project**: Ichor ERP System
**Goal**: Identify and close all gaps in the workflow action system relative to ERP automation needs
**Timeline**: 9 phases
**Status**: Planning Complete - Ready for Phase Execution

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Gap Analysis Summary](#gap-analysis-summary)
3. [Implementation Strategy](#implementation-strategy)
4. [Phase Overview](#phase-overview)
5. [Technology Stack](#technology-stack)
6. [Success Criteria](#success-criteria)
7. [Quick Reference Guide](#quick-reference-guide)

---

## Executive Summary

### The Goal

Close all identified gaps in the Ichor workflow action system so that end-to-end ERP automations — reorder triggering, PO approval, inventory receipt, email notifications, and outbound integrations — work correctly in production.

### The Approach

A discovery phase identified 9 concrete gaps ranging from one-line whitelist additions to full async-approval infrastructure. Phases are ordered by effort-to-impact ratio: quick wins first, complex features last.

1. Add missing table entries to the data whitelist (10 lines — unblocks procurement/HR/assets)
2. Fix FieldChanges propagation in the delegate handler (enables change-based workflow triggers)
3. Implement `send_notification` via RabbitMQ (already have the infrastructure)
4. Implement `send_email` with real SMTP delivery
5. Implement `seek_approval` with persistence and async Temporal loop
6. Add `create_purchase_order` action (closes the reorder automation chain)
7. Add `receive_inventory` action (closes PO receipt → stock update)
8. Add `call_webhook` action (outbound third-party integrations)
9. Add arithmetic to the template processor

### Why This Matters

- **Approval gates are broken**: `seek_approval` always auto-approves silently
- **Email never sends**: `send_email` only logs — zero delivery
- **Procurement automation is blocked**: procurement tables are not in the whitelist
- **Reorder chain has no endpoint**: `check_reorder_point → needs_reorder` dead-ends at an alert with no way to create a PO
- **FieldChanges always empty**: `changed_from`/`changed_to` conditions never fire, making update-based workflows imprecise

---

## Gap Analysis Summary

Full analysis available in `.claude/research/workflow-gaps-analysis.md` (see session context).

| Gap | Severity | Phase | Effort |
|-----|----------|-------|--------|
| Missing procurement/HR/assets tables in whitelist | **Critical** | 1 | Low |
| FieldChanges never populated in DelegateHandler | **Critical** | 2 | Medium |
| `send_notification` is a stub | Medium | 3 | Low |
| `send_email` never sends | **High** | 4 | Medium |
| `seek_approval` always auto-approves | **Critical** | 5 | High |
| No `create_purchase_order` action | **High** | 6 | High |
| No `receive_inventory` action | **High** | 7 | Medium |
| No `call_webhook` action | Medium | 8 | Medium |
| No arithmetic in template processor | Medium | 9 | Medium |

### What's Already Working

These action handlers are fully implemented and production-ready:

- **All 6 inventory actions**: allocate, reserve, commit, release, check, reorder check
- **All 5 data actions**: update_field, lookup_entity, create_entity, transition_status, log_audit_entry
- **Control flow**: evaluate_condition, delay
- **create_alert**: full implementation with RabbitMQ WebSocket delivery

---

## Implementation Strategy

### Approach

Sequential phases ordered by effort-to-impact ratio. Phase 1 is a prerequisite for several later phases (procurement automation won't work until procurement tables are whitelisted). Phases 2–4 are independent quick wins. Phases 5–9 are larger features that build on the stable foundation.

### Key Principles

1. **Zero breaking changes** — All new handlers add to the registry without modifying existing handlers. The whitelist additions and FieldChanges fix do not change API contracts.
2. **Stub → Real pattern** — Phases 3/4/5 replace stub `Execute()` bodies with real implementations. The handler struct, config validation, and output ports remain unchanged.
3. **Test with `make test`** — Every phase validates with `go build ./...` and `go test ./...`. No phase is marked complete without a green build.

### Phase Dependencies

- Phase 1 (whitelist) is independent — no prerequisites
- Phase 2 (FieldChanges) is independent — no prerequisites
- Phases 3, 4 are independent — no prerequisites
- Phase 5 (seek_approval) needs DB migration — independent but high-effort
- Phase 6 (create_purchase_order) benefits from Phase 1 (procurement tables whitelisted)
- Phase 7 (receive_inventory) is independent
- Phase 8 (call_webhook) is fully independent
- Phase 9 (template arithmetic) is fully independent

---

## Phase Overview

| Phase | Name | Category | Description | Key Deliverables |
|-------|------|----------|-------------|------------------|
| 1 | Add Missing Tables to Whitelist | backend | Add procurement, HR, assets tables to data action whitelist | Updated tables.go with 15+ new entries |
| 2 | Fix FieldChanges in DelegateHandler | backend | Populate FieldChanges on update events for change-based conditions | Updated delegatehandler.go + DelegateEventParams |
| 3 | Implement send_notification | backend | Wire send_notification to RabbitMQ WebSocket delivery (no persistence) | Updated notification.go handler |
| 4 | Implement send_email SMTP | backend | Real SMTP email delivery via configurable server | Updated email.go + SMTP config |
| 5 | Implement seek_approval | backend+database | Full async approval: DB persistence, approver notification, Temporal pause/resume | Migration, approvalrequestbus, API endpoint, async handler |
| 6 | Add create_purchase_order Action | backend | New action to auto-create POs from reorder workflows | New handler + procurement bus integration |
| 7 | Add receive_inventory Action | backend | New action to receive PO inventory: increase stock + create transaction | New handler + inventory bus integration |
| 8 | Add call_webhook Action | backend | Outbound HTTP webhook for third-party integrations | New handler with URL/body template resolution |
| 9 | Add Arithmetic to Template Processor | backend | Enable math expressions like `{{expr: quantity * unit_price}}` | Updated template.go with expression evaluator |

---

## Technology Stack

### Backend
- **Language/Runtime**: Go 1.23
- **Framework**: Ardan Labs Service Architecture (Domain-Driven, Data-Oriented Design)
- **Database**: PostgreSQL 16.4
- **Additional**: sqlx, Temporal, RabbitMQ

### Tools
- **Development**: make, KIND (Kubernetes)
- **Testing**: go test, integration tests
- **Build**: Docker, Kubernetes

---

## Success Criteria

### Functional Requirements
- All 9 gaps closed: every phase validated green
- `seek_approval` pauses workflow until real approver responds (approved/rejected/timed_out)
- `send_email` delivers to inbox (verified with test SMTP server)
- `check_reorder_point → needs_reorder → create_purchase_order` chain executes end-to-end
- `changed_from`/`changed_to` conditions fire correctly in integration tests
- `call_webhook` delivers HTTP POST to test endpoint

### Non-Functional Requirements
- `go build ./...` passes after every phase
- `go test ./...` passes after every phase (no regressions)
- No security vulnerabilities introduced (parameterized queries, no SQL injection)
- No breaking changes to existing action handler contracts

### Quality Metrics
- **Code Quality**: All lint checks pass (`make lint`)
- **Testing**: `make test` passes after each phase
- **Performance**: No execution time regression on existing handlers

---

## Quick Reference Guide

### Commands

Execute phases:
```bash
/workflow-gaps-status      # View current progress
/workflow-gaps-next        # Execute next pending phase
/workflow-gaps-phase N     # Jump to specific phase
/workflow-gaps-validate    # Run validation checks
```

Code review:
```bash
/workflow-gaps-review N    # Manual code review for phase N
```

Planning:
```bash
/workflow-gaps-build-phase # Generate next phase documentation
/workflow-gaps-summary     # Generate executive summary
/workflow-gaps-dependencies # Show cross-plan dependencies
```

### Phase Documentation

Each phase has detailed documentation in `phases/PHASE_N_NAME.md` with:
- Overview and goals
- Task breakdown with specific file paths
- Implementation guide with code examples
- Validation criteria
- Testing strategy

### PROGRESS.yaml

Track progress in [PROGRESS.yaml](./PROGRESS.yaml):
- Real-time phase and task status
- Context (current focus, next task)
- Blockers and decisions
- Deliverables tracking

---

## Phase Documentation

1. [Phase 1: Add Missing Tables to Whitelist](./phases/PHASE_1_MISSING_TABLES.md)
2. [Phase 2: Fix FieldChanges in DelegateHandler](./phases/PHASE_2_FIELD_CHANGES.md)
3. [Phase 3: Implement send_notification](./phases/PHASE_3_SEND_NOTIFICATION.md)
4. [Phase 4: Implement send_email SMTP](./phases/PHASE_4_SEND_EMAIL.md)
5. [Phase 5: Implement seek_approval](./phases/PHASE_5_SEEK_APPROVAL.md)
6. [Phase 6: Add create_purchase_order Action](./phases/PHASE_6_CREATE_PO.md)
7. [Phase 7: Add receive_inventory Action](./phases/PHASE_7_RECEIVE_INVENTORY.md)
8. [Phase 8: Add call_webhook Action](./phases/PHASE_8_CALL_WEBHOOK.md)
9. [Phase 9: Add Arithmetic to Template Processor](./phases/PHASE_9_TEMPLATE_ARITHMETIC.md)

---

## Related Documentation

- [Workflow Engine Documentation](../../../docs/workflow/README.md)
- [Action Handler Architecture](../../../docs/workflow/actions/overview.md)
- [Template Processor](../../../business/sdk/workflow/template.go)
- [Tables Whitelist](../../../business/sdk/workflow/workflowactions/data/tables.go)

---

## Notes

### Assumptions
- SMTP credentials will be provided via environment variables (`ICHOR_SMTP_*`)
- Approval requests use a new `workflow.approval_requests` table (not the existing `hr.user_approval_statuses` which is HR-specific)
- `call_webhook` uses a timeout and does not retry (Temporal provides retries at the workflow level)
- Template arithmetic uses a safe expression evaluator (no `eval`, no arbitrary code)

### Constraints
- No breaking changes to existing action handler interfaces
- No new required dependencies without updating `register.go` config structs
- All new DB tables follow existing migration versioning (append only)

### Future Enhancements
- `generate_document` action (PDF invoice/PO generation)
- `apply_discount` action (pricing rule automation)
- `validate_address` action (geocoding integration)
- SMS/push notification channels for `send_notification`

---

**Last Updated**: 2026-02-18
**Created By**: Claude Code
**Status**: Planning Complete
