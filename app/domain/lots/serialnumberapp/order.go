package serialnumberapp

import (
	"github.com/timmaaaz/ichor/business/domain/lot/serialnumberbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":            serialnumberbus.OrderBySerialID,
	"lot_id":        serialnumberbus.OrderByLotID,
	"product_id":    serialnumberbus.OrderByProductID,
	"location_id":   serialnumberbus.OrderByLocationID,
	"serial_number": serialnumberbus.OrderBySerialNumber,
	"status":        serialnumberbus.OrderByStatus,
	"created_date":  serialnumberbus.OrderByCreatedDate,
	"updated_date":  serialnumberbus.OrderByUpdatedDate,
}
