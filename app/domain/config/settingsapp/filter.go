package settingsapp

import (
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
)

func parseFilter(qp QueryParams) (settingsbus.QueryFilter, error) {
	var filter settingsbus.QueryFilter

	if qp.Key != "" {
		filter.Key = &qp.Key
	}

	if qp.Prefix != "" {
		filter.Prefix = &qp.Prefix
	}

	return filter, nil
}
