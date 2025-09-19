package supplierproductdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
)

func applyFilter(filter supplierproductbus.QueryFilter, data map[string]any, buf *bytes.Buffer) error {
	var wc []string

	if filter.SupplierProductID != nil {
		data["id"] = *filter.SupplierProductID
		wc = append(wc, "id = :id")
	}

	if filter.SupplierID != nil {
		data["supplier_id"] = *filter.SupplierID
		wc = append(wc, "supplier_id = :supplier_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.IsPrimarySupplier != nil {
		data["is_primary_supplier"] = *filter.IsPrimarySupplier
		wc = append(wc, "is_primary_supplier = :is_primary_supplier")
	}

	if filter.MaxOrderQuantity != nil {
		data["max_order_quantity"] = *filter.MaxOrderQuantity
		wc = append(wc, "max_order_quantity = :max_order_quantity")
	}

	if filter.MinOrderQuantity != nil {
		data["min_order_quantity"] = *filter.MinOrderQuantity
		wc = append(wc, "min_order_quantity = :min_order_quantity")
	}

	if filter.LeadTimeDays != nil {
		data["lead_time_days"] = *filter.LeadTimeDays
		wc = append(wc, "lead_time_days = :lead_time_days")
	}

	if filter.UnitCost != nil {
		data["unit_cost"] = *filter.UnitCost
		wc = append(wc, "unit_cost = :unit_cost")
	}

	if filter.SupplierPartNumber != nil {
		data["supplier_part_number"] = *filter.SupplierPartNumber
		wc = append(wc, "supplier_part_number = :supplier_part_number")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}

	return nil
}
