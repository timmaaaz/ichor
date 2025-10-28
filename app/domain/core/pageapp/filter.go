package pageapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
)

func parseFilter(qp QueryParams) (pagebus.QueryFilter, error) {
	var filter pagebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return pagebus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Path != "" {
		filter.Path = &qp.Path
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Module != "" {
		filter.Module = &qp.Module
	}

	if qp.IsActive != "" {
		isActive, err := strconv.ParseBool(qp.IsActive)
		if err != nil {
			return pagebus.QueryFilter{}, errs.NewFieldsError("is_active", err)
		}
		filter.IsActive = &isActive
	}

	return filter, nil
}
