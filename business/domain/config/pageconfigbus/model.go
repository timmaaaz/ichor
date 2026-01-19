package pageconfigbus

import (
	"encoding/json"

	"github.com/google/uuid"
)

// PageConfig represents a page configuration that can be user-specific or default
type PageConfig struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	UserID    uuid.UUID `json:"user_id"` // Zero value means this is a default config
	IsDefault bool      `json:"is_default"`
}

// NewPageConfig contains data required to create a new page configuration
type NewPageConfig struct {
	Name      string    `json:"name"`
	UserID    uuid.UUID `json:"user_id"`
	IsDefault bool      `json:"is_default"`
}

// UpdatePageConfig contains data for updating an existing page configuration
type UpdatePageConfig struct {
	Name      *string    `json:"name,omitempty"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	IsDefault *bool      `json:"is_default,omitempty"`
}

// PageConfigWithRelations represents a page config with its content and actions.
type PageConfigWithRelations struct {
	PageConfig PageConfig          `json:"page_config"`
	Contents   []PageContentExport `json:"contents"`
	Actions    PageActionsExport   `json:"actions"`
}

// PageContentExport represents page content for export (simplified structure).
type PageContentExport struct {
	ID            uuid.UUID       `json:"id"`
	PageConfigID  uuid.UUID       `json:"page_config_id"`
	ContentType   string          `json:"content_type"`
	Label         string          `json:"label"`
	TableConfigID uuid.UUID       `json:"table_config_id"`
	FormID        uuid.UUID       `json:"form_id"`
	ChartConfigID uuid.UUID       `json:"chart_config_id"`
	OrderIndex    int             `json:"order_index"`
	ParentID      uuid.UUID       `json:"parent_id"`
	Layout        json.RawMessage `json:"layout"`
	IsVisible     bool            `json:"is_visible"`
	IsDefault     bool            `json:"is_default"`
}

// PageActionsExport represents page actions for export (grouped by type).
type PageActionsExport struct {
	Buttons    []PageActionExport `json:"buttons"`
	Dropdowns  []PageActionExport `json:"dropdowns"`
	Separators []PageActionExport `json:"separators"`
}

// PageActionExport represents a single page action for export.
type PageActionExport struct {
	ID           uuid.UUID             `json:"id"`
	PageConfigID uuid.UUID             `json:"page_config_id"`
	ActionType   string                `json:"action_type"`
	ActionOrder  int                   `json:"action_order"`
	IsActive     bool                  `json:"is_active"`
	Button       *ButtonActionExport   `json:"button,omitempty"`
	Dropdown     *DropdownActionExport `json:"dropdown,omitempty"`
}

// ButtonActionExport represents button-specific data for export.
type ButtonActionExport struct {
	Label              string `json:"label"`
	Icon               string `json:"icon"`
	TargetPath         string `json:"target_path"`
	Variant            string `json:"variant"`
	Alignment          string `json:"alignment"`
	ConfirmationPrompt string `json:"confirmation_prompt"`
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
	TargetPath string    `json:"target_path"`
	ItemOrder  int       `json:"item_order"`
}

// ImportStats represents statistics from an import operation.
type ImportStats struct {
	ImportedCount int `json:"imported_count"`
	SkippedCount  int `json:"skipped_count"`
	UpdatedCount  int `json:"updated_count"`
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
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// ValidationError represents a single validation error with JSON path (business layer)
type ValidationError struct {
	Field   string `json:"field"`   // JSON path: "contents[0].config.tableConfigId"
	Message string `json:"message"` // User-friendly error message
	Code    string `json:"code"`    // Error code for i18n
}
