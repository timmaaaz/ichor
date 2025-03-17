package zonebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByID          = "zone_id"
	OrderByWarehouseID = "warehouse_id"
	OrderByName        = "name"
	OrderByDescription = "description"
	OrderByIsActive    = "is_active"
	OrderByDateCreated = "date_created"
	OrderByDateUpdated = "date_updated"
	OrderByCreatedBy   = "created_by"
	OrderByUpdatedBy   = "updated_by"
)
