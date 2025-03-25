package supplierproductbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByProductID, order.ASC)

const (
	OrderBySupplierProductID  = "supplier_product_id"
	OrderBySupplierID         = "supplier_id"
	OrderByProductID          = "product_id"
	OrderBySupplierPartNumber = "supplier_part_number"
	OrderByMinOrderQuantity   = "min_order_quantity"
	OrderByMaxOrderQuantity   = "max_order_quantity"
	OrderByLeadTimeDays       = "lead_time_days"
	OrderByUnitCost           = "unit_cost"
	OrderByIsPrimarySupplier  = "is_primary_supplier"
	OrderByCreatedDate        = "created_date"
	OrderByUpdatedDate        = "updated_date"
)
