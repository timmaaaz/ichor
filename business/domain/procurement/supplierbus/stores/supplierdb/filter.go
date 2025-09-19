package supplierdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
)

func applyFilter(filter supplierbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.SupplierID != nil {
		data["id"] = *filter.SupplierID
		wc = append(wc, "id = :id")
	}

	if filter.ContactInfosID != nil {
		data["contact_infos_id"] = *filter.ContactInfosID
		wc = append(wc, "contact_infos_id = :contact_infos_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.LeadTimeDays != nil {
		data["lead_time_days"] = *filter.LeadTimeDays
		wc = append(wc, "lead_time_days = :lead_time_days")
	}

	if filter.Rating != nil {
		data["rating"] = *filter.Rating
		wc = append(wc, "rating = :rating")
	}

	if filter.IsActive != nil {
		data["is_active"] = *filter.IsActive
		wc = append(wc, "is_active = :is_active")
	}

	if filter.PaymentTerms != nil {
		data["payment_terms"] = *filter.PaymentTerms
		wc = append(wc, "payment_terms = :payment_terms")
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
