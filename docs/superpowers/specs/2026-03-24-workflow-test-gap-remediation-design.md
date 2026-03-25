# Workflow System Test Gap Remediation

**Date:** 2026-03-24
**Status:** Approved
**Scope:** Add missing test coverage across the workflow system

## Context

An exploration of the workflow system identified 9 testing gaps ranging from HIGH to LOW severity. The Temporal package and most action handlers are well-tested (155+ unit tests in temporal, thorough condition/webhook/delay/inventory handler coverage). The gaps concentrate in: permissions, orchestration, DAG validation, one large action handler, and integration test completeness.

## Execution Model

All 5 phases are **fully independent** — no overlapping files, no shared state. Each phase creates new `*_test.go` files without modifying existing source.

**Each phase runs in its own git worktree**, allowing all 5 to execute in parallel. Each produces its own PR for review.

| Phase | Packages Touched | Conflicts With |
|-------|-----------------|----------------|
| 1 | `actionpermissionsbus/`, `workflowactions/data/tables.go` | None |
| 2 | `workflow/actionservice.go` | None |
| 3 | `workflowactions/procurement/createpo.go` | None |
| 4 | `workflowsaveapp/` | None |
| 5 | `notificationsapi/`, `communication/resend.go` | None |

## Testing Philosophy

These principles govern all test code written across every phase:

- **Tests get what they need to run and nothing more** — if it doesn't need DB, don't give it DB. If it needs DB, use `dbtest` helpers with real Postgres.
- **Test real functionality, don't mock** — mocks only where external systems make real calls impossible (email APIs, HTTP clients). Internal interfaces use real implementations.
- **Use `cmp.diff` for structural comparison** — compare full structs, not individual fields. Catches unexpected mutations and missing fields.
- **Tests hunt for real bugs** — every test should target a plausible failure mode. No tests that exist solely for coverage metrics.
- **App layer tests are integration tests** — they exercise the bus-to-app seam through the HTTP API, not as isolated unit tests. Focus on gaps between layers.
- **Follow existing patterns** — `dbtest` for DB-backed tests, `apitest.Table` for HTTP integration tests, Temporal test environment for workflow tests.

## Phase 1: Security & Permissions

**Priority:** HIGH
**Packages:** `business/domain/workflow/actionpermissionsbus/`, `business/sdk/workflow/workflowactions/data/tables.go`

### actionpermissionsbus Unit Tests

New file: `business/domain/workflow/actionpermissionsbus/actionpermissionsbus_test.go`

Uses `dbtest` helpers (real Postgres). Tests:

- **CRUD lifecycle** — Create, Update, Delete, Query, QueryByID, Count. Full `cmp.diff` on returned structs.
- **QueryByRoleAndAction** — exact match lookup, not-found case.
- **CanUserExecuteAction** — the core permission check:
  - User with role that has permission → true
  - User with role that lacks permission → false
  - User with multiple roles, one has permission → true
  - User with no roles → false
  - Permission for different action type → false
- **GetAllowedActionsForRoles** — returns correct action types for given role IDs, empty roles returns empty list, multiple roles aggregate correctly.
- **NewWithTx** — verify transactional behavior (create in tx, rollback, verify not persisted).

### data/tables.go Unit Tests

New file: `business/sdk/workflow/workflowactions/data/tables_test.go`

Pure unit tests, no DB needed. These are SQL injection prevention functions:

- **IsValidColumnName** — valid column names, names with SQL injection attempts (`; DROP TABLE`, `1=1 --`), empty string, special characters, unicode.
- **IsValidTableName** — whitelist-based lookup. Test known valid entries from the whitelist return true, unknown table names return false, empty string returns false.
- **IsValidOperator** — each valid operator, invalid operators, empty string, SQL fragments.

## Phase 2: ActionService Orchestration

**Priority:** HIGH
**Package:** `business/sdk/workflow/actionservice.go`

New file: `business/sdk/workflow/actionservice_test.go` (use `package workflow_test` — matches existing test files in this directory like `workflow_crud_test.go`).

The ActionService wires together: registry lookup → template resolution → permission checks → handler dispatch → execution recording. It needs a real DB (for execution recording) and real registry (for handler lookup), but can use test action handlers.

Tests:

- **Execute — happy path** — registered action, valid config, verify execution recorded in DB with correct fields via `cmp.diff`.
- **Execute — unknown action type** — returns appropriate error.
- **Execute — handler validation failure** — handler's `Validate` rejects config, verify error propagation.
- **Execute — handler execution failure** — handler's `Execute` returns error, verify execution recorded with error status.
- **Execute — template resolution** — config with `{{variables}}`, verify templates resolved before handler receives config.
- **ExecuteForAutomation** — executes with rule context (ruleID, ruleName, eventType). Note: this method does NOT call `recordExecution` — no DB write assertions. Verify the `ActionExecutionContext` passed to the handler has `RuleID`, `RuleName`, `EventType`, and `TriggerSource=automation` populated correctly (use a spy/test handler).
- **GetExecutionStatus** — query by execution ID, not-found case.
- **ListAvailableActions / ListManuallyExecutableActions** — verify registry enumeration returns correct action info.
- **recordExecution** — verify all fields persisted correctly (execution ID, action type, status, timestamps, trigger source).

## Phase 3: CreatePurchaseOrderHandler

**Priority:** HIGH
**Package:** `business/sdk/workflow/workflowactions/procurement/createpo.go`

New file: `business/sdk/workflow/workflowactions/procurement/createpo_test.go`

503-line handler with supplier lookups, event extraction, PO creation. Needs DB for real supplier/product/PO queries.

Tests:

- **Validate — valid config** — all required fields present.
- **Validate — missing required fields** — product ID, quantity, each returns clear error.
- **Validate — invalid values** — negative quantity, zero quantity.
- **Execute — happy path** — product exists, supplier found, PO created. Verify PO fields via `cmp.diff`.
- **Execute — supplier product not found** — graceful error, no PO created.
- **Execute — extractFromEvent** — config references event data, verify correct extraction from `ActionExecutionContext`.
- **Execute — lookupSupplierProduct** — product has supplier mapping → returns it; no mapping → returns false.
- **GetType, GetDescription, SupportsManualExecution, IsAsync** — verify metadata methods.
- **GetOutputPorts** — verify declared output ports match actual behavior.
- **GetEntityModifications** — verify declared modifications match what Execute actually does.

## Phase 4: WorkflowSaveApp Validation

**Priority:** MEDIUM
**Package:** `app/domain/workflow/workflowsaveapp/`

### graph.go Unit Tests

New file: `app/domain/workflow/workflowsaveapp/graph_test.go`

Pure unit tests — these functions take slices of request structs and return errors. No DB needed.

- **ValidateGraph — valid DAG** — linear chain, diamond, parallel branches with convergence.
- **ValidateGraph — cycle detection** — A→B→C→A, self-loop, complex cycle in larger graph.
- **ValidateGraph — unreachable actions** — action with no incoming edge (not start), orphaned subgraph.
- **ValidateGraph — missing action references** — edge references action ID that doesn't exist.
- **ValidateGraph — empty graph** — no actions, no edges.
- **resolveActionRef — by name** — resolves action reference correctly.
- **resolveActionRef — by index** — numeric ref resolves to correct action.
- **resolveActionRef — invalid ref** — returns error for unknown reference.
- **detectCycles — Kahn's algorithm edge cases** — single node, two nodes no cycle, two nodes with cycle, large acyclic graph.
- **checkReachability — all reachable** — every action reachable from start edges.
- **checkReachability — disconnected component** — subset of actions unreachable.

### validation.go Unit Tests

New file: `app/domain/workflow/workflowsaveapp/validation_test.go`

Pure unit tests for config validation:

- **validateOutputPorts** — action with condition outputs wired correctly, missing true_branch edge, missing false_branch edge, non-condition action with extra edges.
- **ValidateActionConfigs** — each of the 7 config validators:
  - `validateCreateAlertConfig` — valid, missing severity, missing message template.
  - `validateSendEmailConfig` — valid, missing recipients, missing subject.
  - `validateSendNotificationConfig` — valid, missing recipients, missing channels.
  - `validateUpdateFieldConfig` — valid, missing field name, missing value.
  - `validateSeekApprovalConfig` — valid, missing approver roles.
  - `validateAllocateInventoryConfig` — valid, missing product reference.
  - `validateEvaluateConditionConfig` — valid, empty conditions array.
- Each validator tested with invalid JSON input.
- **Unknown action type** — `validateActionConfig` with unregistered type returns error (guards against new types silently bypassing validation).

### Integration Test Gap Check

Review existing `workflowsaveapi/` integration tests for coverage of:
- `SaveWorkflow` — create + update paths
- `CreateWorkflow` — new workflow from scratch
- `DuplicateWorkflow` — copy existing workflow
- `syncActions` — action add/update/remove during save

Add integration tests only where the bus↔app seam has gaps not covered by existing `workflowsaveapi/` tests.

## Phase 5: Integration Test Completeness

**Priority:** MEDIUM/LOW
**Packages:** Various

### notificationsapi Integration Tests

New directory: `api/cmd/services/ichor/tests/workflow/notificationsapi/`

Uses `apitest.Table` pattern:
- Query notification deliveries (list, pagination, filters).
- Query by execution ID.
- Verify response structure via `cmp.diff`.

### ResendEmailClient Tests

New file: `business/sdk/workflow/workflowactions/communication/resend_test.go`

- **NewResendEmailClient — empty API key** — returns nil (graceful degradation).
- **NewResendEmailClient — empty From** — returns nil.
- **NewResendEmailClient — valid config** — returns non-nil client.
- Note: `MockEmailClient` is already exercised by `email_test.go` — no additional mock tests needed.

### Conditional Skip Audit

Review CI configuration to determine if these skipped tests actually run in the full test suite:
- `actionhandlers/inventory_test.go` — skips without seeded products
- `workflowsaveapi/errors_test.go` — skips with < 3 trigger types
- `workflow_replay_test.go` — skips in `-short` mode

Document findings. If CI always runs in short mode or without seed data, these tests are effectively dead code — flag for follow-up.

### actionpermissionsapi Assessment

Determine if `actionpermissionsapi` HTTP endpoints exist. If they do, add integration tests. If no API layer exists, document that permissions are internal-only and integration tests are not applicable.

## Success Criteria

Each phase is complete when:
1. All listed tests pass (`go test ./path/to/package/...`)
2. Tests use `cmp.diff` for structural comparison where applicable
3. No mocks of internal interfaces — real implementations only
4. Tests target plausible failure modes, not just happy paths
5. PR passes CI and is reviewed independently

## Discovered Work

If a phase discovers that source changes are needed to make tests possible (e.g., Phase 5 finds `actionpermissionsapi` HTTP endpoints don't exist and need to be built, or a handler needs an exported constructor), that work gets absorbed into that phase. The phase owns the full scope of making its tests real — don't punt discovered prerequisites to a separate effort.

However, if a phase discovers **bugs** in existing logic (not missing infrastructure), document them and file separately. The phase's job is to add test coverage, not fix pre-existing bugs.

## Out of Scope

- App layer unit tests for `actionapp` and `actionpermissionsapp` — their logic is thin wrappers; integration tests cover the bus↔app seam adequately.
- Large-scale refactoring unrelated to testability.
