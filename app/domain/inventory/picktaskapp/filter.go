package picktaskapp

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
)

func parseFilter(qp QueryParams) (picktaskbus.QueryFilter, error) {
	var filter picktaskbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.SalesOrderID != "" {
		id, err := uuid.Parse(qp.SalesOrderID)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.SalesOrderID = &id
	}

	if qp.SalesOrderLineItemID != "" {
		id, err := uuid.Parse(qp.SalesOrderLineItemID)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.SalesOrderLineItemID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.LocationID = &id
	}

	if qp.Status != "" {
		st, err := picktaskbus.ParseStatus(qp.Status)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.Status = &st
	}

	if qp.AssignedTo != "" {
		id, err := uuid.Parse(qp.AssignedTo)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.AssignedTo = &id
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return picktaskbus.QueryFilter{}, err
		}
		filter.CreatedBy = &id
	}

	return filter, nil
}
