package inventorylocationbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
)

type InventoryLocation struct {
	LocationID         uuid.UUID
	WarehouseID        uuid.UUID
	ZoneID             uuid.UUID
	Aisle              string
	Rack               string
	Shelf              string
	Bin                string
	IsPickLocation     bool // DO we need both of these? seems like if its not pick location then its reserve location
	IsReserveLocation  bool
	MaxCapacity        int
	CurrentUtilization types.RoundedFloat
	CreatedDate        time.Time
	UpdatedDate        time.Time
}

type NewInventoryLocation struct {
	WarehouseID        uuid.UUID
	ZoneID             uuid.UUID
	Aisle              string
	Rack               string
	Shelf              string
	Bin                string
	IsPickLocation     bool
	IsReserveLocation  bool
	MaxCapacity        int
	CurrentUtilization types.RoundedFloat
}

type UpdateInventoryLocation struct {
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
}
