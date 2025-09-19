package homedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/hr/homebus"
)

func (s *Store) applyFilter(filter homebus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.Type != nil {
		data["type"] = filter.Type.String()
		wc = append(wc, "type = :type")
	}

	if filter.StartCreatedDate != nil {
		data["start_created_date"] = filter.StartCreatedDate.UTC()
		wc = append(wc, "created_date >= :start_created_date")
	}

	if filter.EndCreatedDate != nil {
		data["end_created_date"] = filter.EndCreatedDate.UTC()
		wc = append(wc, "created_date <= :end_created_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
