package ordersdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/order/ordersbus"
)

// TODO: Switch these over to use string.builder?

func applyFilter(filter ordersbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Number != nil {
		data["number"] = *filter.Number
		wc = append(wc, "number ILIKE :number")
	}

	if filter.CustomerID != nil {
		data["customer_id"] = *filter.CustomerID
		wc = append(wc, "customer_id = :customer_id")
	}

	if filter.FulfillmentStatusID != nil {
		data["fulfillment_status_id"] = *filter.FulfillmentStatusID
		wc = append(wc, "fulfillment_status_id = :fulfillment_status_id")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.UpdatedBy != nil {
		data["updated_by"] = *filter.UpdatedBy
		wc = append(wc, "updated_by = :updated_by")
	}

	if filter.StartDueDate != nil {
		data["start_due_date"] = *filter.StartDueDate
		wc = append(wc, "due_date >= :start_due_date")
	}

	if filter.EndDueDate != nil {
		data["end_due_date"] = *filter.EndDueDate
		wc = append(wc, "due_date <= :end_due_date")
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
