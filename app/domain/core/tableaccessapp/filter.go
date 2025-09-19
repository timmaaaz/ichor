package tableaccessapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
)

func parseFilter(qp QueryParams) (tableaccessbus.QueryFilter, error) {
	var filter tableaccessbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return tableaccessbus.QueryFilter{}, errs.NewFieldsError("table_access_id", err)
		}
		filter.ID = &id
	}

	if qp.RoleID != "" {
		id, err := uuid.Parse(qp.RoleID)
		if err != nil {
			return tableaccessbus.QueryFilter{}, errs.NewFieldsError("role_id", err)
		}
		filter.RoleID = &id
	}

	if qp.TableName != "" {
		filter.TableName = &qp.TableName
	}

	if qp.CanCreate != "" {
		canCreate, err := strconv.ParseBool(qp.CanCreate)
		if err != nil {
			return tableaccessbus.QueryFilter{}, errs.NewFieldsError("can_create", err)
		}
		filter.CanCreate = &canCreate
	}

	if qp.CanRead != "" {
		canRead, err := strconv.ParseBool(qp.CanRead)
		if err != nil {
			return tableaccessbus.QueryFilter{}, errs.NewFieldsError("can_read", err)
		}
		filter.CanRead = &canRead
	}

	if qp.CanUpdate != "" {
		canUpdate, err := strconv.ParseBool(qp.CanUpdate)
		if err != nil {
			return tableaccessbus.QueryFilter{}, errs.NewFieldsError("can_update", err)
		}
		filter.CanUpdate = &canUpdate
	}

	if qp.CanDelete != "" {
		canDelete, err := strconv.ParseBool(qp.CanDelete)
		if err != nil {
			return tableaccessbus.QueryFilter{}, errs.NewFieldsError("can_delete", err)
		}
		filter.CanDelete = &canDelete
	}

	return filter, nil
}
