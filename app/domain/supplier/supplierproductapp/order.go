package supplierproductapp

import (
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("product_id", order.ASC)

var orderByFields = map[string]string{
	"supplier_product_id":  supplierproductbus.OrderBySupplierProductID,
	"supplier_id":          supplierproductbus.OrderBySupplierID,
	"product_id":           supplierproductbus.OrderByProductID,
	"supplier_part_number": supplierproductbus.OrderBySupplierPartNumber,
	"min_order_quantity":   supplierproductbus.OrderByMinOrderQuantity,
	"max_order_quantity":   supplierproductbus.OrderByMaxOrderQuantity,
	"lead_time_days":       supplierproductbus.OrderByLeadTimeDays,
	"unit_cost":            supplierproductbus.OrderByUnitCost,
	"is_primary_supplier":  supplierproductbus.OrderByIsPrimarySupplier,
	"created_date":         supplierproductbus.OrderByCreatedDate,
	"updated_date":         supplierproductbus.OrderByUpdatedDate,
}
