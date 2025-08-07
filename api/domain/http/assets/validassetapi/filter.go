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
		StartDateCreated:    values.Get("start_date_created"),
		EndDateCreated:      values.Get("end_date_created"),
		StartDateUpdated:    values.Get("start_date_updated"),
		EndDateUpdated:      values.Get("end_date_updated"),
		CreatedBy:           values.Get("created_by"),
		UpdatedBy:           values.Get("updated_by"),
	}

	return filter, nil
}
