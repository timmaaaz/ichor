package titleapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/titlebus"
)

func parseFilter(qp QueryParams) (titlebus.QueryFilter, error) {
	var filter titlebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return titlebus.QueryFilter{}, errs.NewFieldsError("id", err)
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
