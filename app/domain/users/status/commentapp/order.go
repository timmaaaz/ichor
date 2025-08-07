package commentapp

import (
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":           commentbus.OrderByID,
	"user_id":      commentbus.OrderByUserID,
	"commenter_id": commentbus.OrderByCommenterID,
	"comment":      commentbus.OrderByComment,
	"created_date": commentbus.OrderByCreatedDate,
}
