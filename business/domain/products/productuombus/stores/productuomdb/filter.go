package productuomdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
)

func applyFilter(filter productuombus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.IsBase != nil {
		data["is_base"] = *filter.IsBase
		wc = append(wc, "is_base = :is_base")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name LIKE :name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
