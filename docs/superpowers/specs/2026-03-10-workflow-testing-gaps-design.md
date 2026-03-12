# Workflow Testing Gaps — Design Spec

**Date:** 2026-03-10
**Status:** Approved
**Source plan:** `docs/plans/workflow-testing-gaps/PLAN.md`

---

## Overview

Restore and extend workflow integration and unit test coverage after the Temporal migration (Phase 13/15). Three `//go:build ignore` files need rewriting, one file needs wiring, and new happy-path and unit tests need adding.

---

## Phase Structure

| Phase | Name | Type | Target | Runs In |
|-------|------|------|--------|---------|
| 1 | Wire trigger_test.go | Wiring | `Test_WorkflowSaveAPI` | Temporal container |
| 2 | Rewrite execution_test.go | Rewrite | `Test_WorkflowSaveAPI` | Temporal container |
| 3a | Actions: alert + approval | Rewrite | `Test_WorkflowSaveAPI` | Temporal container |
| 3b | Actions: inventory + procurement | New standalone | Own `TestXxx` funcs | Temporal container |
| 3c | Actions: comms + webhook | New standalone + `InitWorkflowInfra` extension | Own `TestXxx` funcs | Temporal container |
| 4 | Rewrite errors_test.go | Rewrite | `Test_WorkflowSaveAPI` | Temporal container |
| 5 | actionapi execute200 | New helpers | `Test_ActionAPI` | No Temporal needed |
| 6 | Business layer unit tests | New unit tests | `go test ./business/sdk/workflow/...` | No Temporal needed |

Phase order: 1 → 2 → 3a → (3b, 3c, 4 independent) → 5 → 6

---

## Architecture

### Real Temporal Infrastructure

All Temporal-dependent tests use **real Temporal**, not mocks:

- `foundationtemporal.GetTestContainer(t)` — spins up `temporalio/temporal:latest` with SQLite, shared across the test binary run
- `apitest.InitWorkflowInfra(t, db)` — connects a real client, starts a real worker, returns `*WorkflowInfra`
- `wf.WorkflowTrigger.OnEntityEvent(ctx, event)` — dispatches real workflows via `tc.ExecuteWorkflow()`

This exercises the full Temporal pipeline: dispatch → schedule → activity → routing → completion.

### Test File Layout

```
Test_WorkflowSaveAPI (save_test.go)
  ├── insertSeedData()                 ← existing
  ├── insertExecutionSeedData()        ← existing (InitWorkflowInfra + templates)
  ├── [existing] create/update/validation/dryrun
  ├── [Phase 1]  insertTriggerSeedData() → runTriggerTests()
  ├── [Phase 2]  runExecutionTests()
  ├── [Phase 3a] runActionTests_Alert()
  └── [Phase 4]  runErrorTests()

Standalone tests (own DB + WorkflowInfra per test):
  ├── [Phase 3b] TestReceiveInventoryAction, TestCreatePurchaseOrderAction
  └── [Phase 3c] TestCallWebhookAction, TestSendEmailAction, TestSendNotificationAction

Separate suites:
  ├── [Phase 5]  Test_ActionAPI — execute200, getExecutionStatus200 helpers
  └── [Phase 6]  go test ./business/sdk/workflow/ — pure unit tests
```

---

## Execution Pattern (Temporal tests)

Every Temporal-dependent test follows this pattern:

```
1. Seed rule + template + action + edge via wf.WorkflowBus.*
2. wf.TriggerProcessor.RefreshRules(ctx)
3. wf.WorkflowTrigger.OnEntityEvent(ctx, event)   ← dispatches to real Temporal
4. Poll DB in loop (30 × 500ms = 15s max)
5. Assert on DB state (execution record, alert row, approval request, etc.)
```

Tests assert on **DB state**, not return values.

---

## Temporal Isolation

- `Test_WorkflowSaveAPI` subtests run **sequentially** (shared `WorkflowInfra`, shared DB)
- No `t.Parallel()` on subtests within the master runner
- Task queue isolation: `"test-workflow-" + t.Name()` — each top-level test gets its own queue
- Each subtest seeds a uniquely-named rule (UUID suffix) to avoid cross-test interference
- `t.Cleanup` in `InitWorkflowInfra` handles worker stop + client close
- Standalone tests (3b, 3c) each call `dbtest.NewDatabase` + `InitWorkflowInfra` independently

---

## Handler Registration Constraint

`InitWorkflowInfra` currently registers:
- Sync: `create_alert`, `send_email` (nil client), `send_notification` (nil queue), `evaluate_condition`
- Async: `seek_approval`

**Phase 3a** — no changes needed (handlers already registered)
**Phase 3b** — standalone tests with local infra registering only the handler under test (needs `inventoryBus`, `purchaseOrderBus`)
**Phase 3c** — add `call_webhook` to `InitWorkflowInfra` (no bus deps); `send_email`/`send_notification` already registered

---

## Error Handling

**Temporal timeouts:** Every polling loop fails with a descriptive message:
```
"timeout: no X found after 15s — workflow may have failed to dispatch"
```

**Seed failures:** `t.Fatalf` on any seed error — no cascading assertion failures.

**Phase 3b minimal infra:** Each standalone test registers only the one handler under test to keep bus dependency surface minimal.

---

## Key Files

| File | Role |
|------|------|
| `api/sdk/http/apitest/workflow.go` | `InitWorkflowInfra` — extend here for `call_webhook` (Phase 3c) |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go` | Master runner — add `run*Tests` calls here |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_seed_test.go` | `ExecutionTestData` struct + workflow seeding helpers |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go` | Already written, needs wiring (Phase 1) |
| `api/cmd/services/ichor/tests/workflow/approvalapi/approval_test.go` | Reference pattern for all new Temporal tests |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_test.go` | `//go:build ignore` — rewrite (Phase 2) |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go` | `//go:build ignore` — rewrite (Phases 3a/3b/3c) |
| `api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go` | `//go:build ignore` — rewrite (Phase 4) |
