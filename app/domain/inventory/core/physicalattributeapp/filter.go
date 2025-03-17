package physicalattributeapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (physicalattributebus.QueryFilter, error) {
	var filter physicalattributebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}

	if qp.Color != "" {
		filter.Color = &qp.Color
	}

	if qp.Material != "" {
		filter.Material = &qp.Material
	}

	if qp.Size != "" {
		filter.Size = &qp.Size
	}

	if qp.Weight != "" {
		weight, err := strconv.ParseFloat(qp.Weight, 64)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("weight", err)
		}

		w := float64(weight)

		filter.Weight = &w
	}

	if qp.WeightUnit != "" {
		filter.WeightUnit = &qp.WeightUnit
	}

	if qp.Length != "" {
		length, err := strconv.ParseFloat(qp.Length, 64)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("length", err)
		}

		l := float64(length)

		filter.Length = &l
	}

	if qp.Width != "" {
		width, err := strconv.ParseFloat(qp.Width, 64)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("width", err)
		}

		w := float64(width)

		filter.Width = &w
	}

	if qp.Height != "" {
		height, err := strconv.ParseFloat(qp.Height, 64)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("height", err)
		}

		h := float64(height)
		filter.Height = &h
	}

	if qp.ShelfLifeDays != "" {
		shelfLifeDays, err := strconv.ParseInt(qp.ShelfLifeDays, 10, 0)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("shelf_life_days", err)
		}

		s := int(shelfLifeDays)

		filter.ShelfLifeDays = &s
	}

	if qp.HazmatClass != "" {
		filter.HazmatClass = &qp.HazmatClass
	}

	if qp.StorageRequirements != "" {
		filter.StorageRequirements = &qp.StorageRequirements
	}

	if qp.CreatedDate != "" {
		createdDate, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &createdDate
	}

	if qp.UpdatedDate != "" {
		updatedDate, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return physicalattributebus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &updatedDate
	}

	return filter, nil
}
