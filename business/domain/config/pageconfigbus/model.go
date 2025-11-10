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

// PageTabConfig represents a legacy tab configuration (kept for backward compatibility)
// This is being superseded by the page_content system
type PageTabConfig struct {
	ID           uuid.UUID
	PageConfigID uuid.UUID
	Label        string
	ConfigID     uuid.UUID
	IsDefault    bool
	TabOrder     int
}

// NewPageTabConfig contains data required to create a new page tab configuration
type NewPageTabConfig struct {
	PageConfigID uuid.UUID
	Label        string
	ConfigID     uuid.UUID
	IsDefault    bool
	TabOrder     int
}

// UpdatePageTabConfig contains data for updating an existing page tab configuration
type UpdatePageTabConfig struct {
	Label        *string
	PageConfigID *uuid.UUID
	ConfigID     *uuid.UUID
	IsDefault    *bool
	TabOrder     *int
}
