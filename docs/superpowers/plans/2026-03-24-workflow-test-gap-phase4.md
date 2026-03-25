# Phase 4: WorkflowSaveApp Validation Tests

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add unit tests for `workflowsaveapp` graph validation (Kahn's algorithm DAG cycle detection, reachability) and action config validators.

**Architecture:** Pure unit tests — all functions take request structs and return errors. No DB needed. These are the most algorithmically complex untested code paths in the workflow system.

**Tech Stack:** Go testing (standard library only — no `dbtest` needed)

**Spec:** `docs/superpowers/specs/2026-03-24-workflow-test-gap-remediation-design.md` (Phase 4)

---

### Task 1: Graph Validation Tests

**Files:**
- Create: `app/domain/workflow/workflowsaveapp/graph_test.go`
- Reference: `app/domain/workflow/workflowsaveapp/graph.go`
- Reference: `app/domain/workflow/workflowsaveapp/model.go` (SaveActionRequest, SaveEdgeRequest)

- [ ] **Step 1: Write graph validation test file**

```go
package workflowsaveapp

import (
	"strings"
	"testing"
)

// helper to create a SaveActionRequest with the given name and type.
func action(name string) SaveActionRequest {
	return SaveActionRequest{
		Name:         name,
		ActionType:   "send_email",
		ActionConfig: []byte(`{"recipients":["a@b.com"],"subject":"s","body":"b"}`),
		IsActive:     true,
	}
}

// helper to create an action with an existing ID.
func actionWithID(name, id string) SaveActionRequest {
	a := action(name)
	a.ID = &id
	return a
}

func edge(edgeType, source, target string) SaveEdgeRequest {
	return SaveEdgeRequest{
		EdgeType:       edgeType,
		SourceActionID: source,
		TargetActionID: target,
	}
}

func startEdge(target string) SaveEdgeRequest {
	return edge("start", "", target)
}

func seqEdge(source, target string) SaveEdgeRequest {
	return edge("sequence", source, target)
}

func TestValidateGraph_EmptyActions(t *testing.T) {
	err := ValidateGraph(nil, nil)
	if err != nil {
		t.Fatalf("empty graph should be valid, got: %s", err)
	}
}

func TestValidateGraph_ActionsWithoutEdges(t *testing.T) {
	actions := []SaveActionRequest{action("a")}
	err := ValidateGraph(actions, nil)
	if err == nil {
		t.Fatal("expected error for actions without edges")
	}
	if !strings.Contains(err.Error(), "at least one edge") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_NoStartEdge(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b")}
	edges := []SaveEdgeRequest{seqEdge("temp:0", "temp:1")}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected error for missing start edge")
	}
	if !strings.Contains(err.Error(), "start edge is required") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_MultipleStartEdges(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		startEdge("temp:1"),
	}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected error for multiple start edges")
	}
	if !strings.Contains(err.Error(), "only one start edge") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_LinearChain(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b"), action("c")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:1"),
		seqEdge("temp:1", "temp:2"),
	}
	if err := ValidateGraph(actions, edges); err != nil {
		t.Fatalf("linear chain should be valid: %s", err)
	}
}

func TestValidateGraph_Diamond(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b"), action("c"), action("d")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:1"),
		seqEdge("temp:0", "temp:2"),
		seqEdge("temp:1", "temp:3"),
		seqEdge("temp:2", "temp:3"),
	}
	if err := ValidateGraph(actions, edges); err != nil {
		t.Fatalf("diamond should be valid: %s", err)
	}
}

func TestValidateGraph_CycleDetection(t *testing.T) {
	tests := []struct {
		name    string
		actions []SaveActionRequest
		edges   []SaveEdgeRequest
	}{
		{
			name:    "simple cycle A->B->C->A",
			actions: []SaveActionRequest{action("a"), action("b"), action("c")},
			edges: []SaveEdgeRequest{
				startEdge("temp:0"),
				seqEdge("temp:0", "temp:1"),
				seqEdge("temp:1", "temp:2"),
				seqEdge("temp:2", "temp:0"),
			},
		},
		{
			name:    "self-loop",
			actions: []SaveActionRequest{action("a")},
			edges: []SaveEdgeRequest{
				startEdge("temp:0"),
				seqEdge("temp:0", "temp:0"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGraph(tt.actions, tt.edges)
			if err == nil {
				t.Fatal("expected cycle detection error")
			}
			if !strings.Contains(err.Error(), "cycle") {
				t.Fatalf("expected cycle error, got: %s", err)
			}
		})
	}
}

func TestValidateGraph_UnreachableAction(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b"), action("orphan")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:1"),
		// temp:2 ("orphan") has no incoming edge and is not the start target
	}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected unreachable action error")
	}
	if !strings.Contains(err.Error(), "not reachable") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateGraph_InvalidActionRef(t *testing.T) {
	actions := []SaveActionRequest{action("a")}
	edges := []SaveEdgeRequest{
		startEdge("temp:0"),
		seqEdge("temp:0", "temp:99"),
	}
	err := ValidateGraph(actions, edges)
	if err == nil {
		t.Fatal("expected error for invalid action reference")
	}
}

func TestValidateGraph_UUIDReferences(t *testing.T) {
	existingID := "550e8400-e29b-41d4-a716-446655440000"
	actions := []SaveActionRequest{
		actionWithID("existing", existingID),
		action("new"),
	}
	edges := []SaveEdgeRequest{
		startEdge(existingID),
		seqEdge(existingID, "temp:1"),
	}
	if err := ValidateGraph(actions, edges); err != nil {
		t.Fatalf("UUID references should work: %s", err)
	}
}

func TestResolveActionRef(t *testing.T) {
	refMap := map[string]string{
		"temp:0":   "temp:0",
		"temp:1":   "temp:1",
		"some-uuid": "temp:0",
	}

	tests := []struct {
		name    string
		ref     string
		want    string
		wantErr bool
	}{
		{"temp ref", "temp:0", "temp:0", false},
		{"uuid ref", "some-uuid", "temp:0", false},
		{"empty ref", "", "", true},
		{"unknown ref", "unknown-uuid", "", true},
		{"temp out of range", "temp:99", "", true},
		{"invalid temp format", "temp:abc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveActionRef(tt.ref, refMap, 2)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
```

Note: Since the test file uses `package workflowsaveapp` (internal test), it can access unexported functions like `resolveActionRef`, `detectCycles`, and `checkReachability` directly.

- [ ] **Step 2: Run graph tests**

Run: `go test ./app/domain/workflow/workflowsaveapp/... -run "TestValidateGraph|TestResolveActionRef" -v -count=1`
Expected: All PASS.

- [ ] **Step 3: Commit**

```
git add app/domain/workflow/workflowsaveapp/graph_test.go
git commit -m "test(workflowsaveapp): add graph validation unit tests (cycle detection, reachability)"
```

---

### Task 2: Action Config Validation Tests

**Files:**
- Create: `app/domain/workflow/workflowsaveapp/validation_test.go`
- Reference: `app/domain/workflow/workflowsaveapp/validation.go`

- [ ] **Step 1: Write config validator tests**

```go
package workflowsaveapp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateActionConfig_CreateAlert(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"valid", `{"alert_type":"low_stock","severity":"high","title":"Alert","message":"msg"}`, ""},
		{"missing alert_type", `{"severity":"high","title":"Alert","message":"msg"}`, "alert_type is required"},
		{"missing severity", `{"alert_type":"low_stock","title":"Alert","message":"msg"}`, "severity is required"},
		{"missing title", `{"alert_type":"low_stock","severity":"high","message":"msg"}`, "title is required"},
		{"missing message", `{"alert_type":"low_stock","severity":"high","title":"Alert"}`, "message is required"},
		{"invalid json", `{bad`, "invalid config JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActionConfig(ActionTypeCreateAlert, json.RawMessage(tt.config))
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateActionConfig_SendEmail(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"valid", `{"recipients":["a@b.com"],"subject":"Hi","body":"Hello"}`, ""},
		{"missing recipients", `{"subject":"Hi","body":"Hello"}`, "recipients is required"},
		{"empty recipients", `{"recipients":[],"subject":"Hi","body":"Hello"}`, "recipients is required"},
		{"missing subject", `{"recipients":["a@b.com"],"body":"Hello"}`, "subject is required"},
		{"missing body", `{"recipients":["a@b.com"],"subject":"Hi"}`, "body is required"},
		{"invalid json", `{bad`, "invalid config JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActionConfig(ActionTypeSendEmail, json.RawMessage(tt.config))
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateActionConfig_SendNotification(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"valid", `{"recipients":["user1"],"channels":["email"]}`, ""},
		{"missing recipients", `{"channels":["email"]}`, "recipients is required"},
		{"empty recipients", `{"recipients":[],"channels":["email"]}`, "recipients is required"},
		{"missing channels", `{"recipients":["user1"]}`, "channels is required"},
		{"empty channels", `{"recipients":["user1"],"channels":[]}`, "channels is required"},
		{"invalid json", `{bad`, "invalid config JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActionConfig(ActionTypeSendNotification, json.RawMessage(tt.config))
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateActionConfig_UpdateField(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"valid", `{"target_entity":"sales.orders","target_field":"status"}`, ""},
		{"missing target_entity", `{"target_field":"status"}`, "target_entity is required"},
		{"missing target_field", `{"target_entity":"sales.orders"}`, "target_field is required"},
		{"invalid json", `{bad`, "invalid config JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActionConfig(ActionTypeUpdateField, json.RawMessage(tt.config))
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateActionConfig_SeekApproval(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"valid", `{"approvers":["role1"],"approval_type":"manager"}`, ""},
		{"missing approvers", `{"approval_type":"manager"}`, "approvers is required"},
		{"empty approvers", `{"approvers":[],"approval_type":"manager"}`, "approvers is required"},
		{"missing approval_type", `{"approvers":["role1"]}`, "approval_type is required"},
		{"invalid json", `{bad`, "invalid config JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActionConfig(ActionTypeSeekApproval, json.RawMessage(tt.config))
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateActionConfig_AllocateInventory(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"valid", `{"inventory_items":[{"id":"x"}],"allocation_mode":"fifo"}`, ""},
		{"source_from_line_item no items needed", `{"source_from_line_item":true,"allocation_mode":"fifo"}`, ""},
		{"no items without source_from_line_item", `{"allocation_mode":"fifo"}`, "inventory_items is required"},
		{"missing allocation_mode", `{"inventory_items":[{"id":"x"}]}`, "allocation_mode is required"},
		{"invalid json", `{bad`, "invalid config JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActionConfig(ActionTypeAllocateInventory, json.RawMessage(tt.config))
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateActionConfig_EvaluateCondition(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"valid", `{"conditions":[{"field":"status","op":"equals","value":"active"}]}`, ""},
		{"empty conditions", `{"conditions":[]}`, "conditions is required"},
		{"missing conditions", `{}`, "conditions is required"},
		{"invalid json", `{bad`, "invalid config JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActionConfig(ActionTypeEvaluateCondition, json.RawMessage(tt.config))
			assertValidationError(t, err, tt.wantErr)
		})
	}
}

func TestValidateActionConfig_UnknownType(t *testing.T) {
	err := validateActionConfig("totally_unknown_action", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for unknown action type")
	}
	if !strings.Contains(err.Error(), "unknown action type") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateActionConfig_EmptyConfig(t *testing.T) {
	err := validateActionConfig(ActionTypeCreateAlert, nil)
	if err == nil {
		t.Fatal("expected error for empty config")
	}
	if !strings.Contains(err.Error(), "action_config is required") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateActionConfig_PassthroughTypes(t *testing.T) {
	passthroughTypes := []string{
		ActionTypeCheckInventory,
		ActionTypeCheckReorderPoint,
		ActionTypeCommitAllocation,
		ActionTypeReleaseReservation,
		ActionTypeReserveInventory,
		ActionTypeDelay,
		ActionTypeLookupEntity,
		ActionTypeCreateEntity,
		ActionTypeTransitionStatus,
		ActionTypeLogAuditEntry,
	}

	for _, at := range passthroughTypes {
		t.Run(at, func(t *testing.T) {
			err := validateActionConfig(at, json.RawMessage(`{"any":"config"}`))
			if err != nil {
				t.Fatalf("passthrough type %s should not fail validation: %s", at, err)
			}
		})
	}
}

func TestValidateActionConfigs_Wrapper(t *testing.T) {
	actions := []SaveActionRequest{
		{
			Name:         "good",
			ActionType:   ActionTypeCreateAlert,
			ActionConfig: json.RawMessage(`{"alert_type":"x","severity":"y","title":"t","message":"m"}`),
		},
		{
			Name:         "bad",
			ActionType:   ActionTypeCreateAlert,
			ActionConfig: json.RawMessage(`{}`), // missing required fields
		},
	}

	err := ValidateActionConfigs(actions)
	if err == nil {
		t.Fatal("expected error from second action")
	}
	if !strings.Contains(err.Error(), "action[1]") {
		t.Fatalf("error should reference action index: %s", err)
	}
}

// assertValidationError checks that an error matches expectations.
func assertValidationError(t *testing.T, err error, wantErr string) {
	t.Helper()
	if wantErr == "" {
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantErr)
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("error %q does not contain %q", err.Error(), wantErr)
	}
}
```

- [ ] **Step 2: Run validation tests**

Run: `go test ./app/domain/workflow/workflowsaveapp/... -run "TestValidateActionConfig" -v -count=1`
Expected: All PASS.

- [ ] **Step 3: Commit**

```
git add app/domain/workflow/workflowsaveapp/validation_test.go
git commit -m "test(workflowsaveapp): add action config validation unit tests"
```

---

### Task 3: Output Port Validation Tests

- [ ] **Step 1: Add output port validation tests to validation_test.go**

`validateOutputPorts` requires a `workflow.ActionRegistry` to look up output ports. The `workflowsaveapp` package already imports `workflow`, so you can create a registry and register a test handler.

Read `business/sdk/workflow/interfaces.go` for:
- `NewActionRegistry()` constructor
- `Register(handler)` method
- `GetOutputPorts(actionType)` method
- The `OutputPortProvider` interface that handlers implement

Create a minimal test handler struct that implements `ActionHandler` + `OutputPortProvider` and declares output ports (e.g., "success" and "failure"). Then test:

1. **start edge with source_output** → error ("start edges must not have source_output")
2. **always edge with source_output** → error ("always edges must not have source_output")
3. **sequence edge with valid source_output** → ok
4. **sequence edge with invalid source_output** → error ("source_output X is not valid")
5. **sequence edge without source_output** → ok (defaults at runtime)
6. **source action not in action list** → skipped (graph validation catches this)

If constructing the test handler is complex due to the full `ActionHandler` interface, you can embed a minimal struct. The `workflowsaveapp` package already depends on `workflow`, so there are no import cycle issues.

- [ ] **Step 2: Run all Phase 4 tests**

Run: `go test ./app/domain/workflow/workflowsaveapp/... -v -count=1`
Expected: All PASS.

- [ ] **Step 3: Commit**

```
git add app/domain/workflow/workflowsaveapp/validation_test.go
git commit -m "test(workflowsaveapp): add output port validation tests"
```

---

### Task 4: Integration Test Gap Check

- [ ] **Step 1: Review existing workflowsaveapi integration tests**

Read the test files in `api/cmd/services/ichor/tests/workflow/workflowsaveapi/`:
- `create_test.go` — covers CreateWorkflow
- `update_test.go` — covers SaveWorkflow (update path)
- `save_test.go` — main test runner
- `validation_test.go` — validation error cases
- `dryrun_test.go` — DryRunValidate
- `actions_test.go` — action-level tests

Check if `DuplicateWorkflow` is tested. Check if `syncActions` (add/update/remove actions during save) has explicit test coverage.

- [ ] **Step 2: Document any gaps found**

If gaps exist, add integration test cases to the appropriate existing test files. If no gaps exist, document that in a commit message.

- [ ] **Step 3: Commit any additions**

```
git commit -m "test(workflowsaveapp): verify integration test coverage, add missing cases"
```
