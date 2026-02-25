package lotlocationapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
)

func parseFilter(qp QueryParams) (lotlocationbus.QueryFilter, error) {
	var filter lotlocationbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return lotlocationbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.LotID != "" {
		id, err := uuid.Parse(qp.LotID)
		if err != nil {
			return lotlocationbus.QueryFilter{}, errs.NewFieldsError("lot_id", err)
		}
		filter.LotID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return lotlocationbus.QueryFilter{}, errs.NewFieldsError("location_id", err)
		}
		filter.LocationID = &id
	}

	return filter, nil
}
