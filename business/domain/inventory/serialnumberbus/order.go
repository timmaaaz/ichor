package serialnumberbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderBySerialID, order.ASC)

const (
	OrderBySerialID     = "id"
	OrderByLotID        = "lot_id"
	OrderByProductID    = "product_id"
	OrderByLocationID   = "location_id"
	OrderBySerialNumber = "serial_number"
	OrderByStatus       = "status"
	OrderByCreatedDate  = "created_date"
	OrderByUpdatedDate  = "updated_date"
)
