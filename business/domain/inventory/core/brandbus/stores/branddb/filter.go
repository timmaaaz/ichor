package branddb

import (
	"bytes"
	"strings"

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

	if filter.ContactInfoID != nil {
		data["contact_info_id"] = *filter.ContactInfoID
		wc = append(wc, "contact_info_id = :contact_info_id")
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
