package reportstobus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID         = "reports_to_id"
	OrderByReporterID = "reporter_id"
	OrderByBossID     = "boss_id"
)
