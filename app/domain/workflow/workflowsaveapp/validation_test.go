package workflowsaveapp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
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

// =============================================================================
// Output Port Validation Tests
// =============================================================================

// testHandler is a minimal ActionHandler + OutputPortProvider for testing.
type testHandler struct {
	actionType  string
	outputPorts []workflow.OutputPort
}

func (h testHandler) Execute(_ context.Context, _ json.RawMessage, _ workflow.ActionExecutionContext) (any, error) {
	return nil, nil
}
func (h testHandler) Validate(_ json.RawMessage) error { return nil }
func (h testHandler) GetType() string                  { return h.actionType }
func (h testHandler) SupportsManualExecution() bool     { return false }
func (h testHandler) IsAsync() bool                     { return false }
func (h testHandler) GetDescription() string            { return "test handler" }
func (h testHandler) GetOutputPorts() []workflow.OutputPort {
	return h.outputPorts
}

func newTestRegistry() *workflow.ActionRegistry {
	reg := workflow.NewActionRegistry()
	reg.Register(testHandler{
		actionType: "test_action",
		outputPorts: []workflow.OutputPort{
			{Name: "success", Description: "ok", IsDefault: true},
			{Name: "failure", Description: "fail"},
		},
	})
	return reg
}

func TestValidateOutputPorts_StartEdgeWithSourceOutput(t *testing.T) {
	actions := []SaveActionRequest{action("a")}
	edges := []SaveEdgeRequest{
		{EdgeType: "start", TargetActionID: "temp:0", SourceOutput: "success"},
	}
	err := validateOutputPorts(actions, edges, newTestRegistry())
	if err == nil {
		t.Fatal("expected error for start edge with source_output")
	}
	if !strings.Contains(err.Error(), "must not have source_output") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateOutputPorts_AlwaysEdgeWithSourceOutput(t *testing.T) {
	actions := []SaveActionRequest{action("a"), action("b")}
	edges := []SaveEdgeRequest{
		{EdgeType: "always", SourceActionID: "temp:0", TargetActionID: "temp:1", SourceOutput: "success"},
	}
	err := validateOutputPorts(actions, edges, newTestRegistry())
	if err == nil {
		t.Fatal("expected error for always edge with source_output")
	}
	if !strings.Contains(err.Error(), "must not have source_output") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateOutputPorts_SequenceEdgeValidOutput(t *testing.T) {
	actions := []SaveActionRequest{{
		Name:         "a",
		ActionType:   "test_action",
		ActionConfig: json.RawMessage(`{}`),
		IsActive:     true,
	}, action("b")}
	edges := []SaveEdgeRequest{
		{EdgeType: "sequence", SourceActionID: "temp:0", TargetActionID: "temp:1", SourceOutput: "success"},
	}
	err := validateOutputPorts(actions, edges, newTestRegistry())
	if err != nil {
		t.Fatalf("valid source_output should pass: %s", err)
	}
}

func TestValidateOutputPorts_SequenceEdgeInvalidOutput(t *testing.T) {
	actions := []SaveActionRequest{{
		Name:         "a",
		ActionType:   "test_action",
		ActionConfig: json.RawMessage(`{}`),
		IsActive:     true,
	}, action("b")}
	edges := []SaveEdgeRequest{
		{EdgeType: "sequence", SourceActionID: "temp:0", TargetActionID: "temp:1", SourceOutput: "nonexistent"},
	}
	err := validateOutputPorts(actions, edges, newTestRegistry())
	if err == nil {
		t.Fatal("expected error for invalid source_output")
	}
	if !strings.Contains(err.Error(), "is not valid") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestValidateOutputPorts_SequenceEdgeNoOutput(t *testing.T) {
	actions := []SaveActionRequest{{
		Name:         "a",
		ActionType:   "test_action",
		ActionConfig: json.RawMessage(`{}`),
		IsActive:     true,
	}, action("b")}
	edges := []SaveEdgeRequest{
		{EdgeType: "sequence", SourceActionID: "temp:0", TargetActionID: "temp:1"},
	}
	err := validateOutputPorts(actions, edges, newTestRegistry())
	if err != nil {
		t.Fatalf("empty source_output should default at runtime: %s", err)
	}
}

func TestValidateOutputPorts_UnknownSourceAction(t *testing.T) {
	actions := []SaveActionRequest{action("a")}
	edges := []SaveEdgeRequest{
		{EdgeType: "sequence", SourceActionID: "unknown-uuid", TargetActionID: "temp:0", SourceOutput: "success"},
	}
	// Should skip validation (graph validation catches missing refs)
	err := validateOutputPorts(actions, edges, newTestRegistry())
	if err != nil {
		t.Fatalf("unknown source action should be skipped: %s", err)
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
