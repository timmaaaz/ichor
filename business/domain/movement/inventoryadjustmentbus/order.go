package inventoryadjustmentbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByInventoryAdjustmentID, order.ASC)

const (
	OrderByInventoryAdjustmentID = "adjustment_id"
	OrderByProductID             = "product_id"
	OrderByLocationID            = "location_id"
	OrderByAdjustedBy            = "adjusted_by"
	OrderByApprovedBy            = "approved_by"
	OrderByQuantityChange        = "quantity_change"
	OrderByReasonCode            = "reason_code"
	OrderByNotes                 = "notes"
	OrderByAdjustmentDate        = "adjustment_date"
	OrderByCreatedDate           = "created_date"
	OrderByUpdatedDate           = "updated_date"
)
