package pagecontentbus

import "github.com/google/uuid"

// QueryFilter represents filtering options for page content queries
type QueryFilter struct {
	ID           *uuid.UUID
	PageConfigID *uuid.UUID
	ContentType  *string
	ParentID     *uuid.UUID
	IsVisible    *bool
}
