package pagecontentapp

import (
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(pagecontentbus.OrderByOrderIndex, order.ASC)

var orderByFields = map[string]string{
	pagecontentbus.OrderByID:          pagecontentbus.OrderByID,
	pagecontentbus.OrderByOrderIndex:  pagecontentbus.OrderByOrderIndex,
	pagecontentbus.OrderByLabel:       pagecontentbus.OrderByLabel,
	pagecontentbus.OrderByContentType: pagecontentbus.OrderByContentType,
}
