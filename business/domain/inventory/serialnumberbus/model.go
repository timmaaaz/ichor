package serialnumberbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type SerialNumber struct {
	SerialID     uuid.UUID `json:"serial_id"`
	LotID        uuid.UUID `json:"lot_id"`
	ProductID    uuid.UUID `json:"product_id"`
	LocationID   uuid.UUID `json:"location_id"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
	CreatedDate  time.Time `json:"created_date"`
	UpdatedDate  time.Time `json:"updated_date"`
}

type NewSerialNumber struct {
	LotID        uuid.UUID `json:"lot_id"`
	ProductID    uuid.UUID `json:"product_id"`
	LocationID   uuid.UUID `json:"location_id"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
}

type UpdateSerialNumber struct {
	LotID        *uuid.UUID `json:"lot_id,omitempty"`
	ProductID    *uuid.UUID `json:"product_id,omitempty"`
	LocationID   *uuid.UUID `json:"location_id,omitempty"`
	SerialNumber *string    `json:"serial_number,omitempty"`
	Status       *string    `json:"status,omitempty"`
}

// SerialLocation represents the storage location of a serial number.
type SerialLocation struct {
	LocationID    uuid.UUID `json:"location_id"`
	LocationCode  string    `json:"location_code"`
	Aisle         string    `json:"aisle"`
	Rack          string    `json:"rack"`
	Shelf         string    `json:"shelf"`
	Bin           string    `json:"bin"`
	WarehouseName string    `json:"warehouse_name"`
	ZoneName      string    `json:"zone_name"`
}
