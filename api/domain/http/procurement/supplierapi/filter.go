package supplierapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/procurement/supplierapp"
)

func parseQueryParams(r *http.Request) (supplierapp.QueryParams, error) {
	values := r.URL.Query()

	filter := supplierapp.QueryParams{
		Page:           values.Get("page"),
		Rows:           values.Get("rows"),
		OrderBy:        values.Get("orderBy"),
		SupplierID:     values.Get("supplier_id"),
		ContactInfosID: values.Get("contact_infos_id"),
		Name:           values.Get("name"),
		PaymentTermID:  values.Get("payment_term_id"),
		LeadTimeDays:   values.Get("lead_time_days"),
		Rating:         values.Get("rating"),
		IsActive:       values.Get("is_active"),
		CreatedDate:    values.Get("created_date"),
		UpdatedDate:    values.Get("updated_date"),
	}

	return filter, nil
}
