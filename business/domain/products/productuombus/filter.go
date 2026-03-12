package productuombus

import "github.com/google/uuid"

// QueryFilter holds the available filters for querying product UOMs.
type QueryFilter struct {
	ID        *uuid.UUID
	ProductID *uuid.UUID
	IsBase    *bool
	Name      *string
}
