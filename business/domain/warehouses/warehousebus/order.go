package warehousebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID          = "warehouse_id"
	OrderByStreetID    = "street_id"
	OrderByName        = "name"
	OrderByIsActive    = "is_active"
	OrderByDateCreated = "date_created"
	OrderByDateUpdated = "date_updated"
	OrderByCreatedBy   = "created_by"
	OrderByUpdatedBy   = "updated_by"
)
