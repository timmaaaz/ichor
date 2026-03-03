package settingsbus

import (
	"encoding/json"
	"time"
)

// Setting represents a system configuration key-value pair.
// The value is stored as raw JSON to support any JSON type (string, number, object).
type Setting struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
	CreatedDate time.Time       `json:"created_date"`
	UpdatedDate time.Time       `json:"updated_date"`
}

// NewSetting contains fields required to create a new setting.
type NewSetting struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
}

// UpdateSetting contains fields for updating an existing setting.
type UpdateSetting struct {
	Value       json.RawMessage `json:"value,omitempty"`
	Description *string         `json:"description,omitempty"`
}
