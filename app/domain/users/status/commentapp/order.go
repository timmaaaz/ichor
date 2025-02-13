package commentapp

import (
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("comment_id", order.ASC)

var orderByFields = map[string]string{
	"comment_id":   commentbus.OrderByID,
	"user_id":      commentbus.OrderByUserID,
	"commenter_id": commentbus.OrderByCommenterID,
	"comment":      commentbus.OrderByComment,
	"created_date": commentbus.OrderByCreatedDate,
}
