package alert_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/alertapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// AlertSeedData holds alert-specific test data.
type AlertSeedData struct {
	apitest.SeedData
	Alerts           []alertapi.Alert
	AlertIDs         []uuid.UUID
	NonRecipientID   uuid.UUID
	LowSeverityID    uuid.UUID
	MediumSeverityID uuid.UUID
	HighSeverityID   uuid.UUID
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (AlertSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Create regular user
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return AlertSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	// Create admin user
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return AlertSeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	now := time.Now().UTC().Truncate(time.Second)
	contextData, _ := json.Marshal(map[string]any{"test": "data"})

	// Create alerts with different severities
	var alerts []alertbus.Alert
	var alertIDs []uuid.UUID
	var appAlerts []alertapi.Alert

	// Alert 1: low severity, active, user is direct recipient
	alert1ID := uuid.New()
	alert1 := alertbus.Alert{
		ID:          alert1ID,
		AlertType:   "test_alert",
		Severity:    alertbus.SeverityLow,
		Title:       "Low Severity Alert",
		Message:     "This is a low severity test alert",
		Context:     contextData,
		Status:      alertbus.StatusActive,
		CreatedDate: now,
		UpdatedDate: now,
	}
	alerts = append(alerts, alert1)
	alertIDs = append(alertIDs, alert1ID)

	// Alert 2: medium severity, active, user is direct recipient
	alert2ID := uuid.New()
	alert2 := alertbus.Alert{
		ID:          alert2ID,
		AlertType:   "test_alert",
		Severity:    alertbus.SeverityMedium,
		Title:       "Medium Severity Alert",
		Message:     "This is a medium severity test alert",
		Context:     contextData,
		Status:      alertbus.StatusActive,
		CreatedDate: now.Add(1 * time.Second),
		UpdatedDate: now.Add(1 * time.Second),
	}
	alerts = append(alerts, alert2)
	alertIDs = append(alertIDs, alert2ID)

	// Alert 3: high severity, active, user is direct recipient
	alert3ID := uuid.New()
	alert3 := alertbus.Alert{
		ID:          alert3ID,
		AlertType:   "test_alert",
		Severity:    alertbus.SeverityHigh,
		Title:       "High Severity Alert",
		Message:     "This is a high severity test alert",
		Context:     contextData,
		Status:      alertbus.StatusActive,
		CreatedDate: now.Add(2 * time.Second),
		UpdatedDate: now.Add(2 * time.Second),
	}
	alerts = append(alerts, alert3)
	alertIDs = append(alertIDs, alert3ID)

	// Alert 4: critical severity, active, user is NOT a recipient (for skip testing)
	alert4ID := uuid.New()
	alert4 := alertbus.Alert{
		ID:          alert4ID,
		AlertType:   "test_alert",
		Severity:    alertbus.SeverityCritical,
		Title:       "Critical Severity Alert - Not Recipient",
		Message:     "This user is not a recipient of this alert",
		Context:     contextData,
		Status:      alertbus.StatusActive,
		CreatedDate: now.Add(3 * time.Second),
		UpdatedDate: now.Add(3 * time.Second),
	}
	alerts = append(alerts, alert4)

	// Create all alerts
	for _, a := range alerts {
		if err := busDomain.Alert.Create(ctx, a); err != nil {
			return AlertSeedData{}, fmt.Errorf("creating alert %s: %w", a.ID, err)
		}
	}

	// Create recipients - user is recipient of alerts 1, 2, 3 but NOT 4
	recipients := []alertbus.AlertRecipient{
		{
			ID:            uuid.New(),
			AlertID:       alert1ID,
			RecipientType: "user",
			RecipientID:   tu1.ID,
			CreatedDate:   now,
		},
		{
			ID:            uuid.New(),
			AlertID:       alert2ID,
			RecipientType: "user",
			RecipientID:   tu1.ID,
			CreatedDate:   now,
		},
		{
			ID:            uuid.New(),
			AlertID:       alert3ID,
			RecipientType: "user",
			RecipientID:   tu1.ID,
			CreatedDate:   now,
		},
		// Alert 4 has a different recipient (admin)
		{
			ID:            uuid.New(),
			AlertID:       alert4ID,
			RecipientType: "user",
			RecipientID:   tu2.ID,
			CreatedDate:   now,
		},
	}

	if err := busDomain.Alert.CreateRecipients(ctx, recipients); err != nil {
		return AlertSeedData{}, fmt.Errorf("creating recipients: %w", err)
	}

	// Convert to app models for test comparisons
	for _, a := range alerts[:3] { // Only first 3 alerts are for the user
		appAlerts = append(appAlerts, toAppAlert(a))
	}

	return AlertSeedData{
		SeedData: apitest.SeedData{
			Users:  []apitest.User{tu1},
			Admins: []apitest.User{tu2},
		},
		Alerts:           appAlerts,
		AlertIDs:         alertIDs,
		NonRecipientID:   alert4ID,
		LowSeverityID:    alert1ID,
		MediumSeverityID: alert2ID,
		HighSeverityID:   alert3ID,
	}, nil
}

func toAppAlert(bus alertbus.Alert) alertapi.Alert {
	app := alertapi.Alert{
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
