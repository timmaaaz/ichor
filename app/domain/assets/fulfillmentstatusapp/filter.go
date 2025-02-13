package fulfillmentstatusapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
)

func parseFilter(qp QueryParams) (fulfillmentstatusbus.QueryFilter, error) {
	var filter fulfillmentstatusbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return fulfillmentstatusbus.QueryFilter{}, errs.NewFieldsError("fulfillment_status_id", err)
		}
		filter.ID = &id
	}

	if qp.IconID != "" {
		id, err := uuid.Parse(qp.IconID)
		if err != nil {
			return fulfillmentstatusbus.QueryFilter{}, errs.NewFieldsError("icon_id", err)
		}
		filter.IconID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	return filter, nil
}
