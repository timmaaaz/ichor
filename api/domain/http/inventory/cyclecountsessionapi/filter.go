package cyclecountsessionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
)

func parseQueryParams(r *http.Request) (cyclecountsessionapp.QueryParams, error) {
	values := r.URL.Query()

	qp := cyclecountsessionapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Status:      values.Get("status"),
		CreatedBy:   values.Get("created_by"),
		CreatedDate: values.Get("created_date"),
	}

	return qp, nil
}
