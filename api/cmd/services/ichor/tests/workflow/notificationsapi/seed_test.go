package notificationsapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// NotificationSeedData holds seed state for notification summary tests.
type NotificationSeedData struct {
	apitest.SeedData

	// Expected alert counts by severity (only active alerts for the user).
	ExpectedLow      int
	ExpectedMedium   int
	ExpectedHigh     int
	ExpectedCritical int
	ExpectedTotal    int

	// Expected pending approval count for the user.
	ExpectedPending int
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (NotificationSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// -------------------------------------------------------------------------
	// Seed users

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return NotificationSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return NotificationSeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// -------------------------------------------------------------------------
	// Seed alerts with varying severities

	now := time.Now().UTC().Truncate(time.Second)
	contextData, _ := json.Marshal(map[string]any{"test": "notification"})

	alerts := []alertbus.Alert{
		{
			ID:          uuid.New(),
			AlertType:   "test_alert",
			Severity:    alertbus.SeverityLow,
			Title:       "Low Alert",
			Message:     "A low severity alert",
			Context:     contextData,
			Status:      alertbus.StatusActive,
			CreatedDate: now,
			UpdatedDate: now,
		},
		{
			ID:          uuid.New(),
			AlertType:   "test_alert",
			Severity:    alertbus.SeverityMedium,
			Title:       "Medium Alert",
			Message:     "A medium severity alert",
			Context:     contextData,
			Status:      alertbus.StatusActive,
			CreatedDate: now.Add(1 * time.Second),
			UpdatedDate: now.Add(1 * time.Second),
		},
		{
			ID:          uuid.New(),
			AlertType:   "test_alert",
			Severity:    alertbus.SeverityHigh,
			Title:       "High Alert",
			Message:     "A high severity alert",
			Context:     contextData,
			Status:      alertbus.StatusActive,
			CreatedDate: now.Add(2 * time.Second),
			UpdatedDate: now.Add(2 * time.Second),
		},
	}

	for _, a := range alerts {
		if err := busDomain.Alert.Create(ctx, a); err != nil {
			return NotificationSeedData{}, fmt.Errorf("creating alert %s: %w", a.ID, err)
		}
	}

	// Create recipients — user is recipient of all 3 alerts.
	var recipients []alertbus.AlertRecipient
	for _, a := range alerts {
		recipients = append(recipients, alertbus.AlertRecipient{
			ID:            uuid.New(),
			AlertID:       a.ID,
			RecipientType: "user",
			RecipientID:   tu1.ID,
			CreatedDate:   now,
		})
	}
	if err := busDomain.Alert.CreateRecipients(ctx, recipients); err != nil {
		return NotificationSeedData{}, fmt.Errorf("creating recipients: %w", err)
	}

	// -------------------------------------------------------------------------
	// Seed approval requests (requires valid workflow FKs)

	wfData, err := workflow.TestSeedFullWorkflow(ctx, admins[0].ID, busDomain.Workflow)
	if err != nil {
		return NotificationSeedData{}, fmt.Errorf("seeding workflow: %w", err)
	}

	approvalStore := approvalrequestdb.NewStore(db.Log, db.DB)

	// Pending approval 1: user is approver.
	req1 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     wfData.AutomationExecutions[0].ID,
		RuleID:          wfData.AutomationRules[0].ID,
		ActionName:      "seek_approval_0",
		Approvers:       []uuid.UUID{tu1.ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		ApprovalMessage: "Please approve this",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req1); err != nil {
		return NotificationSeedData{}, fmt.Errorf("creating approval 1: %w", err)
	}

	// Pending approval 2: user is approver.
	req2 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     wfData.AutomationExecutions[1].ID,
		RuleID:          wfData.AutomationRules[0].ID,
		ActionName:      "seek_approval_1",
		Approvers:       []uuid.UUID{tu1.ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		ApprovalMessage: "Another approval needed",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req2); err != nil {
		return NotificationSeedData{}, fmt.Errorf("creating approval 2: %w", err)
	}

	// Resolved approval: user was approver but it's already approved (should NOT count).
	req3 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     wfData.AutomationExecutions[0].ID,
		RuleID:          wfData.AutomationRules[0].ID,
		ActionName:      "seek_approval_2",
		Approvers:       []uuid.UUID{tu1.ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusApproved,
		TimeoutHours:    72,
		ApprovalMessage: "Already resolved",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req3); err != nil {
		return NotificationSeedData{}, fmt.Errorf("creating resolved approval: %w", err)
	}

	return NotificationSeedData{
		SeedData: apitest.SeedData{
			Users:  []apitest.User{tu1},
			Admins: []apitest.User{tu2},
		},
		ExpectedLow:      1,
		ExpectedMedium:   1,
		ExpectedHigh:     1,
		ExpectedCritical: 0,
		ExpectedTotal:    3,
		ExpectedPending:  2, // Only the 2 pending approvals, not the resolved one.
	}, nil
}
