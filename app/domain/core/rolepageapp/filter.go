package rolepageapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
)

func parseFilter(qp QueryParams) (rolepagebus.QueryFilter, error) {
	var filter rolepagebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return rolepagebus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.RoleID != "" {
		roleID, err := uuid.Parse(qp.RoleID)
		if err != nil {
			return rolepagebus.QueryFilter{}, errs.NewFieldsError("role_id", err)
		}
		filter.RoleID = &roleID
	}

	if qp.PageID != "" {
		pageID, err := uuid.Parse(qp.PageID)
		if err != nil {
			return rolepagebus.QueryFilter{}, errs.NewFieldsError("page_id", err)
		}
		filter.PageID = &pageID
	}

	if qp.CanAccess != "" {
		canAccess, err := strconv.ParseBool(qp.CanAccess)
		if err != nil {
			return rolepagebus.QueryFilter{}, errs.NewFieldsError("can_access", err)
		}
		filter.CanAccess = &canAccess
	}

	return filter, nil
}
