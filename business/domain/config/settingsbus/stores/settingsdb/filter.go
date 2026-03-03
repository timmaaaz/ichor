package settingsdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
)

func applyFilter(filter settingsbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.Key != nil {
		data["key"] = *filter.Key
		wc = append(wc, "key = :key")
	}

	if filter.Prefix != nil {
		data["prefix"] = *filter.Prefix + "%"
		wc = append(wc, "key LIKE :prefix")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
