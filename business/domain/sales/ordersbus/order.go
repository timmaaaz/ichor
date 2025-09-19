package ordersbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByNumber, order.ASC)

const (
	OrderByID                         = "id"
	OrderByNumber                     = "number"
	OrderByCustomerID                 = "customer_id"
	OrderByDueDate                    = "due_date"
	OrderByOrderByFulfillmentStatusID = "fulfillment_status_id"
	OrderByCreatedBy                  = "created_by"
	OrderByUpdatedBy                  = "updated_by"
	OrderByCreatedDate                = "created_date"
	OrderByUpdatedDate                = "updated_date"
)
