package metricsdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
)

func applyFilter(filter metricsbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.MetricID != nil {
		data["quality_metric_id"] = *filter.MetricID
		wc = append(wc, "quality_metric_id = :quality_metric_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}
	if filter.ReturnRate != nil {
		data["return_rate"] = *filter.ReturnRate
		wc = append(wc, "return_rate = :return_rate")
	}

	if filter.DefectRate != nil {
		data["defect_rate"] = *filter.DefectRate
		wc = append(wc, "defect_rate = :defect_rate")
	}

	if filter.MeasurementPeriod != nil {
		data["measurement_period"] = *filter.MeasurementPeriod
		wc = append(wc, "measurement_period = :measurement_period")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}

}
