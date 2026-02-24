package putawaytaskbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds optional filters for querying put-away tasks.
type QueryFilter struct {
	ID              *uuid.UUID
	ProductID       *uuid.UUID
	LocationID      *uuid.UUID
	Status          *Status
	AssignedTo      *uuid.UUID
	CreatedBy       *uuid.UUID
	ReferenceNumber *string
	CreatedDate     *time.Time
	UpdatedDate     *time.Time
}
