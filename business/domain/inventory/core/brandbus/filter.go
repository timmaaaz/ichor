package brandbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID            *uuid.UUID
	Name          *string
	ContactInfoID *uuid.UUID
	CreatedDate   *time.Time
	UpdatedDate   *time.Time
}
