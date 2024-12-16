package userassetapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/userassetapp"
)

func parseQueryParams(r *http.Request) (userassetapp.QueryParams, error) {
	values := r.URL.Query()

	filter := userassetapp.QueryParams{
		Page:                values.Get("page"),
		Rows:                values.Get("rows"),
		OrderBy:             values.Get("orderBy"),
		ID:                  values.Get("user_asset_id"),
		UserID:              values.Get("user_id"),
		AssetID:             values.Get("asset_id"),
		ApprovedBy:          values.Get("approved_by"),
		ConditionID:         values.Get("condition_id"),
		ApprovalStatusID:    values.Get("approved_status_id"),
		FulfillmentStatusID: values.Get("fulfillment_status_id"),
		DateReceived:        values.Get("date_received"),
		LastMaintenance:     values.Get("last_maintenance"),
	}

	return filter, nil
}
