package tagapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
)

func parseFilter(qp QueryParams) (tagbus.QueryFilter, error) {
	var filter tagbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return tagbus.QueryFilter{}, errs.NewFieldsError("id", err)
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
