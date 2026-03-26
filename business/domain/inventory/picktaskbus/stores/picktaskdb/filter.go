package picktaskdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
)

func applyFilter(filter picktaskbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.SalesOrderID != nil {
		data["sales_order_id"] = *filter.SalesOrderID
		wc = append(wc, "sales_order_id = :sales_order_id")
	}

	if filter.SalesOrderLineItemID != nil {
		data["sales_order_line_item_id"] = *filter.SalesOrderLineItemID
		wc = append(wc, "sales_order_line_item_id = :sales_order_line_item_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()
		wc = append(wc, "status = :status")
	}

	if filter.AssignedTo != nil {
		data["assigned_to"] = *filter.AssignedTo
		wc = append(wc, "assigned_to = :assigned_to")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
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
