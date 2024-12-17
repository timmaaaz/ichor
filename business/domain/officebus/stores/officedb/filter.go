package officedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/officebus"
)

func applyFilter(filter officebus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["office_id"] = *filter.ID
		wc = append(wc, "office_id = :office_id")
	}

	if filter.StreetID != nil {
		data["street_id"] = *filter.ID
		wc = append(wc, "street_id = :street_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
