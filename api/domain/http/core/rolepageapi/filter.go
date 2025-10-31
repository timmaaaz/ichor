package rolepageapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/core/rolepageapp"
)

func parseQueryParams(r *http.Request) (rolepageapp.QueryParams, error) {
	values := r.URL.Query()

	filter := rolepageapp.QueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("id"),
		RoleID:    values.Get("roleId"),
		PageID:    values.Get("pageId"),
		CanAccess: values.Get("canAccess"),
	}

	return filter, nil
}
