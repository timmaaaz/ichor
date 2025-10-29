package warehousebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID          = "id"
	OrderByCode        = "code"
	OrderByStreetID    = "street_id"
	OrderByName        = "name"
	OrderByIsActive    = "is_active"
	OrderByCreatedDate = "created_date"
	OrderByUpdatedDate = "updated_date"
	OrderByCreatedBy   = "created_by"
	OrderByUpdatedBy   = "updated_by"
)
