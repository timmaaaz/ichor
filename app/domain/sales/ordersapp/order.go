package ordersapp

import (
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("number", order.ASC)

var orderByFields = map[string]string{
	"id":                    ordersbus.OrderByID,
	"number":                ordersbus.OrderByNumber,
	"customer_id":           ordersbus.OrderByCustomerID,
	"due_date":              ordersbus.OrderByDueDate,
	"fulfillment_status_id": ordersbus.OrderByOrderByFulfillmentStatusID,
	"order_date":            ordersbus.OrderByOrderDate,
	"subtotal":              ordersbus.OrderBySubtotal,
	"total_amount":          ordersbus.OrderByTotalAmount,
	"currency":              ordersbus.OrderByCurrency,
	"created_by":            ordersbus.OrderByCreatedBy,
	"updated_by":            ordersbus.OrderByUpdatedBy,
	"created_date":          ordersbus.OrderByCreatedDate,
	"updated_date":          ordersbus.OrderByUpdatedDate,
}
