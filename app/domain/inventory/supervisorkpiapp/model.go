package supervisorkpiapp

import "encoding/json"

// KPIs represents the aggregated supervisor dashboard counts.
type KPIs struct {
	PendingApprovals    int `json:"pending_approvals"`
	PendingAdjustments  int `json:"pending_adjustments"`
	PendingTransfers    int `json:"pending_transfers"`
	OpenInspections     int `json:"open_inspections"`
	PendingPutAwayTasks int `json:"pending_put_away_tasks"`
	ActiveAlerts        int `json:"active_alerts"`
}

// Encode implements the web.Encoder interface.
func (k KPIs) Encode() ([]byte, string, error) {
	data, err := json.Marshal(k)
	return data, "application/json", err
}
