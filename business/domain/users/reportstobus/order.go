package reportstobus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID         = "id"
	OrderByReporterID = "reporter_id"
	OrderByBossID     = "boss_id"
)
