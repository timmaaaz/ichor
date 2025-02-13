package assetconditionapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
)

func parseFilter(qp QueryParams) (assetconditionbus.QueryFilter, error) {
	var filter assetconditionbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return assetconditionbus.QueryFilter{}, errs.NewFieldsError("id", err)
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
