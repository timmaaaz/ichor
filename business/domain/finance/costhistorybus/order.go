package costhistorybus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByAmount, order.ASC)

const (
	OrderByCostHistoryID = "history_id"
	OrderByProductID     = "product_id"
	OrderByCostType      = "cost_type"
	OrderByAmount        = "amount"
	OrderByCurrency      = "currency"
	OrderByEffectiveDate = "effective_date"
	OrderByEndDate       = "end_date"
	OrderByCreatedDate   = "created_date"
	OrderByUpdatedDate   = "updated_date"
)
