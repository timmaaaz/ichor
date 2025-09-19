package costhistoryapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/products/costhistoryapp"
)

func parseQueryParams(r *http.Request) (costhistoryapp.QueryParams, error) {
	values := r.URL.Query()

	filter := costhistoryapp.QueryParams{
		Page:          values.Get("page"),
		Rows:          values.Get("rows"),
		OrderBy:       values.Get("orderBy"),
		CostHistoryID: values.Get("cost_history_id"),
		ProductID:     values.Get("product_id"),
		CostType:      values.Get("cost_type"),
		Amount:        values.Get("amount"),
		Currency:      values.Get("currency"),
		CreatedDate:   values.Get("created_date"),
		UpdatedDate:   values.Get("updated_date"),
		EffectiveDate: values.Get("effective_date"),
		EndDate:       values.Get("end_date"),
	}

	return filter, nil
}
