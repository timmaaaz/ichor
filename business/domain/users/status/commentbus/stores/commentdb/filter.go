package commentdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
)

func applyFilter(filter commentbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Comment != nil {
		data["comment"] = "%" + *filter.Comment + "%"
		wc = append(wc, "comment LIKE :comment")
	}

	if filter.CommenterID != nil {
		data["commenter_id"] = *filter.CommenterID
		wc = append(wc, "commenter_id = :commenter_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
