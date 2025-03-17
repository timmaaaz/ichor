package productapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (productbus.QueryFilter, error) {
	var filter productbus.QueryFilter

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.BrandID != "" {
		id, err := uuid.Parse(qp.BrandID)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("brand_id", err)
		}
		filter.BrandID = &id
	}

	if qp.ProductCategoryID != "" {
		id, err := uuid.Parse(qp.ProductCategoryID)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("product_category_id", err)
		}
		filter.ProductCategoryID = &id
	}

	if qp.Status != "" {
		filter.Status = &qp.Status
	}

	if qp.IsActive != "" {
		isActive := qp.IsActive == "true"
		filter.IsActive = &isActive
	}

	if qp.IsPerishable != "" {
		isPerishable := qp.IsPerishable == "true"
		filter.IsPerishable = &isPerishable
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &date
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	if qp.Description != "" {
		filter.Description = &qp.Description
	}

	if qp.HandlingInstructions != "" {
		filter.HandlingInstructions = &qp.HandlingInstructions
	}

	if qp.ModelNumber != "" {
		filter.ModelNumber = &qp.ModelNumber
	}

	if qp.UnitsPerCase != "" {
		unitsPerCase, err := strconv.Atoi(qp.UnitsPerCase)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldsError("units_per_case", err)
		}
		filter.UnitsPerCase = &unitsPerCase
	}

	if qp.UpcCode != "" {
		filter.UpcCode = &qp.UpcCode
	}

	return filter, nil
}
