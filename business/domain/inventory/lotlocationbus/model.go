package lotlocationbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization.

type LotLocation struct {
	ID          uuid.UUID `json:"id"`
	LotID       uuid.UUID `json:"lot_id"`
	LocationID  uuid.UUID `json:"location_id"`
	Quantity    int       `json:"quantity"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

type NewLotLocation struct {
	LotID      uuid.UUID `json:"lot_id"`
	LocationID uuid.UUID `json:"location_id"`
	Quantity   int       `json:"quantity"`
}

type UpdateLotLocation struct {
	LotID      *uuid.UUID `json:"lot_id,omitempty"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	Quantity   *int       `json:"quantity,omitempty"`
}
