package orderlineitemsbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID                            = "id"
	OrderByOrderID                       = "order_id"
	OrderByProductID                     = "product_id"
	OrderByQuantity                      = "quantity"
	OrderByDiscount                      = "discount"
	OrderByLineItemFulfillmentStatusesID = "line_item_fulfillment_statuses_id"
	OrderByCreatedBy                     = "created_by"
	OrderByCreatedDate                   = "created_date"
	OrderByUpdatedBy                     = "updated_by"
	OrderByUpdatedDate                   = "updated_date"
)
