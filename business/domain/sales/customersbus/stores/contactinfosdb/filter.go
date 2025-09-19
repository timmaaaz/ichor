package customersdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
)

// type QueryFilter struct {
// 	ID                *uuid.UUID
// 	ContactID         *uuid.UUID
// 	DeliveryAddressID *uuid.UUID
// 	Notes             *string
// 	CreatedBy         *uuid.UUID
// 	UpdatedBy         *uuid.UUID

// 	// DateFilters
// 	StartCreatedDate *time.Time
// 	EndCreatedDate   *time.Time
// 	StartUpdatedDate *time.Time
// 	EndUpdatedDate   *time.Time
// }

func applyFilter(filter customersbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = *filter.Name
		wc = append(wc, "name = :name")
	}

	if filter.ContactID != nil {
		data["contact_id"] = *filter.ContactID
		wc = append(wc, "contact_id = :contact_id")
	}

	if filter.DeliveryAddressID != nil {
		data["delivery_address_id"] = *filter.DeliveryAddressID
		wc = append(wc, "delivery_address_id = :delivery_address_id")
	}

	if filter.Notes != nil {
		data["notes"] = *filter.Notes
		wc = append(wc, "notes ILIKE :notes")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.UpdatedBy != nil {
		data["updated_by"] = *filter.UpdatedBy
		wc = append(wc, "updated_by = :updated_by")
	}

	// Date range filters
	if filter.StartCreatedDate != nil {
		data["start_created_date"] = *filter.StartCreatedDate
		wc = append(wc, "created_at >= :start_created_date")
	}

	if filter.EndCreatedDate != nil {
		data["end_created_date"] = *filter.EndCreatedDate
		wc = append(wc, "created_at <= :end_created_date")
	}

	if filter.StartUpdatedDate != nil {
		data["start_updated_date"] = *filter.StartUpdatedDate
		wc = append(wc, "updated_at >= :start_updated_date")
	}

	if filter.EndUpdatedDate != nil {
		data["end_updated_date"] = *filter.EndUpdatedDate
		wc = append(wc, "updated_at <= :end_updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
