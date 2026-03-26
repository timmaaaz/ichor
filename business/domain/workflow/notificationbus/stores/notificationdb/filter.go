package notificationdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
)

func applyFilter(filter notificationbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.IsRead != nil {
		data["is_read"] = *filter.IsRead
		wc = append(wc, "is_read = :is_read")
	}

	if filter.Priority != nil {
		data["priority"] = *filter.Priority
		wc = append(wc, "priority = :priority")
	}

	if filter.SourceEntityName != nil {
		data["source_entity_name"] = *filter.SourceEntityName
		wc = append(wc, "source_entity_name = :source_entity_name")
	}

	if filter.SourceEntityID != nil {
		data["source_entity_id"] = *filter.SourceEntityID
		wc = append(wc, "source_entity_id = :source_entity_id")
	}

	if filter.CreatedAfter != nil {
		data["created_after"] = *filter.CreatedAfter
		wc = append(wc, "created_date >= :created_after")
	}

	if filter.CreatedBefore != nil {
		data["created_before"] = *filter.CreatedBefore
		wc = append(wc, "created_date <= :created_before")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
