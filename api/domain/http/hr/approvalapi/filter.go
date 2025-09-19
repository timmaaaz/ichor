package approvalapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/hr/approvalapp"
)

func parseQueryParams(r *http.Request) (approvalapp.QueryParams, error) {
	values := r.URL.Query()

	filter := approvalapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		Name:    values.Get("name"),
		IconID:  values.Get("icon_id"),
	}

	return filter, nil
}
