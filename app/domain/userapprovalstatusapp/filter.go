package userapprovalstatusapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
)

func parseFilter(qp QueryParams) (userapprovalstatusbus.QueryFilter, error) {
	var filter userapprovalstatusbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return userapprovalstatusbus.QueryFilter{}, errs.NewFieldsError("user_approval_status_id", err)
		}
		filter.ID = &id
	}

	if qp.IconID != "" {
		id, err := uuid.Parse(qp.IconID)
		if err != nil {
			return userapprovalstatusbus.QueryFilter{}, errs.NewFieldsError("icon_id", err)
		}
		filter.IconID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	return filter, nil
}
