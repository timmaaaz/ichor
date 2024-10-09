package assetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assetapp"
)

func parseQueryParams(r *http.Request) (assetapp.QueryParams, error) {
	values := r.URL.Query()

	filter := assetapp.QueryParams{
		Page:                values.Get("page"),
		Rows:                values.Get("row"),
		OrderBy:             values.Get("orderBy"),
		ID:                  values.Get("asset_id"),
		TypeID:              values.Get("type_id"),
		ConditionID:         values.Get("condition_id"),
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
