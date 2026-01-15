package paymenttermapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
)

func parseFilter(qp QueryParams) (paymenttermbus.QueryFilter, error) {
	var filter paymenttermbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return paymenttermbus.QueryFilter{}, errs.NewFieldsError("id", err)
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
