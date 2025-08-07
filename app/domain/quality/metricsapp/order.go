package metricsapp

import (
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":                 metricsbus.OrderByMetricID,
	"product_id":         metricsbus.OrderByProductID,
	"return_rate":        metricsbus.OrderByReturnRate,
	"defect_rate":        metricsbus.OrderByDefectRate,
	"measurement_period": metricsbus.OrderByMeasurementPeriod,
	"created_date":       metricsbus.OrderByCreatedDate,
	"updated_date":       metricsbus.OrderByUpdatedDate,
}
