package pagedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
)

func applyFilter(filter pagebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Path != nil {
		data["path"] = "%" + *filter.Path + "%"
		wc = append(wc, "path ILIKE :path")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.Module != nil {
		data["module"] = *filter.Module
		wc = append(wc, "module = :module")
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
