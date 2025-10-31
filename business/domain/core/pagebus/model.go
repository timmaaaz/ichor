package pagebus

import "github.com/google/uuid"

// Page represents information about an individual page.
type Page struct {
	ID         uuid.UUID
	Path       string
	Name       string
	Module     string
	Icon       string
	SortOrder  int
	IsActive   bool
	ShowInMenu bool
}

// NewPage contains information needed to create a new page.
type NewPage struct {
	Path       string
	Name       string
	Module     string
	Icon       string
	SortOrder  int
	IsActive   bool
	ShowInMenu bool
}

// UpdatePage contains information needed to update a page.
type UpdatePage struct {
	Path       *string
	Name       *string
	Module     *string
	Icon       *string
	SortOrder  *int
	IsActive   *bool
	ShowInMenu *bool
}
