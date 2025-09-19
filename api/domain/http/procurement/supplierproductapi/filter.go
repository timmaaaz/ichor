package supplierproductapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/procurement/supplierproductapp"
)

func parseQueryParams(r *http.Request) (supplierproductapp.QueryParams, error) {
	values := r.URL.Query()

	filter := supplierproductapp.QueryParams{
		Page:               values.Get("page"),
		Rows:               values.Get("rows"),
		OrderBy:            values.Get("orderBy"),
		SupplierID:         values.Get("supplier_id"),
		ProductID:          values.Get("product_id"),
		SupplierProductID:  values.Get("supplier_product_id"),
		IsPrimarySupplier:  values.Get("is_primary_supplier"),
		SupplierPartNumber: values.Get("supplier_part_number"),
		MinOrderQuantity:   values.Get("min_order_quantity"),
		MaxOrderQuantity:   values.Get("max_order_quantity"),
		LeadTimeDays:       values.Get("lead_time_days"),
		UnitCost:           values.Get("unit_cost"),
		CreatedDate:        values.Get("created_date"),
		UpdatedDate:        values.Get("updated_date"),
	}

	return filter, nil
}
