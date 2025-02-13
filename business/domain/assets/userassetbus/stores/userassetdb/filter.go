package userassetdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/assets/userassetbus"
)

func applyFilter(filter userassetbus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["user_asset_id"] = *filter.ID
		wc = append(wc, "user_asset_id = :user_asset_id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.AssetID != nil {
		data["asset_id"] = *filter.AssetID
		wc = append(wc, "asset_id = :asset_id")
	}

	if filter.ApprovalStatusID != nil {
		data["approval_status_id"] = *filter.ApprovalStatusID
		wc = append(wc, "approval_status_id = :approval_status_id")
	}

	if filter.ApprovedBy != nil {
		data["approved_by"] = *filter.ApprovedBy
		wc = append(wc, "approved_by = :approved_by")
	}

	if filter.ConditionID != nil {
		data["condition_id"] = *filter.ConditionID
		wc = append(wc, "condition_id = :condition_id")
	}

	if filter.DateReceived != nil {
		data["date_received"] = *filter.DateReceived
		wc = append(wc, "date_received = :date_received")
	}

	if filter.FulfillmentStatusID != nil {
		data["fulfillment_status_id"] = *filter.FulfillmentStatusID
		wc = append(wc, "fulfillment_status_id = :fulfillment_status_id")
	}

	if filter.LastMaintenance != nil {
		data["last_maintenance"] = *filter.LastMaintenance
		wc = append(wc, "last_maintenance = :last_maintenance")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
