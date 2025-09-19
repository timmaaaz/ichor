package commentapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/hr/commentapp"
)

func parseQueryParams(r *http.Request) (commentapp.QueryParams, error) {
	values := r.URL.Query()

	filter := commentapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		UserID:      values.Get("user_id"),
		CommenterID: values.Get("commenter_id"),
		Comment:     values.Get("comment"),
		CreatedDate: values.Get("created_date"),
	}

	return filter, nil
}
