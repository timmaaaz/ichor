package tableaccessapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/permissions/tableaccessapp"
)

func parseQueryParams(r *http.Request) (tableaccessapp.QueryParams, error) {
	values := r.URL.Query()

	filter := tableaccessapp.QueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("id"),
		TableName: values.Get("table_name"),
		CanCreate: values.Get("can_create"),
		CanRead:   values.Get("can_read"),
		CanUpdate: values.Get("can_update"),
		CanDelete: values.Get("can_delete"),
	}

	return filter, nil
}
