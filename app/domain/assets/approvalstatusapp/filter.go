package approvalstatusapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
)

func parseFilter(qp QueryParams) (approvalstatusbus.QueryFilter, error) {
	var filter approvalstatusbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return approvalstatusbus.QueryFilter{}, errs.NewFieldsError("approval_status_id", err)
		}
		filter.ID = &id
	}

	if qp.IconID != "" {
		id, err := uuid.Parse(qp.IconID)
		if err != nil {
			return approvalstatusbus.QueryFilter{}, errs.NewFieldsError("icon_id", err)
		}
		filter.IconID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	return filter, nil
}
