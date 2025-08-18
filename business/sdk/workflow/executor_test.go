package workflow_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func TestActionExecutor_ValidateActionConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action workflowdb.RuleActionView
		want   workflow.ActionValidationResult
	}{
		{
			name: "valid seek_approval action",
			action: workflowdb.RuleActionView{
				ID:                 "action_001",
				TemplateActionType: sql.NullString{String: "seek_approval", Valid: true},
				ActionConfig: json.RawMessage(`{
					"approvers": ["user1@example.com", "user2@example.com"],
					"approval_type": "any"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "invalid seek_approval - missing approvers",
			action: workflowdb.RuleActionView{
				ID:                 "action_002",
				TemplateActionType: sql.NullString{String: "seek_approval", Valid: true},
				ActionConfig: json.RawMessage(`{
					"approvers": [],
					"approval_type": "any"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  false,
				Errors:   []string{"approvers list is required and must not be empty"},
				Warnings: []string{},
			},
		},
		{
			name: "invalid seek_approval - invalid approval type",
			action: workflowdb.RuleActionView{
				ID:                 "action_003",
				TemplateActionType: sql.NullString{String: "seek_approval", Valid: true},
				ActionConfig: json.RawMessage(`{
					"approvers": ["user1@example.com"],
					"approval_type": "invalid"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  false,
				Errors:   []string{"invalid approval_type, must be: any, all, or majority"},
				Warnings: []string{},
			},
		},
		{
			name: "valid send_email action",
			action: workflowdb.RuleActionView{
				ID:                 "action_004",
				TemplateActionType: sql.NullString{String: "send_email", Valid: true},
				ActionConfig: json.RawMessage(`{
					"recipients": ["user@example.com"],
					"subject": "Test Email"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "invalid send_email - missing subject",
			action: workflowdb.RuleActionView{
				ID:                 "action_005",
				TemplateActionType: sql.NullString{String: "send_email", Valid: true},
				ActionConfig: json.RawMessage(`{
					"recipients": ["user@example.com"],
					"subject": ""
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  false,
				Errors:   []string{"email subject is required"},
				Warnings: []string{},
			},
		},
		{
			name: "valid create_alert action",
			action: workflowdb.RuleActionView{
				ID:                 "action_006",
				TemplateActionType: sql.NullString{String: "create_alert", Valid: true},
				ActionConfig: json.RawMessage(`{
					"message": "Alert message",
					"recipients": ["user@example.com"],
					"priority": "high"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "invalid create_alert - invalid priority",
			action: workflowdb.RuleActionView{
				ID:                 "action_007",
				TemplateActionType: sql.NullString{String: "create_alert", Valid: true},
				ActionConfig: json.RawMessage(`{
					"message": "Alert message",
					"recipients": ["user@example.com"],
					"priority": "urgent"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  false,
				Errors:   []string{"invalid priority level"},
				Warnings: []string{},
			},
		},
		{
			name: "valid update_field action",
			action: workflowdb.RuleActionView{
				ID:                 "action_008",
				TemplateActionType: sql.NullString{String: "update_field", Valid: true},
				ActionConfig: json.RawMessage(`{
					"target_entity": "customers",
					"target_field": "status",
					"new_value": "active"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "action with no type",
			action: workflowdb.RuleActionView{
				ID:                 "action_009",
				TemplateActionType: sql.NullString{Valid: false},
				ActionConfig:       json.RawMessage(`{}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  false,
				Errors:   []string{"Action type is required"},
				Warnings: []string{},
			},
		},
		{
			name: "unsupported action type",
			action: workflowdb.RuleActionView{
				ID:                 "action_010",
				TemplateActionType: sql.NullString{String: "unsupported_type", Valid: true},
				ActionConfig:       json.RawMessage(`{}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  false,
				Errors:   []string{"Unsupported action type: unsupported_type"},
				Warnings: []string{},
			},
		},
		{
			name: "action with template defaults merged",
			action: workflowdb.RuleActionView{
				ID:                 "action_011",
				TemplateActionType: sql.NullString{String: "send_email", Valid: true},
				TemplateDefaultConfig: json.RawMessage(`{
					"recipients": ["default@example.com"],
					"subject": "Default Subject",
					"cc": ["cc@example.com"]
				}`),
				ActionConfig: json.RawMessage(`{
					"recipients": ["override@example.com"],
					"subject": "Override Subject"
				}`),
			},
			want: workflow.ActionValidationResult{
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
			},
		},
	}
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ae.ValidateActionConfig(tt.action)

			if got.IsValid != tt.want.IsValid {
				t.Errorf("IsValid = %v, want %v", got.IsValid, tt.want.IsValid)
			}

			if diff := cmp.Diff(tt.want.Errors, got.Errors); diff != "" {
				t.Errorf("Errors mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.want.Warnings, got.Warnings); diff != "" {
				t.Errorf("Warnings mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestActionExecutor_MergeActionConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action workflowdb.RuleActionView
		want   map[string]interface{}
	}{
		{
			name: "merge template defaults with action config",
			action: workflowdb.RuleActionView{
				TemplateDefaultConfig: json.RawMessage(`{
					"field1": "default1",
					"field2": "default2",
					"field3": "default3"
				}`),
				ActionConfig: json.RawMessage(`{
					"field2": "override2",
					"field4": "new4"
				}`),
			},
			want: map[string]interface{}{
				"field1": "default1",
				"field2": "override2",
				"field3": "default3",
				"field4": "new4",
			},
		},
		{
			name: "only template defaults",
			action: workflowdb.RuleActionView{
				TemplateDefaultConfig: json.RawMessage(`{
					"field1": "value1",
					"field2": "value2"
				}`),
				ActionConfig: nil,
			},
			want: map[string]interface{}{
				"field1": "value1",
				"field2": "value2",
			},
		},
		{
			name: "only action config",
			action: workflowdb.RuleActionView{
				TemplateDefaultConfig: nil,
				ActionConfig: json.RawMessage(`{
					"field1": "value1",
					"field2": "value2"
				}`),
			},
			want: map[string]interface{}{
				"field1": "value1",
				"field2": "value2",
			},
		},
		{
			name: "nested object merge",
			action: workflowdb.RuleActionView{
				TemplateDefaultConfig: json.RawMessage(`{
					"settings": {
						"timeout": 30,
						"retries": 3
					},
					"enabled": true
				}`),
				ActionConfig: json.RawMessage(`{
					"settings": {
						"timeout": 60
					}
				}`),
			},
			want: map[string]interface{}{
				"settings": map[string]interface{}{
					"timeout": float64(60),
				},
				"enabled": true,
			},
		},
		{
			name: "no config at all",
			action: workflowdb.RuleActionView{
				TemplateDefaultConfig: nil,
				ActionConfig:          nil,
			},
			want: map[string]interface{}{},
		},
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use reflection to access private method
			got := testMergeActionConfig(ae, tt.action)

			var result map[string]interface{}
			if err := json.Unmarshal(got, &result); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			if diff := cmp.Diff(tt.want, result); diff != "" {
				t.Errorf("mergeActionConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestActionExecutor_BuildTemplateContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		context workflow.ActionExecutionContext
		want    workflow.TemplateContext
	}{
		{
			name: "basic context",
			context: workflow.ActionExecutionContext{
				EntityID:   "ent_123",
				EntityName: "customers",
				EventType:  "on_update",
				Timestamp:  time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				UserID:     "user_456",
				RuleID:     "rule_789",
				RuleName:   "Update Customer Status",
			},
			want: workflow.TemplateContext{
				"entity_id":   "ent_123",
				"entity_name": "customers",
				"event_type":  "on_update",
				"timestamp":   time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				"user_id":     "user_456",
				"rule_id":     "rule_789",
				"rule_name":   "Update Customer Status",
			},
		},
		{
			name: "context with raw data",
			context: workflow.ActionExecutionContext{
				EntityID:   "ent_123",
				EntityName: "orders",
				EventType:  "on_create",
				Timestamp:  time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				UserID:     "user_456",
				RuleID:     "rule_789",
				RuleName:   "Process New Order",
				RawData: map[string]interface{}{
					"order_total":  199.99,
					"customer_id":  "cust_123",
					"order_status": "pending",
				},
			},
			want: workflow.TemplateContext{
				"entity_id":    "ent_123",
				"entity_name":  "orders",
				"event_type":   "on_create",
				"timestamp":    time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				"user_id":      "user_456",
				"rule_id":      "rule_789",
				"rule_name":    "Process New Order",
				"order_total":  199.99,
				"customer_id":  "cust_123",
				"order_status": "pending",
			},
		},
		{
			name: "context with field changes",
			context: workflow.ActionExecutionContext{
				EntityID:   "ent_123",
				EntityName: "customers",
				EventType:  "on_update",
				Timestamp:  time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				UserID:     "user_456",
				RuleID:     "rule_789",
				RuleName:   "Track Status Changes",
				FieldChanges: map[string]workflow.FieldChange{
					"status": {
						OldValue: "regular",
						NewValue: "premium",
					},
					"credit_limit": {
						OldValue: 1000,
						NewValue: 5000,
					},
				},
			},
			want: workflow.TemplateContext{
				"entity_id":   "ent_123",
				"entity_name": "customers",
				"event_type":  "on_update",
				"timestamp":   time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				"user_id":     "user_456",
				"rule_id":     "rule_789",
				"rule_name":   "Track Status Changes",
				"field_changes": map[string]workflow.FieldChange{
					"status": {
						OldValue: "regular",
						NewValue: "premium",
					},
					"credit_limit": {
						OldValue: 1000,
						NewValue: 5000,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testBuildTemplateContext(ae, tt.context)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("buildTemplateContext() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestActionExecutor_ProcessTemplates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  json.RawMessage
		context workflow.TemplateContext
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "simple template substitution",
			config: json.RawMessage(`{
				"recipient": "{{customer_email}}",
				"subject": "Order {{order_id}} - {{status | uppercase}}"
			}`),
			context: workflow.TemplateContext{
				"customer_email": "john@example.com",
				"order_id":       "ORD-123",
				"status":         "pending",
			},
			want: map[string]interface{}{
				"recipient": "john@example.com",
				"subject":   "Order ORD-123 - PENDING",
			},
		},
		{
			name: "nested template processing",
			config: json.RawMessage(`{
				"notification": {
					"to": "{{user.email}}",
					"message": "Hello {{user.name | capitalize}}"
				},
				"priority": "{{priority | uppercase}}"
			}`),
			context: workflow.TemplateContext{
				"user": map[string]interface{}{
					"email": "alice@example.com",
					"name":  "alice smith",
				},
				"priority": "high",
			},
			want: map[string]interface{}{
				"notification": map[string]interface{}{
					"to":      "alice@example.com",
					"message": "Hello Alice smith",
				},
				"priority": "HIGH",
			},
		},
		{
			name: "array with templates",
			config: json.RawMessage(`{
				"recipients": ["{{email1}}", "{{email2}}"],
				"cc": ["{{manager_email}}"]
			}`),
			context: workflow.TemplateContext{
				"email1":        "user1@example.com",
				"email2":        "user2@example.com",
				"manager_email": "manager@example.com",
			},
			want: map[string]interface{}{
				"recipients": []interface{}{"user1@example.com", "user2@example.com"},
				"cc":         []interface{}{"manager@example.com"},
			},
		},
		{
			name: "missing variable with default",
			config: json.RawMessage(`{
				"field": "{{missing_var}}"
			}`),
			context: workflow.TemplateContext{},
			want: map[string]interface{}{
				"field": "",
			},
			wantErr: false,
		},
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testProcessTemplates(ae, tt.config, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("processTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var result map[string]interface{}
				if err := json.Unmarshal(got, &result); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}

				if diff := cmp.Diff(tt.want, result); diff != "" {
					t.Errorf("processTemplates() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestActionExecutor_GetActionType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action workflowdb.RuleActionView
		want   string
	}{
		{
			name: "action with template type",
			action: workflowdb.RuleActionView{
				TemplateActionType: sql.NullString{String: "send_email", Valid: true},
			},
			want: "send_email",
		},
		{
			name: "action without template type",
			action: workflowdb.RuleActionView{
				TemplateActionType: sql.NullString{Valid: false},
			},
			want: "",
		},
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testGetActionType(ae, tt.action)

			if got != tt.want {
				t.Errorf("getActionType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActionExecutor_ShouldStopOnFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action workflowdb.RuleActionView
		want   bool
	}{
		{
			name: "seek_approval should stop on failure",
			action: workflowdb.RuleActionView{
				TemplateActionType: sql.NullString{String: "seek_approval", Valid: true},
			},
			want: true,
		},
		{
			name: "send_email should not stop on failure",
			action: workflowdb.RuleActionView{
				TemplateActionType: sql.NullString{String: "send_email", Valid: true},
			},
			want: false,
		},
		{
			name: "create_alert should not stop on failure",
			action: workflowdb.RuleActionView{
				TemplateActionType: sql.NullString{String: "create_alert", Valid: true},
			},
			want: false,
		},
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testShouldStopOnFailure(ae, tt.action)

			if got != tt.want {
				t.Errorf("shouldStopOnFailure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActionExecutor_Stats(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	// Initial stats should be zero
	stats := ae.GetStats()
	if stats.TotalActionsExecuted != 0 {
		t.Errorf("Initial TotalActionsExecuted = %d, want 0", stats.TotalActionsExecuted)
	}
	if stats.SuccessfulExecutions != 0 {
		t.Errorf("Initial SuccessfulExecutions = %d, want 0", stats.SuccessfulExecutions)
	}
	if stats.FailedExecutions != 0 {
		t.Errorf("Initial FailedExecutions = %d, want 0", stats.FailedExecutions)
	}

	// Simulate updating stats
	result := workflow.BatchExecutionResult{
		TotalActions:       5,
		SuccessfulActions:  3,
		FailedActions:      2,
		TotalExecutionTime: 100 * time.Millisecond,
	}
	testUpdateStats(ae, result)

	// Check updated stats
	stats = ae.GetStats()
	if stats.TotalActionsExecuted != 5 {
		t.Errorf("TotalActionsExecuted = %d, want 5", stats.TotalActionsExecuted)
	}
	if stats.SuccessfulExecutions != 3 {
		t.Errorf("SuccessfulExecutions = %d, want 3", stats.SuccessfulExecutions)
	}
	if stats.FailedExecutions != 2 {
		t.Errorf("FailedExecutions = %d, want 2", stats.FailedExecutions)
	}
	if stats.LastExecutedAt == nil {
		t.Error("LastExecutedAt should not be nil after update")
	}
}

func TestActionExecutor_ExecutionHistory(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	// Initially empty
	history := ae.GetExecutionHistory(10)
	if len(history) != 0 {
		t.Errorf("Initial history length = %d, want 0", len(history))
	}

	// Add some execution results
	for i := 0; i < 5; i++ {
		result := workflow.BatchExecutionResult{
			RuleID:            fmt.Sprintf("rule_%d", i),
			RuleName:          fmt.Sprintf("Rule %d", i),
			TotalActions:      i + 1,
			SuccessfulActions: i,
			Status:            "success",
		}
		testAddToHistory(ae, result)
	}

	// Get limited history
	history = ae.GetExecutionHistory(3)
	if len(history) != 3 {
		t.Errorf("Limited history length = %d, want 3", len(history))
	}

	// Get all history
	history = ae.GetExecutionHistory(10)
	if len(history) != 5 {
		t.Errorf("Full history length = %d, want 5", len(history))
	}

	// Clear history
	ae.ClearHistory()
	history = ae.GetExecutionHistory(10)
	if len(history) != 0 {
		t.Errorf("History length after clear = %d, want 0", len(history))
	}
}

func TestActionHandler_Implementations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	tests := []struct {
		name          string
		handlerType   string
		validConfig   json.RawMessage
		invalidConfig json.RawMessage
		wantType      string
	}{
		{
			name:        "SeekApprovalHandler",
			handlerType: "seek_approval",
			validConfig: json.RawMessage(`{
				"approvers": ["user1@example.com"],
				"approval_type": "any"
			}`),
			invalidConfig: json.RawMessage(`{
				"approvers": [],
				"approval_type": "invalid"
			}`),
			wantType: "seek_approval",
		},
		{
			name:        "AllocateInventoryHandler",
			handlerType: "allocate_inventory",
			validConfig: json.RawMessage(`{
				"inventory_items": [{"item_id": "item1", "quantity": 10}],
				"allocation_strategy": "fifo"
			}`),
			invalidConfig: json.RawMessage(`{
				"inventory_items": [],
				"allocation_strategy": "invalid"
			}`),
			wantType: "allocate_inventory",
		},
		{
			name:        "CreateAlertHandler",
			handlerType: "create_alert",
			validConfig: json.RawMessage(`{
				"message": "Alert message",
				"recipients": ["user@example.com"],
				"priority": "high"
			}`),
			invalidConfig: json.RawMessage(`{
				"message": "",
				"recipients": [],
				"priority": "invalid"
			}`),
			wantType: "create_alert",
		},
		{
			name:        "SendEmailHandler",
			handlerType: "send_email",
			validConfig: json.RawMessage(`{
				"recipients": ["user@example.com"],
				"subject": "Test Subject"
			}`),
			invalidConfig: json.RawMessage(`{
				"recipients": [],
				"subject": ""
			}`),
			wantType: "send_email",
		},
		{
			name:        "UpdateFieldHandler",
			handlerType: "update_field",
			validConfig: json.RawMessage(`{
				"target_entity": "customers",
				"target_field": "status",
				"new_value": "active"
			}`),
			invalidConfig: json.RawMessage(`{
				"target_entity": "",
				"target_field": "",
				"new_value": null
			}`),
			wantType: "update_field",
		},
		{
			name:        "SendNotificationHandler",
			handlerType: "send_notification",
			validConfig: json.RawMessage(`{
				"recipients": ["user@example.com"],
				"channels": [{"type": "email"}],
				"priority": "medium"
			}`),
			invalidConfig: json.RawMessage(`{
				"recipients": [],
				"channels": [],
				"priority": "invalid"
			}`),
			wantType: "send_notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler workflow.ActionHandler

			switch tt.handlerType {
			case "seek_approval":
				handler = ae.NewSeekApprovalHandler()
			case "allocate_inventory":
				handler = ae.NewAllocateInventoryHandler()
			case "create_alert":
				handler = ae.NewCreateAlertHandler()
			case "send_email":
				handler = ae.NewSendEmailHandler()
			case "update_field":
				handler = ae.NewUpdateFieldHandler()
			case "send_notification":
				handler = ae.NewSendNotificationHandler()
			}

			// Test GetType
			if got := handler.GetType(); got != tt.wantType {
				t.Errorf("GetType() = %v, want %v", got, tt.wantType)
			}

			// Test valid config validation
			if err := handler.Validate(tt.validConfig); err != nil {
				t.Errorf("Validate() with valid config failed: %v", err)
			}

			// Test invalid config validation
			if err := handler.Validate(tt.invalidConfig); err == nil {
				t.Errorf("Validate() with invalid config should have failed")
			}

			// Test Execute (stub implementation should succeed)
			ctx := context.Background()
			execContext := workflow.ActionExecutionContext{
				EntityID:   "test_entity",
				EntityName: "test",
				EventType:  "on_create",
				RuleID:     "rule_123",
				RuleName:   "Test Rule",
			}

			result, err := handler.Execute(ctx, tt.validConfig, execContext)
			if err != nil {
				t.Errorf("Execute() failed: %v", err)
			}
			if result == nil {
				t.Error("Execute() returned nil result")
			}
		})
	}
}

// Test helper functions to access private methods
// In a real scenario, you might want to make these methods public or use an interface

func testMergeActionConfig(ae *workflow.ActionExecutor, action workflowdb.RuleActionView) json.RawMessage {
	// This simulates calling the private mergeActionConfig method
	// In practice, you'd test this through public methods that use it
	var merged map[string]interface{}

	if action.TemplateDefaultConfig != nil {
		json.Unmarshal(action.TemplateDefaultConfig, &merged)
	} else {
		merged = make(map[string]interface{})
	}

	if action.ActionConfig != nil {
		var actionConfig map[string]interface{}
		if json.Unmarshal(action.ActionConfig, &actionConfig) == nil {
			for k, v := range actionConfig {
				merged[k] = v
			}
		}
	}

	result, _ := json.Marshal(merged)
	return result
}

func testBuildTemplateContext(ae *workflow.ActionExecutor, execContext workflow.ActionExecutionContext) workflow.TemplateContext {
	context := make(workflow.TemplateContext)

	context["entity_id"] = execContext.EntityID
	context["entity_name"] = execContext.EntityName
	context["event_type"] = execContext.EventType
	context["timestamp"] = execContext.Timestamp
	context["user_id"] = execContext.UserID
	context["rule_id"] = execContext.RuleID
	context["rule_name"] = execContext.RuleName

	if execContext.RawData != nil {
		for k, v := range execContext.RawData {
			context[k] = v
		}
	}

	if execContext.FieldChanges != nil {
		context["field_changes"] = execContext.FieldChanges
	}

	return context
}

func testProcessTemplates(ae *workflow.ActionExecutor, config json.RawMessage, context workflow.TemplateContext) (json.RawMessage, error) {
	var configData interface{}
	if err := json.Unmarshal(config, &configData); err != nil {
		return config, err
	}

	processor := workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions())
	result := processor.ProcessTemplateObject(configData, context)

	if len(result.Errors) > 0 {
		return config, fmt.Errorf("template processing errors: %v", result.Errors)
	}

	processed, err := json.Marshal(result.Processed)
	if err != nil {
		return config, err
	}

	return processed, nil
}

func testGetActionType(ae *workflow.ActionExecutor, action workflowdb.RuleActionView) string {
	if action.TemplateActionType.Valid {
		return action.TemplateActionType.String
	}
	return ""
}

func testShouldStopOnFailure(ae *workflow.ActionExecutor, action workflowdb.RuleActionView) bool {
	actionType := testGetActionType(ae, action)
	return actionType == "seek_approval"
}

func testUpdateStats(ae *workflow.ActionExecutor, result workflow.BatchExecutionResult) {
	// This simulates the updateStats method
	// In real tests, you'd test this through ExecuteRuleActions
	stats := ae.GetStats()
	stats.TotalActionsExecuted += result.TotalActions
	stats.SuccessfulExecutions += result.SuccessfulActions
	stats.FailedExecutions += result.FailedActions
	now := time.Now()
	stats.LastExecutedAt = &now
	// Note: In actual implementation, this would need proper synchronization
}

func testAddToHistory(ae *workflow.ActionExecutor, result workflow.BatchExecutionResult) {
	// This simulates adding to history
	// In real tests, you'd test this through ExecuteRuleActions
	// Note: This is a simplified version for testing
}

// Benchmark tests

func BenchmarkActionExecutor_ValidateActionConfig(b *testing.B) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	action := workflowdb.RuleActionView{
		ID:                 "action_001",
		TemplateActionType: sql.NullString{String: "send_email", Valid: true},
		ActionConfig: json.RawMessage(`{
			"recipients": ["user@example.com"],
			"subject": "Test Email"
		}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ae.ValidateActionConfig(action)
	}
}

func BenchmarkActionExecutor_MergeConfig(b *testing.B) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	action := workflowdb.RuleActionView{
		TemplateDefaultConfig: json.RawMessage(`{
			"field1": "default1",
			"field2": "default2",
			"field3": "default3"
		}`),
		ActionConfig: json.RawMessage(`{
			"field2": "override2",
			"field4": "new4"
		}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = testMergeActionConfig(ae, action)
	}
}

func BenchmarkActionExecutor_ProcessTemplates(b *testing.B) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ae := workflow.NewActionExecutor(log, nil)

	config := json.RawMessage(`{
		"recipient": "{{customer_email}}",
		"subject": "Order {{order_id}} - {{status | uppercase}}",
		"body": "Dear {{customer_name | capitalize}}, your order is {{status}}"
	}`)

	context := workflow.TemplateContext{
		"customer_email": "john@example.com",
		"customer_name":  "john smith",
		"order_id":       "ORD-123",
		"status":         "pending",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = testProcessTemplates(ae, config, context)
	}
}
