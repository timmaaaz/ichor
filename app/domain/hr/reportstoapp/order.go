package reportstoapp

import (
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":          reportstobus.OrderByID,
	"boss_id":     reportstobus.OrderByBossID,
	"reporter_id": reportstobus.OrderByReporterID,
}
