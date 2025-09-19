package streetapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
)

func parseFilter(qp QueryParams) (streetbus.QueryFilter, error) {
	var filter streetbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return streetbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.CityID != "" {
		id, err := uuid.Parse(qp.CityID)
		if err != nil {
			return streetbus.QueryFilter{}, errs.NewFieldsError("city_id", err)
		}
		filter.CityID = &id
	}

	if qp.Line1 != "" {
		filter.Line1 = &qp.Line1
	}

	if qp.Line2 != "" {
		filter.Line2 = &qp.Line2
	}

	if qp.PostalCode != "" {
		filter.PostalCode = &qp.PostalCode
	}

	return filter, nil
}
