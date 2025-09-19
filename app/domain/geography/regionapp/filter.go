package regionapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
)

func parseFilter(qp QueryParams) (regionbus.QueryFilter, error) {
	var filter regionbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return regionbus.QueryFilter{}, errs.NewFieldsError("region_id", err)
		}
		filter.ID = &id
	}

	if qp.CountryID != "" {
		countryID, err := uuid.Parse(qp.CountryID)
		if err != nil {
			return regionbus.QueryFilter{}, errs.NewFieldsError("country_id", err)
		}
		filter.CountryID = &countryID
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Code != "" {
		filter.Code = &qp.Code
	}

	return filter, nil
}
