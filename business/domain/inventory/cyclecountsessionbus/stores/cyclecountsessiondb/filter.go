package cyclecountsessiondb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
)

func applyFilter(filter cyclecountsessionbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name ILIKE :name")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()
		wc = append(wc, "status = :status")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
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
