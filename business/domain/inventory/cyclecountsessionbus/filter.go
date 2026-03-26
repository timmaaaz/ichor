package cyclecountsessionbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds optional filters for querying cycle count sessions.
type QueryFilter struct {
	ID          *uuid.UUID
	Name        *string
	Status      *Status
	CreatedBy   *uuid.UUID
	CreatedDate *time.Time
}
