package purchaseorderlineitembus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID                   = "id"
	OrderByPurchaseOrderID      = "purchase_order_id"
	OrderBySupplierProductID    = "supplier_product_id"
	OrderByQuantityOrdered      = "quantity_ordered"
	OrderByQuantityReceived     = "quantity_received"
	OrderByLineTotal            = "line_total"
	OrderByExpectedDeliveryDate = "expected_delivery_date"
	OrderByActualDeliveryDate   = "actual_delivery_date"
	OrderByCreatedDate          = "created_date"
	OrderByUpdatedDate          = "updated_date"
)