package commentbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByCreatedDate, order.DESC)

const (
	OrderByID          = "comment_id"
	OrderByCommenterID = "commenter_id"
	OrderByUserID      = "user_id"
	OrderByCreatedDate = "created_date"
	OrderByComment     = "comment"
)
