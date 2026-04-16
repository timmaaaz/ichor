package labelapp

import (
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
)

func parseFilter(qp QueryParams) (labelbus.QueryFilter, error) {
	var filter labelbus.QueryFilter
	if qp.Type != "" {
		t := qp.Type
		filter.Type = &t
	}
	return filter, nil
}
