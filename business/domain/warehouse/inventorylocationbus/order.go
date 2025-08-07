package inventorylocationbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByLocationID, order.ASC)

const (
	OrderByLocationID         = "id"
	OrderByWarehouseID        = "zone_id"
	OrderByZoneID             = "warehouse_id"
	OrderByAisle              = "aisle"
	OrderByRack               = "rack"
	OrderByShelf              = "shelf"
	OrderByBin                = "bin"
	OrderByIsPickLocation     = "is_pick_location"
	OrderByIsReserveLocation  = "is_reserve_location"
	OrderByMaxCapacity        = "max_capacity"
	OrderByCurrentUtilization = "current_utilization"
	OrderByCreatedDate        = "created_date"
	OrderByUpdatedDate        = "updated_date"
)
