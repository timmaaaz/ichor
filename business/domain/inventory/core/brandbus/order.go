package brandbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

const (
	OrderByID             = "brand_id"
	OrderByName           = "name"
	OrderByManufacturerID = "manufacturer_id"
	OrderByCreatedDate    = "created_date"
	OrderByUpdatedDate    = "updated_date"
)
