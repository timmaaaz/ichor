package purchaseorderdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
)

func applyFilter(filter purchaseorderbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.OrderNumber != nil {
		data["order_number"] = "%" + *filter.OrderNumber + "%"
		wc = append(wc, "order_number ILIKE :order_number")
	}

	if filter.SupplierID != nil {
		data["supplier_id"] = *filter.SupplierID
		wc = append(wc, "supplier_id = :supplier_id")
	}

	if filter.PurchaseOrderStatusID != nil {
		data["purchase_order_status_id"] = *filter.PurchaseOrderStatusID
		wc = append(wc, "purchase_order_status_id = :purchase_order_status_id")
	}

	if filter.DeliveryWarehouseID != nil {
		data["delivery_warehouse_id"] = *filter.DeliveryWarehouseID
		wc = append(wc, "delivery_warehouse_id = :delivery_warehouse_id")
	}

	if filter.RequestedBy != nil {
		data["requested_by"] = *filter.RequestedBy
		wc = append(wc, "requested_by = :requested_by")
	}

	if filter.ApprovedBy != nil {
		data["approved_by"] = *filter.ApprovedBy
		wc = append(wc, "approved_by = :approved_by")
	}

	if filter.StartOrderDate != nil {
		data["start_order_date"] = *filter.StartOrderDate
		wc = append(wc, "order_date >= :start_order_date")
	}

	if filter.EndOrderDate != nil {
		data["end_order_date"] = *filter.EndOrderDate
		wc = append(wc, "order_date <= :end_order_date")
	}

	if filter.StartExpectedDelivery != nil {
		data["start_expected_delivery"] = *filter.StartExpectedDelivery
		wc = append(wc, "expected_delivery_date >= :start_expected_delivery")
	}

	if filter.EndExpectedDelivery != nil {
		data["end_expected_delivery"] = *filter.EndExpectedDelivery
		wc = append(wc, "expected_delivery_date <= :end_expected_delivery")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}