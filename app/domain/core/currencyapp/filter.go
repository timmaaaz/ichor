package currencyapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
)

func parseFilter(qp QueryParams) (currencybus.QueryFilter, error) {
	var filter currencybus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return currencybus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Code != "" {
		filter.Code = &qp.Code
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.IsActive != "" {
		isActive, err := strconv.ParseBool(qp.IsActive)
		if err != nil {
			return currencybus.QueryFilter{}, errs.NewFieldsError("is_active", err)
		}
		filter.IsActive = &isActive
	}

	return filter, nil
}
