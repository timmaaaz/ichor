package purchaseorderlineitemdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
)

func applyFilter(filter purchaseorderlineitembus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.PurchaseOrderID != nil {
		data["purchase_order_id"] = *filter.PurchaseOrderID
		wc = append(wc, "purchase_order_id = :purchase_order_id")
	}

	if filter.SupplierProductID != nil {
		data["supplier_product_id"] = *filter.SupplierProductID
		wc = append(wc, "supplier_product_id = :supplier_product_id")
	}

	if filter.LineItemStatusID != nil {
		data["line_item_status_id"] = *filter.LineItemStatusID
		wc = append(wc, "line_item_status_id = :line_item_status_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}