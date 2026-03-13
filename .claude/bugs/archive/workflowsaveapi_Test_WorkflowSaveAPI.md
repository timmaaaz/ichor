# Test Failure: Test_WorkflowSaveAPI

- **Package**: `github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/workflowsaveapi`
- **Duration**: 43.48s

## Failure Output

```
2026/03/13 11:43:47 INFO  No logger configured for temporal client. Created default one.
    save_test.go:64: Temporal Started: servicetest-temporal at localhost:56424
2026/03/13 11:43:47 INFO  No logger configured for temporal client. Created default one.
2026/03/13 11:43:48 INFO  Started Worker Namespace default TaskQueue test-workflow-Test_WorkflowSaveAPI WorkerID 44522@Jakes-MacBook-Pro.local@
    save_test.go:64: Temporal workflow infrastructure initialized
2026/03/13 11:44:25 INFO  Stopped Worker Namespace default TaskQueue test-workflow-Test_WorkflowSaveAPI WorkerID 44522@Jakes-MacBook-Pro.local@
--- FAIL: Test_WorkflowSaveAPI (43.48s)
```

## Failing Subtests
1. `exec-branch-true` (15.61s timeout)
2. `error-condition-field-not-found` (3.07s — `expected no new alerts (graceful no-op), got 5 new`)

---

## Investigation: exec-branch-true

**Root cause CONFIRMED** — JSON key mismatch in seed data.

`execution_seed_test.go:332` seeds condition config as:
```json
{"conditions":[{"field":"amount","operator":"greater_than","value":1000}]}
```
`FieldCondition` struct (`control/condition.go:17`) uses `json:"field_name"`, not `json:"field"`.
So `cond.FieldName = ""` → `data[""] = nil` → `compareValues(nil, 1000, ">")` = false.
Condition always routes to false branch → "Normal Value Alert" fires, not "High Value Alert".
Test polls for `high_value` alert, never appears, times out after 15s.

**Fix**: `execution_seed_test.go:332` — `"field":"amount"` → `"field_name":"amount"`

`actions_test.go` condition tests all use `"field_name":` correctly (lines 407, 515, 623, 732-733). No change needed there.

**Log noise** (not causing this failure): Many rules with no `template_id` seeded via
`TestSeedRuleActions(ctx, 3, ruleIDs, nil, ...)` fire for TriggerTypes[0]. `edgedb.toActionNode` leaves
`ActionType=""` when template LEFT JOIN is NULL. Activity fails: "action_type is required" after 3 retries.
Pre-existing issue, does not affect this subtest's assertion.

---

## Investigation: error-condition-field-not-found

**Root cause PARTIALLY CONFIRMED** — stale Temporal workflows from prior action tests complete during 3s wait.

Failure: `errors_test.go:267: expected no new alerts (graceful no-op), got 5 new`

`error-condition-field-not-found` fires TriggerTypes[2] ("on_delete") for Entities[0].
Only the test's own "Missing Field Test" rule uses TriggerTypes[2].
That condition (`field_name: nonexistent_field_xyz`) evaluates false → no alerts from this rule.

**Hypothesis**: The 5 alerts come from Temporal workflows that were started by earlier `runActionTests`
tests (all use TriggerTypes[0]) and haven't completed yet when the 3s sleep starts. Specifically:
- `action-condition-greater-than` (testConditionGreaterThan) creates a condition rule that succeeds,
  then follows a `create_alert` branch. If that async workflow completes mid-sleep, alert count rises.
- `error-action-fails-sequence-stops` fires TriggerTypes[0], triggering ALL accumulated TriggerTypes[0]
  rules (inc. action test rules), some may still be completing during the 3s window.

**Still need to verify**: read `actions_test.go` condition tests to confirm which create alerts,
and check whether `action-condition-*` tests each create a rule that leads to an alert branch.

**Key facts**:
- `RefreshRules` works correctly — resets `lastLoadTime` to zero, forces `loadMetadata` reload.
- TriggerProcessor `checkRuleMatch` (`trigger.go:216`) DOES check `rule.TriggerTypeName != event.EventType`.
  So TriggerTypes[0] rules do NOT fire for the "on_delete" event. Hypothesis is timing, not wrong matching.
- `before` count is taken AFTER firing the event but BEFORE the 3s sleep, so any alert from a
  concurrent Temporal workflow completing mid-sleep increments `after - before`.

---

## Fix Plan
1. `execution_seed_test.go:332`: `"field":"amount"` → `"field_name":"amount"` ← fixes exec-branch-true
2. For error-condition-field-not-found: investigate action_test condition workflows to confirm
   timing issue, then either: (a) take `before` count BEFORE firing, or (b) lengthen sleep,
   or (c) filter alert query to only this test's specific rule/entity.

---

## Session 2 Investigation (2026-03-13)

### Status: Both originally-identified fixes already applied in PR #89

**Fix #1 (exec-branch-true)**: `execution_seed_test.go:332` already uses `"field_name":"amount"` (not `"field"`).
Git log shows last touches: `ead95fc3` (PR #74), `36aba3d8` (PR #73). Fix was in the codebase before this session.

**Fix #2 (error-condition-field-not-found)**: `errors_test.go` already uses `QueryFilter{SourceRuleID: &ruleID}`
to scope the alert query to only this test's rule. This was introduced in `0f9e2588` (PR #89).
- `alertbus.QueryFilter.SourceRuleID *uuid.UUID` exists at `business/domain/workflow/alertbus/filter.go:18`
- DB store WHERE clause: `source_rule_id = :source_rule_id` in both `applyFilter` and `applyFilterWithJoin`
- `CreateAlertHandler` sets `SourceRuleID: sourceRuleID` where `sourceRuleID = *execCtx.RuleID` (or `uuid.Nil` if nil)

### Open Questions (need test run to confirm)
1. Does `exec-branch-true` still time out (15s poll for high_value alert)?
2. Does `error-condition-field-not-found` still get 5 unexpected alerts?

If both pass → move bug to complete.
If still failing → root cause has shifted since original investigation.

### Key Architecture Note
The `testConditionFieldNotFound` test uses `TriggerTypes[2]`, not `TriggerTypes[0]`.
The 8 action tests (`runActionTests`) ALL use `TriggerTypes[0]` and NONE clean up their rules.
So the 5-alert burst could not come from those rules — unless `source_rule_id` was nil on the spurious alerts
(making them appear under ANY rule's filter). That was fixed in PR #89.

### Test Run Results (2026-03-13)
- `exec-branch-true` — **PASS** (0.06s) — `execution_test.go:176: SUCCESS: branch-true path created high_value alert`
- `error-condition-field-not-found` — **PASS** (3.04s) — `errors_test.go:265: SUCCESS: missing field condition handled gracefully`
- **Full suite: ALL 35 subtests PASS** in 32.66s

Both bugs were already fixed in PR #89 (`0f9e2588`). No code changes needed this session.

---

## Temporal Container Isolation Investigation (2026-03-13)

User proposed: apply per-test Temporal container isolation analogous to `dbtest` Postgres containers.

### Current architecture (foundation/temporal/temporal.go)
- Docker image: `temporalio/temporal:latest` with `["server", "start-dev", "--ip", "0.0.0.0"]` (SQLite dev server)
- Container management: `foundation/docker/docker.go` — raw `exec.Command("docker", ...)`, same as Postgres
- Reuse pattern: package-level `sync.Mutex + testStarted bool + testContainer *Container` singleton
  → One container per test binary process, never stopped
- `InitWorkflowInfra` (`api/sdk/http/apitest/workflow.go:38–127`) creates per-test isolation via:
  - Fresh `temporalclient.Client` per call
  - **Unique task queue per test**: `"test-workflow-" + t.Name()`
  - Worker + client stopped in `t.Cleanup`

### Postgres pattern (business/sdk/dbtest/dbtest.go)
- Shared container (named `"servicetest"`, never stopped) — same singleton approach
- Isolation achieved by: fresh randomly-named DB per `NewDatabase` call, dropped in `t.Cleanup`

### Assessment: Per-subtest Temporal containers are NOT viable
- Temporal `waitForReady` takes **10–15s** per container start (60s timeout, 500ms poll)
- 10 subtests × 12s = **2 minutes of pure startup overhead** added
- Unique task queue per `t.Name()` already provides the same isolation guarantee without the overhead:
  - Workflows from subtest A are dispatched to task queue `test-workflow-Test_X/subtest-a`
  - Subtest B worker only listens on `test-workflow-Test_X/subtest-b` — never sees A's workflows
- **Conclusion**: Current shared-container + unique-task-queue design is the correct approach.
  The original timing bug (5 spurious alerts) was caused by nil `source_rule_id` on alerts, fixed in PR #89.

---

## Fix

- **File**: `execution_seed_test.go:332` + `errors_test.go:258`
- **Classification**: test bug (both)
- **Change**: Fix #1 applied in PR #74/#73 (`"field"` → `"field_name"` in seed JSON). Fix #2 applied in PR #89 (`SourceRuleID` filter on alert query + `source_rule_id` propagation in `CreateAlertHandler`).
- **Verified**: `go test -v -run Test_WorkflowSaveAPI github.com/timmaaaz/ichor/api/cmd/services/ichor/tests/workflow/workflowsaveapi` → **35/35 PASS** ✓
