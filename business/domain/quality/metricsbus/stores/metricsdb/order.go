package metricsdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	metricsbus.OrderByMetricID:          "id",
	metricsbus.OrderByProductID:         "product_id",
	metricsbus.OrderByReturnRate:        "return_rate",
	metricsbus.OrderByDefectRate:        "defect_rate",
	metricsbus.OrderByMeasurementPeriod: "measurement_period",
	metricsbus.OrderByCreatedDate:       "created_date",
	metricsbus.OrderByUpdatedDate:       "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
