package lottrackingsapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (lottrackingsbus.QueryFilter, error) {
	var filter lottrackingsbus.QueryFilter

	if qp.LotID != "" {
		id, err := uuid.Parse(qp.LotID)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("lot_id", err)
		}
		filter.LotID = &id
	}

	if qp.SupplierProductID != "" {
		id, err := uuid.Parse(qp.SupplierProductID)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("supplier_product_id", err)
		}
		filter.SupplierProductID = &id
	}

	if qp.LotNumber != "" {
		filter.LotNumber = &qp.LotNumber
	}

	if qp.ManufactureDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.ManufactureDate)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("manufacture_date", err)
		}
		filter.ManufactureDate = &date
	}

	if qp.ExpirationDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.ExpirationDate)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("expiration_date", err)
		}
		filter.ExpirationDate = &date
	}

	if qp.RecievedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.RecievedDate)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("received_date", err)
		}
		filter.RecievedDate = &date
	}

	if qp.Quantity != "" {
		quantity, err := strconv.Atoi(qp.Quantity)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("quantity", err)
		}
		filter.Quantity = &quantity
	}

	if qp.QualityStatus != "" {
		filter.QualityStatus = &qp.QualityStatus
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &date
	}

	if qp.UpdatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return lottrackingsbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &date
	}

	return filter, nil
}
