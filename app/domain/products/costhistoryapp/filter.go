package costhistoryapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (costhistorybus.QueryFilter, error) {
	var filter costhistorybus.QueryFilter

	if qp.Amount != "" {
		amount, err := types.ParseMoney(qp.Amount)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.Amount = &amount
	}

	if qp.CostHistoryID != "" {
		id, err := uuid.Parse(qp.CostHistoryID)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.CostHistoryID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.ProductID = &id
	}

	if qp.CostType != "" {
		filter.CostType = &qp.CostType
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.UpdatedDate = &t
	}

	if qp.EffectiveDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EffectiveDate)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.EffectiveDate = &t
	}

	if qp.CurrencyID != "" {
		id, err := uuid.Parse(qp.CurrencyID)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.CurrencyID = &id
	}

	if qp.EffectiveDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EffectiveDate)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.EffectiveDate = &t
	}

	if qp.EndDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndDate)
		if err != nil {
			return costhistorybus.QueryFilter{}, err
		}
		filter.EndDate = &t
	}

	return filter, nil

}
