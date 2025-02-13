package officeapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/location/officebus"
)

func parseFilter(qp QueryParams) (officebus.QueryFilter, error) {

	var filter officebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return officebus.QueryFilter{}, errs.NewFieldsError("office_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.StreetID != "" {
		id, err := uuid.Parse(qp.StreetID)
		if err != nil {
			return officebus.QueryFilter{}, errs.NewFieldsError("street_id", err)
		}
		filter.StreetID = &id
	}

	return filter, nil
}
