package inspectionapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/quality/inspectionbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (inspectionbus.QueryFilter, error) {

	var filter inspectionbus.QueryFilter

	if qp.InspectionID != "" {
		id, err := uuid.Parse(qp.InspectionID)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("inspection_id", err)
		}
		filter.InspectionID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id

	}

	if qp.InspectorID != "" {
		id, err := uuid.Parse(qp.InspectorID)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("inspector_id", err)
		}
		filter.InspectorID = &id
	}

	if qp.LotID != "" {
		id, err := uuid.Parse(qp.LotID)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("lot_id", err)
		}
		filter.LotID = &id
	}

	if qp.Status != "" {
		filter.Status = &qp.Status
	}

	if qp.Notes != "" {
		filter.Notes = &qp.Notes
	}

	if qp.InspectionDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.InspectionDate)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("inspection_date", err)
		}
		filter.InspectionDate = &date
	}

	if qp.NextInspectionDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.NextInspectionDate)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("next_inspection_date", err)
		}
		filter.NextInspectionDate = &date
	}

	if qp.UpdatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &date
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return inspectionbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &date
	}

	return filter, nil

}
