package timezoneapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
)

func parseFilter(qp QueryParams) (timezonebus.QueryFilter, error) {
	var filter timezonebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return timezonebus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.DisplayName != "" {
		filter.DisplayName = &qp.DisplayName
	}

	if qp.IsActive != "" {
		isActive, err := strconv.ParseBool(qp.IsActive)
		if err != nil {
			return timezonebus.QueryFilter{}, errs.NewFieldsError("is_active", err)
		}
		filter.IsActive = &isActive
	}

	return filter, nil
}
