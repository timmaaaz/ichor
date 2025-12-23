package timezonedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
)

func applyFilter(filter timezonebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.DisplayName != nil {
		data["display_name"] = "%" + *filter.DisplayName + "%"
		wc = append(wc, "display_name ILIKE :display_name")
	}

	if filter.IsActive != nil {
		data["is_active"] = *filter.IsActive
		wc = append(wc, "is_active = :is_active")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
