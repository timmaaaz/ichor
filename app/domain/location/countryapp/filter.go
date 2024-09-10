package countryapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
)

func parseFilter(qp QueryParams) (countrybus.QueryFilter, error) {
	var filter countrybus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return countrybus.QueryFilter{}, errs.NewFieldsError("country_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Alpha2 != "" {
		filter.Alpha2 = &qp.Alpha2
	}

	if qp.Alpha3 != "" {
		filter.Alpha3 = &qp.Alpha3
	}

	return filter, nil
}
