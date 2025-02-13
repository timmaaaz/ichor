package commentbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID          *uuid.UUID
	CommenterID *uuid.UUID
	UserID      *uuid.UUID
	CreatedDate *time.Time
	Comment     *string
}
