package brandbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID          *uuid.UUID
	Name        *string
	ContactInfo *uuid.UUID
	CreatedDate *time.Time
	UpdatedDate *time.Time
}
