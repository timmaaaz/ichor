package userroleapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/permissions/userroleapp.go"
)

func parseQueryParams(r *http.Request) (userroleapp.QueryParams, error) {
	values := r.URL.Query()

	filter := userroleapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("id"),
		UserID:  values.Get("user_id"),
		RoleID:  values.Get("role_id"),
	}

	return filter, nil
}
