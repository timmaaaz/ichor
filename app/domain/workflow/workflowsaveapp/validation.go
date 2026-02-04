package workflowsaveapp

import (
	"encoding/json"
	"fmt"
)

// Action type constants
const (
	ActionTypeCreateAlert       = "create_alert"
	ActionTypeSendEmail         = "send_email"
	ActionTypeSendNotification  = "send_notification"
	ActionTypeUpdateField       = "update_field"
	ActionTypeSeekApproval      = "seek_approval"
	ActionTypeAllocateInventory = "allocate_inventory"
	ActionTypeEvaluateCondition = "evaluate_condition"
)

// ValidateActionConfigs validates the action configuration for each action
// based on its action type.
func ValidateActionConfigs(actions []SaveActionRequest) error {
	for i, action := range actions {
		if err := validateActionConfig(action.ActionType, action.ActionConfig); err != nil {
			return fmt.Errorf("action[%d] (%s): %w", i, action.Name, err)
		}
	}
	return nil
}

func validateActionConfig(actionType string, config json.RawMessage) error {
	if len(config) == 0 {
		return fmt.Errorf("action_config is required")
	}

	switch actionType {
	case ActionTypeCreateAlert:
		return validateCreateAlertConfig(config)
	case ActionTypeSendEmail:
		return validateSendEmailConfig(config)
	case ActionTypeSendNotification:
		return validateSendNotificationConfig(config)
	case ActionTypeUpdateField:
		return validateUpdateFieldConfig(config)
	case ActionTypeSeekApproval:
		return validateSeekApprovalConfig(config)
	case ActionTypeAllocateInventory:
		return validateAllocateInventoryConfig(config)
	case ActionTypeEvaluateCondition:
		return validateEvaluateConditionConfig(config)
	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}
}

// CreateAlertConfig defines the required fields for create_alert action.
type CreateAlertConfig struct {
	AlertType string `json:"alert_type"`
	Severity  string `json:"severity"`
	Title     string `json:"title"`
	Message   string `json:"message"`
}

func validateCreateAlertConfig(config json.RawMessage) error {
	var c CreateAlertConfig
	if err := json.Unmarshal(config, &c); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}
	if c.AlertType == "" {
		return fmt.Errorf("alert_type is required")
	}
	if c.Severity == "" {
		return fmt.Errorf("severity is required")
	}
	if c.Title == "" {
		return fmt.Errorf("title is required")
	}
	if c.Message == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

// SendEmailConfig defines the required fields for send_email action.
type SendEmailConfig struct {
	Recipients []string `json:"recipients"`
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
}

func validateSendEmailConfig(config json.RawMessage) error {
	var c SendEmailConfig
	if err := json.Unmarshal(config, &c); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}
	if len(c.Recipients) == 0 {
		return fmt.Errorf("recipients is required")
	}
	if c.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if c.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// SendNotificationConfig defines the required fields for send_notification action.
type SendNotificationConfig struct {
	Recipients []string `json:"recipients"`
	Channels   []string `json:"channels"`
}

func validateSendNotificationConfig(config json.RawMessage) error {
	var c SendNotificationConfig
	if err := json.Unmarshal(config, &c); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}
	if len(c.Recipients) == 0 {
		return fmt.Errorf("recipients is required")
	}
	if len(c.Channels) == 0 {
		return fmt.Errorf("channels is required")
	}
	return nil
}

// UpdateFieldConfig defines the required fields for update_field action.
type UpdateFieldConfig struct {
	TargetEntity string `json:"target_entity"`
	TargetField  string `json:"target_field"`
}

func validateUpdateFieldConfig(config json.RawMessage) error {
	var c UpdateFieldConfig
	if err := json.Unmarshal(config, &c); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}
	if c.TargetEntity == "" {
		return fmt.Errorf("target_entity is required")
	}
	if c.TargetField == "" {
		return fmt.Errorf("target_field is required")
	}
	return nil
}

// SeekApprovalConfig defines the required fields for seek_approval action.
type SeekApprovalConfig struct {
	Approvers    []string `json:"approvers"`
	ApprovalType string   `json:"approval_type"`
}

func validateSeekApprovalConfig(config json.RawMessage) error {
	var c SeekApprovalConfig
	if err := json.Unmarshal(config, &c); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}
	if len(c.Approvers) == 0 {
		return fmt.Errorf("approvers is required")
	}
	if c.ApprovalType == "" {
		return fmt.Errorf("approval_type is required")
	}
	return nil
}

// AllocateInventoryConfig defines the required fields for allocate_inventory action.
type AllocateInventoryConfig struct {
	InventoryItems []any  `json:"inventory_items"`
	AllocationMode string `json:"allocation_mode"`
}

func validateAllocateInventoryConfig(config json.RawMessage) error {
	var c AllocateInventoryConfig
	if err := json.Unmarshal(config, &c); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}
	if len(c.InventoryItems) == 0 {
		return fmt.Errorf("inventory_items is required")
	}
	if c.AllocationMode == "" {
		return fmt.Errorf("allocation_mode is required")
	}
	return nil
}

// EvaluateConditionConfig defines the required fields for evaluate_condition action.
type EvaluateConditionConfig struct {
	Conditions []any `json:"conditions"`
}

func validateEvaluateConditionConfig(config json.RawMessage) error {
	var c EvaluateConditionConfig
	if err := json.Unmarshal(config, &c); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}
	if len(c.Conditions) == 0 {
		return fmt.Errorf("conditions is required")
	}
	return nil
}
