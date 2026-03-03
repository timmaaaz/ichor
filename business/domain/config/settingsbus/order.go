package settingsbus

import (
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(OrderByKey, order.ASC)

const (
	OrderByKey         = "key"
	OrderByCreatedDate = "created_date"
	OrderByUpdatedDate = "updated_date"
)

var orderByFields = map[string]string{
	OrderByKey:         "key",
	OrderByCreatedDate: "created_date",
	OrderByUpdatedDate: "updated_date",
}
