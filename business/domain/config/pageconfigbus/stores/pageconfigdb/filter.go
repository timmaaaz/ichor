package pageconfigdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
)

func applyFilter(filter pageconfigbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = *filter.Name
		wc = append(wc, "name = :name")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.IsDefault != nil {
		data["is_default"] = *filter.IsDefault
		wc = append(wc, "is_default = :is_default")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
