package serialnumberapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (serialnumberbus.QueryFilter, error) {
	var filter serialnumberbus.QueryFilter

	if qp.SerialID != "" {
		id, err := uuid.Parse(qp.SerialID)
		if err != nil {
			return serialnumberbus.QueryFilter{}, errs.NewFieldsError("serial_id", err)
		}
		filter.SerialID = &id
	}

	if qp.LotID != "" {
		id, err := uuid.Parse(qp.LotID)
		if err != nil {
			return serialnumberbus.QueryFilter{}, errs.NewFieldsError("lot_id", err)
		}
		filter.LotID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return serialnumberbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return serialnumberbus.QueryFilter{}, errs.NewFieldsError("location_id", err)
		}
		filter.LocationID = &id
	}

	if qp.SerialNumber != "" {
		filter.SerialNumber = &qp.SerialNumber
	}

	if qp.Status != "" {
		filter.Status = &qp.Status
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return serialnumberbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return serialnumberbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	return filter, nil
}
