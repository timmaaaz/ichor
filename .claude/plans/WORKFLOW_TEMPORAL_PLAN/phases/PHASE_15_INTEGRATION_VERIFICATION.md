# Phase 15: Integration Verification

**Category**: testing
**Status**: Pending
**Dependencies**: Phase 13 (Dead Code Removal), Phases 11-12 (test patterns established)

---

## Overview

Rewrite the test infrastructure (`apitest/workflow.go`) for Temporal, update 3 integration test files that use the old `WorkflowInfra` (Engine/QueueManager/EventPublisher), add worker health checks with K8s liveness/readiness probes, and run full end-to-end verification.

**Scope**: 1 infrastructure file rewrite + 3 integration test rewrites + 2 health check files + verification commands.

## Goals

1. Rewrite `apitest/workflow.go` for Temporal (replace RabbitMQ-based `WorkflowInfra`)
2. Rewrite 3 integration test files that use old Engine/QueueManager/EventPublisher verification
3. Add worker health checks (`/healthz/live`, `/healthz/ready`) and K8s probes

## Prerequisites

- Phase 13 completed (old engine removed, TemporalDelegateHandler created, all.go rewired)
- Phases 11-12 completed (Temporal test patterns established in `temporal/` package)
- `foundation/temporal/` test container infrastructure available (Phase 1)
- Understanding of Temporal SDK test framework (`testsuite.WorkflowTestSuite`)

### Key Signatures Reference

```go
// foundation/temporal/temporal.go (test container)
func GetTestContainer(t *testing.T) *Container
type Container struct { HostPort string }

// Temporal SDK worker
func worker.New(client client.Client, taskQueue string, options worker.Options) worker.Worker

// temporal package (our code)
func NewWorkflowTrigger(log, starter, matcher, store) *WorkflowTrigger
func NewDelegateHandler(log, trigger) *DelegateHandler
type Activities struct { Registry *workflow.ActionRegistry; AsyncRegistry *AsyncRegistry }

// Old WorkflowInfra (being rewritten)
type WorkflowInfra struct {
    QueueManager *workflow.QueueManager  // REMOVE
    Engine       *workflow.Engine        // REMOVE
    WorkflowBus  *workflow.Business      // KEEP
    Client       *rabbitmq.Client        // REMOVE
}
```

### Callers of Old `WorkflowInfra` Fields

Grep results show exactly **3 files** that call `InitWorkflowInfra`:

| File | Uses | Old Fields Accessed |
|------|------|---------------------|
| `workflowsaveapi/execution_seed_test.go` | `wf.QueueManager.ResetMetrics()`, `wf.Engine.Initialize()`, `wf.Engine.GetRegistry()` | QueueManager, Engine |
| `workflowsaveapi/trigger_test.go` | `wf.QueueManager.GetMetrics()`, `wf.Engine.GetExecutionHistory()`, `workflow.NewEventPublisher(db.Log, esd.WF.QueueManager)`, `workflow.NewDelegateHandler()` | QueueManager, Engine, EventPublisher, DelegateHandler |
| `formdata/formdataapi/workflow_test.go` | `wf.QueueManager.GetMetrics()`, `wf.QueueManager.QueueEvent()`, `wf.Engine.GetExecutionHistory()`, `wf.Engine.Initialize()` | QueueManager, Engine |
| `sales/ordersapi/workflow_test.go` | `wf.QueueManager.GetMetrics()`, `wf.Engine.GetExecutionHistory()`, `wf.Engine.Initialize()`, `workflow.NewEventPublisher()`, `workflow.NewDelegateHandler()` | QueueManager, Engine, EventPublisher, DelegateHandler |

**Note**: PROGRESS.yaml also listed `errors_test.go` but grep shows it does NOT call `InitWorkflowInfra`. It only exists in the workflowsaveapi test directory as a separate test file that doesn't use workflow infrastructure directly.

---

## Task Breakdown

### Task 1: Rewrite `apitest/workflow.go`

**Status**: Pending

**Description**: Replace the RabbitMQ-based `WorkflowInfra` struct with a Temporal-based version. The new infrastructure spins up a Temporal test worker with workflows and activities registered, creates a `WorkflowTrigger` + `TemporalDelegateHandler` for dispatch, and exposes a Temporal client for verification.

**Current State** (117 lines):
- `WorkflowInfra` struct: `QueueManager`, `Engine`, `WorkflowBus`, `Client` (RabbitMQ)
- `InitWorkflowInfra`: creates RabbitMQ client → queue → workflowBus → Engine → registry → QueueManager
- Cleanup: `qm.Stop()`, `client.Close()` via `t.Cleanup()`
- Registers 4 handlers: send_email, send_notification, create_alert, evaluate_condition

**New Implementation**:

```go
package apitest

import (
    "context"
    "testing"

    "github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
    "github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
    "github.com/timmaaaz/ichor/business/sdk/dbtest"
    "github.com/timmaaaz/ichor/business/sdk/workflow"
    "github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
    "github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
    "github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
    "github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
    foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
    temporalclient "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

// WorkflowInfra holds the Temporal-based workflow infrastructure for tests.
type WorkflowInfra struct {
    WorkflowBus      *workflow.Business
    TemporalClient   temporalclient.Client
    WorkflowTrigger  *temporal.WorkflowTrigger
    DelegateHandler  *temporal.DelegateHandler
    TriggerProcessor *workflow.TriggerProcessor
    Worker           worker.Worker
}

// InitWorkflowInfra sets up Temporal workflow infrastructure for testing.
func InitWorkflowInfra(t *testing.T, db *dbtest.Database) *WorkflowInfra {
    t.Helper()
    ctx := context.Background()

    // 1. Get shared Temporal test container
    container := foundationtemporal.GetTestContainer(t)
    tc, err := temporalclient.Dial(temporalclient.Options{
        HostPort: container.HostPort,
    })
    if err != nil {
        t.Fatalf("connecting to temporal: %s", err)
    }

    // 2. Create workflow business layer
    workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowdb.NewStore(db.Log, db.DB))

    // 3. Build action registry (same 4 handlers as before)
    registry := workflow.NewActionRegistry()
    registry.Register(communication.NewSendEmailHandler(db.Log, db.DB))
    registry.Register(communication.NewSendNotificationHandler(db.Log, db.DB))
    alertBus := alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))
    registry.Register(communication.NewCreateAlertHandler(db.Log, alertBus, nil))
    registry.Register(control.NewEvaluateConditionHandler(db.Log))

    // 4. Create and start test worker
    taskQueue := "test-workflow-" + t.Name()
    w := worker.New(tc, taskQueue, worker.Options{})
    w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
    w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)
    activities := &temporal.Activities{
        Registry:      registry,
        AsyncRegistry: temporal.NewAsyncRegistry(),
    }
    w.RegisterActivity(activities)

    if err := w.Start(); err != nil {
        tc.Close()
        t.Fatalf("starting temporal worker: %s", err)
    }

    // 5. Create trigger infrastructure
    edgeStore := edgedb.NewStore(db.Log, db.DB)
    triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
    if err := triggerProcessor.Initialize(ctx); err != nil {
        w.Stop()
        tc.Close()
        t.Fatalf("initializing trigger processor: %s", err)
    }

    workflowTrigger := temporal.NewWorkflowTrigger(
        db.Log, tc, triggerProcessor, edgeStore,
    )

    // 6. Create delegate handler
    delegateHandler := temporal.NewDelegateHandler(db.Log, workflowTrigger)

    // 7. Register cleanup
    t.Cleanup(func() {
        w.Stop()
        tc.Close()
    })

    t.Log("Temporal workflow infrastructure initialized")

    return &WorkflowInfra{
        WorkflowBus:      workflowBus,
        TemporalClient:   tc,
        WorkflowTrigger:  workflowTrigger,
        DelegateHandler:  delegateHandler,
        TriggerProcessor: triggerProcessor,
        Worker:           w,
    }
}
```

**Key Design Decisions**:
- Unique task queue per test (`test-workflow-` + `t.Name()`) prevents cross-test interference
- `GetTestContainer(t)` returns shared singleton (same as RabbitMQ pattern) — container starts once per test run
- `NewAsyncRegistry()` returns empty registry (async workflows dispatch but fail gracefully)
- `DelegateHandler` exposed so integration tests can register domains directly
- `TriggerProcessor` exposed so tests can call `RefreshRules()` after creating rules
- `TemporalClient` exposed for direct workflow status queries in verification

**Files**:
- `api/sdk/http/apitest/workflow.go` (REWRITE: 117 lines → ~100 lines)

---

### Task 2: Rewrite `workflowsaveapi/execution_seed_test.go`

**Status**: Pending

**Description**: Update the `ExecutionTestData` struct and `insertExecutionSeedData` function to use the new Temporal-based `WorkflowInfra`. Remove references to `Engine`, `QueueManager`, `ResetEngineForTesting`, `ResetMetrics`.

**Current State** (510 lines):
- `ExecutionTestData` struct: `WF *apitest.WorkflowInfra` (accesses `WF.QueueManager`, `WF.Engine`)
- `insertExecutionSeedData`: calls `workflow.ResetEngineForTesting()`, `wf.QueueManager.ResetMetrics()`, `wf.Engine.Initialize()`, `wf.Engine.Initialize()` (twice)
- `waitForProcessing`: polls `QueueManager.GetMetrics()` — needs complete replacement
- `createTriggerEvent`: fine as-is (just builds TriggerEvent struct)

**Changes Required**:

1. **`ExecutionTestData` struct** — no change needed (still stores `WF *apitest.WorkflowInfra`), but field accesses change since WorkflowInfra fields changed

2. **`insertExecutionSeedData`** — remove:
   - `workflow.ResetEngineForTesting()` (no engine singleton)
   - `wf.QueueManager.ResetMetrics()` (no queue manager)
   - `wf.Engine.Initialize()` calls (no engine)
   - `wf.Engine.GetRegistry()` (registry now built inside InitWorkflowInfra)
   - Add: `wf.TriggerProcessor.RefreshRules(ctx)` after creating rules (replaces engine re-init)

3. **`waitForProcessing`** — complete rewrite. Two approaches:
   - **Option A**: Poll Temporal client for workflow completion: `tc.GetWorkflow(ctx, workflowID, "").Get(ctx, nil)`
   - **Option B**: Sleep-based wait (simpler, works for non-deterministic workflow IDs)
   - **Option C**: Count completed workflows via Temporal visibility API
   - **Recommended**: Option B (sleep 2s) for simplicity, since these tests verify dispatch happened not execution results

4. **`createTriggerEvent`** — keep as-is

**Files**:
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_seed_test.go` (MODIFY: ~50 lines changed)

---

### Task 3: Rewrite `workflowsaveapi/trigger_test.go`

**Status**: Pending

**Description**: Update the `TriggerTestData` struct and trigger tests to use Temporal instead of EventPublisher/QueueManager/Engine verification.

**Current State** (525 lines):
- `TriggerTestData` struct: stores `EventPublisher *workflow.EventPublisher`, `DelegateHandler *workflow.DelegateHandler`
- `insertTriggerSeedData`: creates old `workflow.NewEventPublisher()` + `workflow.NewDelegateHandler()`
- `testCustomerCreateTriggersWorkflow`: verifies via `QueueManager.GetMetrics()`, `Engine.GetExecutionHistory()`
- `testCustomerUpdateTriggersWorkflow`: same verification pattern
- `testInactiveRuleNoTrigger`: verifies via `Engine.GetExecutionHistory()` + `countRuleMatches()`
- `countRuleMatches`: iterates `WorkflowExecution.BatchResults.RuleResults` (old engine types — DELETED in Phase 13)
- `waitForProcessing`: polls `QueueManager.GetMetrics()` (reused from execution_seed_test.go)

**Changes Required**:

1. **`TriggerTestData` struct** — replace:
   ```go
   // Old:
   EventPublisher  *workflow.EventPublisher
   DelegateHandler *workflow.DelegateHandler
   // New: (DelegateHandler already on WorkflowInfra)
   // Remove both fields — use wf.DelegateHandler directly
   ```

2. **`insertTriggerSeedData`** — replace:
   - Remove `workflow.NewEventPublisher()` + `workflow.NewDelegateHandler()`
   - Use `esd.WF.DelegateHandler.RegisterDomain(busDomain.Delegate, customersbus.DomainName, customersbus.EntityName)`
   - Call `esd.WF.TriggerProcessor.RefreshRules(ctx)` after creating rules

3. **Verification strategy** — replace Engine.GetExecutionHistory/QueueManager.GetMetrics with:
   - **For positive tests** (create/update trigger): Use `time.Sleep(3*time.Second)` then query `workflow.automation_executions` table via workflowBus, OR simply verify the Temporal workflow was dispatched by checking for workflow completion via Temporal client
   - **For negative test** (inactive rule): Verify no workflow dispatched by checking Temporal visibility API or DB execution records after a wait period
   - **Simplest approach**: Sleep-based wait + DB query for execution records (automation_executions table). This is the most reliable since it doesn't depend on Temporal SDK visibility features.

4. **`countRuleMatches`** — delete entirely (references deleted `WorkflowExecution`, `BatchResults`, `RuleResults` types)

5. **`waitForProcessing`** — rewrite to time-based wait (no QueueManager metrics available)

**Files**:
- `api/cmd/services/ichor/tests/workflow/workflowsaveapi/trigger_test.go` (REWRITE: ~200 lines changed)

---

### Task 4: Rewrite `formdata/formdataapi/workflow_test.go`

**Status**: Pending

**Description**: Update the formdata workflow test to use Temporal verification instead of QueueManager/Engine.

**Current State** (329 lines):
- `TestWorkflow_OrderCreateEvent`: creates rules, fires events via `wf.QueueManager.QueueEvent()`, verifies via `wf.QueueManager.GetMetrics()` + `wf.Engine.GetExecutionHistory()`
- `waitForProcessing`: local copy polling QueueManager.GetMetrics
- `verifyEmailActionExecuted`: iterates `WorkflowExecution.BatchResults.RuleResults.ActionResults` (old types)

**Changes Required**:

1. **Event dispatch** — replace `wf.QueueManager.QueueEvent(ctx, event)` with:
   ```go
   wf.WorkflowTrigger.OnEntityEvent(ctx, event)
   ```
   This directly dispatches to Temporal instead of queueing to RabbitMQ.

2. **Verification** — replace `wf.Engine.GetExecutionHistory()` with:
   - Sleep-based wait for Temporal workflow completion
   - OR: Query automation_executions table
   - OR: Verify Temporal workflow status via client
   - **Recommended**: Since these tests create rules then dispatch events directly, use `time.Sleep(3*time.Second)` then verify action side effects (e.g., alert created in DB for create_alert actions, or just verify no error from dispatch)

3. **`wf.Engine.Initialize()` calls** — replace with `wf.TriggerProcessor.RefreshRules(ctx)` after creating rules

4. **`waitForProcessing`** — rewrite to time-based wait

5. **`verifyEmailActionExecuted`** — delete (references deleted types). Replace with simpler verification (workflow dispatched without error)

**Files**:
- `api/cmd/services/ichor/tests/formdata/formdataapi/workflow_test.go` (REWRITE: ~150 lines changed)

---

### Task 5: Rewrite `sales/ordersapi/workflow_test.go`

**Status**: Pending

**Description**: Update the orders workflow test to use Temporal verification. This test is the most comprehensive — it tests create/update/delete delegate events end-to-end.

**Current State** (512 lines):
- `TestWorkflow_OrdersDelegateEvents`: creates 3 rules (create/update/delete), creates EventPublisher + DelegateHandler, verifies via QueueManager.GetMetrics + Engine.GetExecutionHistory
- `waitForProcessing`: local copy
- `verifyActionExecuted`: iterates old WorkflowExecution types

**Changes Required**:

1. **DelegateHandler setup** — replace:
   ```go
   // Old:
   eventPublisher := workflow.NewEventPublisher(db.Log, wf.QueueManager)
   delegateHandler := workflow.NewDelegateHandler(db.Log, eventPublisher)
   delegateHandler.RegisterDomain(...)
   // New:
   wf.DelegateHandler.RegisterDomain(db.BusDomain.Delegate, ordersbus.DomainName, ordersbus.EntityName)
   ```

2. **Engine re-init** — replace `wf.Engine.Initialize(ctx, wf.WorkflowBus)` with `wf.TriggerProcessor.RefreshRules(ctx)`

3. **Verification** — same strategy as Task 4: sleep-based wait + verify no dispatch error, or query DB for execution records

4. **`waitForProcessing`** — rewrite to time-based wait

5. **`verifyActionExecuted`** — delete (references deleted types)

6. **Remove imports**: `workflow.NewEventPublisher`, `workflow.NewDelegateHandler`, `workflow.QueueMetrics`, `workflow.WorkflowExecution`

**Files**:
- `api/cmd/services/ichor/tests/sales/ordersapi/workflow_test.go` (REWRITE: ~200 lines changed)

---

### Task 6: Add Worker Health Checks

**Status**: Pending

**Description**: Add HTTP health server to the workflow-worker service and update K8s manifest with liveness/readiness probes.

**Notes**:
- `/healthz/live` on port 4001 — always returns 200 (process alive)
- `/healthz/ready` on port 4001 — returns 200 when worker is started, 503 otherwise
- Run health server in background goroutine alongside Temporal worker
- Use atomic bool or channel to signal readiness after `worker.Start()` succeeds

**Implementation Guide for `main.go`**:

```go
// Add to main.go after worker creation, before worker.Start():

// Start health server.
ready := make(chan struct{})
go func() {
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })
    mux.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
        select {
        case <-ready:
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("ok"))
        default:
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte("not ready"))
        }
    })
    if err := http.ListenAndServe(":4001", mux); err != nil {
        log.Error(context.Background(), "health server error", "error", err)
    }
}()

// After worker.Start() succeeds:
close(ready)
```

**Implementation Guide for K8s manifest**:

Add to `base-workflow-worker.yaml` under containers[0]:

```yaml
ports:
  - name: health
    containerPort: 4001
livenessProbe:
  httpGet:
    path: /healthz/live
    port: 4001
  initialDelaySeconds: 5
  periodSeconds: 15
  timeoutSeconds: 3
  failureThreshold: 3
readinessProbe:
  httpGet:
    path: /healthz/ready
    port: 4001
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3
```

**Files**:
- `api/cmd/services/workflow-worker/main.go` (MODIFY: ~25 lines added)
- `zarf/k8s/base/workflow-worker/base-workflow-worker.yaml` (MODIFY: ~15 lines added)

---

### Task 7: End-to-End Verification

**Status**: Pending

**Description**: Run all verification commands to confirm everything works end-to-end.

**Automated Verification**:
```bash
go build ./...                                              # Full compilation
go vet ./...                                                # Static analysis
go test ./business/sdk/workflow/...                         # Surviving workflow tests
go test ./business/sdk/workflow/temporal/...                # All 155 Temporal tests
go test ./api/cmd/services/ichor/tests/workflow/...        # Integration tests
go test ./api/cmd/services/ichor/tests/formdata/...        # FormData workflow test
go test ./api/cmd/services/ichor/tests/sales/ordersapi/... # Orders workflow test
```

**Manual Verification** (requires KIND cluster):
```bash
make dev-bounce                                             # K8s deployment
kubectl get pods -n ichor-system | grep workflow-worker     # Verify pod running
kubectl exec -n ichor-system deploy/workflow-worker -- wget -qO- http://localhost:4001/healthz/live    # Liveness
kubectl exec -n ichor-system deploy/workflow-worker -- wget -qO- http://localhost:4001/healthz/ready   # Readiness
make dev-logs-workflow-worker                                # Check worker logs
```

**Files**: None (verification only)

---

## Validation Criteria

- [ ] `go build ./...` compiles cleanly
- [ ] `go vet ./...` passes
- [ ] `go test ./business/sdk/workflow/...` passes (surviving workflow tests)
- [ ] `go test ./business/sdk/workflow/temporal/...` passes (all 155 Temporal tests)
- [ ] `go test ./api/cmd/services/ichor/tests/workflow/...` passes (integration tests)
- [ ] `go test ./api/cmd/services/ichor/tests/formdata/formdataapi/...` passes (formdata workflow test)
- [ ] `go test ./api/cmd/services/ichor/tests/sales/ordersapi/...` passes (orders workflow test)
- [ ] Worker health checks respond on port 4001 (`/healthz/live` = 200, `/healthz/ready` = 200 after start)
- [ ] K8s manifest has liveness/readiness probes configured
- [ ] No remaining references to `QueueManager`, `Engine`, `EventPublisher`, `rabbitmq` in test infrastructure
- [ ] No remaining references to `WorkflowExecution`, `BatchResults`, `RuleResults`, `ActionResult` in test files

---

## Deliverables

- Rewritten `api/sdk/http/apitest/workflow.go` (~100 lines, Temporal-based)
- Rewritten `workflowsaveapi/execution_seed_test.go` (~50 lines changed)
- Rewritten `workflowsaveapi/trigger_test.go` (~200 lines changed)
- Rewritten `formdata/formdataapi/workflow_test.go` (~150 lines changed)
- Rewritten `sales/ordersapi/workflow_test.go` (~200 lines changed)
- Updated `api/cmd/services/workflow-worker/main.go` (health server ~25 lines)
- Updated `zarf/k8s/base/workflow-worker/base-workflow-worker.yaml` (probes ~15 lines)

---

## Gotchas & Tips

### Common Pitfalls

1. **Temporal test container is slow to start**: The first test in a run spins up the Docker container (~2-5s). Subsequent tests reuse the singleton. If tests timeout, increase the deadline.

2. **Task queue name conflicts**: Each test MUST use a unique task queue name (e.g., `test-workflow-` + `t.Name()`). If two tests share a task queue, their workers compete for tasks and tests become flaky.

3. **Rule cache invalidation**: After creating rules in a test, call `wf.TriggerProcessor.RefreshRules(ctx)` to load them into the trigger processor's cache. The old pattern of `wf.Engine.Initialize()` no longer works.

4. **`workflow.ResetEngineForTesting()` is deleted**: Phase 13 removed this function along with the engine singleton. Tests no longer need to reset global state — each test gets its own `TriggerProcessor` instance via `InitWorkflowInfra`.

5. **`workflow.QueueMetrics` type is deleted**: Any test code that references `workflow.QueueMetrics` as a parameter type in `waitForProcessing` must be updated. The new approach uses time-based waits or Temporal client queries.

6. **`workflow.WorkflowExecution` type is deleted**: All code that iterates `.BatchResults[].RuleResults[].ActionResults[]` must be removed. There is no equivalent in the Temporal model — verification uses workflow completion status or DB queries.

7. **`workflow.NewEventPublisher()` is deleted**: Tests that created their own EventPublisher (trigger_test.go, ordersapi/workflow_test.go) must use `wf.DelegateHandler.RegisterDomain()` instead.

8. **Three local `waitForProcessing` copies**: The function exists in `execution_seed_test.go`, `formdataapi/workflow_test.go`, and `ordersapi/workflow_test.go`. All three need updating independently (they're in different packages).

9. **Health check port 4001**: Must not conflict with any existing service ports. The main ichor service uses 3000 (API) and 4000 (debug). Port 4001 is safe.

10. **Import alias needed**: `foundation/temporal` vs `business/sdk/workflow/temporal` — use `foundationtemporal` for foundation and `temporal` (or `temporalpkg`) for business SDK, consistent with Phase 9's `all.go` pattern.

### Tips

- Start with Task 1 (apitest/workflow.go) — everything else depends on it compiling
- Run `go build ./api/sdk/http/apitest/...` after Task 1 before touching test files
- For verification in rewritten tests, simplest approach: dispatch event → `time.Sleep(3s)` → assert no panic/error. Full execution verification (action side effects) is nice-to-have but not required
- Task 6 (health checks) is independent of Tasks 1-5 — can be done in parallel
- Run `grep -r "QueueManager\|Engine\.\|EventPublisher\|WorkflowExecution\|BatchResults" api/cmd/services/ichor/tests/` after all rewrites to verify cleanup
- The `errors_test.go` file in workflowsaveapi does NOT use `InitWorkflowInfra` — no changes needed there

### Task Execution Order

1. **Task 1** (apitest/workflow.go) — must be first, others depend on it
2. **Task 6** (health checks) — independent, can run in parallel with Tasks 2-5
3. **Tasks 2-5** (test rewrites) — independent of each other, can be done in any order
4. **Task 7** (verification) — must be last

---

## Testing Strategy

### Test Infrastructure Tests

After rewriting `apitest/workflow.go`, verify the infrastructure itself works:

```bash
# Verify Temporal test container starts
go test -v ./foundation/temporal/... -run TestGetTestContainer

# Verify a simple integration test passes
go test -v ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -run TestSave -count=1
```

### Integration Test Verification

For each rewritten test file, run individually:

```bash
go test -v -timeout=120s ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -count=1
go test -v -timeout=120s ./api/cmd/services/ichor/tests/formdata/formdataapi/... -run TestWorkflow -count=1
go test -v -timeout=120s ./api/cmd/services/ichor/tests/sales/ordersapi/... -run TestWorkflow -count=1
```

### Race Detection

```bash
go test -race ./api/cmd/services/ichor/tests/workflow/...
go test -race ./api/cmd/services/ichor/tests/formdata/formdataapi/...
go test -race ./api/cmd/services/ichor/tests/sales/ordersapi/...
```

### Health Check Test

```bash
# Build worker
go build ./api/cmd/services/workflow-worker/...

# If KIND cluster running:
make dev-update-apply
kubectl exec -n ichor-system deploy/workflow-worker -- wget -qO- http://localhost:4001/healthz/live
```

---

## Commands Reference

```bash
# Start this phase
/workflow-temporal-next

# Validate this phase
/workflow-temporal-validate 15

# Review plan before implementing
/workflow-temporal-plan-review 15

# Review code after implementing
/workflow-temporal-review 15
```
