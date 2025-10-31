package purchaseorderlineitemapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (purchaseorderlineitembus.QueryFilter, error) {
	var filter purchaseorderlineitembus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.PurchaseOrderID != "" {
		id, err := uuid.Parse(qp.PurchaseOrderID)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("purchaseOrderId", err)
		}
		filter.PurchaseOrderID = &id
	}

	if qp.SupplierProductID != "" {
		id, err := uuid.Parse(qp.SupplierProductID)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("supplierProductId", err)
		}
		filter.SupplierProductID = &id
	}

	if qp.LineItemStatusID != "" {
		id, err := uuid.Parse(qp.LineItemStatusID)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("lineItemStatusId", err)
		}
		filter.LineItemStatusID = &id
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("createdBy", err)
		}
		filter.CreatedBy = &id
	}

	if qp.UpdatedBy != "" {
		id, err := uuid.Parse(qp.UpdatedBy)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("updatedBy", err)
		}
		filter.UpdatedBy = &id
	}

	if qp.StartExpectedDeliveryDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.StartExpectedDeliveryDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("startExpectedDeliveryDate", err)
		}
		filter.StartExpectedDeliveryDate = &t
	}

	if qp.EndExpectedDeliveryDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndExpectedDeliveryDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("endExpectedDeliveryDate", err)
		}
		filter.EndExpectedDeliveryDate = &t
	}

	if qp.StartActualDeliveryDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.StartActualDeliveryDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("startActualDeliveryDate", err)
		}
		filter.StartActualDeliveryDate = &t
	}

	if qp.EndActualDeliveryDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndActualDeliveryDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("endActualDeliveryDate", err)
		}
		filter.EndActualDeliveryDate = &t
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.StartCreatedDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("startCreatedDate", err)
		}
		filter.StartCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndCreatedDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("endCreatedDate", err)
		}
		filter.EndCreatedDate = &t
	}

	if qp.StartUpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.StartUpdatedDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("startUpdatedDate", err)
		}
		filter.StartUpdatedDate = &t
	}

	if qp.EndUpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndUpdatedDate)
		if err != nil {
			return purchaseorderlineitembus.QueryFilter{}, errs.NewFieldsError("endUpdatedDate", err)
		}
		filter.EndUpdatedDate = &t
	}

	return filter, nil
}
