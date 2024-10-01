package citydb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/location/citybus"
)

func applyFilter(filter citybus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["city_id"] = *filter.ID
		wc = append(wc, "city_id = :city_id")
	}

	if filter.RegionID != nil {
		data["region_id"] = *filter.RegionID
		wc = append(wc, "region_id = :region_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
