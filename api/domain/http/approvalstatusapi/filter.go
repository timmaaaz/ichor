package approvalstatusapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/approvalstatusapp"
)

func parseQueryParams(r *http.Request) (approvalstatusapp.QueryParams, error) {
	values := r.URL.Query()

	filter := approvalstatusapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("approval_status_id"),
		Name:    values.Get("name"),
		IconID:  values.Get("icon_id"),
	}

	return filter, nil
}
