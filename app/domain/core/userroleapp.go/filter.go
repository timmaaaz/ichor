package userroleapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
)

func parseFilter(qp QueryParams) (userrolebus.QueryFilter, error) {
	var filter userrolebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return userrolebus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.UserID != "" {
		userID, err := uuid.Parse(qp.UserID)
		if err != nil {
			return userrolebus.QueryFilter{}, errs.NewFieldsError("user_id", err)
		}
		filter.UserID = &userID
	}

	if qp.RoleID != "" {
		roleID, err := uuid.Parse(qp.RoleID)
		if err != nil {
			return userrolebus.QueryFilter{}, errs.NewFieldsError("role_id", err)
		}
		filter.RoleID = &roleID
	}

	return filter, nil
}
