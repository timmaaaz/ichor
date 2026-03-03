package settingsapp

import (
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(settingsbus.OrderByKey, order.ASC)

var orderByFields = map[string]string{
	"key":          settingsbus.OrderByKey,
	"created_date": settingsbus.OrderByCreatedDate,
	"updated_date": settingsbus.OrderByUpdatedDate,
}
