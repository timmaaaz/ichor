package cityapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
)

func parseFilter(qp QueryParams) (citybus.QueryFilter, error) {
	var filter citybus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return citybus.QueryFilter{}, errs.NewFieldsError("city_id", err)
		}
		filter.ID = &id
	}

	if qp.RegionID != "" {
		id, err := uuid.Parse(qp.RegionID)
		if err != nil {
			return citybus.QueryFilter{}, errs.NewFieldsError("region_id", err)
		}
		filter.RegionID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	return filter, nil
}
