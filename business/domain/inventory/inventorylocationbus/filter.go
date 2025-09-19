package inventorylocationbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
)

type QueryFilter struct {
	LocationID         *uuid.UUID
	WarehouseID        *uuid.UUID
	ZoneID             *uuid.UUID
	Aisle              *string
	Rack               *string
	Shelf              *string
	Bin                *string
	IsPickLocation     *bool
	IsReserveLocation  *bool
	MaxCapacity        *int
	CurrentUtilization *types.RoundedFloat
	CreatedDate        *time.Time
	UpdatedDate        *time.Time
}
