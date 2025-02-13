package userassetbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                  *uuid.UUID
	UserID              *uuid.UUID
	AssetID             *uuid.UUID
	ApprovedBy          *uuid.UUID
	ConditionID         *uuid.UUID
	ApprovalStatusID    *uuid.UUID
	FulfillmentStatusID *uuid.UUID

	DateReceived    *time.Time
	LastMaintenance *time.Time
}
