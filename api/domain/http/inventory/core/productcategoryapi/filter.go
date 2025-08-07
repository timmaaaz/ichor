package productcategoryapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/core/productcategoryapp"
)

func parseQueryParams(r *http.Request) (productcategoryapp.QueryParams, error) {
	values := r.URL.Query()

	filter := productcategoryapp.QueryParams{
		Page:        values.Get("page"),
		Rows:        values.Get("rows"),
		OrderBy:     values.Get("orderBy"),
		ID:          values.Get("id"),
		Name:        values.Get("name"),
		Description: values.Get("description"),
		CreatedDate: values.Get("created_date"),
		UpdatedDate: values.Get("updated_date"),
	}

	return filter, nil
}
