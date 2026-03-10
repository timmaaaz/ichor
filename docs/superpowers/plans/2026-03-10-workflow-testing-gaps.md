# Workflow Testing Gaps Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore and extend workflow integration and unit test coverage after the Temporal migration, using real Temporal infrastructure for all pipeline tests.

**Architecture:** All Temporal-dependent tests use `apitest.InitWorkflowInfra` to spin up a real Temporal test container and worker. Tests fire events via `wf.WorkflowTrigger.OnEntityEvent()`, then poll the DB for side effects. Subtests within `Test_WorkflowSaveAPI` run sequentially and share one `WorkflowInfra` instance.

**Tech Stack:** Go 1.23, Temporal SDK, PostgreSQL (via dbtest), `apitest.Table` pattern for HTTP tests, `t.Run` for integration subtests.

**Design spec:** `docs/superpowers/specs/2026-03-10-workflow-testing-gaps-design.md`

---

## Chunk 1: Phase 1 — Wire trigger_test.go

### Context

`trigger_test.go` is fully written and compiles (no build tag). It defines `insertTriggerSeedData` and `runTriggerTests`. They just need to be called from `Test_WorkflowSaveAPI` in `save_test.go`.

`execution_seed_test.go` defines `insertExecutionSeedData` and `ExecutionTestData` (includes `WF *apitest.WorkflowInfra`). This must be called before `insertTriggerSeedData` since `TriggerTestData` embeds `ExecutionTestData`.

**Files:**
- Modify: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go`

### Task 1: Add trigger test wiring to Test_WorkflowSaveAPI

- [ ] **Step 1: Read save_test.go to confirm current structure**

  Open `save_test.go`. Confirm it ends with:
  ```go
  // Phase 13: Execution, trigger, action, and error tests excluded.
  // These tests depend on the old workflow.Engine which was removed.
  // Phase 15 will rewrite them for Temporal.
  ```

- [ ] **Step 2: Replace the Phase 13 comment with trigger test wiring**

  Replace the comment block at the end of `Test_WorkflowSaveAPI` with:
  ```go
  // ============================================================
  // Temporal integration tests (requires Temporal container)
  // ============================================================

  esd := insertExecutionSeedData(t, test, sd)
  tsd := insertTriggerSeedData(t, test, esd)
  runTriggerTests(t, tsd)
  ```

- [ ] **Step 3: Build to verify compilation**

  ```bash
  go build ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...
  ```
  Expected: no errors.

- [ ] **Step 4: Run only the trigger subtests**

  ```bash
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... \
    -run "Test_WorkflowSaveAPI/trigger" -v -timeout 120s
  ```
  Expected: 3 subtests pass:
  - `trigger-customer-create`
  - `trigger-customer-update`
  - `trigger-inactive-rule-no-trigger`

- [ ] **Step 5: Commit**

  ```bash
  git add api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go
  git commit -m "test(workflow): wire trigger_test.go into Test_WorkflowSaveAPI"
  ```

---

## Chunk 2: Phase 2 — Rewrite execution_test.go

### Context

`execution_test.go` has `//go:build ignore` and uses the deleted `Engine.ExecuteWorkflow()` API. It tests 7 scenarios. The new Temporal pattern:
1. `wf.TriggerProcessor.RefreshRules(ctx)` to load the rule
2. `wf.WorkflowTrigger.OnEntityEvent(ctx, event)` to dispatch
3. Poll `alertBus.Query()` for expected side effects (all seeded workflows use `create_alert`)

Drop `exec-record-created` and `exec-history-tracking` — these tested old in-memory Engine state (GetExecutionHistory, GetStats) which no longer exists.

The seeded workflows in `execution_seed_test.go`:
- `SimpleWorkflow`: 1 create_alert action, triggers on `sd.TriggerTypes[0]` + `sd.Entities[0]`
- `SequenceWorkflow`: 3 create_alert actions, triggers on `sd.TriggerTypes[1]` (on_update) + `sd.Entities[0]`
- `BranchingWorkflow`: evaluate_condition → true: "high_value" alert / false: "normal_value" alert

**Files:**
- Modify: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_test.go`

### Task 2: Rewrite execution_test.go for Temporal

- [ ] **Step 1: Expose AlertBus from WorkflowInfra (prerequisite)**

  `alertBus` is created as a local variable in `InitWorkflowInfra` but not returned. Add it to the struct and return it **before** writing any test code that references it.

  In `api/sdk/http/apitest/workflow.go`:
  ```go
  // Add AlertBus field to WorkflowInfra struct
  type WorkflowInfra struct {
      WorkflowBus        *workflow.Business
      TemporalClient     temporalclient.Client
      WorkflowTrigger    *temporal.WorkflowTrigger
      DelegateHandler    *temporal.DelegateHandler
      TriggerProcessor   *workflow.TriggerProcessor
      Worker             worker.Worker
      ApprovalRequestBus *approvalrequestbus.Business
      AlertBus           *alertbus.Business  // ← add this
  }

  // In InitWorkflowInfra return statement, add:
  return &WorkflowInfra{
      // ...existing fields...
      AlertBus: alertBus,  // ← add this
  }
  ```

  Build to verify:
  ```bash
  go build ./api/sdk/http/apitest/...
  ```

- [ ] **Step 2: Remove the build tag and old imports**

  Replace the file header:
  ```go
  //go:build ignore
  // +build ignore

  // Phase 13: Excluded until Phase 15 rewrites for Temporal.

  package workflowsaveapi_test
  ```
  With:
  ```go
  package workflowsaveapi_test
  ```

  Update imports — remove `"fmt"`, `"strings"`, add `"time"` and `alertbus` import:
  ```go
  import (
      "context"
      "testing"
      "time"

      "github.com/google/uuid"
      "github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
      "github.com/timmaaaz/ichor/business/sdk/page"
      "github.com/timmaaaz/ichor/business/sdk/workflow"
  )
  ```

- [ ] **Step 2: Rewrite runExecutionTests**

  Replace the entire `runExecutionTests` function:
  ```go
  func runExecutionTests(t *testing.T, sd ExecutionTestData) {
      t.Run("exec-single-alert", func(t *testing.T) {
          testExecuteSingleCreateAlert(t, sd)
      })
      t.Run("exec-sequence", func(t *testing.T) {
          testExecuteSequence3Actions(t, sd)
      })
      t.Run("exec-branch-true", func(t *testing.T) {
          testExecuteBranchTrue(t, sd)
      })
      t.Run("exec-branch-false", func(t *testing.T) {
          testExecuteBranchFalse(t, sd)
      })
      t.Run("exec-no-matching-rules", func(t *testing.T) {
          testNoMatchingRules(t, sd)
      })
  }
  ```

- [ ] **Step 3: Rewrite testExecuteSingleCreateAlert**

  ```go
  func testExecuteSingleCreateAlert(t *testing.T, sd ExecutionTestData) {
      ctx := context.Background()
      if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
          t.Fatal("insufficient seed data")
      }

      if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
          t.Fatalf("refreshing rules: %v", err)
      }

      event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, map[string]any{})
      if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
          t.Fatalf("firing trigger: %v", err)
      }

      // Poll alertbus for alert created by SimpleWorkflow's create_alert action.
      alertBus := alertbus.NewBusiness(sd.WF.WorkflowBus /* use db from test */)
      // NOTE: alertBus is already available via sd.WF - check WorkflowInfra for alertBus field.
      // If not, create it from db: alertBus := alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))
      // Use the same alertBus registered in InitWorkflowInfra.
      var found bool
      for i := 0; i < 30; i++ {
          alerts, err := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "20"))
          if err != nil {
              t.Fatalf("querying alerts: %v", err)
          }
          if len(alerts) > 0 {
              found = true
              break
          }
          time.Sleep(500 * time.Millisecond)
      }
      if !found {
          t.Fatal("timeout: no alert created after 15s — SimpleWorkflow may not have executed")
      }
      t.Log("SUCCESS: single create_alert workflow executed via Temporal")
  }
  ```

  > **Implementation note:** Check whether `WorkflowInfra` exposes `AlertBus`. If not, add it to `WorkflowInfra` struct in `apitest/workflow.go` and return it from `InitWorkflowInfra` (it's already created there as a local variable — just expose it).

- [ ] **Step 4: Rewrite testExecuteSequence3Actions**

  Same pattern as Step 3. Fire `sd.TriggerTypes[1].Name` (on_update) event for `sd.Entities[0]`. SequenceWorkflow creates 3 alerts — poll until count ≥ 3 (accounting for other tests' alerts by filtering on the rule ID or timing).

  > **Simplification:** Since alerts accumulate across subtests in the shared DB, compare alert count before and after dispatch to verify ≥ 3 new alerts were created.

  ```go
  func testExecuteSequence3Actions(t *testing.T, sd ExecutionTestData) {
      ctx := context.Background()

      // Count alerts before
      before, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
      beforeCount := len(before)

      if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
          t.Fatalf("refreshing rules: %v", err)
      }

      triggerType := sd.TriggerTypes[0].Name
      if len(sd.TriggerTypes) > 1 {
          triggerType = sd.TriggerTypes[1].Name
      }

      event := createTriggerEvent(sd.Entities[0].Name, triggerType, sd.Users[0].ID, map[string]any{})
      if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
          t.Fatalf("firing trigger: %v", err)
      }

      // Poll until ≥ 3 new alerts appear.
      for i := 0; i < 30; i++ {
          after, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
          if len(after)-beforeCount >= 3 {
              t.Log("SUCCESS: sequence created ≥3 alerts via Temporal")
              return
          }
          time.Sleep(500 * time.Millisecond)
      }
      t.Fatal("timeout: expected ≥3 new alerts from sequence workflow after 15s")
  }
  ```

- [ ] **Step 5: Rewrite testExecuteBranchTrue**

  Fire with `RawData: map[string]any{"amount": 1500}`. BranchingWorkflow routes to true branch (alert_type "high_value"). Poll for alert with `AlertType == "high_value"`.

- [ ] **Step 6: Rewrite testExecuteBranchFalse**

  Fire with `RawData: map[string]any{"amount": 500}`. Poll for alert with `AlertType == "normal_value"`.

- [ ] **Step 7: Rewrite testNoMatchingRules**

  ```go
  func testNoMatchingRules(t *testing.T, sd ExecutionTestData) {
      ctx := context.Background()

      before, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
      beforeCount := len(before)

      event := workflow.TriggerEvent{
          EventType:  "on_create",
          EntityName: "nonexistent_entity_xyz_" + uuid.New().String()[:8],
          EntityID:   uuid.New(),
          UserID:     sd.Users[0].ID,
      }

      // Should not error — no matching rules means no dispatch.
      if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
          t.Fatalf("unexpected error for no-match event: %v", err)
      }

      time.Sleep(2 * time.Second)

      after, _ := sd.WF.AlertBus.Query(ctx, alertbus.QueryFilter{}, alertbus.DefaultOrderBy, page.MustParse("1", "100"))
      if len(after) != beforeCount {
          t.Errorf("expected no new alerts for no-match event, got %d new", len(after)-beforeCount)
      }
      t.Log("SUCCESS: no matching rules fired no workflows")
  }
  ```

- [ ] **Step 8: Remove old helper functions**

  Delete `formatExecutionErrors` — it references the old `workflow.WorkflowExecution` type which no longer exists. If any other tests reference it, update them.

- [ ] **Step 9: Wire into Test_WorkflowSaveAPI**

  In `save_test.go`, add `runExecutionTests(t, esd)` after `insertExecutionSeedData`:
  ```go
  esd := insertExecutionSeedData(t, test, sd)
  runExecutionTests(t, esd)  // ← add this
  tsd := insertTriggerSeedData(t, test, esd)
  runTriggerTests(t, tsd)
  ```

- [ ] **Step 10: Expose AlertBus from WorkflowInfra if needed**

  If `WorkflowInfra` doesn't expose `AlertBus`, add it:
  - `api/sdk/http/apitest/workflow.go`: add `AlertBus *alertbus.Business` field to `WorkflowInfra` struct, populate in `InitWorkflowInfra` (already created as local `alertBus`).

- [ ] **Step 11: Build and run**

  ```bash
  go build ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... \
    -run "Test_WorkflowSaveAPI/exec" -v -timeout 120s
  ```
  Expected: 5 exec subtests pass.

- [ ] **Step 12: Commit**

  ```bash
  git add api/cmd/services/ichor/tests/workflow/workflowsaveapi/execution_test.go \
          api/sdk/http/apitest/workflow.go \
          api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go
  git commit -m "test(workflow): rewrite execution_test.go for Temporal"
  ```

---

## Chunk 3: Phase 3a — Actions: alert + approval

### Context

`actions_test.go` has `//go:build ignore`. The existing tests use `Engine.Execute` + `Engine.Initialize`. New pattern: use `TriggerProcessor.RefreshRules` + `WorkflowTrigger.OnEntityEvent` then poll alertBus.

Keep from actions_test.go:
- `testCreateAlertBasic` → poll alertBus for `alert_type: "basic_test"`
- `testCreateAlertWithRecipients` → poll alertBus for new alert
- `testCreateAlertTemplateVars` → poll alertBus (template resolution happens server-side)
- `testCreateAlertSeverityLevels` → poll alertBus for severity-specific alerts (4 subtests)
- `testConditionEqualsTrue` → poll alertBus for `alert_type: "condition_true"`
- `testConditionEqualsFalse` → poll alertBus for `alert_type: "condition_false"`
- `testConditionGreaterThan` → poll alertBus for `alert_type: "high_amount"`
- `testConditionMultipleAnd` → poll alertBus for `alert_type: "all_conditions_met"`

Drop:
- `testSendEmailBasic` / `testSendEmailMultipleRecipients` — move to Phase 3c
- Result data assertions like `actionResult.ResultData["alert_id"]` — not available via Temporal; DB record existence is sufficient

The `seek_approval` tests from Phase 3a are covered by the standalone `TestSeekApproval_Approved/Rejected` tests already passing. No new seek_approval tests needed here.

**Files:**
- Modify: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go`

### Task 3: Rewrite actions_test.go (alert + condition tests)

- [ ] **Step 1: Remove the build tag**

  Remove:
  ```go
  //go:build ignore
  // +build ignore

  // Phase 13: Excluded until Phase 15 rewrites for Temporal.
  ```

- [ ] **Step 2: Update imports**

  Replace old imports with:
  ```go
  import (
      "context"
      "encoding/json"
      "testing"
      "time"

      "github.com/google/uuid"
      "github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
      "github.com/timmaaaz/ichor/business/sdk/page"
      "github.com/timmaaaz/ichor/business/sdk/workflow"
  )
  ```

- [ ] **Step 3: Rewrite runActionTests**

  ```go
  func runActionTests(t *testing.T, sd ExecutionTestData) {
      t.Run("action-alert-basic", func(t *testing.T) {
          testCreateAlertBasic(t, sd)
      })
      t.Run("action-alert-with-recipients", func(t *testing.T) {
          testCreateAlertWithRecipients(t, sd)
      })
      t.Run("action-alert-template-vars", func(t *testing.T) {
          testCreateAlertTemplateVars(t, sd)
      })
      t.Run("action-alert-severity-levels", func(t *testing.T) {
          testCreateAlertSeverityLevels(t, sd)
      })
      t.Run("action-condition-equals-true", func(t *testing.T) {
          testConditionEqualsTrue(t, sd)
      })
      t.Run("action-condition-equals-false", func(t *testing.T) {
          testConditionEqualsFalse(t, sd)
      })
      t.Run("action-condition-greater-than", func(t *testing.T) {
          testConditionGreaterThan(t, sd)
      })
      t.Run("action-condition-multiple-and", func(t *testing.T) {
          testConditionMultipleAnd(t, sd)
      })
  }
  ```

- [ ] **Step 4: Rewrite each test function using the Temporal pattern**

  For each test, replace the old `Engine.Initialize` + `Engine.ExecuteWorkflow` + result traversal with:

  ```go
  // Pattern for every action test:

  // 1. Seed rule + action + edge (same as before)
  // 2. Replace Engine.Initialize with:
  if err := sd.WF.TriggerProcessor.RefreshRules(ctx); err != nil {
      t.Fatalf("refreshing rules: %v", err)
  }
  // 3. Replace Engine.ExecuteWorkflow with:
  event := createTriggerEvent(sd.Entities[0].Name, sd.TriggerTypes[0].Name, sd.Users[0].ID, rawData)
  if err := sd.WF.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
      t.Fatalf("firing trigger: %v", err)
  }
  // 4. Replace result traversal with alert DB poll:
  var found bool
  for i := 0; i < 30; i++ {
      alerts, _ := sd.WF.AlertBus.Query(ctx,
          alertbus.QueryFilter{AlertType: stringPtr("basic_test")},
          alertbus.DefaultOrderBy,
          page.MustParse("1", "5"),
      )
      if len(alerts) > 0 {
          found = true
          break
      }
      time.Sleep(500 * time.Millisecond)
  }
  if !found {
      t.Fatal("timeout: alert not created after 15s")
  }
  ```

  Alert types to poll for (positive assertion — wait for alert with this type):
  | Test | alert_type to poll |
  |------|--------------------|
  | testCreateAlertBasic | "basic_test" |
  | testCreateAlertWithRecipients | "multi_recipient_test" |
  | testCreateAlertTemplateVars | "template_test" |
  | testCreateAlertSeverityLevels | "severity_test" (per iteration) |
  | testConditionEqualsTrue | "condition_true" |
  | testConditionGreaterThan | "high_amount" |
  | testConditionMultipleAnd | "all_conditions_met" |

  **False-branch tests (negative assertion):** For `testConditionEqualsFalse`, the false branch creates an alert with `alert_type: "condition_false"`. Poll for that alert type the same way. Do NOT attempt to assert the absence of the true branch — it's simpler and equally correct to assert the false-branch alert exists.

  For `testConditionEqualsFalse` the RawData should be `{"status": "inactive"}` so the `equals "active"` condition evaluates false, triggering the false branch which creates a `"condition_false"` alert. Poll for that alert type.

- [ ] **Step 5: Add strPtr helper if not already present**

  The old file uses `strPtr("true")` for edge `SourceOutput`. Confirm this helper exists in the package or add it:
  ```go
  func strPtr(s string) *string { return &s }
  ```

- [ ] **Step 6: Wire into Test_WorkflowSaveAPI**

  In `save_test.go`, add after `runExecutionTests`:
  ```go
  runActionTests(t, esd)
  ```

- [ ] **Step 7: Build and run**

  ```bash
  go build ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... \
    -run "Test_WorkflowSaveAPI/action" -v -timeout 180s
  ```
  Expected: 8 action subtests pass (4 alert + 4 condition). `severity-levels` expands to 4 sub-subtests.

- [ ] **Step 8: Commit**

  ```bash
  git add api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_test.go \
          api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go
  git commit -m "test(workflow): rewrite actions_test.go (alert + condition) for Temporal"
  ```

---

## Chunk 4: Phases 3b + 3c — Actions: inventory/procurement + comms/webhook

### Context

These tests require handlers not registered in `InitWorkflowInfra`. Two approaches:
- **3b (receive_inventory, create_purchase_order):** Bus dependencies required → standalone tests with local Temporal worker setup, similar to `approvalapi/approval_test.go`
- **3c (send_email, send_notification, call_webhook):** `send_email` and `send_notification` are already in `InitWorkflowInfra` (nil clients, graceful degrade). `call_webhook` needs to be added (no bus deps).

### Task 4a: Phase 3b — Standalone inventory + procurement action tests

**Files:**
- Create: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_inventory_test.go`

- [ ] **Step 1: Create the file with package declaration and local infra helper**

  These tests do NOT use `InitWorkflowInfra`. They set up a minimal local Temporal worker:
  ```go
  package workflowsaveapi_test

  import (
      "context"
      "encoding/json"
      "testing"
      "time"

      "github.com/google/uuid"
      "github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
      "github.com/timmaaaz/ichor/business/sdk/dbtest"
      "github.com/timmaaaz/ichor/business/sdk/page"
      "github.com/timmaaaz/ichor/business/sdk/workflow"
      "github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
      "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
      "github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
      "github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
      foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
      temporalclient "go.temporal.io/sdk/client"
      "go.temporal.io/sdk/worker"
  )
  ```

- [ ] **Step 2: Write TestReceiveInventoryAction**

  Pattern (mirrors `approvalapi/approval_test.go`):
  ```go
  func TestReceiveInventoryAction(t *testing.T) {
      t.Parallel()

      db := dbtest.NewDatabase(t, "Test_ReceiveInventoryAction")

      // Set up minimal Temporal infra with only receive_inventory registered.
      container := foundationtemporal.GetTestContainer(t)
      tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
      // ... (error handling)

      taskQueue := "test-workflow-" + t.Name()
      w := worker.New(tc, taskQueue, worker.Options{})
      w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
      w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)

      registry := workflow.NewActionRegistry()
      registry.Register(inventory.NewReceiveInventoryHandler(
          db.Log, db.DB,
          db.BusDomain.InventoryItem,         // inventoryItemBus
          db.BusDomain.InventoryTransaction,   // inventoryTransactionBus
          db.BusDomain.SupplierProduct,        // supplierProductBus
      ))

      activities := &temporal.Activities{Registry: registry, AsyncRegistry: temporal.NewAsyncRegistry()}
      w.RegisterActivity(activities)
      if err := w.Start(); err != nil { t.Fatalf("starting worker: %v", err) }
      t.Cleanup(func() { w.Stop(); tc.Close() })

      // Create workflow infra
      workflowStore := workflowdb.NewStore(db.Log, db.DB)
      workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
      edgeStore := edgedb.NewStore(db.Log, db.DB)
      triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
      // NOTE: Do NOT call triggerProcessor.Initialize here.
      // Initialize only once at startup for production. In tests, call RefreshRules
      // AFTER seeding the rule so the rule is picked up.
      workflowTrigger := temporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
          WithTaskQueue(taskQueue)

      ctx := context.Background()

      // Seed: find an inventory item to receive against.
      // (Use db.BusDomain.InventoryItem.Query to find a seeded item)
      items, err := db.BusDomain.InventoryItem.Query(ctx, /* filter */)
      // ...

      // Create rule with receive_inventory action.
      // Seed entity = "inventory_items" or appropriate entity name.
      // ...

      // RefreshRules AFTER seeding so the new rule is picked up.
      if err := triggerProcessor.RefreshRules(ctx); err != nil {
          t.Fatalf("refreshing rules: %v", err)
      }

      // Fire event.
      if err := workflowTrigger.OnEntityEvent(ctx, event); err != nil {
          t.Fatalf("firing trigger: %v", err)
      }

      // Poll: verify inventory item quantity increased.
      time.Sleep(3 * time.Second)
      updatedItem, err := db.BusDomain.InventoryItem.QueryByID(ctx, items[0].ID)
      // Assert quantity changed.
  }
  ```

  > **Implementation note:** Check `db.BusDomain` for the correct field names for `InventoryItemBus`, `InventoryTransactionBus`, and `SupplierProductBus`. Run `grep -r "InventoryItem" business/sdk/dbtest/` to find field names.

- [ ] **Step 3: Write TestCreatePurchaseOrderAction**

  Same pattern as Step 2 but with `procurement.NewCreatePurchaseOrderHandler(...)` registered. Assert a new purchase order row exists in DB after workflow fires.

  > **Implementation note:** Check `db.BusDomain` for `PurchaseOrderBus`, `PurchaseOrderLineItemBus`. The handler requires `supplierProductBus` (already found in Step 2 investigation).

- [ ] **Step 4: Build and run**

  ```bash
  go build ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... \
    -run "TestReceiveInventoryAction|TestCreatePurchaseOrderAction" -v -timeout 120s
  ```
  Expected: both tests pass.

- [ ] **Step 5: Commit**

  ```bash
  git add api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_inventory_test.go
  git commit -m "test(workflow): add standalone inventory + procurement action tests"
  ```

### Task 4b: Phase 3c — Add call_webhook to InitWorkflowInfra, standalone comms tests

**Files:**
- Modify: `api/sdk/http/apitest/workflow.go`
- Create: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_comms_test.go`

- [ ] **Step 6: Add call_webhook to InitWorkflowInfra**

  In `workflow.go`, add `integration` package import and register handler:
  ```go
  import "github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/integration"

  // In InitWorkflowInfra, after existing registry.Register calls:
  registry.Register(integration.NewCallWebhookHandler(db.Log))
  ```

  Build to verify:
  ```bash
  go build ./api/sdk/http/apitest/...
  ```

- [ ] **Step 7: Write TestCallWebhookAction**

  Use `apitest.InitWorkflowInfra` (now includes call_webhook). Set up an `httptest.NewServer` to receive the webhook:
  ```go
  func TestCallWebhookAction(t *testing.T) {
      t.Parallel()
      db := dbtest.NewDatabase(t, "Test_CallWebhookAction")
      wf := apitest.InitWorkflowInfra(t, db)
      ctx := context.Background()

      // Start a local test server to receive the webhook.
      received := make(chan struct{}, 1)
      server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          received <- struct{}{}
          w.WriteHeader(http.StatusOK)
      }))
      t.Cleanup(server.Close)

      // Create rule with call_webhook action pointing at test server.
      // (standard rule + action + edge seeding pattern)
      // ...
      if err := wf.TriggerProcessor.RefreshRules(ctx); err != nil { ... }
      if err := wf.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil { ... }

      select {
      case <-received:
          t.Log("SUCCESS: webhook received")
      case <-time.After(15 * time.Second):
          t.Fatal("timeout: webhook not received after 15s")
      }
  }
  ```

  > **Important:** `call_webhook` validates HTTPS URLs only. For the test server (HTTP), you'll need to either use `https` via `httptest.NewTLSServer` OR check if `CallWebhookHandler` has an option to allow HTTP in tests. Check `webhook.go`'s `Validate()` function.

- [ ] **Step 8: Write TestSendEmailAction and TestSendNotificationAction**

  These use `InitWorkflowInfra` (nil client / nil queue). Assertions verify the workflow completes without error since there's no DB side effect:
  ```go
  func TestSendEmailAction(t *testing.T) {
      t.Parallel()
      db := dbtest.NewDatabase(t, "Test_SendEmailAction")
      wf := apitest.InitWorkflowInfra(t, db)
      ctx := context.Background()

      // Seed rule with send_email action (nil client = graceful degrade)
      // Fire event, wait 3 seconds, verify no panics/test failures.
      // No DB assertion possible (no record written on nil client).
      time.Sleep(3 * time.Second)
      t.Log("SUCCESS: send_email with nil client degrades gracefully")
  }
  ```

- [ ] **Step 9: Build and run**

  ```bash
  go build ./api/sdk/http/apitest/... ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... \
    -run "TestCallWebhookAction|TestSendEmailAction|TestSendNotificationAction" -v -timeout 120s
  ```

- [ ] **Step 10: Commit**

  ```bash
  git add api/sdk/http/apitest/workflow.go \
          api/cmd/services/ichor/tests/workflow/workflowsaveapi/actions_comms_test.go
  git commit -m "test(workflow): add call_webhook to infra, standalone comms action tests"
  ```

---

## Chunk 5: Phase 4 — Rewrite errors_test.go

### Context

`errors_test.go` has `//go:build ignore`. Many tests used `Engine.GetExecutionHistory()`, `Engine.GetStats()`, `QueueManager.GetMetrics()` — all deleted.

Keep and rewrite:
- `error-action-fails-sequence-stops` → use `simulate_failure: true` flag in evaluate_condition config to cause failure, verify other actions in sequence still run (Temporal retries 3x by default)
- `error-condition-field-not-found` → fire event without the referenced field → verify workflow completes (no panic, no unhandled error)
- `error-condition-type-mismatch` → fire with wrong type → verify workflow completes
- `error-no-actions-defined` → fire on rule with no start edge → verify no crash
- `error-inactive-action-skipped` → active + inactive actions in sequence → verify only active ones create alerts

Drop (incompatible with Temporal model):
- `error-action-timeout` (Temporal handles internally)
- `error-concurrent-triggers-same-rule` (tested old Engine concurrency, not Temporal)
- `error-queue-retry-success` (QueueManager deleted)

**Files:**
- Modify: `api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go`

### Task 5: Rewrite errors_test.go for Temporal

- [ ] **Step 1: Remove the build tag and update imports**

  Remove `//go:build ignore` block. Remove `"sync"`. Add `alertbus` and `page` imports. Remove `"strings"` and `"fmt"`.

- [ ] **Step 2: Rewrite runErrorTests**

  ```go
  func runErrorTests(t *testing.T, sd ExecutionTestData) {
      t.Run("error-action-fails-sequence-stops", func(t *testing.T) {
          testActionFailsSequenceStops(t, sd)
      })
      t.Run("error-condition-field-not-found", func(t *testing.T) {
          testConditionFieldNotFound(t, sd)
      })
      t.Run("error-condition-type-mismatch", func(t *testing.T) {
          testConditionTypeMismatch(t, sd)
      })
      t.Run("error-no-actions-defined", func(t *testing.T) {
          testNoActionsDefined(t, sd)
      })
      t.Run("error-inactive-action-skipped", func(t *testing.T) {
          testInactiveActionSkipped(t, sd)
      })
  }
  ```

- [ ] **Step 3: Rewrite testActionFailsSequenceStops**

  Replace `Engine.Initialize` + `Engine.ExecuteWorkflow` with `RefreshRules` + `OnEntityEvent`.

  Action 2 uses `evaluate_condition` with deliberately invalid config (missing required `conditions` field). This causes the handler's `Validate()` to return an error, which Temporal treats as an activity failure and retries up to 3x (MaximumAttempts=3).

  ```go
  // Action 2: force failure via invalid config (missing required 'conditions' field)
  action2, err := sd.WF.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
      AutomationRuleID: rule.ID,
      Name:             "Action 2 - Invalid Config",
      ActionConfig:     json.RawMessage(`{"invalid_field": "this should fail validation"}`),
      IsActive:         true,
      TemplateID:       &sd.EvaluateConditionTemplate.ID,
  })
  ```

  After firing event, wait 10 seconds (retries may delay completion) then assert:
  - Action 1's alert (`alert_type: "test"`) was created ✓
  - Log action 3 status — don't hard-fail on this since Temporal retry semantics determine whether the sequence continues past a failed action

  > **Note:** Temporal retries failed activities 3x. After retries exhaust, the workflow continues to the next action in some configurations or halts depending on edge routing. The key assertion is that action 1 succeeded; action 3 behavior is informational.

- [ ] **Step 4: Rewrite testConditionFieldNotFound, testConditionTypeMismatch**

  Replace old Engine calls. New assertions: fire event, wait 3 seconds, no panic, no unhandled errors. Verify no new alert was created for the condition action's downstream branch (since missing field → condition evaluates to false → no branch taken).

- [ ] **Step 5: Rewrite testNoActionsDefined**

  Create rule with no actions (no `CreateRuleAction` calls, no edges). Fire event. Wait 2 seconds. Assert no crash and no new alerts from this rule.

  > **Note:** Temporal may not dispatch a workflow at all if there are no edges from the start. This is fine — the test just verifies graceful handling.

- [ ] **Step 6: Rewrite testInactiveActionSkipped**

  Create: start → action1(active) → action2(inactive) → action3(active). Fire event. Assert action1's alert created, action2's alert NOT created, action3's alert created (sequence continues past inactive action at the edge level).

  > **Note:** Whether action3 executes depends on how the edge store handles inactive actions. If the edge is filtered out at graph load time, action3 may not execute either. Verify behavior empirically and adjust assertions.

- [ ] **Step 7: Remove inapplicable functions**

  Delete: `testActionTimeout`, `testConcurrentTriggersSameRule`, `testQueueRetrySuccess`. Also delete `formatExecutionErrors` if it exists in this file (it was in execution_test.go originally).

- [ ] **Step 8: Wire into Test_WorkflowSaveAPI**

  In `save_test.go`, add `runErrorTests(t, esd)` after `runActionTests`.

- [ ] **Step 9: Build and run**

  ```bash
  go build ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/...
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... \
    -run "Test_WorkflowSaveAPI/error" -v -timeout 120s
  ```
  Expected: 5 error subtests pass.

- [ ] **Step 10: Run the full suite to check for regressions**

  ```bash
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -v -timeout 300s
  ```

- [ ] **Step 11: Commit**

  ```bash
  git add api/cmd/services/ichor/tests/workflow/workflowsaveapi/errors_test.go \
          api/cmd/services/ichor/tests/workflow/workflowsaveapi/save_test.go
  git commit -m "test(workflow): rewrite errors_test.go for Temporal"
  ```

---

## Chunk 6: Phases 5 + 6 — actionapi happy path + business layer unit tests

### Task 6: Phase 5 — actionapi execute200

**Context:** `Test_ActionAPI` currently has only error-path tests. `UserWithAlertPerm` has `create_alert` and `send_notification` permissions (from seed_test.go). The `execute200` test POSTs a valid create_alert config as this user and asserts 200 + response shape.

For status200, the cleanest approach is a sequential function in `Test_ActionAPI` (not a table test) that calls execute, extracts the ID, then calls status.

**Files:**
- Modify: `api/cmd/services/ichor/tests/workflow/actionapi/action_test.go`
- Modify: `api/cmd/services/ichor/tests/workflow/actionapi/execute_test.go`

- [ ] **Step 1: Add execute200 table helper to execute_test.go**

  First check what `actionapp.ExecuteResponse` looks like:
  ```bash
  grep -r "ExecuteResponse" app/domain/workflow/actionapp/
  ```

  Add to `execute_test.go`:
  ```go
  func execute200CreateAlert(sd ActionSeedData) []apitest.Table {
      config := map[string]any{
          "alert_type": "manual_execute_test",
          "severity":   "low",
          "title":      "Manual Execute Test",
          "message":    "Executed manually via test",
          "recipients": map[string]any{
              "users": []string{sd.UserWithAlertPerm.User.ID.String()},
              "roles": []string{},
          },
      }
      configBytes, _ := json.Marshal(config)

      return []apitest.Table{
          {
              Name:       "user-with-alert-perm-executes-create-alert",
              URL:        "/v1/workflow/actions/create_alert/execute",
              Token:      sd.UserWithAlertPerm.Token,
              StatusCode: http.StatusOK,
              Method:     http.MethodPost,
              Input: &actionapp.ExecuteRequest{
                  Config: configBytes,
              },
              GotResp: &actionapp.ExecuteResponse{},
              ExpResp: &actionapp.ExecuteResponse{},
              CmpFunc: func(got any, exp any) string {
                  gotResp := got.(*actionapp.ExecuteResponse)
                  if gotResp.ExecutionID == uuid.Nil.String() || gotResp.ExecutionID == "" {
                      return "expected non-empty execution_id in response"
                  }
                  return ""
              },
          },
      }
  }
  ```

  > **Note:** Check the actual field name in `actionapp.ExecuteResponse` — it may be `ExecutionID uuid.UUID` or `ExecutionID string`. Adjust accordingly.

- [ ] **Step 2: Wire execute200 into Test_ActionAPI**

  In `action_test.go`, add before the status tests:
  ```go
  test.Run(t, execute200CreateAlert(sd), "execute-200-create-alert")
  ```

- [ ] **Step 3: Skip getExecutionStatus200 (covered by table infrastructure)**

  The `execute200CreateAlert` table test in Step 1 already covers the happy path and verifies a non-empty `ExecutionID` in the response. A separate status200 test would require extracting the execution ID from a live HTTP response and making a second chained call — this is not supported by the `apitest.Table` infrastructure (single-shot, no state between table rows).

  **Decision:** Do not add `testExecuteAndCheckStatus200` to this plan. The table entry in Step 1 is sufficient for the happy-path goal. If a status200 chained test is needed in the future, it would require a raw `httptest` client setup analogous to Phase 3b/3c standalone tests.

- [ ] **Step 4: Build and run**

  ```bash
  go build ./api/cmd/services/ichor/tests/workflow/actionapi/...
  go test ./api/cmd/services/ichor/tests/workflow/actionapi/... \
    -run "Test_ActionAPI" -v -timeout 120s
  ```
  Expected: all existing tests + new `execute-200-create-alert` pass.

- [ ] **Step 5: Commit**

  ```bash
  git add api/cmd/services/ichor/tests/workflow/actionapi/execute_test.go \
          api/cmd/services/ichor/tests/workflow/actionapi/status_test.go \
          api/cmd/services/ichor/tests/workflow/actionapi/action_test.go
  git commit -m "test(workflow): add execute200 happy path to Test_ActionAPI"
  ```

### Task 7: Phase 6 — Business layer unit tests

**Context:** Pure Go unit tests, no DB, no Temporal. Add to existing test files in `business/sdk/workflow/`.

**Files:**
- Modify: `business/sdk/workflow/trigger_test.go` (or create new file if it doesn't exist)
- Modify: `business/sdk/workflow/expr_test.go`

- [ ] **Step 1: Check existing trigger test file**

  ```bash
  ls business/sdk/workflow/*_test.go
  ```

  Check what tests already exist. The PLAN notes `temporal/trigger_test.go` (15 tests) and `temporal/delegatehandler_test.go` (2 tests) exist. We need tests for `business/sdk/workflow/trigger.go` (the `TriggerProcessor`).

- [ ] **Step 2: Verify expr_test.go before adding anything**

  > **IMPORTANT:** Before writing any expr tests, verify what already exists:
  ```bash
  grep -n "func Test" business/sdk/workflow/expr_test.go
  ```
  The existing `TestEvalExpr` table-driven test already covers: `"division by zero"`, `"modulo by zero"`, `"division by zero literal"`, `"unknown variable"`, `"missing closing paren"`, `"unexpected character"`. **Do NOT add duplicate tests for these cases** — they will fail due to `t.Parallel()` registration conflicts.

- [ ] **Step 3: Add TriggerProcessor condition evaluation unit tests**

  The genuine gap is in `business/sdk/workflow/trigger.go`: `evaluateFieldCondition` for `contains`, `starts_with`, `ends_with` operators and mixed `and`/`or` condition logic. These are tested at the engine level but not at the `TriggerProcessor` level.

  Check existing files first:
  ```bash
  ls business/sdk/workflow/*_test.go
  grep -n "TestTriggerProcessor\|evaluateField\|evaluateRule" business/sdk/workflow/*_test.go 2>/dev/null
  ```

  If `trigger_test.go` does NOT already have these, create or modify `business/sdk/workflow/trigger_test.go`:
  ```go
  package workflow_test

  import (
      "testing"
      "github.com/timmaaaz/ichor/business/sdk/workflow"
  )

  // TestEvaluateFieldCondition_StringOperators tests contains/starts_with/ends_with
  // which are in trigger.go's evaluateFieldCondition but not covered in other unit tests.
  func TestEvaluateFieldCondition_StringOperators(t *testing.T) {
      t.Parallel()

      tests := []struct {
          name      string
          operator  string
          value     string
          fieldVal  any
          wantMatch bool
      }{
          {"contains match", "contains", "foo", "foobar", true},
          {"contains no match", "contains", "baz", "foobar", false},
          {"starts_with match", "starts_with", "foo", "foobar", true},
          {"starts_with no match", "starts_with", "bar", "foobar", false},
          {"ends_with match", "ends_with", "bar", "foobar", true},
          {"ends_with no match", "ends_with", "foo", "foobar", false},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              cond := workflow.Condition{
                  Field:    "test_field",
                  Operator: tt.operator,
                  Value:    tt.value,
              }
              data := map[string]any{"test_field": tt.fieldVal}
              got := workflow.EvaluateFieldCondition(cond, data)
              if got != tt.wantMatch {
                  t.Errorf("EvaluateFieldCondition(%q, %q, %q) = %v, want %v",
                      tt.operator, tt.value, tt.fieldVal, got, tt.wantMatch)
              }
          })
      }
  }
  ```

  > **Implementation note:** `EvaluateFieldCondition` (or `evaluateFieldCondition`) may be unexported. Check whether it is exported in `trigger.go`. If unexported, either: (a) export it for testability, or (b) test it indirectly by calling `TriggerProcessor.evaluateRuleConditions` through a DB-backed integration test using `dbtest.NewDatabase`. Adjust test approach based on what's actually exported.

- [ ] **Step 4: Add mixed AND/OR logic test**

  ```go
  // TestEvaluateRuleConditions_MixedLogic tests AND + OR condition group evaluation.
  // If the function is unexported, this test exercises it via an exported method.
  func TestEvaluateRuleConditions_MixedLogic(t *testing.T) {
      t.Parallel()
      // AND group: both conditions must match
      andGroup := []workflow.Condition{
          {Field: "status", Operator: "equals", Value: "active", LogicalOp: "and"},
          {Field: "amount", Operator: "greater_than", Value: "100", LogicalOp: "and"},
      }
      data := map[string]any{"status": "active", "amount": float64(200)}
      // Verify AND logic: both must pass
      // (call exported EvaluateConditions or equivalent)
  }
  ```

  > **Note:** Read `trigger.go` to understand the exported surface before writing this test. If condition evaluation is entirely internal, skip this test and add a comment explaining why.

- [ ] **Step 5: Run unit tests**

  ```bash
  go test ./business/sdk/workflow/... -v -run "TestEvalExpr|TestEvaluateField|TestEvaluateRule"
  ```
  Expected: all new tests pass (existing expr tests must still pass).

- [ ] **Step 6: Commit**

  ```bash
  git add business/sdk/workflow/trigger_test.go
  git commit -m "test(workflow): add TriggerProcessor field condition unit tests"
  ```

---

## Final Verification

- [ ] **Run the full workflowsaveapi suite**

  ```bash
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... -v -timeout 300s
  ```

- [ ] **Run actionapi suite**

  ```bash
  go test ./api/cmd/services/ichor/tests/workflow/actionapi/... -v -timeout 120s
  ```

- [ ] **Run workflow unit tests**

  ```bash
  go test ./business/sdk/workflow/... -v -timeout 60s
  ```

- [ ] **Run standalone Temporal action tests**

  ```bash
  go test ./api/cmd/services/ichor/tests/workflow/workflowsaveapi/... \
    -run "TestReceiveInventoryAction|TestCreatePurchaseOrderAction|TestCallWebhookAction|TestSendEmailAction|TestSendNotificationAction" \
    -v -timeout 120s
  ```
