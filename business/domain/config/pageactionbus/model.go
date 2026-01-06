package pageactionbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// ActionType represents the type of page action.
type ActionType string

const (
	ActionTypeButton    ActionType = "button"
	ActionTypeDropdown  ActionType = "dropdown"
	ActionTypeSeparator ActionType = "separator"
)

// PageAction represents a unified page action. Only one of Button, Dropdown,
// or nil (for separator) will be populated based on ActionType.
type PageAction struct {
	ID           uuid.UUID  `json:"id"`
	PageConfigID uuid.UUID  `json:"page_config_id"`
	ActionType   ActionType `json:"action_type"`
	ActionOrder  int        `json:"action_order"`
	IsActive     bool       `json:"is_active"`

	// Type-specific data (only one populated based on ActionType)
	Button   *ButtonAction   `json:"button,omitempty"`
	Dropdown *DropdownAction `json:"dropdown,omitempty"`
}

// ButtonAction contains button-specific configuration.
type ButtonAction struct {
	Label              string `json:"label"`
	Icon               string `json:"icon"`
	TargetPath         string `json:"target_path"`
	Variant            string `json:"variant"`
	Alignment          string `json:"alignment"`
	ConfirmationPrompt string `json:"confirmation_prompt"`
}

// DropdownAction contains dropdown-specific configuration including items.
type DropdownAction struct {
	Label string         `json:"label"`
	Icon  string         `json:"icon"`
	Items []DropdownItem `json:"items"`
}

// DropdownItem represents a single item within a dropdown action.
type DropdownItem struct {
	ID         uuid.UUID `json:"id"`
	Label      string    `json:"label"`
	TargetPath string    `json:"target_path"`
	ItemOrder  int       `json:"item_order"`
}

// NewButtonAction contains information needed to create a button action.
type NewButtonAction struct {
	PageConfigID       uuid.UUID `json:"page_config_id"`
	ActionOrder        int       `json:"action_order"`
	IsActive           bool      `json:"is_active"`
	Label              string    `json:"label"`
	Icon               string    `json:"icon"`
	TargetPath         string    `json:"target_path"`
	Variant            string    `json:"variant"`
	Alignment          string    `json:"alignment"`
	ConfirmationPrompt string    `json:"confirmation_prompt"`
}

// NewDropdownAction contains information needed to create a dropdown action.
type NewDropdownAction struct {
	PageConfigID uuid.UUID         `json:"page_config_id"`
	ActionOrder  int               `json:"action_order"`
	IsActive     bool              `json:"is_active"`
	Label        string            `json:"label"`
	Icon         string            `json:"icon"`
	Items        []NewDropdownItem `json:"items"`
}

// NewDropdownItem contains information needed to create a dropdown item.
type NewDropdownItem struct {
	Label      string `json:"label"`
	TargetPath string `json:"target_path"`
	ItemOrder  int    `json:"item_order"`
}

// NewSeparatorAction contains information needed to create a separator action.
type NewSeparatorAction struct {
	PageConfigID uuid.UUID `json:"page_config_id"`
	ActionOrder  int       `json:"action_order"`
	IsActive     bool      `json:"is_active"`
}

// UpdateButtonAction contains information needed to update a button action.
type UpdateButtonAction struct {
	PageConfigID       *uuid.UUID `json:"page_config_id,omitempty"`
	ActionOrder        *int       `json:"action_order,omitempty"`
	IsActive           *bool      `json:"is_active,omitempty"`
	Label              *string    `json:"label,omitempty"`
	Icon               *string    `json:"icon,omitempty"`
	TargetPath         *string    `json:"target_path,omitempty"`
	Variant            *string    `json:"variant,omitempty"`
	Alignment          *string    `json:"alignment,omitempty"`
	ConfirmationPrompt *string    `json:"confirmation_prompt,omitempty"`
}

// UpdateDropdownAction contains information needed to update a dropdown action.
type UpdateDropdownAction struct {
	PageConfigID *uuid.UUID         `json:"page_config_id,omitempty"`
	ActionOrder  *int               `json:"action_order,omitempty"`
	IsActive     *bool              `json:"is_active,omitempty"`
	Label        *string            `json:"label,omitempty"`
	Icon         *string            `json:"icon,omitempty"`
	Items        *[]NewDropdownItem `json:"items,omitempty"` // Replaces all items
}

// UpdateSeparatorAction contains information needed to update a separator action.
type UpdateSeparatorAction struct {
	PageConfigID *uuid.UUID `json:"page_config_id,omitempty"`
	ActionOrder  *int       `json:"action_order,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// ActionsGroupedByType represents page actions grouped by their type.
type ActionsGroupedByType struct {
	Buttons    []PageAction `json:"buttons"`
	Dropdowns  []PageAction `json:"dropdowns"`
	Separators []PageAction `json:"separators"`
}
