package pageactionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
)

func parseQueryParams(r *http.Request) (pageactionapp.QueryParams, error) {
	values := r.URL.Query()

	qp := pageactionapp.QueryParams{
		Page:         values.Get("page"),
		Rows:         values.Get("rows"),
		OrderBy:      values.Get("orderBy"),
		ID:           values.Get("id"),
		PageConfigID: values.Get("pageConfigId"),
		ActionType:   values.Get("actionType"),
		IsActive:     values.Get("isActive"),
	}

	return qp, nil
}
