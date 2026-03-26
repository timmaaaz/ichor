package cyclecountsessionbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for cycle count session queries.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID          = "id"
	OrderByName        = "name"
	OrderByStatus      = "status"
	OrderByCreatedBy   = "created_by"
	OrderByCreatedDate = "created_date"
)
