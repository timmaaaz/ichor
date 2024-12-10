package reportstoapp

import (
	"github.com/timmaaaz/ichor/business/domain/reportstobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("reports_to_id", order.ASC)

var orderByFields = map[string]string{
	"reports_to_id": reportstobus.OrderByID,
	"boss_id":       reportstobus.OrderByBossID,
	"reporter_id":   reportstobus.OrderByReporterID,
}
