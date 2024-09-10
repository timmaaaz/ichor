package countrydb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
)

func applyFilter(filter countrybus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["country_id"] = *filter.ID
		wc = append(wc, "country_id = :country_id")
	}

	if filter.Number != nil {
		data["number"] = *filter.Number
		wc = append(wc, "number = :number")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Alpha2 != nil {
		data["alpha_2"] = *filter.Alpha2
		wc = append(wc, "alpha_2 = :alpha_2")
	}

	if filter.Alpha3 != nil {
		data["alpha_3"] = *filter.Alpha3
		wc = append(wc, "alpha_3 = :alpha_3")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
