package formfieldapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
)

func parseQueryParams(r *http.Request) (formfieldapp.QueryParams, error) {
	values := r.URL.Query()

	qp := formfieldapp.QueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ID:        values.Get("id"),
		FormID:    values.Get("form_id"),
		Name:      values.Get("name"),
		FieldType: values.Get("field_type"),
		Required:  values.Get("required"),
	}

	return qp, nil
}