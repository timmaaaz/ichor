package inventorylocationapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":                  inventorylocationbus.OrderByLocationID,
	"warehouse_id":        inventorylocationbus.OrderByWarehouseID,
	"zone_id":             inventorylocationbus.OrderByZoneID,
	"aisle":               inventorylocationbus.OrderByAisle,
	"rack":                inventorylocationbus.OrderByRack,
	"shelf":               inventorylocationbus.OrderByShelf,
	"bin":                 inventorylocationbus.OrderByBin,
	"is_pick_location":    inventorylocationbus.OrderByIsPickLocation,
	"is_reserve_location": inventorylocationbus.OrderByIsReserveLocation,
	"max_capacity":        inventorylocationbus.OrderByMaxCapacity,
	"current_utilization": inventorylocationbus.OrderByCurrentUtilization,
	"created_date":        inventorylocationbus.OrderByCreatedDate,
	"updated_date":        inventorylocationbus.OrderByUpdatedDate,
}
