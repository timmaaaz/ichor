package formbus

import "github.com/google/uuid"

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *string
}
