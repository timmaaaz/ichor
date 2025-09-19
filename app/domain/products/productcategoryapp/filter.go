package productcategoryapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (productcategorybus.QueryFilter, error) {
	var filter productcategorybus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return productcategorybus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Description != "" {
		filter.Description = &qp.Description
	}

	if qp.CreatedDate != "" {
		createdDate, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return productcategorybus.QueryFilter{}, err
		}
		filter.CreatedDate = &createdDate
	}

	if qp.UpdatedDate != "" {
		updatedDate, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return productcategorybus.QueryFilter{}, err
		}
		filter.UpdatedDate = &updatedDate
	}

	return filter, nil
}
