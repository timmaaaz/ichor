package actionpermissionsapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
)

func parseFilter(qp QueryParams) (actionpermissionsbus.QueryFilter, error) {
	var filter actionpermissionsbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return actionpermissionsbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.RoleID != "" {
		roleID, err := uuid.Parse(qp.RoleID)
		if err != nil {
			return actionpermissionsbus.QueryFilter{}, errs.NewFieldsError("roleId", err)
		}
		filter.RoleID = &roleID
	}

	if qp.ActionType != "" {
		filter.ActionType = &qp.ActionType
	}

	if qp.IsAllowed != "" {
		isAllowed, err := strconv.ParseBool(qp.IsAllowed)
		if err != nil {
			return actionpermissionsbus.QueryFilter{}, errs.NewFieldsError("isAllowed", err)
		}
		filter.IsAllowed = &isAllowed
	}

	return filter, nil
}
