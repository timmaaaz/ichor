package lottrackingsbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type LotTrackings struct {
	LotID             uuid.UUID `json:"lot_id"`
	SupplierProductID uuid.UUID `json:"supplier_product_id"`
	LotNumber         string    `json:"lot_number"`
	ManufactureDate   time.Time `json:"manufacture_date"`
	ExpirationDate    time.Time `json:"expiration_date"`
	RecievedDate      time.Time `json:"recieved_date"`
	Quantity          int       `json:"quantity"`
	QualityStatus     string    `json:"quality_status"`
	CreatedDate       time.Time `json:"created_date"`
	UpdatedDate       time.Time `json:"updated_date"`
}

type NewLotTrackings struct {
	SupplierProductID uuid.UUID `json:"supplier_product_id"`
	LotNumber         string    `json:"lot_number"`
	ManufactureDate   time.Time `json:"manufacture_date"`
	ExpirationDate    time.Time `json:"expiration_date"`
	RecievedDate      time.Time `json:"recieved_date"`
	Quantity          int       `json:"quantity"`
	QualityStatus     string    `json:"quality_status"`
}

type UpdateLotTrackings struct {
	SupplierProductID *uuid.UUID `json:"supplier_product_id,omitempty"`
	LotNumber         *string    `json:"lot_number,omitempty"`
	ManufactureDate   *time.Time `json:"manufacture_date,omitempty"`
	ExpirationDate    *time.Time `json:"expiration_date,omitempty"`
	RecievedDate      *time.Time `json:"recieved_date,omitempty"`
	Quantity          *int       `json:"quantity,omitempty"`
	QualityStatus     *string    `json:"quality_status,omitempty"`
}

// LotLocation represents a storage location where serial numbers for a lot are stored,
// along with an aggregated count.
type LotLocation struct {
	LocationID   uuid.UUID `json:"location_id"`
	LocationCode string    `json:"location_code"`
	Aisle        string    `json:"aisle"`
	Rack         string    `json:"rack"`
	Shelf        string    `json:"shelf"`
	Bin          string    `json:"bin"`
	Quantity     int       `json:"quantity"`
}
