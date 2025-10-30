package pageactionapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
)

func parseFilter(qp QueryParams) (pageactionbus.QueryFilter, error) {
	var filter pageactionbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return pageactionbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.PageConfigID != "" {
		pageConfigID, err := uuid.Parse(qp.PageConfigID)
		if err != nil {
			return pageactionbus.QueryFilter{}, errs.NewFieldsError("pageConfigId", err)
		}
		filter.PageConfigID = &pageConfigID
	}

	if qp.ActionType != "" {
		actionType := pageactionbus.ActionType(qp.ActionType)
		// Validate action type
		switch actionType {
		case pageactionbus.ActionTypeButton, pageactionbus.ActionTypeDropdown, pageactionbus.ActionTypeSeparator:
			filter.ActionType = &actionType
		default:
			return pageactionbus.QueryFilter{}, errs.NewFieldsError("actionType", errs.Newf(errs.InvalidArgument, "invalid action type: %s", qp.ActionType))
		}
	}

	if qp.IsActive != "" {
		isActive, err := strconv.ParseBool(qp.IsActive)
		if err != nil {
			return pageactionbus.QueryFilter{}, errs.NewFieldsError("isActive", err)
		}
		filter.IsActive = &isActive
	}

	return filter, nil
}
