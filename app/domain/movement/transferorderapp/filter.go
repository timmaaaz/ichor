package transferorderapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/movement/transferorderbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (transferorderbus.QueryFilter, error) {
	var filter transferorderbus.QueryFilter

	if qp.TransferID != "" {
		id, err := uuid.Parse(qp.TransferID)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.TransferID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.ProductID = &id
	}

	if qp.FromLocationID != "" {
		id, err := uuid.Parse(qp.FromLocationID)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.FromLocationID = &id
	}

	if qp.ToLocationID != "" {
		id, err := uuid.Parse(qp.ToLocationID)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.ToLocationID = &id
	}

	if qp.RequestedByID != "" {
		id, err := uuid.Parse(qp.RequestedByID)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.RequestedByID = &id
	}

	if qp.ApprovedByID != "" {
		id, err := uuid.Parse(qp.ApprovedByID)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.ApprovedByID = &id
	}

	if qp.Quantity != "" {
		quantity, err := strconv.Atoi(qp.Quantity)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.Quantity = &quantity
	}

	if qp.Status != "" {
		filter.Status = &qp.Status
	}

	if qp.TransferDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.TransferDate)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.TransferDate = &date
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.CreatedDate = &date
	}

	if qp.UpdatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return transferorderbus.QueryFilter{}, err
		}
		filter.UpdatedDate = &date
	}

	return filter, nil

}
