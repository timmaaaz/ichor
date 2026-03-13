# Workflow Testing Gaps Plan

**Created:** 2026-03-10
**Purpose:** Track and fix gaps in workflow integration and unit test coverage
**Status:** Research complete — phases ready to execute

---

## Research Summary

Analysis of ~110 workflow test files (500+ test functions) across three layers:
- `business/sdk/workflow/temporal/` — unit tests (155+, well-covered)
- `business/sdk/workflow/workflowactions/` — per-handler unit tests (all handlers covered)
- `api/cmd/services/ichor/tests/workflow/` — integration tests (mostly covered via table-helper pattern)

### Pattern Note
Most `*_test.go` files in integration directories do NOT have top-level `Test*` functions.
Instead they export table-builder helpers (e.g. `createRule201(sd)`) called by a single
master runner (e.g. `Test_RuleAPI`). This is intentional and correct.

---

## Confirmed Gaps

### GAP-1: workflowsaveapi — 3 files excluded since Phase 13 (HIGH PRIORITY)

**Files:** (all have `//go:build ignore` with comment "Phase 13: Excluded until Phase 15 rewrites for Temporal")
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_test.go`
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go`
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go`

**What they test (pre-Temporal):**
- `execution_test.go` — E2E workflow execution tests (run a workflow, verify execution record)
- `actions_test.go` — Per-action-type execution side effects (alert created, email sent, etc.)
- `errors_test.go` — Error propagation, action failure handling, partial workflow failures

**What's needed:**
- Rewrite to use `InitWorkflowInfra` (Temporal + Worker) instead of old `workflow.Engine.ExecuteWorkflow`
- Wire into `Test_WorkflowSaveAPI` master runner
- Focus on: execution record created, Temporal workflow completes, side effects verified

**Reference:** See `api/cmd/services/ichor/tests/workflow/approvalapi/approval_test.go` for the
pattern of Temporal-based integration tests using `InitWorkflowInfra`.

---

### GAP-2: workflowsaveapi/trigger_test.go — written but never wired (HIGH PRIORITY)

**File:** `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go`

**What's there:**
- `TriggerTestData` struct (extends ExecutionTestData with CustomersBus, StreetID, ContactInfoID)
- `insertTriggerSeedData()` — wires delegate handler, seeds FK deps, returns TriggerTestData
- `runTriggerTests()` — runner for 3 subtests:
  - `trigger-customer-create` — create customer → expect workflow dispatch via delegate
  - `trigger-customer-update` — update customer → expect on_update workflow dispatch
  - `trigger-inactive-rule-no-trigger` — inactive rule should NOT trigger
- 3 test helper functions: `testCustomerCreateTriggersWorkflow`, `testCustomerUpdateTriggersWorkflow`, `testInactiveRuleNoTrigger`

**What's missing:**
`Test_WorkflowSaveAPI` in `save_test.go` does not call `runTriggerTests()`. The trigger test
infrastructure is fully written but completely excluded from the test run.

**Fix:**
```go
// In save_test.go, after existing test.Run calls:

// Trigger integration tests (Temporal delegate pipeline)
esd := insertExecutionSeedData(t, test, sd)
tsd := insertTriggerSeedData(t, test, esd)
runTriggerTests(t, tsd)
```

**Note:** Requires `ExecutionTestData` infrastructure (from `execution_seed_test.go` and
`InitWorkflowInfra`) to be hoisted above trigger setup. May need `Test_WorkflowSaveAPI` to
be split into a sub-test structure.

---

### GAP-3: actionapi — no happy-path execute200 test (MEDIUM PRIORITY)

**File:** `api/cmd/services/ichor/tests/workflow/actionapi/action_test.go`

**What's there:** Only error-path tests for execute endpoint:
- `execute401` — missing token
- `execute403NoPermission` — wrong permission
- `execute404UnknownAction` — unknown type (actually 403 due to permission-first check)

**What's missing:**
- `execute200` — successful manual execution of an action (e.g. `create_alert`) by an admin or
  user with the correct `actionpermissions` row, verifying: 200 status + result payload
- `executeAsync200` — successful dispatch of async action (returns task token / status pending)
- `getExecutionStatus200` — polling a valid execution status ID

**Context:** The `getExecutionStatus401` and `getExecutionStatus404` helpers exist in
`status_test.go` but a 200 case (successful status lookup) is missing.

---

### GAP-4: TriggerProcessor business-layer isolation tests (LOW PRIORITY)

**Location:** `business/sdk/workflow/` (trigger.go unit tests are sparse)

**What's there:**
- `temporal/trigger_test.go` — 15 tests for `WorkflowTrigger.OnEntityEvent`
- `temporal/delegatehandler_test.go` — 2 tests (computeFieldChanges, extractEntityData)

**What's missing (for `business/sdk/workflow/trigger.go`):**
- `RegisterCacheInvalidation` → delegate fire → `RefreshRules` round-trip
- `evaluateRuleConditions` with mixed `and`/`or` logic and multiple conditions
- `evaluateFieldCondition` for `contains`, `starts_with`, `ends_with` operators
- `TriggerProcessor.Initialize` failure modes (DB down, empty rule set)
- `isSupportedEventType` boundary conditions

**Note:** The condition operators are tested in
`workflowactions/control/condition_test.go` but that's the action handler, not the trigger
processor's own `evaluateFieldCondition`. These are separate implementations.

---

### GAP-5: EvalExpr and template error paths (LOW PRIORITY)

**Location:** `business/sdk/workflow/expr_test.go`, `business/sdk/workflow/template_test.go`

**What's there:** `TestEvalExpr`, `TestTemplateProcessor_ErrorHandling`

**What's missing:**
- `EvalExpr` division by zero → should return error
- `EvalExpr` unknown variable reference → behavior undefined (silently 0? error?)
- `EvalExpr` malformed expression (unclosed paren, double operator)
- Template variable resolution with deeply nested missing keys
- `resolveTemplateVars` with circular references (if possible)

---

## Not Gaps (False Positives from Initial Research)

These were flagged but are actually covered:

| Area | Why It's Covered |
|------|-----------------|
| ruleapi CRUD | Table helpers in create/update/delete/query/toggle files wired into `Test_RuleAPI` |
| workflowsaveapi create/update/validation/dryrun | Wired into `Test_WorkflowSaveAPI` |
| edgeapi CRUD | Wired into `Test_EdgeAPI` |
| alertapi acknowledge/dismiss/query | Wired into `Test_AlertAPI` |
| actionapi list/execute error paths | Wired into `Test_ActionAPI` |
| Branch convergence | Covered in `temporal/graph_executor_convergence_test.go` (15 tests) |
| ContinueAsNew | Covered in `temporal/workflow_continueasnew_test.go` (9 tests) |
| Retry policies | Covered in `temporal/workflow_errors_test.go` (14 tests) |

---

## Execution Phases

### Phase 1 — Wire trigger_test.go (Quick Win, ~1-2 hours)

**Goal:** The trigger test infrastructure is fully written. Just wire it in.

**Tasks:**
1. Read `save_test.go` and `execution_seed_test.go` to understand ExecutionTestData
2. Determine if `Test_WorkflowSaveAPI` needs to be restructured as subtests
3. Add `insertTriggerSeedData` + `runTriggerTests` call to master runner
4. Run: `go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...`
5. Fix any seeding issues

**Risk:** `Test_WorkflowSaveAPI` already uses `InitWorkflowInfra` for execution_seed. May need
Temporal container running to pass. Check if `execution_seed_test.go` is actually compiled (no build tag).

---

### Phase 2 — Rewrite execution_test.go for Temporal (Complex, ~4-6 hours)

**Goal:** Restore execution integration tests using Temporal instead of old Engine.

**Tasks:**
1. Read the ignored `execution_test.go` to understand what it tested
2. Design replacement: use `InitWorkflowInfra`, seed a rule with 1-2 actions, call
   `WorkflowTrigger.startWorkflowForRule` directly OR trigger via delegate, wait for
   Temporal execution to complete, verify execution record in DB
3. Create `runExecutionTests(t, esd ExecutionTestData)` helper
4. Wire into `Test_WorkflowSaveAPI`

**Key pattern:** See `approvalapi/approval_test.go` — uses `wf.WorkflowTrigger`, seeds a complete
workflow, verifies Temporal execution completed.

---

### Phase 3 — Rewrite actions_test.go for Temporal (Complex, ~4-6 hours)

**Goal:** Restore per-action-type side-effect integration tests.

**Tasks:**
1. Read the ignored `actions_test.go` to understand which actions it tested
2. For each action: create a rule with that action type, trigger it, verify the side effect
   - `create_alert` → verify `alertbus.Query` returns the new alert
   - `seek_approval` → verify `approvalrequestbus.Query` returns pending request
   - `allocate_inventory` → verify allocation result record created
3. Focus on 3-5 most important action types first; others can follow
4. Wire into `Test_WorkflowSaveAPI` via `runActionTests(t, esd)`

---

### Phase 4 — Rewrite errors_test.go for Temporal (Medium, ~2-4 hours)

**Goal:** Restore error scenario integration tests.

**Tasks:**
1. Read the ignored `errors_test.go` to understand what it tested
2. Design replacement scenarios:
   - Invalid action config → workflow fails at activity, execution record marks failed
   - Handler not found → workflow fails immediately
   - Network error in webhook → retry behavior observed
3. Wire into `Test_WorkflowSaveAPI`

---

### Phase 5 — actionapi execute200 happy path (Medium, ~2-3 hours)

**Goal:** Add successful manual action execution integration test.

**Tasks:**
1. Read `actionapi/seed_test.go` to understand `ActionSeedData`
2. Seed an `actionpermissions` row granting a user `create_alert` execution permission
3. Add `execute200CreateAlert(sd)` table helper
4. Add `getExecutionStatus200(sd)` table helper (poll a valid execution)
5. Wire into `Test_ActionAPI`

---

### Phase 6 — TriggerProcessor + EvalExpr unit tests (Low, ~2-3 hours)

**Goal:** Fill business-layer unit test gaps for trigger and expression evaluation.

**Tasks:**
1. Add `TestTriggerProcessor_EvaluateConditions_MixedLogic` to `trigger_test.go` (or new file)
2. Add `TestTriggerProcessor_RegisterCacheInvalidation`
3. Add `TestEvalExpr_DivisionByZero`, `TestEvalExpr_UnknownVariable`, `TestEvalExpr_MalformedExpr`
4. Add `TestTriggerProcessor_Initialize_EmptyRuleSet`

---

## Files Reference

| File | Status | Phase |
|------|--------|-------|
| `workflowsaveapi/trigger_test.go` | Written, unwired | Phase 1 |
| `workflowsaveapi/execution_test.go` | `//go:build ignore` | Phase 2 |
| `workflowsaveapi/actions_test.go` | `//go:build ignore` | Phase 3 |
| `workflowsaveapi/errors_test.go` | `//go:build ignore` | Phase 4 |
| `actionapi/execute_test.go` | Error paths only | Phase 5 |
| `actionapi/status_test.go` | Error paths only | Phase 5 |
| `business/sdk/workflow/trigger.go` | Sparse unit tests | Phase 6 |
| `business/sdk/workflow/expr_test.go` | Missing error paths | Phase 6 |

---

## Key Infrastructure Files

- `api/sdk/http/apitest/workflow.go` — `InitWorkflowInfra` (Temporal worker + infra)
- `business/sdk/workflow/testutil.go` — `TestSeedFullWorkflow` and all seed helpers
- `api/cmd/services/ichor/tests/workflow/approvalapi/approval_test.go` — reference pattern for Temporal integration tests
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_seed_test.go` — `ExecutionTestData` struct

---

## Decision Points

Before Phase 2-4, verify: does `Test_WorkflowSaveAPI` currently pass with the Temporal
container running? The `execution_seed_test.go` seeds `InitWorkflowInfra` but `save_test.go`
never calls `runExecutionTests` — so the infra is seeded but unused. Confirm this compiles
and runs cleanly before adding more Temporal-dependent tests.
