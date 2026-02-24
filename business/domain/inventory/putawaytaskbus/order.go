package putawaytaskbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for put-away task queries.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID              = "id"
	OrderByProductID       = "product_id"
	OrderByLocationID      = "location_id"
	OrderByQuantity        = "quantity"
	OrderByReferenceNumber = "reference_number"
	OrderByStatus          = "status"
	OrderByAssignedTo      = "assigned_to"
	OrderByCreatedBy       = "created_by"
	OrderByCreatedDate     = "created_date"
	OrderByUpdatedDate     = "updated_date"
)
