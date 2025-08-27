package validassetapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"

	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus/types"
)

func parseFilter(qp QueryParams) (validassetbus.QueryFilter, error) {
	var filter validassetbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("asset_id", err)
		}
		filter.ID = &id
	}

	if qp.TypeID != "" {
		id, err := uuid.Parse(qp.TypeID)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("type_id", err)
		}
		filter.TypeID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.EstPrice != "" {
		estPrice, err := types.ParseMoneyPtr(qp.EstPrice)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("est_price", err)
		}

		filter.EstPrice = estPrice
	}

	if qp.Price != "" {
		price, err := types.ParseMoneyPtr(qp.Price)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("price", err)
		}
		filter.Price = price
	}

	if qp.MaintenanceInterval != "" {
		maintenanceInterval, err := types.ParseIntervalPtr(qp.MaintenanceInterval)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("maintenance_interval", err)
		}
		filter.MaintenanceInterval = maintenanceInterval
	}

	if qp.LifeExpectancy != "" {
		lifeExpectancy, err := types.ParseIntervalPtr(qp.LifeExpectancy)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("life_expectancy", err)
		}
		filter.LifeExpectancy = lifeExpectancy
	}

	if qp.ModelNumber != "" {
		filter.ModelNumber = &qp.ModelNumber
	}

	if qp.IsEnabled != "" {
		isEnabled, err := strconv.ParseBool(qp.IsEnabled)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("is_enabled", err)
		}
		filter.IsEnabled = &isEnabled
	}

	if qp.StartCreatedDate != "" {
		startCreatedDate, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &startCreatedDate
	}

	if qp.EndCreatedDate != "" {
		endCreatedDate, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &endCreatedDate
	}

	if qp.StartUpdatedDate != "" {
		startUpdatedDate, err := time.Parse(time.RFC3339, qp.StartUpdatedDate)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("start_updated_date", err)
		}
		filter.StartUpdatedDate = &startUpdatedDate
	}

	if qp.EndUpdatedDate != "" {
		endUpdatedDate, err := time.Parse(time.RFC3339, qp.EndUpdatedDate)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("end_updated_date", err)
		}
		filter.EndUpdatedDate = &endUpdatedDate
	}

	if qp.CreatedBy != "" {
		createdBy, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("created_by", err)
		}
		filter.CreatedBy = &createdBy
	}

	if qp.UpdatedBy != "" {
		updatedBy, err := uuid.Parse(qp.UpdatedBy)
		if err != nil {
			return validassetbus.QueryFilter{}, errs.NewFieldsError("updated_by", err)
		}
		filter.UpdatedBy = &updatedBy
	}

	return filter, nil
}
