package warehouseapp

import (
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":           warehousebus.OrderByID,
	"street_id":    warehousebus.OrderByStreetID,
	"name":         warehousebus.OrderByName,
	"is_active":    warehousebus.OrderByIsActive,
	"created_date": warehousebus.OrderByCreatedDate,
	"updated_date": warehousebus.OrderByUpdatedDate,
	"created_by":   warehousebus.OrderByCreatedBy,
	"updated_by":   warehousebus.OrderByUpdatedBy,
}
