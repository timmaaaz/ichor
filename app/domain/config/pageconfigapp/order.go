package pageconfigapp

import (
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(pageconfigbus.OrderByName, order.ASC)

var orderByFields = map[string]string{
	pageconfigbus.OrderByID:        pageconfigbus.OrderByID,
	pageconfigbus.OrderByName:      pageconfigbus.OrderByName,
	pageconfigbus.OrderByIsDefault: pageconfigbus.OrderByIsDefault,
}
