package inventorytransactionapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (inventorytransactionbus.QueryFilter, error) {

	var filter inventorytransactionbus.QueryFilter

	if qp.InventoryTransactionID != "" {
		id, err := uuid.Parse(qp.InventoryTransactionID)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("transaction_id", err)
		}
		filter.InventoryTransactionID = &id

	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("location_id", err)
		}
		filter.LocationID = &id

	}

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("user_id", err)
		}
		filter.UserID = &id

	}

	if qp.Quantity != "" {
		q, err := strconv.Atoi(qp.Quantity)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("quantity", err)
		}
		filter.Quantity = &q
	}

	if qp.TransactionType != "" {
		filter.TransactionType = &qp.TransactionType
	}
	if qp.ReferenceNumber != "" {
		filter.ReferenceNumber = &qp.ReferenceNumber
	}

	if qp.TransactionDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.TransactionDate)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("transaction_date", err)
		}
		filter.TransactionDate = &t
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return inventorytransactionbus.QueryFilter{}, errs.NewFieldsError("updated_date`", err)
		}
		filter.UpdatedDate = &t
	}

	return filter, nil
}
