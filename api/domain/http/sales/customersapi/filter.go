package customersapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/sales/customersapp"
)

func parseQueryParams(r *http.Request) (customersapp.QueryParams, error) {
	values := r.URL.Query()

	filter := customersapp.QueryParams{
		Page:              values.Get("page"),
		Rows:              values.Get("rows"),
		OrderBy:           values.Get("orderBy"),
		ID:                values.Get("id"),
		Name:              values.Get("name"),
		DeliveryAddressID: values.Get("delivery_address_id"),
		ContactID:         values.Get("contact_id"),
		Notes:             values.Get("notes"),
		CreatedBy:         values.Get("created_by"),
		UpdatedBy:         values.Get("updated_by"),
		StartCreatedDate:  values.Get("start_created_date"),
		EndCreatedDate:    values.Get("end_created_date"),
		StartUpdatedDate:  values.Get("start_updated_date"),
		EndUpdatedDate:    values.Get("end_updated_date"),
	}

	return filter, nil

}
