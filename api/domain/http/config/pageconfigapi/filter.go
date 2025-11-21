package pageconfigapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
)

func parseQueryParams(r *http.Request) (pageconfigapp.QueryParams, error) {
	values := r.URL.Query()

	qp := pageconfigapp.QueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("id"),
		Name:      values.Get("name"),
		UserID:    values.Get("userId"),
		IsDefault: values.Get("isDefault"),
	}

	return qp, nil
}
