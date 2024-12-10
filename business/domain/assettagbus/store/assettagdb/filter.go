package assettagdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assettagbus"
)

func applyFilter(filter assettagbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["asset_tag_id"] = *filter.ID
		wc = append(wc, "asset_tag_id = :asset_tag_id")
	}

	if filter.AssetID != nil {
		data["asset_id"] = *filter.AssetID
		wc = append(wc, "asset_id = :asset_id")
	}

	if filter.TagID != nil {
		data["tag_id"] = *filter.TagID
		wc = append(wc, "tag_id = :tag_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
