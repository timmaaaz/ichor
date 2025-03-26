package lottrackingbus

import (
	"time"

	"github.com/google/uuid"
)

type LotTracking struct {
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

type NewLotTracking struct {
	SupplierProductID uuid.UUID
	LotNumber         string
	ManufactureDate   time.Time
	ExpirationDate    time.Time
	RecievedDate      time.Time
	Quantity          int
	QualityStatus     string
}

type UpdateLotTracking struct {
	SupplierProductID *uuid.UUID
	LotNumber         *string
	ManufactureDate   *time.Time
	ExpirationDate    *time.Time
	RecievedDate      *time.Time
	Quantity          *int
	QualityStatus     *string
}
