package alertapi

import (
	"encoding/json"
	"time"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

// Alert represents the API response model for an alert.
type Alert struct {
	ID               string          `json:"id"`
	AlertType        string          `json:"alertType"`
	Severity         string          `json:"severity"`
	Title            string          `json:"title"`
	Message          string          `json:"message"`
	Context          json.RawMessage `json:"context,omitempty"`
	SourceEntityName string          `json:"sourceEntityName,omitempty"`
	SourceEntityID   string          `json:"sourceEntityId,omitempty"`
	SourceRuleID     string          `json:"sourceRuleId,omitempty"`
	Status           string          `json:"status"`
	ExpiresDate      *string         `json:"expiresDate,omitempty"`
	CreatedDate      string          `json:"createdDate"`
	UpdatedDate      string          `json:"updatedDate"`
}

// Encode implements the web.Encoder interface.
func (app Alert) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// toAppAlert converts a business alert to an API alert.
func toAppAlert(bus alertbus.Alert) Alert {
	app := Alert{
		ID:          bus.ID.String(),
		AlertType:   bus.AlertType,
		Severity:    bus.Severity,
		Title:       bus.Title,
		Message:     bus.Message,
		Context:     bus.Context,
		Status:      bus.Status,
		CreatedDate: bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate: bus.UpdatedDate.Format(time.RFC3339),
	}

	if bus.SourceEntityName != "" {
		app.SourceEntityName = bus.SourceEntityName
	}
	if bus.SourceEntityID.String() != "00000000-0000-0000-0000-000000000000" {
		app.SourceEntityID = bus.SourceEntityID.String()
	}
	if bus.SourceRuleID.String() != "00000000-0000-0000-0000-000000000000" {
		app.SourceRuleID = bus.SourceRuleID.String()
	}
	if bus.ExpiresDate != nil {
		exp := bus.ExpiresDate.Format(time.RFC3339)
		app.ExpiresDate = &exp
	}

	return app
}

// toAppAlerts converts a slice of business alerts to API alerts.
func toAppAlerts(bus []alertbus.Alert) []Alert {
	app := make([]Alert, len(bus))
	for i, v := range bus {
		app[i] = toAppAlert(v)
	}
	return app
}

// Alerts is a collection wrapper that implements the Encoder interface.
type Alerts []Alert

// Encode implements the web.Encoder interface.
func (app Alerts) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// AcknowledgeRequest represents the request body for acknowledging an alert.
type AcknowledgeRequest struct {
	Notes string `json:"notes"`
}

// Decode implements the web.Decoder interface.
func (app *AcknowledgeRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// QueryParams holds query parameters for alert queries.
type QueryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ID               string
	AlertType        string
	Severity         string
	Status           string
	SourceEntityName string
	SourceEntityID   string
	SourceRuleID     string
}
