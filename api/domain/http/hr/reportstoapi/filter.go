package reportstoapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/hr/reportstoapp"
)

func parseQueryParams(r *http.Request) (reportstoapp.QueryParams, error) {
	values := r.URL.Query()

	filter := reportstoapp.QueryParams{
		Page:       values.Get("page"),
		Rows:       values.Get("rows"),
		OrderBy:    values.Get("orderBy"),
		ID:         values.Get("id"),
		BossID:     values.Get("boss_id"),
		ReporterID: values.Get("reports_to"),
	}

	return filter, nil
}
