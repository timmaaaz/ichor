package orderlineitemsbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID                            = "id"
	OrderByOrderID                       = "order_id"
	OrderByProductID                     = "product_id"
	OrderByDescription                   = "description"
	OrderByQuantity                      = "quantity"
	OrderByUnitPrice                     = "unit_price"
	OrderByDiscount                      = "discount"
	OrderByDiscountType                  = "discount_type"
	OrderByLineTotal                     = "line_total"
	OrderByLineItemFulfillmentStatusesID = "line_item_fulfillment_statuses_id"
	OrderByCreatedBy                     = "created_by"
	OrderByCreatedDate                   = "created_date"
	OrderByUpdatedBy                     = "updated_by"
	OrderByUpdatedDate                   = "updated_date"
)
