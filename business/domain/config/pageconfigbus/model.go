package pageconfigbus

import "github.com/google/uuid"

// PageConfig represents a page configuration that can be user-specific or default
type PageConfig struct {
	ID        uuid.UUID
	Name      string
	UserID    uuid.UUID // Zero value means this is a default config
	IsDefault bool
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
	PageConfig PageConfig
	Contents   []PageContentExport
	Actions    PageActionsExport
}

// PageContentExport represents page content for export (simplified structure).
type PageContentExport struct {
	ID            uuid.UUID
	PageConfigID  uuid.UUID
	ContentType   string
	Label         string
	TableConfigID uuid.UUID
	FormID        uuid.UUID
	OrderIndex    int
	ParentID      uuid.UUID
	Layout        []byte
	IsVisible     bool
	IsDefault     bool
}

// PageActionsExport represents page actions for export (grouped by type).
type PageActionsExport struct {
	Buttons    []PageActionExport
	Dropdowns  []PageActionExport
	Separators []PageActionExport
}

// PageActionExport represents a single page action for export.
type PageActionExport struct {
	ID           uuid.UUID
	PageConfigID uuid.UUID
	ActionType   string
	ActionOrder  int
	IsActive     bool
	Button       *ButtonActionExport
	Dropdown     *DropdownActionExport
}

// ButtonActionExport represents button-specific data for export.
type ButtonActionExport struct {
	Label              string
	Icon               string
	TargetPath         string
	Variant            string
	Alignment          string
	ConfirmationPrompt string
}

// DropdownActionExport represents dropdown-specific data for export.
type DropdownActionExport struct {
	Label string
	Icon  string
	Items []DropdownItemExport
}

// DropdownItemExport represents a dropdown item for export.
type DropdownItemExport struct {
	ID         uuid.UUID
	Label      string
	TargetPath string
	ItemOrder  int
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
	ErrCodeInvalidType       = "INVALID_CONTENT_TYPE"
	ErrCodeCircularReference = "CIRCULAR_PARENT_REF"
	ErrCodeMaxDepthExceeded  = "MAX_NESTING_DEPTH"
	ErrCodeMaxSizeExceeded   = "MAX_SIZE_EXCEEDED"
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
