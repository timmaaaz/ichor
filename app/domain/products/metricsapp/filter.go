package metricsapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (metricsbus.QueryFilter, error) {
	var filter metricsbus.QueryFilter

	if qp.MetricID != "" {
		id, err := uuid.Parse(qp.MetricID)
		if err != nil {
			return metricsbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.MetricID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return metricsbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}

	if qp.ReturnRate != "" {
		f, err := types.ParseRoundedFloat(qp.ReturnRate)
		if err != nil {
			return metricsbus.QueryFilter{}, errs.NewFieldsError("return_rate", err)
		}
		filter.ReturnRate = &f
	}

	if qp.DefectRate != "" {
		f, err := types.ParseRoundedFloat(qp.DefectRate)
		if err != nil {
			return metricsbus.QueryFilter{}, errs.NewFieldsError("defect_rate", err)
		}
		filter.DefectRate = &f
	}

	if qp.MeasurementPeriod != "" {
		i, err := types.ParseInterval(qp.MeasurementPeriod)
		if err != nil {
			return metricsbus.QueryFilter{}, errs.NewFieldsError("measurement_period", err)
		}
		filter.MeasurementPeriod = &i
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return metricsbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return metricsbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	return filter, nil
}
