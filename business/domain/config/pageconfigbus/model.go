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
