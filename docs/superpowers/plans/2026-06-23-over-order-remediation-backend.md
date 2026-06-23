# Over-Order Remediation Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make over-orders operator-actionable end to end in the workflow engine — ship a default remediation graph (alert + approval), add a re-run-execution endpoint, give over-orders a first-class alert type, fix the approval heartbeat bug, and emit a typed approval-created WebSocket message.

**Architecture:** Backend/workflow-engine only. The configurable engine already routes on typed output-port edges; we wire the unwired `reserve_inventory.insufficient_stock` port in the production seed, add an operator re-run path that re-fires a single rule with a fresh execution-id (clearing three dedup walls), and make supporting fixes in the alert handler, the approval activity options, and the WebSocket emit path.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, Temporal, Ardan Labs service architecture (bus/app/api/db/sdk layers), RabbitMQ (alert queue), integration-test-primary (dbtest + apitest harness).

## Global Constraints

- **Spec:** `docs/superpowers/specs/2026-06-23-over-order-remediation-backend-design.md` — the authority for all decisions.
- **Never run `go test ./...`** — run only the changed package(s), e.g. `go test ./business/sdk/workflow/temporal/... -run TestX -v`.
- **Always `go build` affected packages** before claiming a task done.
- **No new migration** — `alert_type` is free `VARCHAR(100)`; do NOT add a severity value.
- **`docs/arch/*` are authoritative** — read `workflow-engine.md`, `workflow-alerts.md`, `auth.md` as needed; never guess signatures, grep/read first.
- **Money/decimal, delegate, layer-purity** rules per CLAUDE.md still apply.
- **Commit after each task** with a conventional message; end commit messages with the `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>` trailer.
- **Sequential approval design** (not parallel fan-out) — confirmed in spec §3 ★.

---

### Task 1: Fix the approval heartbeat timeout (Deliverable D)

The async-completion path returns `activity.ErrResultPending` with no heartbeating goroutine (activities.go:210), but `activityOptions` sets `HeartbeatTimeout = 1h` for human actions (workflow.go:655-659). Temporal heartbeat-times-out the held approval after ~1h and orphans it. Remove the heartbeat timeout; the 7-day `StartToCloseTimeout` is the real bound.

**Files:**
- Modify: `business/sdk/workflow/temporal/workflow.go:655-659`
- Test: `business/sdk/workflow/temporal/workflow_test.go` (add a test; create the file only if it doesn't exist — first `ls` the dir)

**Interfaces:**
- Consumes: existing `activityOptions(actionType string) workflow.ActivityOptions`
- Produces: no API change — behavior change only (`HeartbeatTimeout == 0` for human actions)

- [ ] **Step 1: Write the failing test**

Add to `business/sdk/workflow/temporal/workflow_test.go` (package `temporal`):

```go
func TestActivityOptions_HumanAction_NoHeartbeatTimeout(t *testing.T) {
	ao := activityOptions("seek_approval")

	// Human actions hold for days via async completion; they must NOT set a
	// heartbeat timeout (the async activity never heartbeats — see activities.go).
	if ao.HeartbeatTimeout != 0 {
		t.Fatalf("seek_approval HeartbeatTimeout = %v, want 0", ao.HeartbeatTimeout)
	}
	// The real bound stays the 7-day start-to-close.
	if ao.StartToCloseTimeout != 7*24*time.Hour {
		t.Fatalf("seek_approval StartToCloseTimeout = %v, want 168h", ao.StartToCloseTimeout)
	}
	if ao.RetryPolicy.MaximumAttempts != 1 {
		t.Fatalf("seek_approval MaximumAttempts = %d, want 1", ao.RetryPolicy.MaximumAttempts)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./business/sdk/workflow/temporal/... -run TestActivityOptions_HumanAction_NoHeartbeatTimeout -v`
Expected: FAIL — `HeartbeatTimeout = 1h0m0s, want 0`

- [ ] **Step 3: Remove the heartbeat timeout**

In `business/sdk/workflow/temporal/workflow.go`, delete the `ao.HeartbeatTimeout = time.Hour` line inside the `if isHumanAction(actionType)` block (currently line 657). The block becomes:

```go
	if isHumanAction(actionType) {
		ao.StartToCloseTimeout = 7 * 24 * time.Hour // 7 days
		ao.RetryPolicy.MaximumAttempts = 1
	}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./business/sdk/workflow/temporal/... -run TestActivityOptions_HumanAction_NoHeartbeatTimeout -v`
Expected: PASS

- [ ] **Step 5: Build + commit**

```bash
go build ./business/sdk/workflow/temporal/...
git add business/sdk/workflow/temporal/workflow.go business/sdk/workflow/temporal/workflow_test.go
git commit -m "fix(workflow): drop heartbeat timeout on human-action activities

Async-completion approvals return ErrResultPending with no heartbeating
goroutine, so a 1h HeartbeatTimeout orphaned any hold > ~1h. The 7-day
StartToCloseTimeout is the real bound.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: Enrich create_alert with execution_id / rule_id (Deliverable C)

Make `execution_id` and `rule_id` available to alert templates and inject them into the alert `context` JSON, so any alert is deep-linkable to its execution. Extract two pure helpers so the logic is unit-testable without a DB/alertBus.

**Files:**
- Modify: `business/sdk/workflow/workflowactions/communication/alert.go` (add helpers; wire into `Execute` ~lines 114-145)
- Test: `business/sdk/workflow/workflowactions/communication/alert_enrich_test.go` (new)

**Interfaces:**
- Produces:
  - `func buildAlertTemplateData(rawData map[string]any, execID uuid.UUID, ruleID *uuid.UUID) map[string]any`
  - `func enrichAlertContext(ctx json.RawMessage, execID uuid.UUID, ruleID *uuid.UUID) (json.RawMessage, error)`
- Consumes: `ActionExecutionContext{ExecutionID uuid.UUID, RuleID *uuid.UUID, RawData map[string]any}` (models.go:65-78)

- [ ] **Step 1: Write the failing tests**

Create `business/sdk/workflow/workflowactions/communication/alert_enrich_test.go` (package `communication`):

```go
package communication

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestBuildAlertTemplateData_AddsIDsAndCopiesRawData(t *testing.T) {
	execID := uuid.New()
	ruleID := uuid.New()
	raw := map[string]any{"product_id": "abc", "quantity": float64(5)}

	out := buildAlertTemplateData(raw, execID, &ruleID)

	if out["execution_id"] != execID.String() {
		t.Fatalf("execution_id = %v, want %s", out["execution_id"], execID)
	}
	if out["rule_id"] != ruleID.String() {
		t.Fatalf("rule_id = %v, want %s", out["rule_id"], ruleID)
	}
	if out["product_id"] != "abc" || out["quantity"] != float64(5) {
		t.Fatalf("raw data not copied: %+v", out)
	}
	// Must NOT mutate the caller's map.
	if _, ok := raw["execution_id"]; ok {
		t.Fatal("buildAlertTemplateData mutated the input RawData map")
	}
}

func TestBuildAlertTemplateData_NilRuleAndRawData(t *testing.T) {
	execID := uuid.New()
	out := buildAlertTemplateData(nil, execID, nil)
	if out["execution_id"] != execID.String() {
		t.Fatalf("execution_id missing")
	}
	if _, ok := out["rule_id"]; ok {
		t.Fatal("rule_id should be absent when ruleID is nil")
	}
}

func TestEnrichAlertContext_MergesIntoEmptyObject(t *testing.T) {
	execID := uuid.New()
	ruleID := uuid.New()

	out, err := enrichAlertContext(json.RawMessage(`{}`), execID, &ruleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if m["execution_id"] != execID.String() || m["rule_id"] != ruleID.String() {
		t.Fatalf("ids not merged: %+v", m)
	}
}

func TestEnrichAlertContext_PreservesExistingKeys(t *testing.T) {
	execID := uuid.New()
	out, err := enrichAlertContext(json.RawMessage(`{"foo":"bar"}`), execID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]any
	_ = json.Unmarshal(out, &m)
	if m["foo"] != "bar" {
		t.Fatalf("existing key dropped: %+v", m)
	}
	if m["execution_id"] != execID.String() {
		t.Fatalf("execution_id not added: %+v", m)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./business/sdk/workflow/workflowactions/communication/... -run 'TestBuildAlertTemplateData|TestEnrichAlertContext' -v`
Expected: FAIL — `undefined: buildAlertTemplateData` / `enrichAlertContext`

- [ ] **Step 3: Implement the helpers**

Add to `business/sdk/workflow/workflowactions/communication/alert.go` (the package already imports `encoding/json` and `github.com/google/uuid`):

```go
// buildAlertTemplateData returns a copy of rawData with execution_id and
// rule_id added, so {{execution_id}} / {{rule_id}} resolve in alert templates.
// It never mutates the caller's map.
func buildAlertTemplateData(rawData map[string]any, execID uuid.UUID, ruleID *uuid.UUID) map[string]any {
	out := make(map[string]any, len(rawData)+2)
	for k, v := range rawData {
		out[k] = v
	}
	if execID != uuid.Nil {
		out["execution_id"] = execID.String()
	}
	if ruleID != nil {
		out["rule_id"] = ruleID.String()
	}
	return out
}

// enrichAlertContext merges execution_id and rule_id into the alert's context
// JSON object so the frontend can deep-link an alert to its execution.
func enrichAlertContext(ctx json.RawMessage, execID uuid.UUID, ruleID *uuid.UUID) (json.RawMessage, error) {
	m := map[string]any{}
	if len(ctx) > 0 {
		if err := json.Unmarshal(ctx, &m); err != nil {
			return nil, fmt.Errorf("parse alert context: %w", err)
		}
	}
	if execID != uuid.Nil {
		m["execution_id"] = execID.String()
	}
	if ruleID != nil {
		m["rule_id"] = ruleID.String()
	}
	return json.Marshal(m)
}
```

- [ ] **Step 4: Wire the helpers into `Execute`**

In `business/sdk/workflow/workflowactions/communication/alert.go` `Execute`, replace the context default block (lines 114-118) and the template-var calls (lines 131-144) so the augmented data + enriched context are used. After computing `sourceRuleID` (the code already builds `execCtx.RuleID` handling at lines 120-124), do:

```go
	// Augment template data + context with execution_id / rule_id (deep-linking).
	tmplData := buildAlertTemplateData(execCtx.RawData, execCtx.ExecutionID, execCtx.RuleID)

	context := cfg.Context
	if len(context) == 0 {
		context = json.RawMessage(`{}`)
	}
	enrichedContext, err := enrichAlertContext(context, execCtx.ExecutionID, execCtx.RuleID)
	if err != nil {
		return nil, err
	}
```

Then in the `alertbus.Alert{...}` literal, change the three `resolveTemplateVars(cfg.X, execCtx.RawData)` calls to use `tmplData`, and set `Context: enrichedContext`:

```go
		Title:   resolveTemplateVars(cfg.Title, tmplData),
		Message: resolveTemplateVars(cfg.Message, tmplData),
		Context: enrichedContext,
```

And the ActionURL block (lines 143-145):

```go
	if cfg.ActionURL != "" {
		alert.ActionURL = resolveTemplateVars(cfg.ActionURL, tmplData)
	}
```

- [ ] **Step 5: Run the tests + build**

Run: `go test ./business/sdk/workflow/workflowactions/communication/... -run 'TestBuildAlertTemplateData|TestEnrichAlertContext' -v`
Expected: PASS
Run: `go build ./business/sdk/workflow/workflowactions/communication/...`
Expected: no errors. Also run any existing alert tests: `go test ./business/sdk/workflow/workflowactions/communication/... -v` and confirm none regressed.

- [ ] **Step 6: Commit**

```bash
git add business/sdk/workflow/workflowactions/communication/alert.go business/sdk/workflow/workflowactions/communication/alert_enrich_test.go
git commit -m "feat(workflow): expose execution_id/rule_id to create_alert templates+context

Every alert becomes deep-linkable to its execution. Pure helpers
buildAlertTemplateData + enrichAlertContext are unit-tested; Execute wires them.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: Typed approval_request WebSocket emit (Deliverable E)

Approval resolution emits a typed `approval_resolved` WS message but creation only emits a generic `alert`. Add a typed `approval_request` (created) message so supervisors get a real-time "new approval pending" event.

**Files:**
- Modify: `foundation/websocket/message.go:11-22` (new const)
- Modify: `api/domain/http/workflow/alertws/consumer.go:70-79` (new case)
- Modify: `business/sdk/workflow/workflowactions/approval/seek.go` (new `buildApprovalRequestMessages` + `publishApprovalRequest`, called after the existing `PublishAlertToRecipients` ~line 284)
- Test: `business/sdk/workflow/workflowactions/approval/seek_wsemit_test.go` (new), and a case in the existing alertws consumer test (or a new `consumer_test.go`)

**Interfaces:**
- Produces:
  - `websocket.MessageTypeApprovalRequest MessageType = "approval_request"`
  - `func buildApprovalRequestMessages(approvalID, ruleID uuid.UUID, actionName string, approvers []uuid.UUID) []*rabbitmq.Message`
- Consumes: `rabbitmq.Message{Type, EntityName, EntityID, UserID, Payload}` (foundation/rabbitmq/client.go:284-298); `h.workflowQueue *rabbitmq.WorkflowQueue` (already a field on the seek handler)

- [ ] **Step 1: Write the failing message-builder test**

Create `business/sdk/workflow/workflowactions/approval/seek_wsemit_test.go` (package `approval`):

```go
package approval

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuildApprovalRequestMessages_OnePerApprover(t *testing.T) {
	approvalID := uuid.New()
	ruleID := uuid.New()
	a1, a2 := uuid.New(), uuid.New()

	msgs := buildApprovalRequestMessages(approvalID, ruleID, "approval_hold", []uuid.UUID{a1, a2})

	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2 (one per approver)", len(msgs))
	}
	for i, m := range msgs {
		if m.Type != "approval_request" {
			t.Fatalf("msg[%d].Type = %q, want approval_request", i, m.Type)
		}
		if m.EntityID != approvalID {
			t.Fatalf("msg[%d].EntityID = %v, want approval id", i, m.EntityID)
		}
		if m.Payload["approvalId"] != approvalID.String() {
			t.Fatalf("msg[%d] payload approvalId = %v", i, m.Payload["approvalId"])
		}
	}
	if msgs[0].UserID != a1 || msgs[1].UserID != a2 {
		t.Fatalf("per-approver UserID targeting wrong: %v, %v", msgs[0].UserID, msgs[1].UserID)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./business/sdk/workflow/workflowactions/approval/... -run TestBuildApprovalRequestMessages_OnePerApprover -v`
Expected: FAIL — `undefined: buildApprovalRequestMessages`

- [ ] **Step 3: Add the MessageType constant**

In `foundation/websocket/message.go`, add (after `MessageTypeApprovalResolved`):

```go
	MessageTypeApprovalRequest MessageType = "approval_request"
```

- [ ] **Step 4: Implement the builder + publisher in seek.go**

Add to `business/sdk/workflow/workflowactions/approval/seek.go` (mirror `approvalapi.publishApprovalResolved` / `buildApprovalResolvedPayload`; confirm the `rabbitmq` import + `QueueTypeAlert` are already used in the package — they are, via `communication.PublishAlertToRecipients` and `h.workflowQueue`):

```go
// buildApprovalRequestMessages builds one typed "approval_request" WS message
// per approver so each is targeted via msg.UserID (BroadcastToUser path).
func buildApprovalRequestMessages(approvalID, ruleID uuid.UUID, actionName string, approvers []uuid.UUID) []*rabbitmq.Message {
	msgs := make([]*rabbitmq.Message, 0, len(approvers))
	for _, approverID := range approvers {
		msgs = append(msgs, &rabbitmq.Message{
			Type:       "approval_request",
			EntityName: "workflow.approval_requests",
			EntityID:   approvalID,
			UserID:     approverID,
			Payload: map[string]any{
				"approvalId": approvalID.String(),
				"ruleId":     ruleID.String(),
				"actionName": actionName,
			},
		})
	}
	return msgs
}

// publishApprovalRequest emits the typed creation event (best-effort, mirrors
// publishApprovalResolved). Nil queue (Temporal/MQ disabled) is a no-op.
func (h *SeekApprovalHandler) publishApprovalRequest(ctx context.Context, approvalID, ruleID uuid.UUID, actionName string, approvers []uuid.UUID) {
	if h.workflowQueue == nil {
		return
	}
	for _, msg := range buildApprovalRequestMessages(approvalID, ruleID, actionName, approvers) {
		if err := h.workflowQueue.Publish(ctx, rabbitmq.QueueTypeAlert, msg); err != nil {
			h.log.Error(ctx, "failed to publish approval_request event", "approval_id", approvalID, "error", err)
		}
	}
}
```

> Implementer note: confirm the handler struct type name (`SeekApprovalHandler` per the registered handler) and the exact field names for the rule id / approvers / action name available at the call site (the approval request was just created — reuse its `ID`, the rule id from `execCtx.RuleID`, the action name from `execCtx.ActionName`, and the parsed `cfg`/`req` approvers UUIDs). Read seek.go around the `PublishAlertToRecipients` call (~line 284) before wiring.

Then call it right after the existing `communication.PublishAlertToRecipients(...)` (~line 284), passing the created approval's ID, the rule id, action name, and the approver UUIDs.

- [ ] **Step 5: Add the consumer mapping + its test**

In `api/domain/http/workflow/alertws/consumer.go` `messageTypeForAlert`, add a case before `default`:

```go
	case "approval_request":
		return websocket.MessageTypeApprovalRequest
```

Add a consumer mapping test (extend the existing alertws consumer test file, or create `api/domain/http/workflow/alertws/consumer_maptype_test.go`, package `alertws`):

```go
func TestMessageTypeForAlert_ApprovalRequest(t *testing.T) {
	if got := messageTypeForAlert("approval_request"); got != websocket.MessageTypeApprovalRequest {
		t.Fatalf("messageTypeForAlert(approval_request) = %v, want MessageTypeApprovalRequest", got)
	}
}
```

- [ ] **Step 6: Run tests + build**

Run: `go test ./business/sdk/workflow/workflowactions/approval/... -run TestBuildApprovalRequestMessages_OnePerApprover -v`
Run: `go test ./api/domain/http/workflow/alertws/... -run TestMessageTypeForAlert_ApprovalRequest -v`
Run: `go build ./business/sdk/workflow/workflowactions/approval/... ./api/domain/http/workflow/alertws/... ./foundation/websocket/...`
Expected: PASS + clean build.

- [ ] **Step 7: Commit**

```bash
git add foundation/websocket/message.go api/domain/http/workflow/alertws/consumer.go api/domain/http/workflow/alertws/consumer_maptype_test.go business/sdk/workflow/workflowactions/approval/seek.go business/sdk/workflow/workflowactions/approval/seek_wsemit_test.go
git commit -m "feat(workflow): emit typed approval_request WS message on approval creation

Closes the created-vs-resolved asymmetry: supervisors get a real-time typed
'new approval pending' event, one per approver. Additive; shared publish.go
unchanged. FE consumes the new type string.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: RerunExecution on WorkflowTrigger (Deliverable B — engine core)

Add the dispatch primitive: widen the trigger's `ExecutionStore` with a read-by-id, add `reconstructTriggerEvent` (reverses `buildTriggerData`), and `RerunExecution` (loads the execution, rebuilds the event, dispatches the single originating rule with a fresh execution-id via the existing `startWorkflowForRule`).

**Files:**
- Modify: `business/sdk/workflow/temporal/trigger.go` (widen `ExecutionStore` interface at 25-33; add `reconstructTriggerEvent` + `RerunExecution` + `ErrExecutionNotRerunnable`)
- Test: `business/sdk/workflow/temporal/rerun_test.go` (new)

**Interfaces:**
- Produces:
  - `ExecutionStore` gains `QueryExecutionByID(ctx context.Context, id uuid.UUID) (workflow.AutomationExecution, error)`
  - `func (t *WorkflowTrigger) RerunExecution(ctx context.Context, executionID uuid.UUID) (uuid.UUID, error)`
  - `var ErrExecutionNotRerunnable = errors.New("execution has no automation rule to re-run")`
- Consumes: `workflow.AutomationExecution{ID, AutomationRuleID *uuid.UUID, RuleName, EntityType, TriggerData json.RawMessage, ...}` (models.go:344-358); `workflow.TriggerEvent` + `FieldChange` (models.go:14-36); `workflow.RuleMatchResult{Rule workflow.AutomationRuleView}` (trigger.go:38-46); private `startWorkflowForRule(ctx, event, rm, lineage)`; `buildTriggerData`'s key contract (trigger.go:344-371); `CascadeLineageKey`

- [ ] **Step 1: Write the failing tests**

Create `business/sdk/workflow/temporal/rerun_test.go` (package `temporal`). The package already has fakes for `WorkflowStarter`/`EdgeStore`/`ExecutionStore`/`RuleMatcher` in `trigger_test.go` — **first read `trigger_test.go` and reuse those fakes**; the snippet below assumes fakes named `fakeStarter`, `fakeEdgeStore`, `fakeExecStore` with the obvious recording behavior. If a fake lacks `QueryExecutionByID`, add it there.

```go
func TestReconstructTriggerEvent_ReversesBuildTriggerData(t *testing.T) {
	entityID := uuid.New()
	userID := uuid.New()
	orig := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "order_line_items",
		EntityID:   entityID,
		UserID:     userID,
		RawData:    map[string]any{"product_id": "p1", "quantity": float64(7), "order_id": "o1"},
		FieldChanges: map[string]workflow.FieldChange{
			"quantity": {OldValue: float64(0), NewValue: float64(7)},
		},
	}
	// Round-trip through the persisted shape.
	td := buildTriggerData(orig)
	td[CascadeLineageKey] = WorkflowLineage{} // present in persisted JSON; must be ignored
	raw, _ := json.Marshal(td)

	got, err := reconstructTriggerEvent(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.EntityID != entityID || got.EntityName != "order_line_items" || got.EventType != "on_create" {
		t.Fatalf("metadata wrong: %+v", got)
	}
	if got.UserID != userID {
		t.Fatalf("user id wrong: %v", got.UserID)
	}
	if got.EventID != uuid.Nil {
		t.Fatalf("EventID must be zero for a fresh dedup key, got %v", got.EventID)
	}
	if got.RawData["product_id"] != "p1" || got.RawData["quantity"] != float64(7) {
		t.Fatalf("raw data not recovered: %+v", got.RawData)
	}
	if _, ok := got.RawData[CascadeLineageKey]; ok {
		t.Fatal("CascadeLineageKey leaked into RawData")
	}
	if fc, ok := got.FieldChanges["quantity"]; !ok || fc.NewValue != float64(7) {
		t.Fatalf("field changes not recovered: %+v", got.FieldChanges)
	}
}

func TestRerunExecution_FreshIDAndDispatch(t *testing.T) {
	ruleID := uuid.New()
	origExecID := uuid.New()
	td, _ := json.Marshal(buildTriggerData(workflow.TriggerEvent{
		EventType: "on_create", EntityName: "order_line_items", EntityID: uuid.New(),
		RawData: map[string]any{"product_id": "p1", "quantity": float64(7)},
	}))

	execStore := &fakeExecStore{
		byID: map[uuid.UUID]workflow.AutomationExecution{
			origExecID: {ID: origExecID, AutomationRuleID: &ruleID, RuleName: "Granular Inventory Pipeline", EntityType: "order_line_items", TriggerData: td},
		},
	}
	starter := &fakeStarter{}
	edges := &fakeEdgeStore{ /* return one start action + node so the graph is non-empty */ }

	tr := newTestTrigger(t, starter, edges, execStore) // helper in trigger_test.go; adapt as needed

	newID, err := tr.RerunExecution(context.Background(), origExecID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newID == origExecID || newID == uuid.Nil {
		t.Fatalf("expected a fresh execution id, got %v (orig %v)", newID, origExecID)
	}
	if !starter.executeWorkflowCalled {
		t.Fatal("expected ExecuteWorkflow to be called")
	}
	if execStore.lastCreated.ID != newID {
		t.Fatalf("execution record not created with the fresh id: %v", execStore.lastCreated.ID)
	}
}

func TestRerunExecution_NoRule_Errors(t *testing.T) {
	origExecID := uuid.New()
	execStore := &fakeExecStore{byID: map[uuid.UUID]workflow.AutomationExecution{
		origExecID: {ID: origExecID, AutomationRuleID: nil}, // manual execution
	}}
	tr := newTestTrigger(t, &fakeStarter{}, &fakeEdgeStore{}, execStore)

	_, err := tr.RerunExecution(context.Background(), origExecID)
	if !errors.Is(err, ErrExecutionNotRerunnable) {
		t.Fatalf("err = %v, want ErrExecutionNotRerunnable", err)
	}
}
```

> Implementer note: the exact fake names/constructors come from `trigger_test.go`. If there's no `newTestTrigger` helper, construct `&WorkflowTrigger{...}` directly with the fakes (the struct has unexported fields but the test is in-package). The `fakeEdgeStore` must return a non-empty graph (≥1 action) or `startWorkflowForRule` skips dispatch (trigger.go:191-198).

- [ ] **Step 2: Run to verify failure**

Run: `go test ./business/sdk/workflow/temporal/... -run 'TestReconstructTriggerEvent|TestRerunExecution' -v`
Expected: FAIL — undefined `reconstructTriggerEvent`, `RerunExecution`, `ErrExecutionNotRerunnable`, and `QueryExecutionByID` not in `ExecutionStore`.

- [ ] **Step 3: Widen the ExecutionStore interface**

In `business/sdk/workflow/temporal/trigger.go` (interface at 25-33):

```go
type ExecutionStore interface {
	CreateExecution(ctx context.Context, exec workflow.AutomationExecution) error
	DeleteExecution(ctx context.Context, id uuid.UUID) error
	QueryExecutionByID(ctx context.Context, id uuid.UUID) (workflow.AutomationExecution, error)
}
```

The concrete store injected in `all.go` (`workflowStore`) already implements `QueryExecutionByID` (it backs `workflow.Business.QueryExecutionByID`, workflowbus.go:1237) — confirm by building; no store code needed. If the injected concrete type is a narrower wrapper, point it at the workflowdb store's method.

- [ ] **Step 4: Implement reconstructTriggerEvent + RerunExecution**

Add to `business/sdk/workflow/temporal/trigger.go` (ensure `errors`, `time`, `encoding/json` imported):

```go
// ErrExecutionNotRerunnable is returned when an execution has no automation
// rule to re-fire (e.g. a manual execution).
var ErrExecutionNotRerunnable = errors.New("execution has no automation rule to re-run")

// reconstructTriggerEvent reverses buildTriggerData from a persisted
// execution's trigger_data. EventID is left zero so the dispatch mints a fresh
// dedup key, and Timestamp is re-stamped (the stored value uses time.Time.String()).
func reconstructTriggerEvent(triggerData json.RawMessage) (workflow.TriggerEvent, error) {
	var m map[string]any
	if err := json.Unmarshal(triggerData, &m); err != nil {
		return workflow.TriggerEvent{}, fmt.Errorf("parse trigger_data: %w", err)
	}

	ev := workflow.TriggerEvent{Timestamp: time.Now()}
	if s, ok := m["event_type"].(string); ok {
		ev.EventType = s
	}
	if s, ok := m["entity_name"].(string); ok {
		ev.EntityName = s
	}
	if s, ok := m["entity_id"].(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			ev.EntityID = id
		}
	}
	if s, ok := m["user_id"].(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			ev.UserID = id
		}
	}
	if fcRaw, ok := m["field_changes"].(map[string]any); ok {
		fc := make(map[string]workflow.FieldChange, len(fcRaw))
		for field, v := range fcRaw {
			if inner, ok := v.(map[string]any); ok {
				fc[field] = workflow.FieldChange{OldValue: inner["old_value"], NewValue: inner["new_value"]}
			}
		}
		if len(fc) > 0 {
			ev.FieldChanges = fc
		}
	}

	// RawData = everything except the metadata + cascade-lineage keys.
	raw := make(map[string]any, len(m))
	for k, v := range m {
		switch k {
		case "event_type", "entity_name", "entity_id", "user_id", "timestamp", "field_changes", CascadeLineageKey:
			continue
		}
		raw[k] = v
	}
	if len(raw) > 0 {
		ev.RawData = raw
	}
	return ev, nil
}

// RerunExecution re-fires the single rule that produced executionID against a
// fresh event reconstructed from its trigger_data, with a brand-new execution
// id (clears the allocation_results idempotency key, the Temporal workflow-id
// REJECT_DUPLICATE guard, and the execution-record upsert). Returns the new id.
func (t *WorkflowTrigger) RerunExecution(ctx context.Context, executionID uuid.UUID) (uuid.UUID, error) {
	exec, err := t.executionStore.QueryExecutionByID(ctx, executionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("load execution %s: %w", executionID, err)
	}
	if exec.AutomationRuleID == nil {
		return uuid.Nil, ErrExecutionNotRerunnable
	}

	event, err := reconstructTriggerEvent(exec.TriggerData)
	if err != nil {
		return uuid.Nil, fmt.Errorf("reconstruct event: %w", err)
	}

	rm := workflow.RuleMatchResult{
		Rule: workflow.AutomationRuleView{ID: *exec.AutomationRuleID, Name: exec.RuleName},
	}

	// startWorkflowForRule mints executionID := uuid.New() and seeds a fresh
	// lineage root. We can't read that id back through its signature, so capture
	// it by snapshotting the store's created record.
	before := t.lastCreatedExecutionID(ctx) // see note below
	if err := t.startWorkflowForRule(ctx, event, rm, WorkflowLineage{}); err != nil {
		return uuid.Nil, fmt.Errorf("dispatch rerun: %w", err)
	}
	_ = before
	return t.lastCreatedExecutionID(ctx), nil
}
```

> **Design decision for returning the new id:** `startWorkflowForRule` mints the id internally and returns only `error`. Two clean options — pick one and drop the `lastCreatedExecutionID` placeholder above:
> - **(a, recommended) Refactor `startWorkflowForRule` to return the id:** change its signature to `(uuid.UUID, error)` and `return executionID, nil` on success (it already has `executionID` in scope at trigger.go:201). Update its one existing caller in `OnEntityEvent` (trigger.go:163) to ignore the id (`_, err := t.startWorkflowForRule(...)`). Then `RerunExecution` does `newID, err := t.startWorkflowForRule(ctx, event, rm, WorkflowLineage{})`. This is the cleanest; verify there is exactly one caller before editing.
> - **(b) Mint the id in `RerunExecution`** and pass it down via a new private `startWorkflowForRuleWithID`. More surface; only if (a)'s refactor is risky.
>
> Implement option (a): remove the `before`/`lastCreatedExecutionID` lines and return the id from `startWorkflowForRule`.

- [ ] **Step 5: Run tests + build**

Run: `go test ./business/sdk/workflow/temporal/... -run 'TestReconstructTriggerEvent|TestRerunExecution' -v`
Expected: PASS
Run: `go build ./business/sdk/workflow/temporal/...`
Also re-run the existing trigger tests to confirm the `startWorkflowForRule` signature change didn't break them: `go test ./business/sdk/workflow/temporal/... -run TestTrigger -v` (and the cascade/guard suites named in workflow-engine.md).

- [ ] **Step 6: Commit**

```bash
git add business/sdk/workflow/temporal/trigger.go business/sdk/workflow/temporal/rerun_test.go business/sdk/workflow/temporal/trigger_test.go
git commit -m "feat(workflow): add WorkflowTrigger.RerunExecution (fresh-id single-rule replay)

Widens ExecutionStore with QueryExecutionByID, reconstructs the TriggerEvent
from persisted trigger_data (EventID zero, re-stamped timestamp), and reuses
startWorkflowForRule (now returns the minted execution id) to clear all three
dedup walls. ErrExecutionNotRerunnable for ruleless/manual executions.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: executionapp.Rerun (Deliverable B — app layer)

Thin app layer mirroring `actionapp`: orchestrate the trigger and map errors to `errs.*`. Authorization is at the route (Task 6), so this layer holds no permission bus — it depends on a narrow `Reranner` interface (so it's unit-testable and doesn't force an api→temporal import shape on consumers).

**Files:**
- Create: `app/domain/workflow/executionapp/executionapp.go`
- Create: `app/domain/workflow/executionapp/model.go`
- Test: `app/domain/workflow/executionapp/executionapp_test.go`

**Interfaces:**
- Produces:
  - `type Reranner interface { RerunExecution(ctx context.Context, executionID uuid.UUID) (uuid.UUID, error) }`
  - `func NewApp(rerunner Reranner) *App`
  - `func (a *App) Rerun(ctx context.Context, executionID uuid.UUID) (RerunResponse, error)`
  - `type RerunResponse struct { OriginalExecutionID, NewExecutionID uuid.UUID }` with `Encode()`
- Consumes: `temporal.ErrExecutionNotRerunnable`, `workflow.ErrNotFound`, `*temporal.WorkflowTrigger` (satisfies `Reranner`)

- [ ] **Step 1: Write the failing tests**

Create `app/domain/workflow/executionapp/executionapp_test.go` (package `executionapp`):

```go
package executionapp

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

type fakeReranner struct {
	newID uuid.UUID
	err   error
	gotID uuid.UUID
}

func (f *fakeReranner) RerunExecution(_ context.Context, id uuid.UUID) (uuid.UUID, error) {
	f.gotID = id
	return f.newID, f.err
}

func TestRerun_Success(t *testing.T) {
	orig := uuid.New()
	fresh := uuid.New()
	app := NewApp(&fakeReranner{newID: fresh})

	resp, err := app.Rerun(context.Background(), orig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OriginalExecutionID != orig || resp.NewExecutionID != fresh {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestRerun_NotRerunnable_FailedPrecondition(t *testing.T) {
	app := NewApp(&fakeReranner{err: temporal.ErrExecutionNotRerunnable})
	_, err := app.Rerun(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error")
	}
	// errs.FailedPrecondition mapping — assert via errs code helper if available,
	// otherwise assert the error string contains the sentinel.
	if !errors.Is(err, temporal.ErrExecutionNotRerunnable) && !contains(err.Error(), "re-run") {
		t.Fatalf("err = %v", err)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0))) }
func indexOf(s, sub string) int   { return len([]rune(s)) - len([]rune(s)) /*placeholder*/ }
```

> Implementer note: replace the hand-rolled `contains`/`indexOf` with `strings.Contains` (the codebase prefers `strings.Contains` for substring checks). For the error-code assertion, follow the existing actionapp tests' pattern for checking `errs.Code` if one exists; otherwise assert `errors.Is`/message.

- [ ] **Step 2: Run to verify failure**

Run: `go test ./app/domain/workflow/executionapp/... -v`
Expected: FAIL — package/types undefined.

- [ ] **Step 3: Implement model.go**

Create `app/domain/workflow/executionapp/model.go`:

```go
package executionapp

import (
	"encoding/json"

	"github.com/google/uuid"
)

// RerunResponse is returned when an execution is re-run.
type RerunResponse struct {
	OriginalExecutionID uuid.UUID `json:"original_execution_id"`
	NewExecutionID      uuid.UUID `json:"new_execution_id"`
}

// Encode implements web.Encoder.
func (r RerunResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}
```

- [ ] **Step 4: Implement executionapp.go**

Create `app/domain/workflow/executionapp/executionapp.go`:

```go
package executionapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// Reranner re-fires the rule behind an execution with a fresh execution id.
// *temporal.WorkflowTrigger satisfies this.
type Reranner interface {
	RerunExecution(ctx context.Context, executionID uuid.UUID) (uuid.UUID, error)
}

// App is the application layer for execution operations.
type App struct {
	rerunner Reranner
}

// NewApp constructs an App. rerunner may be nil when the workflow engine
// (Temporal) is disabled; Rerun then returns an Internal error.
func NewApp(rerunner Reranner) *App {
	return &App{rerunner: rerunner}
}

// Rerun re-runs the given execution and returns the original + new ids.
func (a *App) Rerun(ctx context.Context, executionID uuid.UUID) (RerunResponse, error) {
	if a.rerunner == nil {
		return RerunResponse{}, errs.Newf(errs.Internal, "workflow engine is not enabled")
	}

	newID, err := a.rerunner.RerunExecution(ctx, executionID)
	if err != nil {
		switch {
		case errors.Is(err, workflow.ErrNotFound):
			return RerunResponse{}, errs.New(errs.NotFound, err)
		case errors.Is(err, temporal.ErrExecutionNotRerunnable):
			return RerunResponse{}, errs.New(errs.FailedPrecondition, err)
		default:
			return RerunResponse{}, errs.Newf(errs.Internal, "rerun execution: %s", err)
		}
	}

	return RerunResponse{OriginalExecutionID: executionID, NewExecutionID: newID}, nil
}
```

> Verify the `errs` API (`errs.New`, `errs.Newf`, `errs.Internal/NotFound/FailedPrecondition`) against `app/sdk/errs` — actionapp uses exactly these (actionapp.go:31-38, 71-73). `QueryExecutionByID` returns `workflow.ErrNotFound` on a missing row (executionapi.go:74), which surfaces through `RerunExecution`'s wrap — `errors.Is` still matches a wrapped sentinel.

- [ ] **Step 5: Run tests + build**

Run: `go test ./app/domain/workflow/executionapp/... -v`
Expected: PASS
Run: `go build ./app/domain/workflow/executionapp/...`

- [ ] **Step 6: Commit**

```bash
git add app/domain/workflow/executionapp/
git commit -m "feat(workflow): executionapp.Rerun app layer over WorkflowTrigger

Thin orchestration + errs mapping (NotFound / FailedPrecondition / Internal)
behind a narrow Reranner interface for testability.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: executionapi rerun route + handler (Deliverable B — HTTP)

Add `POST /v1/workflow/executions/{id}/rerun`, admin-gated, building `executionapp` from a `Reranner` carried on `Config`.

**Files:**
- Modify: `api/domain/http/workflow/executionapi/route.go` (Config gains `Trigger executionapp.Reranner`; register the POST route with Authenticate + Authorize)
- Modify: `api/domain/http/workflow/executionapi/executionapi.go` (api struct gains `executionApp *executionapp.App`; new `rerun` handler)
- Test: `api/cmd/services/ichor/tests/workflow/executionapi/rerun_test.go` (integration; mirror an existing executionapi integration test for setup/auth)

**Interfaces:**
- Consumes: `executionapp.NewApp`, `executionapp.Reranner`, `mid.Authenticate`, `mid.Authorize`, `auth.RuleAdminOnly`
- Produces: `POST /v1/workflow/executions/{id}/rerun` → `RerunResponse` JSON; 404 unknown id; 422/FailedPrecondition ruleless; 401/403 unauthorized

- [ ] **Step 1: Read auth wiring**

Read `docs/arch/auth.md` (Authorize variants) and grep an existing route that wires rule-only `mid.Authorize(...)` to copy the exact factory form:

Run: `grep -rn "mid.Authorize(" api/domain/http | head` and confirm the constant `auth.RuleAdminOnly` exists: `grep -rn "RuleAdminOnly\|RuleAny" app/sdk/auth/`.

- [ ] **Step 2: Write the failing integration test**

Create `api/cmd/services/ichor/tests/workflow/executionapi/rerun_test.go`. Mirror an existing executionapi integration test (find one under `api/cmd/services/ichor/tests/workflow/...`) for the harness (`apitest`, seeded admin token, Temporal infra via `apitest.InitWorkflowInfra`). Skeleton:

```go
// Test_Execution_Rerun seeds an execution, POSTs rerun as admin, and asserts a
// NEW execution id is returned (distinct from the original).
func Test_Execution_Rerun(t *testing.T) {
	// ... standard apitest setup with Temporal infra + admin token ...
	// 1. Seed a rule + an automation_executions row with valid trigger_data
	//    (entity_id/entity_name/event_type + raw_data product_id/quantity).
	// 2. POST /v1/workflow/executions/{id}/rerun with the admin bearer token.
	// 3. Assert HTTP 200 and resp.new_execution_id != original && != uuid.Nil.
	// 4. Negative: POST with a non-admin token -> 401/403.
	// 5. Negative: POST unknown id -> 404.
}
```

> Implementer note: read a sibling integration test for the exact `apitest` bootstrap (this repo's harness, token minting, and how Temporal infra is initialized — see `docs/arch/testing.md` and `apitest/workflow.go`). The assertions above are the contract; the bootstrap follows the established pattern.

- [ ] **Step 3: Run to verify failure**

Run: `go test ./api/cmd/services/ichor/tests/workflow/executionapi/... -run Test_Execution_Rerun -v`
Expected: FAIL — route 404 (not registered) / compile error (no `rerun` handler).

- [ ] **Step 4: Extend Config + register the route**

In `api/domain/http/workflow/executionapi/route.go`:

```go
type Config struct {
	Log         *logger.Logger
	WorkflowBus *workflow.Business
	AuthClient  *authclient.Client
	Trigger     executionapp.Reranner // nil when Temporal disabled
}
```

In `Routes`, after the existing GET registrations:

```go
	// Re-run a prior execution (admin-gated mutating action).
	app.HandlerFunc(http.MethodPost, version, "/workflow/executions/{id}/rerun", api.rerun,
		authen, mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly))
```

> Use the exact `mid.Authorize` factory form confirmed in Step 1. Add the `auth` + `executionapp` imports.

- [ ] **Step 5: Add the handler**

In `api/domain/http/workflow/executionapi/executionapi.go`, add `executionApp *executionapp.App` to the `api` struct and build it in `newAPI`:

```go
	return &api{
		log:          cfg.Log,
		workflowBus:  cfg.WorkflowBus,
		executionApp: executionapp.NewApp(cfg.Trigger),
	}
```

Add the handler (mirror `queryByID`'s param parsing + error style):

```go
func (a *api) rerun(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	resp, err := a.executionApp.Rerun(ctx, id)
	if err != nil {
		return errs.NewError(err) // executionapp already returns *errs.Error-typed errors
	}

	return resp
}
```

> Confirm the error pass-through helper: actionapi uses `errs.NewError(err)` to forward app-layer `errs` errors (actionapi.go). If executionapi doesn't import it yet, add it; otherwise return the `errs.*` value directly as the GET handlers do.

- [ ] **Step 6: Run the test (will still fail until wired in all.go — Task 7)**

The handler compiles now; the integration test needs the route wired with a non-nil `Trigger`, done in Task 7. Run the package build:
Run: `go build ./api/domain/http/workflow/executionapi/...`
Expected: clean build.

- [ ] **Step 7: Commit**

```bash
git add api/domain/http/workflow/executionapi/ api/cmd/services/ichor/tests/workflow/executionapi/rerun_test.go
git commit -m "feat(workflow): POST /workflow/executions/{id}/rerun (admin-gated)

executionapi gains a Reranner-backed rerun handler + Config.Trigger; route
guarded by Authenticate + Authorize(RuleAdminOnly). Integration test added
(green once wired in all.go).

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 7: Wire WorkflowTrigger into executionapi (Deliverable B — composition root)

`workflowTrigger` is `:=`-scoped inside the Temporal guard (all.go:607-656) and unreachable at the `executionapi.Routes` site (all.go:1473). Hoist it to an outer nil-able pointer and pass it.

**Files:**
- Modify: `api/cmd/services/ichor/build/all/all.go` (declare outer `var workflowTrigger`, change line 616 `:=`→`=`, add `Trigger:` to `executionapi.Config`)
- Modify: `api/sdk/http/apitest/workflow.go` (expose the trigger so integration tests can pass it into `executionapi.Config`)

**Interfaces:**
- Consumes: `temporalpkg.NewWorkflowTrigger(...)` (returns `*temporalpkg.WorkflowTrigger`, satisfies `executionapp.Reranner`)

- [ ] **Step 1: Hoist the trigger pointer**

In `api/cmd/services/ichor/build/all/all.go`, before the `if cfg.TemporalClient != nil {` block (line 607), declare:

```go
	var workflowTrigger *temporalpkg.WorkflowTrigger
```

Change the construction at line 616 from `workflowTrigger := temporalpkg.NewWorkflowTrigger(` to `workflowTrigger = temporalpkg.NewWorkflowTrigger(` (assign the outer var; confirm no other `:=` shadow remains in the block, and that the relay still receives `workflowTrigger` as before).

- [ ] **Step 2: Pass it into executionapi.Config**

At all.go:1473-1477:

```go
	executionapi.Routes(app, executionapi.Config{
		Log:         cfg.Log,
		WorkflowBus: workflowBus,
		AuthClient:  cfg.AuthClient,
		Trigger:     workflowTrigger, // nil when Temporal disabled -> rerun returns Internal
	})
```

- [ ] **Step 3: Build the service**

Run: `go build ./api/cmd/services/ichor/...`
Expected: clean build. If `temporalpkg` isn't the alias in scope at 1473, use the same alias imported at all.go:357.

- [ ] **Step 4: Expose the trigger in apitest (for the integration test)**

In `api/sdk/http/apitest/workflow.go`, add the constructed `*temporal.WorkflowTrigger` to the `WorkflowInfra` struct (it already builds the trigger components for `InitWorkflowInfra`) and pass it into `executionapi.Config.Trigger` wherever the test app mounts routes. Follow the existing field/wiring pattern in that file.

- [ ] **Step 5: Run the Task 6 integration test end-to-end**

Run: `go test ./api/cmd/services/ichor/tests/workflow/executionapi/... -run Test_Execution_Rerun -v`
Expected: PASS — new execution id returned; non-admin 401/403; unknown id 404.

- [ ] **Step 6: Commit**

```bash
git add api/cmd/services/ichor/build/all/all.go api/sdk/http/apitest/workflow.go
git commit -m "feat(workflow): wire WorkflowTrigger into executionapi for rerun

Hoist the trigger to an outer nil-able var (asyncCompleter pattern) and pass
it into executionapi.Config; apitest exposes it for integration tests.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 8: Default over-order remediation graph in the seed (Deliverable A)

Extend Rule 5 "Granular Inventory Pipeline" in the production seed: change the unconditional reserve→success edge to the `success` port, and add the over-order alert + approval hold + result alerts + failure alert. Ships to customers via `make seed-platform`.

**Files:**
- Modify: `business/sdk/dbtest/seed_workflow.go` (Rule 5 region ~630-862: add nodes/edges; change Edge 4 at ~823-829)
- Test: `business/sdk/dbtest/seed_workflow_overorder_test.go` (new) OR extend an existing seed/integration test that loads Rule 5's graph

**Interfaces:**
- Consumes: `workflow.NewRuleAction{AutomationRuleID, Name, Description, ActionConfig json.RawMessage, IsActive, TemplateID}`, `workflow.NewActionEdge{RuleID, SourceActionID *uuid.UUID, TargetActionID, EdgeType, SourceOutput *string, EdgeOrder}`; `busDomain.Workflow.CreateRuleAction` / `CreateActionEdge`
- Admin recipient/approver UUID: `5cf37266-3473-4006-984f-9325122678b7` (admin_gopher)

- [ ] **Step 1: Write the failing graph-shape test**

Create `business/sdk/dbtest/seed_workflow_overorder_test.go` (or add to an existing seed integration test). It seeds the platform workflow graph, loads Rule 5's edges, and asserts the new wiring. Skeleton (use the package's existing seed-test bootstrap — read a neighbor test for the exact `dbtest`/`busDomain` setup):

```go
// Asserts: reserve_inventory's success_alert edge is now gated on "success",
// and an "insufficient_stock" edge from reserve_inventory exists targeting the
// over_order alert, which sequences into a seek_approval node.
func Test_Seed_OverOrderGraph(t *testing.T) {
	// ... seed platform workflows; locate Rule 5 by name "Granular Inventory Pipeline" ...
	// edges := load action_edges for Rule 5
	// 1. find the reserve_inventory action id
	// 2. assert NO edge from reserve has nil SourceOutput (the old unconditional edge is gone)
	// 3. assert an edge from reserve has SourceOutput=="insufficient_stock"
	// 4. assert that edge's target is a create_alert action whose config alert_type=="over_order"
	// 5. assert that over_order alert sequences (SourceOutput=="success") into a seek_approval action
	// 6. assert an edge from reserve has SourceOutput=="failure"
}
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./business/sdk/dbtest/... -run Test_Seed_OverOrderGraph -v`
Expected: FAIL — no `insufficient_stock` edge from reserve; the nil edge still present.

- [ ] **Step 3: Change the existing reserve→success_alert edge to the success port**

In `business/sdk/dbtest/seed_workflow.go` (~823-829), the edge from `reserveAction` to `successAlertAction` currently omits `SourceOutput` (nil). Add a port:

```go
	successOutput := "success"
	// ... in the NewActionEdge for reserve -> success_alert:
	SourceOutput: &successOutput,
```

- [ ] **Step 4: Add the new nodes**

In Rule 5's region, after the existing reserve/alert nodes, create the new action nodes (mirror the existing `CreateRuleAction` calls in this file; reuse the `create_alert` and `seek_approval` templates already created at the top, passing `TemplateID`). Use admin_gopher (`5cf37266-3473-4006-984f-9325122678b7`) for recipients/approvers:

```go
	overOrderAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: granularRule.ID,
		Name:             "over_order_alert",
		Description:      "Alert ops that an over-order shortfall was detected",
		ActionConfig: json.RawMessage(`{
			"alert_type": "over_order",
			"severity": "high",
			"title": "Over-order: insufficient stock",
			"message": "Order line {{order_id}}: requested {{quantity}} of product {{product_id}} exceeds available stock.",
			"action_url": "/workflow/executions/{{execution_id}}",
			"recipients": {"users": ["5cf37266-3473-4006-984f-9325122678b7"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &createAlertTemplateID, // the create_alert template id from the top of seedWorkflow
	})
	if err != nil { return fmt.Errorf("create over_order_alert: %w", err) }

	approvalHoldAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: granularRule.ID,
		Name:             "approval_hold",
		Description:      "Hold the over-order for a human decision",
		ActionConfig: json.RawMessage(`{
			"approvers": ["5cf37266-3473-4006-984f-9325122678b7"],
			"approval_type": "any",
			"timeout_hours": 72,
			"approval_message": "Approve or reject this over-order"
		}`),
		IsActive:   true,
		TemplateID: &seekApprovalTemplateID,
	})
	if err != nil { return fmt.Errorf("create approval_hold: %w", err) }

	approvedAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: granularRule.ID,
		Name:             "over_order_approved_alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "over_order",
			"severity": "medium",
			"title": "Over-order approved",
			"message": "Over-order on order line {{order_id}} approved by {{resolved_by}}.",
			"recipients": {"users": ["5cf37266-3473-4006-984f-9325122678b7"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &createAlertTemplateID,
	})
	if err != nil { return fmt.Errorf("create over_order_approved_alert: %w", err) }

	rejectedAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: granularRule.ID,
		Name:             "over_order_rejected_alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "over_order",
			"severity": "medium",
			"title": "Over-order rejected",
			"message": "Over-order on order line {{order_id}} rejected by {{resolved_by}} — hold/cancel the line.",
			"recipients": {"users": ["5cf37266-3473-4006-984f-9325122678b7"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &createAlertTemplateID,
	})
	if err != nil { return fmt.Errorf("create over_order_rejected_alert: %w", err) }

	failureAlertAction, err := busDomain.Workflow.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: granularRule.ID,
		Name:             "reserve_failure_alert",
		ActionConfig: json.RawMessage(`{
			"alert_type": "over_order",
			"severity": "critical",
			"title": "Reservation failed (infrastructure)",
			"message": "Reservation failed for order line {{order_id}} — investigate.",
			"recipients": {"users": ["5cf37266-3473-4006-984f-9325122678b7"], "roles": []}
		}`),
		IsActive:   true,
		TemplateID: &createAlertTemplateID,
	})
	if err != nil { return fmt.Errorf("create reserve_failure_alert: %w", err) }
```

> Implementer note: match the actual variable names already used in `seedWorkflow` for the template ids (`createAlertTemplateID`, `seekApprovalTemplateID` — grep the top of the function for how template results are captured; the agent confirmed templates exist at lines 83 & 216). If templates aren't captured into vars, pass `TemplateID: nil` and put `action_type` in the config (the inline-fallback resolver handles template-less actions — see workflow-engine.md EdgeStore notes), but prefer the template id to match the file's pattern.

- [ ] **Step 5: Add the new edges**

```go
	insufficientStockOutput := "insufficient_stock"
	approvedOutput := "approved"
	rejectedOutput := "rejected"
	failureOutput := "failure"
	overOrderSuccessOutput := "success"

	// reserve --[insufficient_stock]--> over_order_alert
	busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID: granularRule.ID, SourceActionID: &reserveAction.ID, TargetActionID: overOrderAlertAction.ID,
		EdgeType: "sequence", SourceOutput: &insufficientStockOutput, EdgeOrder: 0,
	})
	// over_order_alert --[success]--> approval_hold
	busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID: granularRule.ID, SourceActionID: &overOrderAlertAction.ID, TargetActionID: approvalHoldAction.ID,
		EdgeType: "sequence", SourceOutput: &overOrderSuccessOutput, EdgeOrder: 0,
	})
	// approval_hold --[approved]--> approved_alert
	busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID: granularRule.ID, SourceActionID: &approvalHoldAction.ID, TargetActionID: approvedAlertAction.ID,
		EdgeType: "sequence", SourceOutput: &approvedOutput, EdgeOrder: 0,
	})
	// approval_hold --[rejected]--> rejected_alert
	busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID: granularRule.ID, SourceActionID: &approvalHoldAction.ID, TargetActionID: rejectedAlertAction.ID,
		EdgeType: "sequence", SourceOutput: &rejectedOutput, EdgeOrder: 1,
	})
	// reserve --[failure]--> failure_alert
	busDomain.Workflow.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID: granularRule.ID, SourceActionID: &reserveAction.ID, TargetActionID: failureAlertAction.ID,
		EdgeType: "sequence", SourceOutput: &failureOutput, EdgeOrder: 1,
	})
```

- [ ] **Step 6: Run the test + build**

Run: `go test ./business/sdk/dbtest/... -run Test_Seed_OverOrderGraph -v`
Expected: PASS
Run: `go build ./business/sdk/dbtest/...`
Also: grep for existing tests that assert Rule 5's edge count / the old nil edge and update them (the success-port change + 5 new nodes/edges shift counts): `grep -rn "Granular Inventory\|success_alert\|insufficient" business/sdk/dbtest api/cmd/services/ichor/tests/workflow | head`.

- [ ] **Step 7: Commit**

```bash
git add business/sdk/dbtest/seed_workflow.go business/sdk/dbtest/seed_workflow_overorder_test.go
git commit -m "feat(workflow): default over-order remediation graph in platform seed

Gate the existing reserve->success_alert edge on the success port, and wire
reserve.insufficient_stock -> over_order alert -> seek_approval (approved/
rejected alerts) plus reserve.failure -> critical alert. Ships via seed-platform.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 9: End-to-end re-run integration test (over-order recovery loop)

Prove the headline behavior: an over-order shortfall routes to the alert+approval branch, and re-running the execution after stock is restored actually re-attempts and succeeds (not a cached failure) — exercising the fresh-execution-id walls.

**Files:**
- Test: `api/cmd/services/ichor/tests/workflow/executionapi/rerun_e2e_test.go` (new) — or extend the Task 6 file

**Interfaces:**
- Consumes: full `apitest` Temporal infra; `reserve_inventory` live-stock behavior; `RerunExecution` via the HTTP endpoint

- [ ] **Step 1: Write the end-to-end test**

```go
// Test_OverOrder_RerunSucceedsAfterRestock:
//   1. Seed a product/order-line with requested qty > on-hand.
//   2. Trigger the rule (or directly create an execution whose reserve soft-failed
//      with output=insufficient_stock); assert an over_order alert exists
//      (alertbus QueryFilter{AlertType:"over_order", SourceRuleID:&ruleID}).
//   3. Increase on-hand inventory (simulate receive) so the requested qty now fits.
//   4. POST /v1/workflow/executions/{id}/rerun (admin).
//   5. Poll the NEW execution: assert its reserve action now succeeds
//      (status success / a reservation row exists), proving a fresh attempt
//      against live stock — NOT the cached insufficient result.
```

> Implementer note: scope all alert queries by `SourceRuleID` (workflow-alerts.md ⚠ — never count global alert totals in workflow tests). Use the apitest Temporal worker so the dispatched workflow actually runs. For polling the async result, follow the existing ordersapi/formdataapi workflow integration tests' wait pattern.

- [ ] **Step 2: Run it (red), implement nothing new (all code exists by Task 8), iterate setup until green**

Run: `go test ./api/cmd/services/ichor/tests/workflow/executionapi/... -run Test_OverOrder_RerunSucceedsAfterRestock -v`
Expected: PASS once the seeding/polling is correct. This test exercises Tasks 1–8 together; no new production code should be needed — if it forces a production change, that's a gap to fix in the relevant task.

- [ ] **Step 3: Final full-package test sweep (changed packages only)**

Run, in turn (NEVER `go test ./...`):
```bash
go test ./business/sdk/workflow/temporal/... -v
go test ./business/sdk/workflow/workflowactions/communication/... ./business/sdk/workflow/workflowactions/approval/... -v
go test ./app/domain/workflow/executionapp/... -v
go test ./api/domain/http/workflow/alertws/... -v
go test ./business/sdk/dbtest/... -run Test_Seed_OverOrderGraph -v
go test ./api/cmd/services/ichor/tests/workflow/executionapi/... -v
go build ./...
```
Expected: all PASS, clean build.

- [ ] **Step 4: Commit**

```bash
git add api/cmd/services/ichor/tests/workflow/executionapi/rerun_e2e_test.go
git commit -m "test(workflow): e2e over-order rerun succeeds after restock

Proves the recovery loop: insufficient_stock -> over_order alert+approval, then
rerun after restock re-attempts against live stock (fresh execution-id), not the
cached failure.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Self-Review

**Spec coverage:**
- Deliverable A (default graph) → Task 8 ✓
- Deliverable B (re-run endpoint) → Tasks 4 (engine), 5 (app), 6 (HTTP), 7 (wiring), 9 (e2e) ✓
- Deliverable C (over_order alert + execution_id) → Task 2 (enrichment) + Task 8 (seed config uses `over_order` + `{{execution_id}}`) ✓
- Deliverable D (heartbeat fix) → Task 1 ✓
- Deliverable E (typed approval_request WS) → Task 3 ✓
- Deferred items (timed_out timer, stale-approval cleanup, parent_execution_id) → correctly NOT in plan ✓
- Frontend follow-up → not in plan (separate repo) ✓

**Type consistency:** `RerunExecution(ctx, uuid.UUID) (uuid.UUID, error)` is defined in Task 4, consumed via the `Reranner` interface in Tasks 5/6/7 with the identical signature. `RerunResponse{OriginalExecutionID, NewExecutionID}` defined in Task 5, returned by the Task 6 handler. `ErrExecutionNotRerunnable` defined in Task 4, matched in Task 5. `buildAlertTemplateData`/`enrichAlertContext` defined+used in Task 2. `MessageTypeApprovalRequest` defined in Task 3 (message.go) and consumed in the same task (consumer.go).

**Known verification points the implementer must confirm (flagged inline, not placeholders):** the existing `trigger_test.go` fake names (Task 4); the `mid.Authorize` factory form + `auth.RuleAdminOnly` constant (Task 6 Step 1); the seed template-id variable names + the exact existing reserve→success edge literal (Task 8); the `apitest` Temporal bootstrap pattern (Tasks 6/9). Each is a "read the neighbor and match" instruction with a concrete assertion to satisfy.
