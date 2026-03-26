package notificationapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the raw query parameters from the HTTP request.
type QueryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ID               string
	IsRead           string
	Priority         string
	SourceEntityName string
	SourceEntityID   string
}

// =============================================================================
// Response model
// =============================================================================

// Notification is the app-layer response model.
type Notification struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	Title            string `json:"title"`
	Message          string `json:"message"`
	Priority         string `json:"priority"`
	IsRead           bool   `json:"isRead"`
	ReadDate         string `json:"readDate"`
	SourceEntityName string `json:"sourceEntityName"`
	SourceEntityID   string `json:"sourceEntityId"`
	ActionURL        string `json:"actionUrl"`
	CreatedDate      string `json:"createdDate"`
}

// Encode implements web.Encoder.
func (app Notification) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Notifications is a slice of Notification for list responses.
type Notifications []Notification

// Encode implements web.Encoder.
func (app Notifications) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppNotification converts a bus model to an app-layer response model.
func ToAppNotification(bus notificationbus.Notification) Notification {
	readDate := ""
	if bus.ReadDate != nil {
		readDate = bus.ReadDate.Format(timeutil.FORMAT)
	}

	sourceEntityID := ""
	if bus.SourceEntityID != uuid.Nil {
		sourceEntityID = bus.SourceEntityID.String()
	}

	return Notification{
		ID:               bus.ID.String(),
		UserID:           bus.UserID.String(),
		Title:            bus.Title,
		Message:          bus.Message,
		Priority:         bus.Priority,
		IsRead:           bus.IsRead,
		ReadDate:         readDate,
		SourceEntityName: bus.SourceEntityName,
		SourceEntityID:   sourceEntityID,
		ActionURL:        bus.ActionURL,
		CreatedDate:      bus.CreatedDate.Format(timeutil.FORMAT),
	}
}

// ToAppNotifications converts a slice of bus models to app-layer response models.
func ToAppNotifications(bus []notificationbus.Notification) []Notification {
	app := make([]Notification, len(bus))
	for i, v := range bus {
		app[i] = ToAppNotification(v)
	}
	return app
}
