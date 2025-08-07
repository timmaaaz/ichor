package reportstodb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/users/reportstobus"
)

func applyFilter(filter reportstobus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.BossID != nil {
		data["boss_id"] = *filter.BossID
		wc = append(wc, "boss_id = :boss_id")
	}

	if filter.ReporterID != nil {
		data["reporter_id"] = *filter.ReporterID
		wc = append(wc, "reporter_id = :reporter_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
