package cyclecountitembus

import "github.com/google/uuid"

// QueryFilter holds optional filters for querying cycle count items.
type QueryFilter struct {
	ID         *uuid.UUID
	SessionID  *uuid.UUID
	ProductID  *uuid.UUID
	LocationID *uuid.UUID
	Status     *Status
	CountedBy  *uuid.UUID
}
