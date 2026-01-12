package orderlineitemsapp

import (
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":                                orderlineitemsbus.OrderByID,
	"order_id":                          orderlineitemsbus.OrderByOrderID,
	"product_id":                        orderlineitemsbus.OrderByProductID,
	"description":                       orderlineitemsbus.OrderByDescription,
	"quantity":                          orderlineitemsbus.OrderByQuantity,
	"unit_price":                        orderlineitemsbus.OrderByUnitPrice,
	"discount":                          orderlineitemsbus.OrderByDiscount,
	"discount_type":                     orderlineitemsbus.OrderByDiscountType,
	"line_total":                        orderlineitemsbus.OrderByLineTotal,
	"line_item_fulfillment_statuses_id": orderlineitemsbus.OrderByLineItemFulfillmentStatusesID,
	"created_by":                        orderlineitemsbus.OrderByCreatedBy,
	"created_date":                      orderlineitemsbus.OrderByCreatedDate,
	"updated_by":                        orderlineitemsbus.OrderByUpdatedBy,
	"updated_date":                      orderlineitemsbus.OrderByUpdatedDate,
}
