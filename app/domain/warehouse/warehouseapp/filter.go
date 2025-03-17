package warehouseapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
)

func parseFilter(qp QueryParams) (warehousebus.QueryFilter, error) {
	var filter warehousebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return warehousebus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.StreetID != "" {
		id, err := uuid.Parse(qp.StreetID)
		if err != nil {
			return warehousebus.QueryFilter{}, errs.NewFieldsError("street_id", err)
		}
		filter.StreetID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.IsActive != "" {
		isActive, err := strconv.ParseBool(qp.IsActive)
		if err != nil {
			return warehousebus.QueryFilter{}, errs.NewFieldsError("is_active", err)
		}
		filter.IsActive = &isActive
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return warehousebus.QueryFilter{}, errs.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return warehousebus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &t
	}

	if qp.StartUpdatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartUpdatedDate)
		if err != nil {
			return warehousebus.QueryFilter{}, errs.NewFieldsError("start_updated_date", err)
		}
		filter.StartUpdatedDate = &t
	}

	if qp.EndUpdatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndUpdatedDate)
		if err != nil {
			return warehousebus.QueryFilter{}, errs.NewFieldsError("end_updated_date", err)
		}
		filter.EndUpdatedDate = &t
	}

	return filter, nil
}
