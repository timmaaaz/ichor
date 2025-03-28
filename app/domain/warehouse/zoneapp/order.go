package zoneapp

import (
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("zone_id", order.ASC)

// app level order takes a json binding and maps order to the business level binding
var orderByFields = map[string]string{
	"zone_id":      zonebus.OrderByZoneID,
	"warehouse_id": zonebus.OrderByWarehouseID,
	"name":         zonebus.OrderByName,
	"description":  zonebus.OrderByDescription,
	"created_date": zonebus.OrderByCreatedDate,
	"updated_date": zonebus.OrderByUpdatedDate,
}
