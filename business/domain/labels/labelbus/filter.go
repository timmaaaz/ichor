package labelbus

import "github.com/google/uuid"

// QueryFilter holds the available fields for label catalog queries.
type QueryFilter struct {
	ID   *uuid.UUID
	Code *string
	Type *string
}
