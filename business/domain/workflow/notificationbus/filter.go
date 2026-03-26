package notificationbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields to filter notifications.
type QueryFilter struct {
	ID               *uuid.UUID
	UserID           *uuid.UUID
	IsRead           *bool
	Priority         *string
	SourceEntityName *string
	SourceEntityID   *uuid.UUID
	CreatedAfter     *time.Time
	CreatedBefore    *time.Time
}
