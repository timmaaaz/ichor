package approvalapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/users/status/approvalapp"
)

func parseQueryParams(r *http.Request) (approvalapp.QueryParams, error) {
	values := r.URL.Query()

	filter := approvalapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("approval_status_id"),
		Name:    values.Get("name"),
		IconID:  values.Get("icon_id"),
	}

	return filter, nil
}
