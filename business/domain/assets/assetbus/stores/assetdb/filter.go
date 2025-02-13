package assetdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
)

func applyFilter(filter assetbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["asset_id"] = *filter.ID
		wc = append(wc, "asset_id = :asset_id")
	}

	if filter.LastMaintenance != nil {
		data["last_maintenance_time"] = *filter.LastMaintenance
		wc = append(wc, "last_maintenance_time = :last_maintenance_time")
	}

	if filter.SerialNumber != nil {
		data["serial_number"] = *filter.SerialNumber
		wc = append(wc, "serial_number = :serial_number")
	}

	if filter.ValidAssetID != nil {
		data["valid_asset_id"] = *filter.ValidAssetID
		wc = append(wc, "valid_asset_id = :valid_asset_id")
	}

	if filter.AssetConditionID != nil {
		data["asset_condition_id"] = filter.AssetConditionID
		wc = append(wc, "asset_condition_id = :asset_condition_id")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
