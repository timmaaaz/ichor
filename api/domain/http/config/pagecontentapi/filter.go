package pagecontentapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/config/pagecontentapp"
)

func parseQueryParams(r *http.Request) (pagecontentapp.QueryParams, error) {
	values := r.URL.Query()

	qp := pagecontentapp.QueryParams{
		Page:         values.Get("page"),
		Rows:         values.Get("rows"),
		OrderBy:      values.Get("orderBy"),
		ID:           values.Get("id"),
		PageConfigID: values.Get("pageConfigId"),
		ContentType:  values.Get("contentType"),
		ParentID:     values.Get("parentId"),
		IsVisible:    values.Get("isVisible"),
	}

	return qp, nil
}
