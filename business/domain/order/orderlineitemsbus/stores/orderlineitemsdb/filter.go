package orderlineitemsdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/order/orderlineitemsbus"
)

// TODO: Switch these over to use string.builder?

func applyFilter(filter orderlineitemsbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.OrderID != nil {
		data["order_id"] = *filter.OrderID
		wc = append(wc, "order_id = :order_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	if filter.Discount != nil {
		data["discount"] = *filter.Discount
		wc = append(wc, "discount = :discount")
	}

	if filter.LineItemFulfillmentStatusesID != nil {
		data["line_item_fulfillment_statuses_id"] = *filter.LineItemFulfillmentStatusesID
		wc = append(wc, "line_item_fulfillment_statuses_id = :line_item_fulfillment_statuses_id")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.UpdatedBy != nil {
		data["updated_by"] = *filter.UpdatedBy
		wc = append(wc, "updated_by = :updated_by")
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
