package pagecontentdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)

func applyFilter(filter pagecontentbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.PageConfigID != nil {
		data["page_config_id"] = *filter.PageConfigID
		wc = append(wc, "page_config_id = :page_config_id")
	}

	if filter.ContentType != nil {
		data["content_type"] = *filter.ContentType
		wc = append(wc, "content_type = :content_type")
	}

	if filter.ParentID != nil {
		data["parent_id"] = *filter.ParentID
		wc = append(wc, "parent_id = :parent_id")
	}

	if filter.IsVisible != nil {
		data["is_visible"] = *filter.IsVisible
		wc = append(wc, "is_visible = :is_visible")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
