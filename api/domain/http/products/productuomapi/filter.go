package productuomapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/products/productuomapp"
)

func parseQueryParams(r *http.Request) (productuomapp.QueryParams, error) {
	values := r.URL.Query()

	filter := productuomapp.QueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("id"),
		ProductID: values.Get("product_id"),
		IsBase:    values.Get("is_base"),
		Name:      values.Get("name"),
	}

	return filter, nil
}
