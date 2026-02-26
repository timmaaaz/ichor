package notificationsapi

import "encoding/json"

// NotificationSummary is the top-level response for the notifications summary endpoint.
type NotificationSummary struct {
	Alerts    AlertSummary    `json:"alerts"`
	Approvals ApprovalSummary `json:"approvals"`
}

// Encode implements web.Encoder.
func (n NotificationSummary) Encode() ([]byte, string, error) {
	data, err := json.Marshal(n)
	return data, "application/json", err
}

// AlertSummary provides a count breakdown of active alerts by severity.
type AlertSummary struct {
	TotalActive int `json:"totalActive"`
	Critical    int `json:"critical"`
	High        int `json:"high"`
	Medium      int `json:"medium"`
	Low         int `json:"low"`
	Info        int `json:"info"`
}

// ApprovalSummary provides a count of pending approvals awaiting the user's action.
type ApprovalSummary struct {
	PendingCount int `json:"pendingCount"`
}
