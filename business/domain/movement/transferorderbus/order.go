package transferorderbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByTransferID, order.ASC)

const (
	OrderByTransferID     = "transfer_id"
	OrderByProductID      = "product_id"
	OrderByFromLocationID = "from_location_id"
	OrderByToLocationID   = "to_location_id"
	OrderByRequestedByID  = "requested_by"
	OrderByApprovedByID   = "approved_by"
	OrderByQuantity       = "quantity"
	OrderByStatus         = "status"
	OrderByTransferDate   = "transfer_date"
	OrderByCreatedDate    = "created_date"
	OrderByUpdatedDate    = "updated_date"
)
