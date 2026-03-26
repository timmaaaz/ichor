package cyclecountsessionapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (cyclecountsessionbus.QueryFilter, error) {
	var filter cyclecountsessionbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return cyclecountsessionbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Status != "" {
		st, err := cyclecountsessionbus.ParseStatus(qp.Status)
		if err != nil {
			return cyclecountsessionbus.QueryFilter{}, err
		}
		filter.Status = &st
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return cyclecountsessionbus.QueryFilter{}, err
		}
		filter.CreatedBy = &id
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return cyclecountsessionbus.QueryFilter{}, err
		}
		filter.CreatedDate = &t
	}

	return filter, nil
}
