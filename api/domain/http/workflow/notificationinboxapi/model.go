package notificationinboxapi

import "encoding/json"

// UnreadCount is the response for the count endpoint.
type UnreadCount struct {
	Count int `json:"count"`
}

// Encode implements web.Encoder.
func (u UnreadCount) Encode() ([]byte, string, error) {
	data, err := json.Marshal(u)
	return data, "application/json", err
}

// MarkAllReadResult is the response for the mark-all-read endpoint.
type MarkAllReadResult struct {
	Count int `json:"count"`
}

// Encode implements web.Encoder.
func (m MarkAllReadResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(m)
	return data, "application/json", err
}

// SuccessResult is a generic success response.
type SuccessResult struct {
	Success bool `json:"success"`
}

// Encode implements web.Encoder.
func (s SuccessResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(s)
	return data, "application/json", err
}
