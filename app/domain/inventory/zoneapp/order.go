package zoneapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

// app level order takes a json binding and maps order to the business level binding
var orderByFields = map[string]string{
	"id":           zonebus.OrderByZoneID,
	"warehouse_id": zonebus.OrderByWarehouseID,
	"name":         zonebus.OrderByName,
	"description":  zonebus.OrderByDescription,
	"created_date": zonebus.OrderByCreatedDate,
	"updated_date": zonebus.OrderByUpdatedDate,
}
