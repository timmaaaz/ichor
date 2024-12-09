package titledb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/titlebus"
)

func applyFilter(filter titlebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["title_id"] = *filter.ID
		wc = append(wc, "title_id = :title_id")
	}

	if filter.Description != nil {
		data["description"] = "%" + *filter.Description + "%"
		wc = append(wc, "description = :description")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name LIKE :name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
