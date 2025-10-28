package pageapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
)

func parseQueryParams(r *http.Request) (pageapp.QueryParams, error) {
	values := r.URL.Query()

	filter := pageapp.QueryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("rows"),
		OrderBy:  values.Get("orderBy"),
		ID:       values.Get("id"),
		Path:     values.Get("path"),
		Name:     values.Get("name"),
		Module:   values.Get("module"),
		IsActive: values.Get("isActive"),
	}

	return filter, nil
}
