package inventoryadjustmentapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (inventoryadjustmentbus.QueryFilter, error) {
	var filter inventoryadjustmentbus.QueryFilter

	if qp.InventoryAdjustmentID != "" {
		id, err := uuid.Parse(qp.InventoryAdjustmentID)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("adjustment_id", err)
		}
		filter.InventoryAdjustmentID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("location_id", err)
		}
		filter.LocationID = &id

	}

	if qp.AdjustedBy != "" {
		id, err := uuid.Parse(qp.AdjustedBy)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("adjusted_by", err)
		}
		filter.AdjustedBy = &id
	}

	if qp.ApprovedBy != "" {
		id, err := uuid.Parse(qp.ApprovedBy)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("approved_by", err)
		}
		filter.ApprovedBy = &id
	}

	if qp.ApprovalStatus != "" {
		filter.ApprovalStatus = &qp.ApprovalStatus
	}

	if qp.QuantityChange != "" {
		change, err := strconv.Atoi(qp.QuantityChange)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("quantity_change", err)
		}
		filter.QuantityChange = &change
	}

	if qp.ReasonCode != "" {
		filter.ReasonCode = &qp.ReasonCode
	}

	if qp.Notes != "" {
		filter.Notes = &qp.Notes
	}

	if qp.AdjustmentDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.AdjustmentDate)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("adjustment_date", err)
		}

		filter.AdjustmentDate = &date
	}

	if qp.StartAdjustmentDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.StartAdjustmentDate)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("startAdjustmentDate", err)
		}
		filter.StartAdjustmentDate = &date
	}

	if qp.EndAdjustmentDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.EndAdjustmentDate)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("endAdjustmentDate", err)
		}
		filter.EndAdjustmentDate = &date
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &date
	}

	if qp.StartCreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.StartCreatedDate)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("startCreatedDate", err)
		}
		filter.StartCreatedDate = &date
	}

	if qp.EndCreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.EndCreatedDate)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("endCreatedDate", err)
		}
		filter.EndCreatedDate = &date
	}

	if qp.UpdatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return inventoryadjustmentbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &date
	}

	return filter, nil

}
