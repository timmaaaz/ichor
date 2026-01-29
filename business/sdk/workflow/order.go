package workflow

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy is the default ordering for automation rule queries.
var DefaultOrderBy = order.NewBy(OrderByCreatedDate, order.DESC)

// Order field constants for automation rules.
const (
	OrderByID          = "id"
	OrderByName        = "name"
	OrderByCreatedDate = "created_date"
	OrderByUpdatedDate = "updated_date"
	OrderByIsActive    = "is_active"
)

// Order field constants for rule actions.
const (
	ActionOrderByID             = "id"
	ActionOrderByExecutionOrder = "execution_order"
	ActionOrderByIsActive       = "is_active"
)
