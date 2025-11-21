package pageconfigapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
)

func parseFilter(qp QueryParams) (pageconfigbus.QueryFilter, error) {
	var filter pageconfigbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return pageconfigbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.UserID != "" {
		userID, err := uuid.Parse(qp.UserID)
		if err != nil {
			return pageconfigbus.QueryFilter{}, errs.NewFieldsError("user_id", err)
		}
		filter.UserID = &userID
	}

	if qp.IsDefault != "" {
		isDefault := qp.IsDefault == "true"
		filter.IsDefault = &isDefault
	}

	return filter, nil
}
