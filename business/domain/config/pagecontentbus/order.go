package pagecontentbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy specifies the default ordering for page content queries
var DefaultOrderBy = order.NewBy(OrderByOrderIndex, order.ASC)

// Ordering field constants for page content
const (
	OrderByID         = "id"
	OrderByOrderIndex = "order_index"
	OrderByLabel      = "label"
	OrderByContentType = "content_type"
)
