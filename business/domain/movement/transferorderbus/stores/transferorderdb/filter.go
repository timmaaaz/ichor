package transferorderdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/movement/transferorderbus"
)

func applyFilter(filter transferorderbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.TransferID != nil {
		data["id"] = *filter.TransferID
		wc = append(wc, "id = :id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.FromLocationID != nil {
		data["from_location_id"] = *filter.FromLocationID
		wc = append(wc, "from_location_id = :from_location_id")
	}

	if filter.ToLocationID != nil {
		data["to_location_id"] = *filter.ToLocationID
		wc = append(wc, "to_location_id = :to_location_id")
	}

	if filter.RequestedByID != nil {
		data["requested_by"] = *filter.RequestedByID
		wc = append(wc, "requested_by = :requested_by")
	}

	if filter.ApprovedByID != nil {
		data["approved_by"] = *filter.ApprovedByID
		wc = append(wc, "approved_by = :approved_by")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	if filter.Status != nil {
		data["status"] = *filter.Status
		wc = append(wc, "status = :status")
	}

	if filter.TransferDate != nil {
		data["transfer_date"] = *filter.TransferDate
		wc = append(wc, "transfer_date = :transfer_date")
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
