package lottrackingsapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/lottrackingsapp"
)

func parseQueryParams(r *http.Request) (lottrackingsapp.QueryParams, error) {
	values := r.URL.Query()

	filter := lottrackingsapp.QueryParams{
		Page:              values.Get("page"),
		Rows:              values.Get("rows"),
		OrderBy:           values.Get("orderBy"),
		LotID:             values.Get("lot_id"),
		SupplierProductID: values.Get("supplier_product_id"),
		LotNumber:         values.Get("lot_number"),
		ManufactureDate:   values.Get("manufacture_date"),
		ExpirationDate:    values.Get("expiration_date"),
		ExpiryBefore:      values.Get("expiry_before"),
		ExpiryAfter:       values.Get("expiry_after"),
		RecievedDate:      values.Get("received_date"),
		Quantity:          values.Get("quantity"),
		QualityStatus:     values.Get("quality_status"),
		ProductID:         values.Get("product_id"),
		CreatedDate:       values.Get("created_date"),
		UpdatedDate:       values.Get("updated_date"),
	}

	return filter, nil
}
