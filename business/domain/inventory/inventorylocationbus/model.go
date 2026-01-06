package inventorylocationbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type InventoryLocation struct {
	LocationID         uuid.UUID          `json:"location_id"`
	WarehouseID        uuid.UUID          `json:"warehouse_id"`
	ZoneID             uuid.UUID          `json:"zone_id"`
	Aisle              string             `json:"aisle"`
	Rack               string             `json:"rack"`
	Shelf              string             `json:"shelf"`
	Bin                string             `json:"bin"`
	IsPickLocation     bool               `json:"is_pick_location"` // DO we need both of these? seems like if its not pick location then its reserve location
	IsReserveLocation  bool               `json:"is_reserve_location"`
	MaxCapacity        int                `json:"max_capacity"`
	CurrentUtilization types.RoundedFloat `json:"current_utilization"`
	CreatedDate        time.Time          `json:"created_date"`
	UpdatedDate        time.Time          `json:"updated_date"`
}

type NewInventoryLocation struct {
	WarehouseID        uuid.UUID          `json:"warehouse_id"`
	ZoneID             uuid.UUID          `json:"zone_id"`
	Aisle              string             `json:"aisle"`
	Rack               string             `json:"rack"`
	Shelf              string             `json:"shelf"`
	Bin                string             `json:"bin"`
	IsPickLocation     bool               `json:"is_pick_location"`
	IsReserveLocation  bool               `json:"is_reserve_location"`
	MaxCapacity        int                `json:"max_capacity"`
	CurrentUtilization types.RoundedFloat `json:"current_utilization"`
}

type UpdateInventoryLocation struct {
	WarehouseID        *uuid.UUID          `json:"warehouse_id,omitempty"`
	ZoneID             *uuid.UUID          `json:"zone_id,omitempty"`
	Aisle              *string             `json:"aisle,omitempty"`
	Rack               *string             `json:"rack,omitempty"`
	Shelf              *string             `json:"shelf,omitempty"`
	Bin                *string             `json:"bin,omitempty"`
	IsPickLocation     *bool               `json:"is_pick_location,omitempty"`
	IsReserveLocation  *bool               `json:"is_reserve_location,omitempty"`
	MaxCapacity        *int                `json:"max_capacity,omitempty"`
	CurrentUtilization *types.RoundedFloat `json:"current_utilization,omitempty"`
}
