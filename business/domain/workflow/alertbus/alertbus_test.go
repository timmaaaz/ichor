package alertbus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Alert(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Alert")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, create(db.BusDomain), "create")
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, queryMine(db.BusDomain, sd), "query-mine")
	unitest.Run(t, acknowledge(db.BusDomain, sd), "acknowledge")
	unitest.Run(t, dismiss(db.BusDomain, sd), "dismiss")
	unitest.Run(t, resolveRelatedAlerts(db.BusDomain), "resolve-related-alerts")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	// First, create a test user (required for acknowledgments FK constraint)
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	testUser := users[0]

	// Create a test alert
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(24 * time.Hour)

	contextData, _ := json.Marshal(map[string]any{"key": "value"})

	alert := alertbus.Alert{
		ID:               uuid.New(),
		AlertType:        "test_alert",
		Severity:         alertbus.SeverityMedium,
		Title:            "Test Alert",
		Message:          "This is a test alert message",
		Context:          contextData,
		SourceEntityName: "test_entity",
		SourceEntityID:   uuid.New(),
		Status:           alertbus.StatusActive,
		ExpiresDate:      &expires,
		CreatedDate:      now,
		UpdatedDate:      now,
	}

	if err := busDomain.Alert.Create(ctx, alert); err != nil {
		return unitest.SeedData{}, err
	}

	// Create recipients for the alert using the real user ID
	roleID := uuid.New() // Role ID doesn't need FK validation for recipients

	recipients := []alertbus.AlertRecipient{
		{
			ID:            uuid.New(),
			AlertID:       alert.ID,
			RecipientType: "user",
			RecipientID:   testUser.ID,
			CreatedDate:   now,
		},
		{
			ID:            uuid.New(),
			AlertID:       alert.ID,
			RecipientType: "role",
			RecipientID:   roleID,
			CreatedDate:   now,
		},
	}

	if err := busDomain.Alert.CreateRecipients(ctx, recipients); err != nil {
		return unitest.SeedData{}, err
	}

	return unitest.SeedData{
		Alerts: []alertbus.Alert{alert},
		AlertTestData: &unitest.AlertTestData{
			UserID: testUser.ID,
			RoleID: roleID,
		},
	}, nil
}

func create(busDomain dbtest.BusDomain) []unitest.Table {
	now := time.Now().UTC().Truncate(time.Second)
	contextData, _ := json.Marshal(map[string]any{"foo": "bar"})

	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: alertbus.Alert{
				AlertType:        "new_test_alert",
				Severity:         alertbus.SeverityHigh,
				Title:            "New Test Alert",
				Message:          "This is a new test alert",
				SourceEntityName: "new_entity",
				Status:           alertbus.StatusActive,
			},
			ExcFunc: func(ctx context.Context) any {
				alert := alertbus.Alert{
					ID:               uuid.New(),
					AlertType:        "new_test_alert",
					Severity:         alertbus.SeverityHigh,
					Title:            "New Test Alert",
					Message:          "This is a new test alert",
					Context:          contextData,
					SourceEntityName: "new_entity",
					Status:           alertbus.StatusActive,
					CreatedDate:      now,
					UpdatedDate:      now,
				}

				if err := busDomain.Alert.Create(ctx, alert); err != nil {
					return err
				}

				// Query it back to verify
				got, err := busDomain.Alert.QueryByID(ctx, alert.ID)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				// Ignore dynamic fields
				expResp.ID = gotResp.ID
				expResp.Context = gotResp.Context
				expResp.SourceEntityID = gotResp.SourceEntityID
				expResp.SourceRuleID = gotResp.SourceRuleID
				expResp.ExpiresDate = gotResp.ExpiresDate
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: sd.Alerts,
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Alert.Query(ctx, alertbus.QueryFilter{}, order.NewBy(alertbus.OrderByCreatedDate, order.DESC), page.MustParse("1", "10"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				// Find matching alert by ID
				if len(gotResp) == 0 {
					return "no alerts returned"
				}

				// Just check that we got some alerts back
				found := false
				for _, g := range gotResp {
					for _, e := range expResp {
						if g.ID == e.ID {
							found = true
							break
						}
					}
				}
				if !found {
					return "expected alert not found in results"
				}

				return ""
			},
		},
	}
}

func queryMine(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "QueryMine",
			ExpResp: sd.Alerts,
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Alert.QueryMine(ctx, sd.AlertTestData.UserID, []uuid.UUID{sd.AlertTestData.RoleID}, alertbus.QueryFilter{}, order.NewBy(alertbus.OrderByCreatedDate, order.DESC), page.MustParse("1", "10"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.([]alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				if len(gotResp) != len(expResp) {
					return "alert count mismatch"
				}

				return ""
			},
		},
	}
}

func acknowledge(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now().UTC().Truncate(time.Second)

	return []unitest.Table{
		{
			Name: "Acknowledge",
			ExpResp: alertbus.Alert{
				ID:     sd.Alerts[0].ID,
				Status: alertbus.StatusAcknowledged,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Alert.Acknowledge(ctx, sd.Alerts[0].ID, sd.AlertTestData.UserID, []uuid.UUID{sd.AlertTestData.RoleID}, "Acknowledged via test", now)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				if gotResp.ID != expResp.ID {
					return "ID mismatch"
				}

				if gotResp.Status != expResp.Status {
					return "status should be acknowledged"
				}

				return ""
			},
		},
	}
}

func dismiss(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	now := time.Now().UTC().Truncate(time.Second)

	// Need a fresh alert for dismiss since the one from seed data was acknowledged
	return []unitest.Table{
		{
			Name: "Dismiss",
			ExpResp: alertbus.Alert{
				Status: alertbus.StatusDismissed,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create a new alert for this test
				contextData, _ := json.Marshal(map[string]any{})
				newAlert := alertbus.Alert{
					ID:          uuid.New(),
					AlertType:   "dismiss_test",
					Severity:    alertbus.SeverityLow,
					Title:       "Dismiss Test Alert",
					Message:     "This alert will be dismissed",
					Context:     contextData,
					Status:      alertbus.StatusActive,
					CreatedDate: now,
					UpdatedDate: now,
				}

				if err := busDomain.Alert.Create(ctx, newAlert); err != nil {
					return err
				}

				// Add recipient using the test user ID
				if err := busDomain.Alert.CreateRecipients(ctx, []alertbus.AlertRecipient{
					{
						ID:            uuid.New(),
						AlertID:       newAlert.ID,
						RecipientType: "user",
						RecipientID:   sd.AlertTestData.UserID,
						CreatedDate:   now,
					},
				}); err != nil {
					return err
				}

				// Dismiss it
				resp, err := busDomain.Alert.Dismiss(ctx, newAlert.ID, sd.AlertTestData.UserID, nil, now)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				if gotResp.Status != expResp.Status {
					return "status should be dismissed"
				}

				return ""
			},
		},
	}
}

func resolveRelatedAlerts(busDomain dbtest.BusDomain) []unitest.Table {
	now := time.Now().UTC().Truncate(time.Second)

	return []unitest.Table{
		{
			Name: "ResolveRelatedAlerts",
			ExpResp: alertbus.Alert{
				Status: alertbus.StatusResolved,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create source entity ID to link alerts
				sourceEntityID := uuid.New()
				alertType := "allocation_status"
				contextData, _ := json.Marshal(map[string]any{})

				// Create first "failure" alert (should get resolved)
				failureAlert1 := alertbus.Alert{
					ID:               uuid.New(),
					AlertType:        alertType,
					Severity:         alertbus.SeverityHigh,
					Title:            "Allocation Failed",
					Message:          "First failure",
					Context:          contextData,
					SourceEntityName: "sales_order",
					SourceEntityID:   sourceEntityID,
					Status:           alertbus.StatusActive,
					CreatedDate:      now,
					UpdatedDate:      now,
				}
				if err := busDomain.Alert.Create(ctx, failureAlert1); err != nil {
					return fmt.Errorf("create failure alert 1: %w", err)
				}

				// Create second "failure" alert (should also get resolved)
				failureAlert2 := alertbus.Alert{
					ID:               uuid.New(),
					AlertType:        alertType,
					Severity:         alertbus.SeverityHigh,
					Title:            "Allocation Failed",
					Message:          "Second failure",
					Context:          contextData,
					SourceEntityName: "sales_order",
					SourceEntityID:   sourceEntityID,
					Status:           alertbus.StatusActive,
					CreatedDate:      now.Add(time.Second),
					UpdatedDate:      now.Add(time.Second),
				}
				if err := busDomain.Alert.Create(ctx, failureAlert2); err != nil {
					return fmt.Errorf("create failure alert 2: %w", err)
				}

				// Create "success" alert (should NOT get resolved - it's excluded)
				successAlert := alertbus.Alert{
					ID:               uuid.New(),
					AlertType:        alertType,
					Severity:         alertbus.SeverityLow,
					Title:            "Allocation Successful",
					Message:          "Success!",
					Context:          contextData,
					SourceEntityName: "sales_order",
					SourceEntityID:   sourceEntityID,
					Status:           alertbus.StatusActive,
					CreatedDate:      now.Add(2 * time.Second),
					UpdatedDate:      now.Add(2 * time.Second),
				}
				if err := busDomain.Alert.Create(ctx, successAlert); err != nil {
					return fmt.Errorf("create success alert: %w", err)
				}

				// Now resolve related alerts (excluding the success alert)
				resolvedCount, err := busDomain.Alert.ResolveRelatedAlerts(ctx, sourceEntityID, alertType, successAlert.ID, now.Add(3*time.Second))
				if err != nil {
					return fmt.Errorf("resolve related alerts: %w", err)
				}

				// Verify count
				if resolvedCount != 2 {
					return fmt.Errorf("expected 2 resolved alerts, got %d", resolvedCount)
				}

				// Query the first failure alert to verify it was resolved
				resolved, err := busDomain.Alert.QueryByID(ctx, failureAlert1.ID)
				if err != nil {
					return fmt.Errorf("query resolved alert: %w", err)
				}

				return resolved
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(alertbus.Alert)
				if !exists {
					return fmt.Sprintf("error occurred: %v", got)
				}

				expResp, exists := exp.(alertbus.Alert)
				if !exists {
					return "error occurred"
				}

				if gotResp.Status != expResp.Status {
					return fmt.Sprintf("status should be resolved, got %s", gotResp.Status)
				}

				return ""
			},
		},
	}
}
