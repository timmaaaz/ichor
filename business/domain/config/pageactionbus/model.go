package pageactionbus

import "github.com/google/uuid"

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
	ID           uuid.UUID
	PageConfigID uuid.UUID
	ActionType   ActionType
	ActionOrder  int
	IsActive     bool

	// Type-specific data (only one populated based on ActionType)
	Button   *ButtonAction
	Dropdown *DropdownAction
}

// ButtonAction contains button-specific configuration.
type ButtonAction struct {
	Label              string
	Icon               string
	TargetPath         string
	Variant            string
	Alignment          string
	ConfirmationPrompt string
}

// DropdownAction contains dropdown-specific configuration including items.
type DropdownAction struct {
	Label string
	Icon  string
	Items []DropdownItem
}

// DropdownItem represents a single item within a dropdown action.
type DropdownItem struct {
	ID             uuid.UUID
	Label          string
	TargetPath     string
	ItemOrder      int
}

// NewButtonAction contains information needed to create a button action.
type NewButtonAction struct {
	PageConfigID       uuid.UUID
	ActionOrder        int
	IsActive           bool
	Label              string
	Icon               string
	TargetPath         string
	Variant            string
	Alignment          string
	ConfirmationPrompt string
}

// NewDropdownAction contains information needed to create a dropdown action.
type NewDropdownAction struct {
	PageConfigID uuid.UUID
	ActionOrder  int
	IsActive     bool
	Label        string
	Icon         string
	Items        []NewDropdownItem
}

// NewDropdownItem contains information needed to create a dropdown item.
type NewDropdownItem struct {
	Label      string
	TargetPath string
	ItemOrder  int
}

// NewSeparatorAction contains information needed to create a separator action.
type NewSeparatorAction struct {
	PageConfigID uuid.UUID
	ActionOrder  int
	IsActive     bool
}

// UpdateButtonAction contains information needed to update a button action.
type UpdateButtonAction struct {
	PageConfigID       *uuid.UUID
	ActionOrder        *int
	IsActive           *bool
	Label              *string
	Icon               *string
	TargetPath         *string
	Variant            *string
	Alignment          *string
	ConfirmationPrompt *string
}

// UpdateDropdownAction contains information needed to update a dropdown action.
type UpdateDropdownAction struct {
	PageConfigID *uuid.UUID
	ActionOrder  *int
	IsActive     *bool
	Label        *string
	Icon         *string
	Items        *[]NewDropdownItem // Replaces all items
}

// UpdateSeparatorAction contains information needed to update a separator action.
type UpdateSeparatorAction struct {
	PageConfigID *uuid.UUID
	ActionOrder  *int
	IsActive     *bool
}

// ActionsGroupedByType represents page actions grouped by their type.
type ActionsGroupedByType struct {
	Buttons    []PageAction
	Dropdowns  []PageAction
	Separators []PageAction
}
