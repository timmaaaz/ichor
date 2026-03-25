# Phase 2: ActionService Orchestration Tests

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add unit tests for `ActionService` — the core orchestration layer that wires registry → validation → handler dispatch → execution recording.

**Architecture:** Uses `package workflow_test` (external test package, matching existing pattern in `workflow_crud_test.go`). Needs real Postgres via `dbtest` for execution recording. Uses a simple test action handler registered in the real `ActionRegistry`.

**Tech Stack:** Go testing, `dbtest`, `cmp.diff`, `unitest`

**Spec:** `docs/superpowers/specs/2026-03-24-workflow-test-gap-remediation-design.md` (Phase 2)

---

### Task 1: Test Handler and Setup

**Files:**
- Create: `business/sdk/workflow/actionservice_test.go`
- Reference: `business/sdk/workflow/actionservice.go`
- Reference: `business/sdk/workflow/interfaces.go` (ActionHandler interface)
- Reference: `business/sdk/workflow/workflow_crud_test.go` (test pattern)

- [ ] **Step 1: Read the ActionHandler interface**

Read `business/sdk/workflow/interfaces.go` to understand the full `ActionHandler` interface that the test handler must implement. Also check `ActionRegistry` methods in the same package.

- [ ] **Step 2: Create test file with test handler and seed data**

```go
package workflow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// testHandler is a minimal ActionHandler for testing ActionService.
// It records calls and can be configured to fail.
type testHandler struct {
	actionType     string
	isAsync        bool
	supportsManual bool
	validateErr    error
	executeErr     error
	executeResult  any
	lastExecCtx    workflow.ActionExecutionContext
	executeCalled  bool
}

func (h *testHandler) GetType() string                        { return h.actionType }
func (h *testHandler) IsAsync() bool                          { return h.isAsync }
func (h *testHandler) SupportsManualExecution() bool          { return h.supportsManual }
func (h *testHandler) GetDescription() string                 { return "test handler: " + h.actionType }
func (h *testHandler) Validate(config json.RawMessage) error  { return h.validateErr }
func (h *testHandler) GetOutputPorts() []workflow.OutputPort  { return nil }

func (h *testHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	h.executeCalled = true
	h.lastExecCtx = execCtx
	return h.executeResult, h.executeErr
}

func Test_ActionService(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ActionService")

	sd, err := insertActionServiceSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, executeTests(db, sd), "execute")
	unitest.Run(t, executeForAutomationTests(db, sd), "executeForAutomation")
	unitest.Run(t, listTests(db, sd), "list")
	unitest.Run(t, statusTests(db, sd), "status")
}

type actionServiceSeedData struct {
	Users []userbus.User
}

func insertActionServiceSeedData(busDomain dbtest.BusDomain) (actionServiceSeedData, error) {
	ctx := context.Background()

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return actionServiceSeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	return actionServiceSeedData{Users: users}, nil
}

// newTestService creates an ActionService with the given handlers registered.
func newTestService(db *dbtest.Database, handlers ...*testHandler) *workflow.ActionService {
	registry := workflow.NewActionRegistry()
	for _, h := range handlers {
		registry.Register(h)
	}
	return workflow.NewActionService(db.Log, db.DB, registry)
}
```

Note: Check what fields `dbtest.Database` exposes. You need `db.Log` (logger) and `db.DB` (*sqlx.DB). Read `business/sdk/dbtest/dbtest.go` to confirm field names.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./business/sdk/workflow/...`

---

### Task 2: Execute Tests

- [ ] **Step 1: Write execute subtests**

```go
func executeTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "happy-path",
			ExpFunc: func(ctx context.Context) any {
				return "completed"
			},
			CmpFunc: func(ctx context.Context, got any) string {
				handler := &testHandler{
					actionType:     "test_action",
					supportsManual: true,
					executeResult:  map[string]any{"ok": true},
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "test_action",
					Config:     json.RawMessage(`{"key":"value"}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err != nil {
					return fmt.Sprintf("execute error: %s", err)
				}
				if result.Status != "completed" {
					return fmt.Sprintf("expected status completed, got %s", result.Status)
				}
				if result.ExecutionID == uuid.Nil {
					return "execution ID should not be nil"
				}
				if !handler.executeCalled {
					return "handler.Execute was not called"
				}
				return ""
			},
		},
		{
			Name: "unknown-action-type",
			ExpFunc: func(ctx context.Context) any {
				return workflow.ErrActionNotFound
			},
			CmpFunc: func(ctx context.Context, got any) string {
				svc := newTestService(db) // empty registry

				_, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "nonexistent",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected error for unknown action type"
				}
				return ""
			},
		},
		{
			Name: "validation-failure",
			ExpFunc: func(ctx context.Context) any {
				return "failed"
			},
			CmpFunc: func(ctx context.Context, got any) string {
				handler := &testHandler{
					actionType:     "test_action",
					supportsManual: true,
					validateErr:    fmt.Errorf("bad config"),
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "test_action",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected validation error"
				}
				if result.Status != "failed" {
					return fmt.Sprintf("expected status failed, got %s", result.Status)
				}
				return ""
			},
		},
		{
			Name: "handler-execution-failure",
			ExpFunc: func(ctx context.Context) any {
				return "failed"
			},
			CmpFunc: func(ctx context.Context, got any) string {
				handler := &testHandler{
					actionType:     "test_fail",
					supportsManual: true,
					executeErr:     fmt.Errorf("handler exploded"),
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "test_fail",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected execution error"
				}
				if result.Status != "failed" {
					return fmt.Sprintf("expected status failed, got %s", result.Status)
				}
				return ""
			},
		},
		{
			Name: "manual-execution-not-supported",
			ExpFunc: func(ctx context.Context) any {
				return workflow.ErrManualExecutionNotSupported
			},
			CmpFunc: func(ctx context.Context, got any) string {
				handler := &testHandler{
					actionType:     "auto_only",
					supportsManual: false,
				}
				svc := newTestService(db, handler)

				_, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "auto_only",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected manual execution not supported error"
				}
				return ""
			},
		},
		{
			Name: "async-action-returns-queued",
			ExpFunc: func(ctx context.Context) any {
				return "queued"
			},
			CmpFunc: func(ctx context.Context, got any) string {
				handler := &testHandler{
					actionType:     "async_test",
					supportsManual: true,
					isAsync:        true,
					executeResult:  map[string]any{"task_id": "123"},
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "async_test",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err != nil {
					return fmt.Sprintf("unexpected error: %s", err)
				}
				if result.Status != "queued" {
					return fmt.Sprintf("expected status queued for async, got %s", result.Status)
				}
				return ""
			},
		},
	}
}
```

- [ ] **Step 2: Run execute tests**

Run: `go test ./business/sdk/workflow/... -run Test_ActionService/execute -v -count=1`

- [ ] **Step 3: Commit**

```
git add business/sdk/workflow/actionservice_test.go
git commit -m "test(actionservice): add Execute method tests"
```

---

### Task 3: ExecuteForAutomation and List/Status Tests

- [ ] **Step 1: Write ExecuteForAutomation subtests**

Per the spec: `ExecuteForAutomation` does NOT call `recordExecution`. Test that the `ActionExecutionContext` has the correct rule context.

```go
func executeForAutomationTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "automation-context-populated",
			ExpFunc: func(ctx context.Context) any {
				return "completed"
			},
			CmpFunc: func(ctx context.Context, got any) string {
				handler := &testHandler{
					actionType:     "auto_action",
					supportsManual: true,
					executeResult:  map[string]any{"done": true},
				}
				svc := newTestService(db, handler)

				ruleID := uuid.New()
				result, err := svc.ExecuteForAutomation(ctx, workflow.ExecuteRequest{
					ActionType: "auto_action",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, ruleID, "My Rule", "on_update")
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				if result.Status != "completed" {
					return fmt.Sprintf("expected completed, got %s", result.Status)
				}

				// Verify the execution context passed to the handler
				if handler.lastExecCtx.RuleID == nil || *handler.lastExecCtx.RuleID != ruleID {
					return fmt.Sprintf("expected RuleID %s, got %v", ruleID, handler.lastExecCtx.RuleID)
				}
				if handler.lastExecCtx.RuleName != "My Rule" {
					return fmt.Sprintf("expected RuleName 'My Rule', got %q", handler.lastExecCtx.RuleName)
				}
				if handler.lastExecCtx.EventType != "on_update" {
					return fmt.Sprintf("expected EventType on_update, got %q", handler.lastExecCtx.EventType)
				}
				if handler.lastExecCtx.TriggerSource != workflow.TriggerSourceAutomation {
					return fmt.Sprintf("expected TriggerSource automation, got %q", handler.lastExecCtx.TriggerSource)
				}
				return ""
			},
		},
		{
			Name: "automation-unknown-action",
			ExpFunc: func(ctx context.Context) any {
				return nil
			},
			CmpFunc: func(ctx context.Context, got any) string {
				svc := newTestService(db)
				_, err := svc.ExecuteForAutomation(ctx, workflow.ExecuteRequest{
					ActionType: "nonexistent",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, uuid.New(), "rule", "on_create")
				if err == nil {
					return "expected error for unknown action"
				}
				return ""
			},
		},
	}
}
```

- [ ] **Step 2: Write List and GetExecutionStatus subtests**

```go
func listTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "list-all",
			ExpFunc: func(ctx context.Context) any {
				return 2
			},
			CmpFunc: func(ctx context.Context, got any) string {
				manual := &testHandler{actionType: "manual_a", supportsManual: true}
				autoOnly := &testHandler{actionType: "auto_b", supportsManual: false}
				svc := newTestService(db, manual, autoOnly)

				all := svc.ListAvailableActions()
				if len(all) != 2 {
					return fmt.Sprintf("expected 2 actions, got %d", len(all))
				}
				return ""
			},
		},
		{
			Name: "list-manual-only",
			ExpFunc: func(ctx context.Context) any {
				return 1
			},
			CmpFunc: func(ctx context.Context, got any) string {
				manual := &testHandler{actionType: "manual_a", supportsManual: true}
				autoOnly := &testHandler{actionType: "auto_b", supportsManual: false}
				svc := newTestService(db, manual, autoOnly)

				manualActions := svc.ListManuallyExecutableActions()
				if len(manualActions) != 1 {
					return fmt.Sprintf("expected 1 manual action, got %d", len(manualActions))
				}
				if manualActions[0].Type != "manual_a" {
					return fmt.Sprintf("expected manual_a, got %s", manualActions[0].Type)
				}
				return ""
			},
		},
	}
}

func statusTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "get-execution-status-after-execute",
			ExpFunc: func(ctx context.Context) any {
				return "completed"
			},
			CmpFunc: func(ctx context.Context, got any) string {
				handler := &testHandler{
					actionType:     "status_test",
					supportsManual: true,
					executeResult:  map[string]any{"ok": true},
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "status_test",
					Config:     json.RawMessage(`{}`),
					EntityName: "test_entity",
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err != nil {
					return fmt.Sprintf("execute error: %s", err)
				}

				status, err := svc.GetExecutionStatus(ctx, result.ExecutionID)
				if err != nil {
					return fmt.Sprintf("status error: %s", err)
				}
				if status.Status != "completed" {
					return fmt.Sprintf("expected completed, got %s", status.Status)
				}
				if status.ExecutionID != result.ExecutionID {
					return "execution IDs don't match"
				}
				return ""
			},
		},
		{
			Name: "get-execution-status-not-found",
			ExpFunc: func(ctx context.Context) any {
				return nil
			},
			CmpFunc: func(ctx context.Context, got any) string {
				svc := newTestService(db)
				_, err := svc.GetExecutionStatus(ctx, uuid.New())
				if err == nil {
					return "expected error for nonexistent execution"
				}
				return ""
			},
		},
	}
}
```

- [ ] **Step 3: Run all ActionService tests**

Run: `go test ./business/sdk/workflow/... -run Test_ActionService -v -count=1`
Expected: All tests PASS.

- [ ] **Step 4: Commit**

```
git add business/sdk/workflow/actionservice_test.go
git commit -m "test(actionservice): add ExecuteForAutomation, List, and Status tests"
```
