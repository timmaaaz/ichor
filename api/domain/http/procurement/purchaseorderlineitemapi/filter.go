package purchaseorderlineitemapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/procurement/purchaseorderlineitemapp"
)

func parseQueryParams(r *http.Request) (purchaseorderlineitemapp.QueryParams, error) {
	values := r.URL.Query()

	qp := purchaseorderlineitemapp.QueryParams{
		Page:                      values.Get("page"),
		Rows:                      values.Get("rows"),
		OrderBy:                   values.Get("orderBy"),
		ID:                        values.Get("id"),
		PurchaseOrderID:           values.Get("purchaseOrderId"),
		SupplierProductID:         values.Get("supplierProductId"),
		LineItemStatusID:          values.Get("lineItemStatusId"),
		CreatedBy:                 values.Get("createdBy"),
		UpdatedBy:                 values.Get("updatedBy"),
		StartExpectedDeliveryDate: values.Get("startExpectedDeliveryDate"),
		EndExpectedDeliveryDate:   values.Get("endExpectedDeliveryDate"),
		StartActualDeliveryDate:   values.Get("startActualDeliveryDate"),
		EndActualDeliveryDate:     values.Get("endActualDeliveryDate"),
		StartCreatedDate:          values.Get("startCreatedDate"),
		EndCreatedDate:            values.Get("endCreatedDate"),
		StartUpdatedDate:          values.Get("startUpdatedDate"),
		EndUpdatedDate:            values.Get("endUpdatedDate"),
	}

	return qp, nil
}
