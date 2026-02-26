package lottrackingsdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
)

func applyFilter(filter lottrackingsbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.LotID != nil {
		data["id"] = *filter.LotID
		wc = append(wc, "lt.id = :id")
	}

	if filter.SupplierProductID != nil {
		data["supplier_product_id"] = *filter.SupplierProductID
		wc = append(wc, "lt.supplier_product_id = :supplier_product_id")
	}

	if filter.ExpirationDate != nil {
		data["expiration_date"] = *filter.ExpirationDate
		wc = append(wc, "lt.expiration_date = :expiration_date")
	}

	if filter.ExpirationDateBefore != nil {
		data["expiration_date_before"] = *filter.ExpirationDateBefore
		wc = append(wc, "lt.expiration_date < :expiration_date_before")
	}

	if filter.ExpirationDateAfter != nil {
		data["expiration_date_after"] = *filter.ExpirationDateAfter
		wc = append(wc, "lt.expiration_date > :expiration_date_after")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "lt.quantity = :quantity")
	}

	if filter.LotNumber != nil {
		data["lot_number"] = *filter.LotNumber
		wc = append(wc, "lt.lot_number = :lot_number")
	}

	if filter.ManufactureDate != nil {
		data["manufacture_date"] = *filter.ManufactureDate
		wc = append(wc, "lt.manufacture_date = :manufacture_date")
	}

	if filter.RecievedDate != nil {
		data["received_date"] = *filter.RecievedDate
		wc = append(wc, "lt.received_date = :received_date")
	}

	if filter.QualityStatus != nil {
		data["quality_status"] = *filter.QualityStatus
		wc = append(wc, "lt.quality_status = :quality_status")
	}

	if filter.ProductID != nil {
		data["filter_product_id"] = *filter.ProductID
		wc = append(wc, "lt.supplier_product_id IN (SELECT id FROM procurement.supplier_products WHERE product_id = :filter_product_id)")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "lt.created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "lt.updated_date = :updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}

}
