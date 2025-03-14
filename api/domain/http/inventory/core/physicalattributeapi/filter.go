package physicalattributeapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/core/physicalattributeapp"
)

func parseQueryParams(r *http.Request) (physicalattributeapp.QueryParams, error) {
	values := r.URL.Query()

	filter := physicalattributeapp.QueryParams{
		Page:                values.Get("page"),
		Rows:                values.Get("rows"),
		OrderBy:             values.Get("orderBy"),
		ID:                  values.Get("attribute_id"),
		ProductID:           values.Get("product_id"),
		Length:              values.Get("length"),
		Width:               values.Get("width"),
		Height:              values.Get("height"),
		Weight:              values.Get("weight"),
		Color:               values.Get("color"),
		Material:            values.Get("material"),
		Size:                values.Get("size"),
		WeightUnit:          values.Get("weight_unit"),
		StorageRequirements: values.Get("storage_requirements"),
		HazmatClass:         values.Get("hazmat_class"),
		ShelfLifeDays:       values.Get("shelf_life_days"),
		CreatedDate:         values.Get("created_date"),
		UpdatedDate:         values.Get("updated_date"),
	}

	return filter, nil
}
