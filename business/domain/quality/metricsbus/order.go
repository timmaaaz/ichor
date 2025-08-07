package metricsbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByMetricID, order.ASC)

const (
	OrderByMetricID          = "id"
	OrderByProductID         = "product_id"
	OrderByReturnRate        = "return_rate"
	OrderByDefectRate        = "defect_rate"
	OrderByMeasurementPeriod = "measurement_period"
	OrderByCreatedDate       = "created_date"
	OrderByUpdatedDate       = "updated_date"
)
