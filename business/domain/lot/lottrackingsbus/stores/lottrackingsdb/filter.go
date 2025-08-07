package lottrackingsdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingsbus"
)

func applyFilter(filter lottrackingsbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.LotID != nil {
		data["id"] = *filter.LotID
		wc = append(wc, "id = :id")
	}

	if filter.SupplierProductID != nil {
		data["product_id"] = *filter.SupplierProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.ExpirationDate != nil {
		data["expiration_date"] = *filter.ExpirationDate
		wc = append(wc, "expiration_date = :expiration_date")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	if filter.LotNumber != nil {
		data["lot_number"] = *filter.LotNumber
		wc = append(wc, "lot_number = :lot_number")
	}

	if filter.ManufactureDate != nil {
		data["manufacture_date"] = *filter.ManufactureDate
		wc = append(wc, "manufacture_date = :manufacture_date")
	}

	if filter.RecievedDate != nil {
		data["recieved_date"] = *filter.RecievedDate
		wc = append(wc, "recieved_date = :recieved_date")
	}

	if filter.QualityStatus != nil {
		data["quality_status"] = *filter.QualityStatus
		wc = append(wc, "quality_status = :quality_status")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}

}
