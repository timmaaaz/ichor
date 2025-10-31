package purchaseorderstatusapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
)

func parseFilter(qp QueryParams) (purchaseorderstatusbus.QueryFilter, error) {
	var filter purchaseorderstatusbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return purchaseorderstatusbus.QueryFilter{}, err
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
