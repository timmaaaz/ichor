package assettypeapp

import (
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
)

func parseFilter(qp QueryParams) (assettypebus.QueryFilter, error) {
	var filter assettypebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return assettypebus.QueryFilter{}, errs.NewFieldsError("asset_type_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	return filter, nil
}
