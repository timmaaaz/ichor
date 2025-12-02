package commentbus

import (
	"time"

	"github.com/google/uuid"
)

type UserApprovalComment struct {
	ID          uuid.UUID
	Comment     string
	CommenterID uuid.UUID
	UserID      uuid.UUID
	CreatedDate time.Time
}

type NewUserApprovalComment struct {
	Comment     string
	CommenterID uuid.UUID
	UserID      uuid.UUID
	CreatedDate *time.Time // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateUserApprovalComment struct {
	Comment *string
}
