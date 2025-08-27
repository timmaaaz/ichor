package validassetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
)

func parseQueryParams(r *http.Request) (validassetapp.QueryParams, error) {
	values := r.URL.Query()

	filter := validassetapp.QueryParams{
		Page:                values.Get("page"),
		Rows:                values.Get("rows"),
		OrderBy:             values.Get("orderBy"),
		ID:                  values.Get("id"),
		TypeID:              values.Get("type_id"),
		Name:                values.Get("name"),
		EstPrice:            values.Get("est_price"),
		Price:               values.Get("price"),
		MaintenanceInterval: values.Get("maintenance_interval"),
		LifeExpectancy:      values.Get("life_expectancy"),
		ModelNumber:         values.Get("model_number"),
		IsEnabled:           values.Get("is_enabled"),
		StartCreatedDate:    values.Get("start_created_date"),
		EndCreatedDate:      values.Get("end_created_date"),
		StartUpdatedDate:    values.Get("start_updated_date"),
		EndUpdatedDate:      values.Get("end_updated_date"),
		CreatedBy:           values.Get("created_by"),
		UpdatedBy:           values.Get("updated_by"),
	}

	return filter, nil
}
