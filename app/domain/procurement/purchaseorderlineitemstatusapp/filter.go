package purchaseorderlineitemstatusapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
)

func parseFilter(qp QueryParams) (purchaseorderlineitemstatusbus.QueryFilter, error) {
	var filter purchaseorderlineitemstatusbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return purchaseorderlineitemstatusbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Description != "" {
		filter.Description = &qp.Description
	}

	return filter, nil
}
