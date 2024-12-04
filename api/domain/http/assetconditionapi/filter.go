package assetconditionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/assetconditionapp"
)

func parseQueryParams(r *http.Request) (assetconditionapp.QueryParams, error) {
	values := r.URL.Query()

	filter := assetconditionapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("row"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("asset_condition_id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
	}

	return filter, nil
}
