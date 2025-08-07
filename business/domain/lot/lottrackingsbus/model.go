package lottrackingsbus

import (
	"time"

	"github.com/google/uuid"
)

type LotTrackings struct {
	LotID             uuid.UUID
	SupplierProductID uuid.UUID
	LotNumber         string
	ManufactureDate   time.Time
	ExpirationDate    time.Time
	RecievedDate      time.Time
	Quantity          int
	QualityStatus     string
	CreatedDate       time.Time
	UpdatedDate       time.Time
}

type NewLotTrackings struct {
	SupplierProductID uuid.UUID
	LotNumber         string
	ManufactureDate   time.Time
	ExpirationDate    time.Time
	RecievedDate      time.Time
	Quantity          int
	QualityStatus     string
}

type UpdateLotTrackings struct {
	SupplierProductID *uuid.UUID
	LotNumber         *string
	ManufactureDate   *time.Time
	ExpirationDate    *time.Time
	RecievedDate      *time.Time
	Quantity          *int
	QualityStatus     *string
}
