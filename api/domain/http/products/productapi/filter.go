package productapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/products/productapp"
)

func parseQueryParams(r *http.Request) (productapp.QueryParams, error) {
	values := r.URL.Query()

	filter := productapp.QueryParams{
		Page:                 values.Get("page"),
		Rows:                 values.Get("rows"),
		OrderBy:              values.Get("orderBy"),
		ProductID:            values.Get("product_id"),
		Name:                 values.Get("name"),
		ProductCategoryID:    values.Get("product_category_id"),
		BrandID:              values.Get("brand_id"),
		CreatedDate:          values.Get("created_date"),
		UpdatedDate:          values.Get("updated_date"),
		SKU:                  values.Get("sku"),
		Description:          values.Get("description"),
		ModelNumber:          values.Get("model_number"),
		UpcCode:              values.Get("upc_code"),
		Status:               values.Get("status"),
		IsActive:             values.Get("is_active"),
		IsPerishable:         values.Get("is_perishable"),
		HandlingInstructions: values.Get("handling_instructions"),
		UnitsPerCase:         values.Get("units_per_case"),
		TrackingType:         values.Get("tracking_type"),
	}

	return filter, nil
}
