package pageconfigbus

import "github.com/google/uuid"

// PageConfig represents a page configuration that can be user-specific or default
type PageConfig struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	UserID    uuid.UUID `json:"userId"` // Zero value means this is a default config
	IsDefault bool      `json:"isDefault"`
}

// NewPageConfig contains data required to create a new page configuration
type NewPageConfig struct {
	Name      string
	UserID    uuid.UUID
	IsDefault bool
}

// UpdatePageConfig contains data for updating an existing page configuration
type UpdatePageConfig struct {
	Name      *string
	UserID    *uuid.UUID
	IsDefault *bool
}

// PageConfigWithRelations represents a page config with its content and actions.
type PageConfigWithRelations struct {
	PageConfig PageConfig          `json:"pageConfig"`
	Contents   []PageContentExport `json:"contents"`
	Actions    PageActionsExport   `json:"actions"`
}

// PageContentExport represents page content for export (simplified structure).
type PageContentExport struct {
	ID            uuid.UUID `json:"id"`
	PageConfigID  uuid.UUID `json:"pageConfigId"`
	ContentType   string    `json:"contentType"`
	Label         string    `json:"label"`
	TableConfigID uuid.UUID `json:"tableConfigId"`
	FormID        uuid.UUID `json:"formId"`
	ChartConfigID uuid.UUID `json:"chartConfigId"`
	OrderIndex    int       `json:"orderIndex"`
	ParentID      uuid.UUID `json:"parentId"`
	Layout        []byte    `json:"layout"`
	IsVisible     bool      `json:"isVisible"`
	IsDefault     bool      `json:"isDefault"`
}

// PageActionsExport represents page actions for export (grouped by type).
type PageActionsExport struct {
	Buttons    []PageActionExport `json:"buttons"`
	Dropdowns  []PageActionExport `json:"dropdowns"`
	Separators []PageActionExport `json:"separators"`
}

// PageActionExport represents a single page action for export.
type PageActionExport struct {
	ID           uuid.UUID            `json:"id"`
	PageConfigID uuid.UUID            `json:"pageConfigId"`
	ActionType   string               `json:"actionType"`
	ActionOrder  int                  `json:"actionOrder"`
	IsActive     bool                 `json:"isActive"`
	Button       *ButtonActionExport  `json:"button,omitempty"`
	Dropdown     *DropdownActionExport `json:"dropdown,omitempty"`
}

// ButtonActionExport represents button-specific data for export.
type ButtonActionExport struct {
	Label              string `json:"label"`
	Icon               string `json:"icon"`
	TargetPath         string `json:"targetPath"`
	Variant            string `json:"variant"`
	Alignment          string `json:"alignment"`
	ConfirmationPrompt string `json:"confirmationPrompt"`
}

// DropdownActionExport represents dropdown-specific data for export.
type DropdownActionExport struct {
	Label string               `json:"label"`
	Icon  string               `json:"icon"`
	Items []DropdownItemExport `json:"items"`
}

// DropdownItemExport represents a dropdown item for export.
type DropdownItemExport struct {
	ID         uuid.UUID `json:"id"`
	Label      string    `json:"label"`
	TargetPath string    `json:"targetPath"`
	ItemOrder  int       `json:"itemOrder"`
}

// ImportStats represents statistics from an import operation.
type ImportStats struct {
	ImportedCount int
	SkippedCount  int
	UpdatedCount  int
}

// Error code constants for validation
const (
	ErrCodeInvalidJSON       = "INVALID_JSON"
	ErrCodeRequiredField     = "REQUIRED_FIELD"
	ErrCodeInvalidReference  = "INVALID_REFERENCE"
	ErrCodeDatabaseError     = "DATABASE_ERROR"     // For database query failures during validation
	ErrCodeInvalidType       = "INVALID_CONTENT_TYPE"
	ErrCodeCircularReference = "CIRCULAR_PARENT_REF"
	ErrCodeMaxDepthExceeded  = "MAX_NESTING_DEPTH"
	ErrCodeMaxSizeExceeded   = "MAX_SIZE_EXCEEDED"
	ErrCodeRangeError        = "RANGE_ERROR"        // For values outside valid range (e.g., colSpan 1-12)
	ErrCodeInvalidFormat     = "INVALID_FORMAT"     // For incorrect format (e.g., invalid UUID)
	ErrCodeTimeout           = "VALIDATION_TIMEOUT" // For validation timeout
)

// ValidationResult represents the outcome of validating a page config import (business layer)
// NOTE: Does NOT implement web.Encoder - that's the app layer's job
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidationError represents a single validation error with JSON path (business layer)
type ValidationError struct {
	Field   string // JSON path: "contents[0].config.tableConfigId"
	Message string // User-friendly error message
	Code    string // Error code for i18n
}
