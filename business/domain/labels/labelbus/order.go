package labelbus

import "github.com/timmaaaz/ichor/business/sdk/order"

const (
	OrderByID          = "label_id"
	OrderByCode        = "code"
	OrderByType        = "type"
	OrderByCreatedDate = "created_date"
)

var DefaultOrderBy = order.NewBy(OrderByCode, order.ASC)
