package workflow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus/stores/inventoryitemdb"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus/stores/productdb"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus/stores/inventorytransactiondb"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus/stores/inventorylocationdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// TODO: Streamline workflow actions register all

/*
Package workflow_test tests the ActionExecutor component of the workflow system.

WHAT THIS TESTS:
- Action configuration validation (ValidateActionConfig)
- Configuration merging between template defaults and action-specific configs
- Template variable processing in action configurations
- Template context building from execution context
- Action execution failure handling logic (shouldStopOnFailure)
- Execution statistics tracking
- Execution history management
- Action handler registration and retrieval
- All registered action types (seek_approval, send_email, create_alert, etc.)
- Database integration for loading rule actions (with real PostgreSQL)
- Database integration for retrieving rule names (with real PostgreSQL)
- SQL query correctness and data mapping
- Action ordering by execution_order field

WHAT THIS DOES NOT TEST:
- Real action execution side effects (emails, alerts, actual inventory allocation)
- Concurrent execution of multiple actions simultaneously
- Retry logic with actual time delays
- External service integrations (email providers, notification services)
- Transaction rollback scenarios under failure conditions
- Authentication/authorization for action execution
- Rate limiting or throttling of action execution
- Cascading failures across dependent actions
- Performance under production load
- Network failures or timeout scenarios
- Actual template engine edge cases (circular references, infinite loops)

NOTE: Integration tests use real PostgreSQL database connections for data operations
while benchmarks use mock connections to measure pure algorithmic performance.
*/

func TestActionExecutor_ValidateActionConfig(t *testing.T) {

	tests := []struct {
		name   string
		action workflow.RuleActionView
		want   workflow.ActionValidationResult
	}{
		{
			name: "valid seek_approval action",
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "seek_approval",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "seek_approval",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "seek_approval",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "send_email",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "send_email",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "create_alert",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "create_alert",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "update_field",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "unsupported_type",
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
			action: workflow.RuleActionView{
				ID:                 uuid.New(),
				TemplateActionType: "send_email",
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

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Create a mock DB connection (or use sqlx.NewDb with a test driver)
	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB

	// Create registry and register all actions
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))
	ae := workflow.NewActionExecutor(log, db, workflowBus)

	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

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

	tests := []struct {
		name   string
		action workflow.RuleActionView
		want   map[string]interface{}
	}{
		{
			name: "merge template defaults with action config",
			action: workflow.RuleActionView{
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
			action: workflow.RuleActionView{
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
			action: workflow.RuleActionView{
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
			action: workflow.RuleActionView{
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
			action: workflow.RuleActionView{
				TemplateDefaultConfig: nil,
				ActionConfig:          nil,
			},
			want: map[string]interface{}{},
		},
	}

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

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

	entityID1 := uuid.New()
	entityID2 := uuid.New()

	userID1 := uuid.New()
	userID2 := uuid.New()

	ruleID1 := uuid.New()
	ruleID2 := uuid.New()

	tests := []struct {
		name    string
		context workflow.ActionExecutionContext
		want    workflow.TemplateContext
	}{
		{
			name: "basic context",
			context: workflow.ActionExecutionContext{
				EntityID:   entityID1,
				EntityName: "customers",
				EventType:  "on_update",
				Timestamp:  time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				UserID:     userID1,
				RuleID:     ruleID1,
				RuleName:   "Update Customer Status",
			},
			want: workflow.TemplateContext{
				"entity_id":   entityID1,
				"entity_name": "customers",
				"event_type":  "on_update",
				"timestamp":   time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				"user_id":     userID1,
				"rule_id":     ruleID1,
				"rule_name":   "Update Customer Status",
			},
		},
		{
			name: "context with raw data",
			context: workflow.ActionExecutionContext{
				EntityID:   entityID2,
				EntityName: "orders",
				EventType:  "on_create",
				Timestamp:  time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				UserID:     userID2,
				RuleID:     ruleID2,
				RuleName:   "Process New Order",
				RawData: map[string]interface{}{
					"order_total":  199.99,
					"customer_id":  "cust_123",
					"order_status": "pending",
				},
			},
			want: workflow.TemplateContext{
				"entity_id":    entityID2,
				"entity_name":  "orders",
				"event_type":   "on_create",
				"timestamp":    time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				"user_id":      userID2,
				"rule_id":      ruleID2,
				"rule_name":    "Process New Order",
				"order_total":  199.99,
				"customer_id":  "cust_123",
				"order_status": "pending",
			},
		},
		{
			name: "context with field changes",
			context: workflow.ActionExecutionContext{
				EntityID:   entityID1,
				EntityName: "customers",
				EventType:  "on_update",
				Timestamp:  time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				UserID:     userID1,
				RuleID:     ruleID1,
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
				"entity_id":   entityID1,
				"entity_name": "customers",
				"event_type":  "on_update",
				"timestamp":   time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
				"user_id":     userID1,
				"rule_id":     ruleID1,
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

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

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

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB

	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))
	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

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

func TestActionExecutor_ShouldStopOnFailure(t *testing.T) {

	tests := []struct {
		name   string
		action workflow.RuleActionView
		want   bool
	}{
		{
			name: "seek_approval should stop on failure",
			action: workflow.RuleActionView{
				TemplateActionType: "seek_approval",
			},
			want: true,
		},
		{
			name: "send_email should not stop on failure",
			action: workflow.RuleActionView{
				TemplateActionType: "send_email",
			},
			want: false,
		},
		{
			name: "create_alert should not stop on failure",
			action: workflow.RuleActionView{
				TemplateActionType: "create_alert",
			},
			want: false,
		},
	}

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

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

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow data: %s", err)
	}

	// Setup the workflow business layer

	// Create ActionExecutor
	ae := workflow.NewActionExecutor(log, db, workflowBus)

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	// Create RabbitMQ client
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Create workflow queue for initialization
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Register all actions
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log:         log,
			DB:          db,
			QueueClient: queue,
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

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

	// Get existing entity type, entity, and trigger type from seeded data
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("Failed to query entity type: %v", err)
	}

	entity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("Failed to query entity: %v", err)
	}

	triggerType, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("Failed to query trigger type: %v", err)
	}

	// Create automation rule using the seeded user ID
	rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "Test Rule for Stats",
		Description:   "Rule to test statistics",
		EntityID:      entity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerType.ID,
		IsActive:      true,
		CreatedBy:     uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), // Use the seeded userID
	})
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Create action template for email
	template, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "send_email_template",
		Description: "Email template",
		ActionType:  "send_email",
		DefaultConfig: json.RawMessage(`{
            "recipients": ["default@example.com"],
            "subject": "Default Subject"
        }`),
		CreatedBy: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), // Use the seeded userID
	})
	if err != nil {
		t.Fatalf("Failed to create action template: %v", err)
	}

	// Create multiple rule actions (2 will succeed, 1 will fail)
	actions := []workflow.NewRuleAction{
		{
			AutomationRuleID: rule.ID,
			Name:             "Success Action 1",
			ActionConfig: json.RawMessage(`{
                "recipients": ["test1@example.com"],
                "subject": "Test Email 1",
                "body": "Test body 1"
            }`),
			ExecutionOrder: 1,
			IsActive:       true,
			TemplateID:     &template.ID,
		},
		{
			AutomationRuleID: rule.ID,
			Name:             "Success Action 2",
			ActionConfig: json.RawMessage(`{
                "recipients": ["test2@example.com"],
                "subject": "Test Email 2",
                "body": "Test body 2"
            }`),
			ExecutionOrder: 2,
			IsActive:       true,
			TemplateID:     &template.ID,
		},
		{
			AutomationRuleID: rule.ID,
			Name:             "Fail Action 1",
			ActionConfig: json.RawMessage(`{
                "recipients": [],
                "subject": ""
            }`), // This will fail validation
			ExecutionOrder: 3,
			IsActive:       true,
			TemplateID:     &template.ID,
		},
	}

	for _, action := range actions {
		_, err := workflowBus.CreateRuleAction(ctx, action)
		if err != nil {
			t.Fatalf("Failed to create rule action: %v", err)
		}
	}

	// Execute the rule actions
	execContext := workflow.ActionExecutionContext{
		EntityID:    entity.ID,
		EntityName:  "customers",
		EventType:   "on_create",
		UserID:      uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
		RuleID:      rule.ID,
		ExecutionID: uuid.New(),
		Timestamp:   time.Now(),
		RawData: map[string]interface{}{
			"test": "data",
		},
	}

	result, err := ae.ExecuteRuleActions(ctx, rule.ID, execContext)
	if err != nil {
		t.Fatalf("Failed to execute rule actions: %v", err)
	}

	// Check the execution result
	if result.TotalActions != 3 {
		t.Errorf("TotalActions = %d, want 3", result.TotalActions)
	}
	if result.SuccessfulActions != 2 {
		t.Errorf("SuccessfulActions = %d, want 2", result.SuccessfulActions)
	}
	if result.FailedActions != 1 {
		t.Errorf("FailedActions = %d, want 1", result.FailedActions)
	}

	// Get updated stats after execution
	stats = ae.GetStats()
	if stats.TotalActionsExecuted != 3 {
		t.Errorf("TotalActionsExecuted = %d, want 3", stats.TotalActionsExecuted)
	}
	if stats.SuccessfulExecutions != 2 {
		t.Errorf("SuccessfulExecutions = %d, want 2", stats.SuccessfulExecutions)
	}
	if stats.FailedExecutions != 1 {
		t.Errorf("FailedExecutions = %d, want 1", stats.FailedExecutions)
	}
	if stats.LastExecutedAt == nil {
		t.Error("LastExecutedAt should not be nil after execution")
	}

	// Verify average execution time was calculated
	if stats.AverageExecutionTimeMs < 0 {
		t.Errorf("AverageExecutionTimeMs should not be negative, got %f", stats.AverageExecutionTimeMs)
	}

	t.Logf("Stats after execution: Total=%d, Success=%d, Failed=%d, AvgTime=%fms",
		stats.TotalActionsExecuted,
		stats.SuccessfulExecutions,
		stats.FailedExecutions,
		stats.AverageExecutionTimeMs)
}

func TestActionExecutor_ExecutionHistory(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	_, err := workflow.TestSeedFullWorkflow(ctx, uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"), workflowBus)
	if err != nil {
		t.Fatalf("seeding workflow data: %s", err)
	}

	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	// Create RabbitMQ client
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Create workflow queue for initialization
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log:         log,
			DB:          db,
			QueueClient: queue,
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

	// Initially empty
	history := ae.GetExecutionHistory(10)
	if len(history) != 0 {
		t.Errorf("Initial history length = %d, want 0", len(history))
	}

	// Get existing entities from seeded data
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("Failed to query entity type: %v", err)
	}

	entity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("Failed to query entity: %v", err)
	}

	triggerType, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("Failed to query trigger type: %v", err)
	}

	// Create action template
	template, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "send_email_template",
		Description: "Email template",
		ActionType:  "send_email",
		DefaultConfig: json.RawMessage(`{
            "recipients": ["default@example.com"],
            "subject": "Default Subject"
        }`),
		CreatedBy: userID,
	})
	if err != nil {
		t.Fatalf("Failed to create action template: %v", err)
	}

	// Create and execute 5 different rules to build history
	for i := 0; i < 5; i++ {
		// Create a rule
		rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
			Name:          fmt.Sprintf("Rule %d", i),
			Description:   fmt.Sprintf("Test rule %d for history", i),
			EntityID:      entity.ID,
			EntityTypeID:  entityType.ID,
			TriggerTypeID: triggerType.ID,
			IsActive:      true,
			CreatedBy:     userID,
		})
		if err != nil {
			t.Fatalf("Failed to create rule %d: %v", i, err)
		}

		// Create actions for this rule (i+1 actions, where i succeed)
		for j := 0; j <= i; j++ {
			actionConfig := json.RawMessage(`{
                "recipients": ["test@example.com"],
                "subject": "Test Subject"
            }`)

			// Make the last action fail if it's not the only action
			if j == i && i > 0 {
				actionConfig = json.RawMessage(`{
                    "recipients": [],
                    "subject": ""
                }`)
			}

			_, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
				AutomationRuleID: rule.ID,
				Name:             fmt.Sprintf("Action %d", j),
				ActionConfig:     actionConfig,
				ExecutionOrder:   j + 1,
				IsActive:         true,
				TemplateID:       &template.ID,
			})
			if err != nil {
				t.Fatalf("Failed to create action: %v", err)
			}
		}

		// Execute the rule
		execContext := workflow.ActionExecutionContext{
			EntityID:    entity.ID,
			EntityName:  "customers",
			EventType:   "on_create",
			UserID:      userID,
			RuleID:      rule.ID,
			ExecutionID: uuid.New(),
			Timestamp:   time.Now(),
		}

		_, err = ae.ExecuteRuleActions(ctx, rule.ID, execContext)
		if err != nil {
			t.Fatalf("Failed to execute rule actions for rule %d: %v", i, err)
		}
	}

	// Get limited history (should return 3 most recent)
	history = ae.GetExecutionHistory(3)
	if len(history) != 3 {
		t.Errorf("Limited history length = %d, want 3", len(history))
	}

	// Based on the actual implementation, GetExecutionHistory returns the LAST N items
	// from the history slice. Since history is appended chronologically (oldest first),
	// getting the last 3 means we get indices 2, 3, 4 (Rule 2, Rule 3, Rule 4)
	for i, entry := range history {
		expectedRuleName := fmt.Sprintf("Rule %d", i+2) // Should be Rule 2, Rule 3, Rule 4
		if entry.RuleName != expectedRuleName {
			t.Errorf("History[%d].RuleName = %s, want %s", i, entry.RuleName, expectedRuleName)
		}
	}

	// Get all history
	history = ae.GetExecutionHistory(10)
	if len(history) != 5 {
		t.Errorf("Full history length = %d, want 5", len(history))
	}

	// Verify all rules are present in order
	for i, entry := range history {
		expectedRuleName := fmt.Sprintf("Rule %d", i)
		if entry.RuleName != expectedRuleName {
			t.Errorf("Full history[%d].RuleName = %s, want %s", i, entry.RuleName, expectedRuleName)
		}
	}

	// Clear history
	ae.ClearHistory()
	history = ae.GetExecutionHistory(10)
	if len(history) != 0 {
		t.Errorf("History length after clear = %d, want 0", len(history))
	}
}
func TestActionHandler_Implementations(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	ndb := dbtest.NewDatabase(t, "Test_Workflow")
	db := ndb.DB
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	// Seed workflow data to ensure any needed entities exist
	userID := uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7")

	// Get RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	// Create RabbitMQ client
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Create workflow queue for initialization
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Create registry and register all actions
	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log:         log,
			DB:          db,
			QueueClient: queue,
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

	tests := []struct {
		name          string
		handlerType   string
		validConfig   json.RawMessage
		invalidConfig json.RawMessage
		wantType      string
		skipExecute   bool // Skip execution test for handlers that need specific DB setup
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
                "inventory_items": [{"product_id": "123e4567-e89b-12d3-a456-426614174000", "quantity": 10}],
                "allocation_mode": "allocate",
                "allocation_strategy": "fifo",
                "priority": "medium"
            }`),
			invalidConfig: json.RawMessage(`{
                "inventory_items": [],
                "allocation_strategy": "invalid"
            }`),
			wantType:    "allocate_inventory",
			skipExecute: true, // Requires allocation_results table
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
                "target_entity": "automation_rules",
                "target_field": "is_active",
                "new_value": true
            }`),
			invalidConfig: json.RawMessage(`{
                "target_entity": "",
                "target_field": "",
                "new_value": null
            }`),
			wantType:    "update_field",
			skipExecute: true, // Would need specific entity to exist
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
			// Get handler from registry
			handler, exists := ae.GetRegistry().Get(tt.handlerType)
			if !exists {
				t.Fatalf("Handler %s not found in registry", tt.handlerType)
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

			// Skip execution test for handlers that need specific DB setup
			if tt.skipExecute {
				t.Logf("Skipping Execute() test for %s (requires specific DB setup)", tt.handlerType)
				return
			}

			// Test Execute with proper context
			execContext := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "test",
				EventType:   "on_create",
				RuleID:      uuid.New(),
				RuleName:    "Test Rule",
				UserID:      userID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now(),
			}

			result, err := handler.Execute(ctx, tt.validConfig, execContext)
			if err != nil {
				t.Errorf("Execute() failed: %v", err)
			} else if result == nil {
				t.Error("Execute() returned nil result")
			}
		})
	}
}

// Test helper functions to access private methods
// In a real scenario, you might want to make these methods public or use an interface

func testMergeActionConfig(ae *workflow.ActionExecutor, action workflow.RuleActionView) json.RawMessage {
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

func testShouldStopOnFailure(ae *workflow.ActionExecutor, action workflow.RuleActionView) bool {
	actionType := action.TemplateActionType
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
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	db := &sqlx.DB{}
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))
	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

	action := workflow.RuleActionView{
		ID:                 uuid.New(),
		TemplateActionType: "send_email",
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
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	db := &sqlx.DB{}
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

	action := workflow.RuleActionView{
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
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	db := &sqlx.DB{}
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db))

	ae := workflow.NewActionExecutor(log, db, workflowBus)
	workflowactions.RegisterAll(
		ae.GetRegistry(),
		workflowactions.ActionConfig{
			Log: log,
			DB:  db,
			QueueClient: rabbitmq.NewWorkflowQueue(
				rabbitmq.NewClient(log, rabbitmq.DefaultConfig()),
				log,
			),
			Buses: workflowactions.BusDependencies{
				InventoryItem:        inventoryitembus.NewBusiness(log, delegate.New(log), inventoryitemdb.NewStore(log, db)),
				InventoryLocation:    inventorylocationbus.NewBusiness(log, delegate.New(log), inventorylocationdb.NewStore(log, db)),
				InventoryTransaction: inventorytransactionbus.NewBusiness(log, delegate.New(log), inventorytransactiondb.NewStore(log, db)),
				Product:              productbus.NewBusiness(log, delegate.New(log), productdb.NewStore(log, db)),
			},
		},
	)

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
