package regiondb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
)

func applyFilter(filter regionbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["region_id"] = *filter.ID
		wc = append(wc, "region_id = :region_id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Code != nil {
		data["code"] = *filter.Code
		wc = append(wc, "code = :code")
	}

	if filter.CountryID != nil {
		data["country_id"] = *filter.CountryID
		wc = append(wc, "country_id = :country_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
