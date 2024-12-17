package userassetbus

import (
	"time"

	"github.com/google/uuid"
)

type UserAsset struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	AssetID             uuid.UUID
	ApprovedBy          uuid.UUID
	ApprovalStatusID    uuid.UUID
	FulfillmentStatusID uuid.UUID

	DateReceived    time.Time
	LastMaintenance time.Time
}

type NewUserAsset struct {
	UserID              uuid.UUID
	AssetID             uuid.UUID
	ApprovedBy          uuid.UUID
	ApprovalStatusID    uuid.UUID
	FulfillmentStatusID uuid.UUID

	DateReceived    time.Time
	LastMaintenance time.Time
}

type UpdateUserAsset struct {
	UserID              *uuid.UUID
	AssetID             *uuid.UUID
	ApprovedBy          *uuid.UUID
	ApprovalStatusID    *uuid.UUID
	FulfillmentStatusID *uuid.UUID

	DateReceived    *time.Time
	LastMaintenance *time.Time
}
