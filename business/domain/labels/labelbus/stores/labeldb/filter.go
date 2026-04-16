package labeldb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
)

func applyFilter(filter labelbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Code != nil {
		data["code"] = *filter.Code
		wc = append(wc, "code = :code")
	}

	if filter.Type != nil {
		data["type"] = *filter.Type
		wc = append(wc, "type = :type")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
