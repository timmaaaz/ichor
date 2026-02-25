package lotlocationbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID          = "id"
	OrderByLotID       = "lot_id"
	OrderByLocationID  = "location_id"
	OrderByQuantity    = "quantity"
	OrderByCreatedDate = "created_date"
	OrderByUpdatedDate = "updated_date"
)
