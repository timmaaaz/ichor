package putawaytaskapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
)

func parseFilter(qp QueryParams) (putawaytaskbus.QueryFilter, error) {
	var filter putawaytaskbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return putawaytaskbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return putawaytaskbus.QueryFilter{}, err
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return putawaytaskbus.QueryFilter{}, err
		}
		filter.LocationID = &id
	}

	if qp.Status != "" {
		st, err := putawaytaskbus.ParseStatus(qp.Status)
		if err != nil {
			return putawaytaskbus.QueryFilter{}, err
		}
		filter.Status = &st
	}

	if qp.AssignedTo != "" {
		id, err := uuid.Parse(qp.AssignedTo)
		if err != nil {
			return putawaytaskbus.QueryFilter{}, err
		}
		filter.AssignedTo = &id
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return putawaytaskbus.QueryFilter{}, err
		}
		filter.CreatedBy = &id
	}

	if qp.ReferenceNumber != "" {
		filter.ReferenceNumber = &qp.ReferenceNumber
	}

	return filter, nil
}
