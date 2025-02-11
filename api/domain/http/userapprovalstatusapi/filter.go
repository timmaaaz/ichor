package userapprovalstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/userapprovalstatusapp"
)

func parseQueryParams(r *http.Request) (userapprovalstatusapp.QueryParams, error) {
	values := r.URL.Query()

	filter := userapprovalstatusapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("approval_status_id"),
		Name:    values.Get("name"),
		IconID:  values.Get("icon_id"),
	}

	return filter, nil
}
