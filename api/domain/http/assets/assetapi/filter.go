package assetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assets/assetapp"
)

func parseQueryParams(r *http.Request) (assetapp.QueryParams, error) {
	values := r.URL.Query()

	filter := assetapp.QueryParams{
		Page:            values.Get("page"),
		Rows:            values.Get("rows"),
		OrderBy:         values.Get("orderBy"),
		ID:              values.Get("id"),
		ValidAssetID:    values.Get("valid_asset_id"),
		ConditionID:     values.Get("asset_condition_id"),
		SerialNumber:    values.Get("serial_number"),
		LastMaintenance: values.Get("last_maintenance"),
	}

	return filter, nil
}
