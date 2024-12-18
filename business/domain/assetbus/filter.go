package assetbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID               *uuid.UUID
	ValidAssetID     *uuid.UUID
	AssetConditionID *uuid.UUID
	LastMaintenance  *time.Time
	SerialNumber     *string
}
