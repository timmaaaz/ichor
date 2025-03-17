package warehouseapp

import (
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("warehouse_id", order.ASC)

var orderByFields = map[string]string{
	"warehouse_id": warehousebus.OrderByID,
	"street_id":    warehousebus.OrderByStreetID,
	"name":         warehousebus.OrderByName,
	"is_active":    warehousebus.OrderByIsActive,
	"date_created": warehousebus.OrderByDateCreated,
	"date_updated": warehousebus.OrderByDateUpdated,
	"created_by":   warehousebus.OrderByCreatedBy,
	"updated_by":   warehousebus.OrderByUpdatedBy,
}
