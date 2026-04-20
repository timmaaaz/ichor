package scenarioapp

import (
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
)

func parseFilter(qp QueryParams) (scenariobus.QueryFilter, error) {
	var filter scenariobus.QueryFilter
	if qp.Name != "" {
		n := qp.Name
		filter.Name = &n
	}
	return filter, nil
}
