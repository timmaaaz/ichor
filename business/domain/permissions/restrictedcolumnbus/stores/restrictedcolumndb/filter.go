package restrictedcolumndb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/permissions/restrictedcolumnbus"
)

func applyFilter(filter restrictedcolumnbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["restricted_column_id"] = *filter.ID
		wc = append(wc, "restricted_column_id = :restricted_column_id")
	}

	if filter.TableName != nil {
		data["table_name"] = *filter.TableName
		wc = append(wc, "table_name ILIKE :table_name")
	}

	if filter.ColumnName != nil {
		data["column_name"] = *filter.ColumnName
		wc = append(wc, "column_name ILIKE :column_name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
