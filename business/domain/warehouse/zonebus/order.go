package zonebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByZoneID, order.ASC)

const (
	OrderByZoneID      = "id"
	OrderByWarehouseID = "warehouse_id"
	OrderByName        = "name"
	OrderByDescription = "description"
	OrderByCreatedDate = "created_date"
	OrderByUpdatedDate = "updated_date"
)
