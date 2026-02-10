package commands

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
)

// workflowEntry pairs a workflow action config with its metadata for validation reporting
type workflowEntry struct {
	name       string
	actionType string
	config     json.RawMessage
}

// ValidateWorkflows validates sample workflow action configurations.
// Returns nil if all configs are valid, otherwise returns validation errors.
// This command does not require a database connection.
func ValidateWorkflows() error {
	// Create action registry with handlers in validation-only mode (nil dependencies)
	registry := createValidationRegistry()

	// Collect all sample action configs to validate
	configs := collectWorkflowConfigs()

	var (
		hasErrors    bool
		validCount   int
		invalidCount int
		warnCount    int
	)

	fmt.Println("Validating workflow action configurations...")
	fmt.Println()

	for _, entry := range configs {
		handler, exists := registry.Get(entry.actionType)
		if !exists {
			invalidCount++
			fmt.Printf("❌ %s: unknown action type %q\n", entry.name, entry.actionType)
			hasErrors = true
			continue
		}

		if err := handler.Validate(entry.config); err != nil {
			hasErrors = true
			invalidCount++
			fmt.Printf("❌ %s:\n", entry.name)
			fmt.Printf("   • %s\n", err.Error())
		} else {
			validCount++
			fmt.Printf("✓ %s\n", entry.name)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d valid, %d invalid, %d warnings\n", validCount, invalidCount, warnCount)

	if hasErrors {
		return fmt.Errorf("validation failed: %d workflow config(s) have errors", invalidCount)
	}

	fmt.Println("\nAll workflow configurations valid!")
	return nil
}

// createValidationRegistry creates an action registry with handlers
// initialized for validation-only mode (nil dependencies are acceptable
// because Validate() doesn't use them).
func createValidationRegistry() *workflow.ActionRegistry {
	registry := workflow.NewActionRegistry()

	// Register handlers with nil dependencies - Validate() doesn't use them
	registry.Register(data.NewUpdateFieldHandler(nil, nil))
	registry.Register(approval.NewSeekApprovalHandler(nil, nil))
	registry.Register(communication.NewSendEmailHandler(nil, nil))
	registry.Register(communication.NewSendNotificationHandler(nil, nil))
	registry.Register(communication.NewCreateAlertHandler(nil, nil, nil))
	registry.Register(inventory.NewAllocateInventoryHandler(nil, nil, nil, nil, nil, nil, nil))

	return registry
}

// collectWorkflowConfigs returns sample action configurations for validation.
// These represent valid configs that demonstrate the supported action types.
// Add your production workflow action configs here for validation.
func collectWorkflowConfigs() []workflowEntry {
	return []workflowEntry{
		// =====================================================================
		// create_alert action configs
		// =====================================================================
		{
			name:       "AlertWithUserRecipient",
			actionType: "create_alert",
			config: json.RawMessage(`{
				"alert_type": "low_inventory",
				"severity": "high",
				"title": "Low Inventory Alert",
				"message": "Product {{product_name}} is below threshold",
				"recipients": {
					"users": ["00000000-0000-0000-0000-000000000001"],
					"roles": []
				}
			}`),
		},
		{
			name:       "AlertWithRoleRecipient",
			actionType: "create_alert",
			config: json.RawMessage(`{
				"alert_type": "order_created",
				"severity": "medium",
				"message": "New order created: {{order_id}}",
				"recipients": {
					"users": [],
					"roles": ["00000000-0000-0000-0000-000000000002"]
				}
			}`),
		},
		{
			name:       "AlertWithResolvePrior",
			actionType: "create_alert",
			config: json.RawMessage(`{
				"alert_type": "stock_replenished",
				"severity": "low",
				"message": "Stock has been replenished for {{product_name}}",
				"recipients": {
					"users": ["00000000-0000-0000-0000-000000000001"]
				},
				"resolve_prior": true
			}`),
		},

		// =====================================================================
		// update_field action configs
		// =====================================================================
		{
			name:       "UpdateFieldBasic",
			actionType: "update_field",
			config: json.RawMessage(`{
				"target_entity": "sales.orders",
				"target_field": "status",
				"new_value": "processing"
			}`),
		},
		{
			name:       "UpdateFieldWithConditions",
			actionType: "update_field",
			config: json.RawMessage(`{
				"target_entity": "inventory.inventory_items",
				"target_field": "quantity",
				"new_value": "{{new_quantity}}",
				"conditions": [
					{"field_name": "id", "operator": "equals", "value": "{{entity_id}}"}
				]
			}`),
		},
		{
			name:       "UpdateFieldWithForeignKey",
			actionType: "update_field",
			config: json.RawMessage(`{
				"target_entity": "sales.orders",
				"target_field": "customer_id",
				"new_value": "{{customer_name}}",
				"field_type": "foreign_key",
				"foreign_key_config": {
					"reference_table": "sales.customers",
					"lookup_field": "name",
					"id_field": "id"
				}
			}`),
		},

		// =====================================================================
		// send_email action configs
		// =====================================================================
		{
			name:       "EmailBasic",
			actionType: "send_email",
			config: json.RawMessage(`{
				"recipients": ["user@example.com"],
				"subject": "Order Confirmation",
				"body": "Your order {{order_id}} has been confirmed."
			}`),
		},
		{
			name:       "EmailMultipleRecipients",
			actionType: "send_email",
			config: json.RawMessage(`{
				"recipients": ["admin@example.com", "manager@example.com"],
				"subject": "Critical Alert: {{alert_type}}",
				"body": "Please review the following alert immediately."
			}`),
		},

		// =====================================================================
		// send_notification action configs
		// =====================================================================
		{
			name:       "NotificationMultiChannel",
			actionType: "send_notification",
			config: json.RawMessage(`{
				"recipients": ["user-123", "user-456"],
				"channels": [
					{"type": "email"},
					{"type": "in_app"}
				],
				"priority": "high"
			}`),
		},
		{
			name:       "NotificationCritical",
			actionType: "send_notification",
			config: json.RawMessage(`{
				"recipients": ["admin-001"],
				"channels": [{"type": "sms"}, {"type": "push"}],
				"priority": "critical"
			}`),
		},

		// =====================================================================
		// seek_approval action configs
		// =====================================================================
		{
			name:       "ApprovalAny",
			actionType: "seek_approval",
			config: json.RawMessage(`{
				"approvers": ["00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"],
				"approval_type": "any"
			}`),
		},
		{
			name:       "ApprovalAll",
			actionType: "seek_approval",
			config: json.RawMessage(`{
				"approvers": ["00000000-0000-0000-0000-000000000001"],
				"approval_type": "all"
			}`),
		},
		{
			name:       "ApprovalMajority",
			actionType: "seek_approval",
			config: json.RawMessage(`{
				"approvers": ["user-1", "user-2", "user-3"],
				"approval_type": "majority"
			}`),
		},

		// =====================================================================
		// allocate_inventory action configs
		// =====================================================================
		{
			name:       "AllocationFIFO",
			actionType: "allocate_inventory",
			config: json.RawMessage(`{
				"inventory_items": [
					{"product_id": "00000000-0000-0000-0000-000000000001", "quantity": 10}
				],
				"allocation_mode": "allocate",
				"allocation_strategy": "fifo",
				"priority": "medium"
			}`),
		},
		{
			name:       "ReservationWithExpiry",
			actionType: "allocate_inventory",
			config: json.RawMessage(`{
				"inventory_items": [
					{"product_id": "00000000-0000-0000-0000-000000000001", "quantity": 5}
				],
				"allocation_mode": "reserve",
				"allocation_strategy": "nearest_expiry",
				"reservation_duration_hours": 48,
				"priority": "high",
				"allow_partial": true
			}`),
		},
		{
			name:       "AllocationFromLineItem",
			actionType: "allocate_inventory",
			config: json.RawMessage(`{
				"source_from_line_item": true,
				"allocation_mode": "allocate",
				"allocation_strategy": "lifo",
				"priority": "critical"
			}`),
		},
	}
}
