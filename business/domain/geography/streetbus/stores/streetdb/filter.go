package streetdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
)

func applyFilter(filter streetbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.CityID != nil {
		data["city_id"] = *filter.CityID
		wc = append(wc, "city_id = :city_id")
	}

	if filter.Line1 != nil {
		data["line_1"] = "%" + *filter.Line1 + "%"
		wc = append(wc, "line_1 ILIKE :line_1")
	}

	if filter.Line2 != nil {
		data["line_2"] = "%" + *filter.Line2 + "%"
		wc = append(wc, "line_2 ILIKE :line_2")
	}

	if filter.PostalCode != nil {
		data["postal_code"] = "%" + *filter.PostalCode + "%"
		wc = append(wc, "postal_code ILIKE :postal_code")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
