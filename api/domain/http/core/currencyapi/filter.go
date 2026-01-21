package currencyapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
)

func parseQueryParams(r *http.Request) (currencyapp.QueryParams, error) {
	values := r.URL.Query()

	filter := currencyapp.QueryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("rows"),
		OrderBy:  values.Get("orderBy"),
		ID:       values.Get("id"),
		Code:     values.Get("code"),
		Name:     values.Get("name"),
		IsActive: values.Get("is_active"),
	}

	return filter, nil
}
