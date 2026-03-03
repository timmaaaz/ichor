package settingsapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/config/settingsapp"
)

func parseQueryParams(r *http.Request) (settingsapp.QueryParams, error) {
	values := r.URL.Query()

	filter := settingsapp.QueryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("order_by"),
		Key:     values.Get("key"),
		Prefix:  values.Get("prefix"),
	}

	return filter, nil
}
