package userassetapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
)

func parseFilter(qp QueryParams) (userassetbus.QueryFilter, error) {
	var filter userassetbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("user_asset_id", err)
		}
		filter.ID = &id
	}

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("user_id", err)
		}
		filter.UserID = &id
	}

	if qp.AssetID != "" {
		id, err := uuid.Parse(qp.AssetID)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("asset_id", err)
		}
		filter.AssetID = &id
	}

	if qp.ApprovalStatusID != "" {
		id, err := uuid.Parse(qp.ApprovalStatusID)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("approval_status_id", err)
		}
		filter.ApprovalStatusID = &id
	}

	if qp.ApprovedBy != "" {
		id, err := uuid.Parse(qp.ApprovedBy)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("approved_by", err)
		}
		filter.ApprovedBy = &id
	}

	if qp.ConditionID != "" {
		id, err := uuid.Parse(qp.ConditionID)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("condition_id", err)
		}
		filter.ConditionID = &id
	}

	if qp.FulfillmentStatusID != "" {
		id, err := uuid.Parse(qp.FulfillmentStatusID)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("fulfillment_status_id", err)
		}
		filter.FulfillmentStatusID = &id
	}

	if qp.DateReceived != "" {
		dateReceived, err := time.Parse(time.RFC3339, qp.DateReceived)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("date_received", err)
		}
		filter.DateReceived = &dateReceived
	}

	if qp.LastMaintenance != "" {
		lastMaintenance, err := time.Parse(time.RFC3339, qp.LastMaintenance)
		if err != nil {
			return userassetbus.QueryFilter{}, errs.NewFieldsError("last_maintenance", err)
		}
		filter.LastMaintenance = &lastMaintenance
	}

	return filter, nil
}
