package lineitemfulfillmentstatusapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
)

func parseFilter(qp QueryParams) (lineitemfulfillmentstatusbus.QueryFilter, error) {
	var filter lineitemfulfillmentstatusbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return lineitemfulfillmentstatusbus.QueryFilter{}, errs.NewFieldsError("id", err)
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
