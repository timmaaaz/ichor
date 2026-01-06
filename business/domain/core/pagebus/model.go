package pagebus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// Page represents information about an individual page.
type Page struct {
	ID         uuid.UUID `json:"id"`
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Module     string    `json:"module"`
	Icon       string    `json:"icon"`
	SortOrder  int       `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
	ShowInMenu bool      `json:"show_in_menu"`
}

// NewPage contains information needed to create a new page.
type NewPage struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Module     string `json:"module"`
	Icon       string `json:"icon"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
	ShowInMenu bool   `json:"show_in_menu"`
}

// UpdatePage contains information needed to update a page.
type UpdatePage struct {
	Path       *string `json:"path,omitempty"`
	Name       *string `json:"name,omitempty"`
	Module     *string `json:"module,omitempty"`
	Icon       *string `json:"icon,omitempty"`
	SortOrder  *int    `json:"sort_order,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
	ShowInMenu *bool   `json:"show_in_menu,omitempty"`
}
