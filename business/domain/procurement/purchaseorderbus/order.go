package purchaseorderbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByOrderDate, order.DESC)

// Set of fields that the results can be ordered by.
const (
	OrderByID                   = "id"
	OrderByOrderNumber          = "order_number"
	OrderBySupplierID           = "supplier_id"
	OrderByOrderDate            = "order_date"
	OrderByExpectedDeliveryDate = "expected_delivery_date"
	OrderByActualDeliveryDate   = "actual_delivery_date"
	OrderByTotalAmount          = "total_amount"
	OrderByRequestedBy          = "requested_by"
	OrderByApprovedBy           = "approved_by"
	OrderByCreatedDate          = "created_date"
	OrderByUpdatedDate          = "updated_date"
)