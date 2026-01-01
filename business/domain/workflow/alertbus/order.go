package alertbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByCreatedDate, order.DESC)

// Set of fields that the results can be ordered by.
const (
	OrderByID          = "id"
	OrderByAlertType   = "alert_type"
	OrderBySeverity    = "severity"
	OrderByStatus      = "status"
	OrderByCreatedDate = "created_date"
	OrderByUpdatedDate = "updated_date"
)
