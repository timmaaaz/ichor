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

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.UpdatedBy != nil {
		data["updated_by"] = *filter.UpdatedBy
		wc = append(wc, "updated_by = :updated_by")
	}

	if filter.StartExpectedDeliveryDate != nil {
		data["start_expected_delivery_date"] = *filter.StartExpectedDeliveryDate
		wc = append(wc, "expected_delivery_date >= :start_expected_delivery_date")
	}

	if filter.EndExpectedDeliveryDate != nil {
		data["end_expected_delivery_date"] = *filter.EndExpectedDeliveryDate
		wc = append(wc, "expected_delivery_date <= :end_expected_delivery_date")
	}

	if filter.StartActualDeliveryDate != nil {
		data["start_actual_delivery_date"] = *filter.StartActualDeliveryDate
		wc = append(wc, "actual_delivery_date >= :start_actual_delivery_date")
	}

	if filter.EndActualDeliveryDate != nil {
		data["end_actual_delivery_date"] = *filter.EndActualDeliveryDate
		wc = append(wc, "actual_delivery_date <= :end_actual_delivery_date")
	}

	if filter.StartCreatedDate != nil {
		data["start_created_date"] = *filter.StartCreatedDate
		wc = append(wc, "created_date >= :start_created_date")
	}

	if filter.EndCreatedDate != nil {
		data["end_created_date"] = *filter.EndCreatedDate
		wc = append(wc, "created_date <= :end_created_date")
	}

	if filter.StartUpdatedDate != nil {
		data["start_updated_date"] = *filter.StartUpdatedDate
		wc = append(wc, "updated_date >= :start_updated_date")
	}

	if filter.EndUpdatedDate != nil {
		data["end_updated_date"] = *filter.EndUpdatedDate
		wc = append(wc, "updated_date <= :end_updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}