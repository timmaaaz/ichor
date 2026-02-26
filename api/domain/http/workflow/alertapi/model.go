package alertapi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

// Alert represents the API response model for an alert.
type Alert struct {
	ID                  string                    `json:"id"`
	AlertType           string                    `json:"alertType"`
	Severity            string                    `json:"severity"`
	Title               string                    `json:"title"`
	Message             string                    `json:"message"`
	Context             json.RawMessage           `json:"context,omitempty"`
	SourceEntityName    string                    `json:"sourceEntityName,omitempty"`
	SourceEntityID      string                    `json:"sourceEntityId,omitempty"`
	SourceRuleID        string                    `json:"sourceRuleId,omitempty"`
	SourceRuleName      string                    `json:"sourceRuleName,omitempty"`
	Status              string                    `json:"status"`
	ExpiresDate         *string                   `json:"expiresDate,omitempty"`
	CreatedDate         string                    `json:"createdDate"`
	UpdatedDate         string                    `json:"updatedDate"`
	Recipients          []AlertRecipientVM        `json:"recipients,omitempty"`
	Acknowledgments     []AlertAcknowledgmentVM   `json:"acknowledgments,omitempty"`
}

// AlertRecipientVM represents an enriched alert recipient with human-readable names.
type AlertRecipientVM struct {
	RecipientType string `json:"recipientType"`
	RecipientID   string `json:"recipientId"`
	Name          string `json:"name"`
	Email         string `json:"email,omitempty"`
}

// AlertAcknowledgmentVM represents an acknowledgment record with acknowledger details.
type AlertAcknowledgmentVM struct {
	ID               string `json:"id"`
	AcknowledgedBy   string `json:"acknowledgedBy"`
	AcknowledgerName string `json:"acknowledgerName,omitempty"`
	AcknowledgedDate string `json:"acknowledgedDate"`
	Notes            string `json:"notes,omitempty"`
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
	if bus.SourceRuleName != "" {
		app.SourceRuleName = bus.SourceRuleName
	}
	if bus.ExpiresDate != nil {
		exp := bus.ExpiresDate.Format(time.RFC3339)
		app.ExpiresDate = &exp
	}

	return app
}

// toAppAcknowledgment converts a business acknowledgment to an API view model.
func toAppAcknowledgment(bus alertbus.AlertAcknowledgment) AlertAcknowledgmentVM {
	vm := AlertAcknowledgmentVM{
		ID:               bus.ID.String(),
		AcknowledgedBy:   bus.AcknowledgedBy.String(),
		AcknowledgedDate: bus.AcknowledgedDate.Format(time.RFC3339),
	}
	if bus.AcknowledgerName != "" {
		vm.AcknowledgerName = bus.AcknowledgerName
	}
	if bus.Notes != "" {
		vm.Notes = bus.Notes
	}
	return vm
}

// toAppAcknowledgments converts a slice of business acknowledgments to API view models.
func toAppAcknowledgments(bus []alertbus.AlertAcknowledgment) []AlertAcknowledgmentVM {
	vms := make([]AlertAcknowledgmentVM, len(bus))
	for i, a := range bus {
		vms[i] = toAppAcknowledgment(a)
	}
	return vms
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
	CreatedAfter     string
	CreatedBefore    string
}

// BulkSelectedRequest represents the request body for bulk operations by IDs.
type BulkSelectedRequest struct {
	IDs   []string `json:"ids" validate:"required,min=1"`
	Notes string   `json:"notes"`
}

// Decode implements the web.Decoder interface.
func (app *BulkSelectedRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the request.
func (app BulkSelectedRequest) Validate() error {
	if len(app.IDs) == 0 {
		return fmt.Errorf("ids is required and must have at least one element")
	}
	return nil
}

// BulkAllRequest represents the request body for bulk all operations.
type BulkAllRequest struct {
	Notes string `json:"notes"`
}

// Decode implements the web.Decoder interface.
func (app *BulkAllRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// BulkActionResult represents the response for bulk operations.
type BulkActionResult struct {
	Count   int `json:"count"`
	Skipped int `json:"skipped"`
}

// Encode implements the web.Encoder interface.
func (app BulkActionResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}
