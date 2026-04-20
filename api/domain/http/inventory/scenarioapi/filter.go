package scenarioapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/scenarioapp"
)

func parseQueryParams(r *http.Request) (scenarioapp.QueryParams, error) {
	values := r.URL.Query()
	return scenarioapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		Name:    values.Get("name"),
	}, nil
}
