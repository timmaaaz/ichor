package inspectionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/quality/inspectionapp"
)

func parseQueryParams(r *http.Request) (inspectionapp.QueryParams, error) {
	values := r.URL.Query()

	filter := inspectionapp.QueryParams{
		Page:               values.Get("page"),
		Rows:               values.Get("rows"),
		OrderBy:            values.Get("orderBy"),
		InspectionID:       values.Get("inspection_id"),
		ProductID:          values.Get("product_id"),
		InspectorID:        values.Get("inspector_id"),
		Status:             values.Get("status"),
		LotID:              values.Get("lot_id"),
		Notes:              values.Get("notes"),
		InspectionDate:     values.Get("inspection_date"),
		NextInspectionDate: values.Get("next_inspection_date"),
		UpdatedDate:        values.Get("updated_date"),
		CreatedDate:        values.Get("created_date"),
	}

	return filter, nil

}
