package serialnumberbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	SerialID     *uuid.UUID
	LotID        *uuid.UUID
	ProductID    *uuid.UUID
	LocationID   *uuid.UUID
	SerialNumber *string
	Status       *string
	CreatedDate  *time.Time
	UpdatedDate  *time.Time
}
