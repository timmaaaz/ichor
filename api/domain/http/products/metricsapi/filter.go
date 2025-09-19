package metricsapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/products/metricsapp"
)

func parseQueryParams(r *http.Request) (metricsapp.QueryParams, error) {
	values := r.URL.Query()

	filter := metricsapp.QueryParams{
		Page:              values.Get("page"),
		Rows:              values.Get("rows"),
		OrderBy:           values.Get("orderBy"),
		MetricID:          values.Get("metric_id"),
		ProductID:         values.Get("product_id"),
		ReturnRate:        values.Get("return_rate"),
		DefectRate:        values.Get("defect_rate"),
		MeasurementPeriod: values.Get("measurement_period"),
		CreatedDate:       values.Get("created_date"),
		UpdatedDate:       values.Get("updated_date"),
	}

	return filter, nil
}
