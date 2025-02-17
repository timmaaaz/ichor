package branddb

import (
	"bytes"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
)

func applyFilter(filter brandbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["brand_id"] = *filter.ID
		wc = append(wc, "brand_id = :brand_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.ContactInfo != nil {
		data["contact_info"] = *filter.ContactInfo
		wc = append(wc, "contact_info = :contact_info")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}
}
