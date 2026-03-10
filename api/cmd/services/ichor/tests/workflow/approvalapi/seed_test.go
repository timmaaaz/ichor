package approvalapi_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ApproveSeedData holds test-specific state for resolve tests.
type ApproveSeedData struct {
	apitest.SeedData

	// ApprovalID is the pending approval request seeded for testing (no task token).
	ApprovalID uuid.UUID

	// ApprovalWithTokenID is a pending approval request with a non-empty task_token.
	// Used to test the retry path behavior when Temporal has a token.
	ApprovalWithTokenID uuid.UUID
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (ApproveSeedData, error) {
	ctx := context.Background()

	// -------------------------------------------------------------------------
	// Seed users

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		return ApproveSeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	admin := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// -------------------------------------------------------------------------
	// Seed workflow data (provides valid execution_id and rule_id FKs)

	wfData, err := workflow.TestSeedFullWorkflow(ctx, admins[0].ID, db.BusDomain.Workflow)
	if err != nil {
		return ApproveSeedData{}, fmt.Errorf("seeding workflow: %w", err)
	}

	executionID1 := wfData.AutomationExecutions[0].ID
	executionID2 := wfData.AutomationExecutions[1].ID
	ruleID := wfData.AutomationRules[0].ID

	// -------------------------------------------------------------------------
	// Create approval requests directly via the store

	approvalStore := approvalrequestdb.NewStore(db.Log, db.DB)

	// Approval 1: no task_token (simulates manual/non-Temporal approval)
	req1 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     executionID1,
		RuleID:          ruleID,
		ActionName:      "seek_approval_0",
		Approvers:       []uuid.UUID{admins[0].ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		TaskToken:       "",
		ApprovalMessage: "Please approve",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req1); err != nil {
		return ApproveSeedData{}, fmt.Errorf("creating approval: %w", err)
	}

	// Approval 2: with task_token (simulates Temporal-backed approval).
	// Since the test server has no Temporal client configured, retryTemporalCompletion
	// will return the approval without calling Complete — exercising the nil-completer path.
	fakeToken := base64.StdEncoding.EncodeToString([]byte("temporal-task-token-abc"))
	req2 := approvalrequestbus.ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     executionID2,
		RuleID:          ruleID,
		ActionName:      "seek_approval_1",
		Approvers:       []uuid.UUID{admins[0].ID},
		ApprovalType:    "any",
		Status:          approvalrequestbus.StatusPending,
		TimeoutHours:    72,
		TaskToken:       fakeToken,
		ApprovalMessage: "Please approve with token",
		CreatedDate:     time.Now(),
	}
	if err := approvalStore.Create(ctx, req2); err != nil {
		return ApproveSeedData{}, fmt.Errorf("creating approval with token: %w", err)
	}

	return ApproveSeedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.User{admin},
		},
		ApprovalID:          req1.ID,
		ApprovalWithTokenID: req2.ID,
	}, nil
}
