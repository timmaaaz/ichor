package warehouseapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/warehouse/warehouseapp"
)

func parseQueryParams(r *http.Request) (warehouseapp.QueryParams, error) {
	values := r.URL.Query()

	filter := warehouseapp.QueryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("rows"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("warehouse_id"),
		StreetID:         values.Get("street_id"),
		Name:             values.Get("name"),
		IsActive:         values.Get("is_active"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
		StartUpdatedDate: values.Get("start_updated_date"),
		EndUpdatedDate:   values.Get("end_updated_date"),
		CreatedBy:        values.Get("created_by"),
		UpdatedBy:        values.Get("updated_by"),
	}

	return filter, nil
}
