package lottrackingsbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	LotID             *uuid.UUID
	SupplierProductID *uuid.UUID
	LotNumber         *string
	ManufactureDate   *time.Time
	ExpirationDate    *time.Time
	RecievedDate      *time.Time
	Quantity          *int
	QualityStatus     *string
	CreatedDate       *time.Time
	UpdatedDate       *time.Time
}
