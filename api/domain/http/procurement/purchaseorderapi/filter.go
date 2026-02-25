package purchaseorderapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderapp"
)

func parseQueryParams(r *http.Request) (purchaseorderapp.QueryParams, error) {
	values := r.URL.Query()

	qp := purchaseorderapp.QueryParams{
		Page:                  values.Get("page"),
		Rows:                  values.Get("rows"),
		OrderBy:               values.Get("orderBy"),
		ID:                    values.Get("id"),
		OrderNumber:           values.Get("orderNumber"),
		SupplierID:            values.Get("supplierId"),
		PurchaseOrderStatusID: values.Get("purchaseOrderStatusId"),
		DeliveryWarehouseID:   values.Get("deliveryWarehouseId"),
		RequestedBy:           values.Get("requestedBy"),
		ApprovedBy:            values.Get("approvedBy"),
		StartOrderDate:          values.Get("startOrderDate"),
		EndOrderDate:            values.Get("endOrderDate"),
		StartExpectedDelivery:   values.Get("startExpectedDelivery"),
		EndExpectedDelivery:     values.Get("endExpectedDelivery"),
		StartActualDeliveryDate: values.Get("startActualDeliveryDate"),
		EndActualDeliveryDate:   values.Get("endActualDeliveryDate"),
		IsUndelivered:           values.Get("isUndelivered"),
	}

	return qp, nil
}
