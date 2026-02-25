package purchaseorderapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (purchaseorderbus.QueryFilter, error) {
	var filter purchaseorderbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.OrderNumber != "" {
		filter.OrderNumber = &qp.OrderNumber
	}

	if qp.SupplierID != "" {
		id, err := uuid.Parse(qp.SupplierID)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("supplierId", err)
		}
		filter.SupplierID = &id
	}

	if qp.PurchaseOrderStatusID != "" {
		id, err := uuid.Parse(qp.PurchaseOrderStatusID)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("purchaseOrderStatusId", err)
		}
		filter.PurchaseOrderStatusID = &id
	}

	if qp.DeliveryWarehouseID != "" {
		id, err := uuid.Parse(qp.DeliveryWarehouseID)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("deliveryWarehouseId", err)
		}
		filter.DeliveryWarehouseID = &id
	}

	if qp.RequestedBy != "" {
		id, err := uuid.Parse(qp.RequestedBy)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("requestedBy", err)
		}
		filter.RequestedBy = &id
	}

	if qp.ApprovedBy != "" {
		id, err := uuid.Parse(qp.ApprovedBy)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("approvedBy", err)
		}
		filter.ApprovedBy = &id
	}

	if qp.StartOrderDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.StartOrderDate)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("startOrderDate", err)
		}
		filter.StartOrderDate = &t
	}

	if qp.EndOrderDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndOrderDate)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("endOrderDate", err)
		}
		filter.EndOrderDate = &t
	}

	if qp.StartExpectedDelivery != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.StartExpectedDelivery)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("startExpectedDelivery", err)
		}
		filter.StartExpectedDelivery = &t
	}

	if qp.EndExpectedDelivery != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndExpectedDelivery)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("endExpectedDelivery", err)
		}
		filter.EndExpectedDelivery = &t
	}

	if qp.StartActualDeliveryDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.StartActualDeliveryDate)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("startActualDeliveryDate", err)
		}
		filter.StartActualDeliveryDate = &t
	}

	if qp.EndActualDeliveryDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.EndActualDeliveryDate)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("endActualDeliveryDate", err)
		}
		filter.EndActualDeliveryDate = &t
	}

	if qp.IsUndelivered != "" {
		b, err := strconv.ParseBool(qp.IsUndelivered)
		if err != nil {
			return purchaseorderbus.QueryFilter{}, errs.NewFieldsError("isUndelivered", err)
		}
		filter.IsUndelivered = &b
	}

	return filter, nil
}
