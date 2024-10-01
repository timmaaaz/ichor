package streetbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByLine1, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByID         = "street_id"
	OrderByCityID     = "city_id"
	OrderByLine1      = "line_1"
	OrderByLine2      = "line_2"
	OrderByPostalCode = "postal_code"
)
