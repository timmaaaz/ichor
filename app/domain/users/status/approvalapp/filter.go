package approvalapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
)

func parseFilter(qp QueryParams) (approvalbus.QueryFilter, error) {
	var filter approvalbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return approvalbus.QueryFilter{}, errs.NewFieldsError("user_approval_status_id", err)
		}
		filter.ID = &id
	}

	if qp.IconID != "" {
		id, err := uuid.Parse(qp.IconID)
		if err != nil {
			return approvalbus.QueryFilter{}, errs.NewFieldsError("icon_id", err)
		}
		filter.IconID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	return filter, nil
}
