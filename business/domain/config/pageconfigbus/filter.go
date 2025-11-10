package pageconfigbus

import "github.com/google/uuid"

// QueryFilter represents filtering options for page config queries
type QueryFilter struct {
	ID        *uuid.UUID
	Name      *string
	UserID    *uuid.UUID
	IsDefault *bool
}
