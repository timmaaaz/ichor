package productcategorybus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID          *uuid.UUID
	Name        *string
	Description *string
	CreatedDate *time.Time
	UpdatedDate *time.Time
}
