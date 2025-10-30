package pageactiondb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
)

func applyFilter(filter pageactionbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.PageConfigID != nil {
		data["page_config_id"] = *filter.PageConfigID
		wc = append(wc, "page_config_id = :page_config_id")
	}

	if filter.ActionType != nil {
		data["action_type"] = string(*filter.ActionType)
		wc = append(wc, "action_type = :action_type")
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
