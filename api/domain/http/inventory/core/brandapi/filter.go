package brandapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/core/brandapp"
)

func parseQueryParams(r *http.Request) (brandapp.QueryParams, error) {
	values := r.URL.Query()

	filter := brandapp.QueryParams{
		Page:          values.Get("page"),
		Rows:          values.Get("rows"),
		OrderBy:       values.Get("orderBy"),
		ID:            values.Get("brand_id"),
		Name:          values.Get("name"),
		ContactInfoID: values.Get("contact_info_id"),
		CreatedDate:   values.Get("created_date"),
		UpdatedDate:   values.Get("updated_date"),
	}

	return filter, nil
}
