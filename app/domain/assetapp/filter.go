package assetapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (assetbus.QueryFilter, error) {
	var filter assetbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return assetbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.ConditionID != "" {
		id, err := uuid.Parse(qp.ConditionID)
		if err != nil {
			return assetbus.QueryFilter{}, errs.NewFieldsError("condition_id", err)
		}
		filter.AssetConditionID = &id
	}

	if qp.LastMaintenance != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.LastMaintenance)
		if err != nil {
			return assetbus.QueryFilter{}, errs.NewFieldsError("last_maintenance", err)
		}
		filter.LastMaintenance = &t
	}

	if qp.SerialNumber != "" {
		filter.SerialNumber = &qp.SerialNumber
	}

	if qp.ValidAssetID != "" {
		id, err := uuid.Parse(qp.ValidAssetID)
		if err != nil {
			return assetbus.QueryFilter{}, errs.NewFieldsError("valid_asset_id", err)
		}
		filter.ValidAssetID = &id
	}

	return filter, nil
}
