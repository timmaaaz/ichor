package inventorytransactiondb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
)

func applyFilter(filter inventorytransactionbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.InventoryTransactionID != nil {
		data["id"] = *filter.InventoryTransactionID
		wc = append(wc, "id = :id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if filter.TransactionType != nil {
		data["transaction_type"] = *filter.TransactionType
		wc = append(wc, "transaction_type = :transaction_type")
	}

	if filter.ReferenceNumber != nil {
		data["reference_number"] = *filter.ReferenceNumber
		wc = append(wc, "reference_number = :reference_number")
	}

	if filter.TransactionDate != nil {
		data["transaction_date"] = *filter.TransactionDate
		wc = append(wc, "transaction_date = :transaction_date")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
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
