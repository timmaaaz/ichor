package commentbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type UserApprovalComment struct {
	ID          uuid.UUID `json:"id"`
	Comment     string    `json:"comment"`
	CommenterID uuid.UUID `json:"commenter_id"`
	UserID      uuid.UUID `json:"user_id"`
	CreatedDate time.Time `json:"created_date"`
}

type NewUserApprovalComment struct {
	Comment     string     `json:"comment"`
	CommenterID uuid.UUID  `json:"commenter_id"`
	UserID      uuid.UUID  `json:"user_id"`
	CreatedDate *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateUserApprovalComment struct {
	Comment *string `json:"comment,omitempty"`
}
