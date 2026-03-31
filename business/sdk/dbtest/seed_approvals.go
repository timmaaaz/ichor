package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

func seedApprovals(ctx context.Context, busDomain BusDomain, foundation FoundationSeed) error {
	// Query automation rules created by seedWorkflow to get real rule IDs.
	rules, err := busDomain.Workflow.QueryAutomationRulesView(ctx)
	if err != nil {
		return fmt.Errorf("querying automation rules for approval seeding: %w", err)
	}
	if len(rules) == 0 {
		return fmt.Errorf("no automation rules found — seedWorkflow must run first")
	}

	ruleIDs := make(uuid.UUIDs, len(rules))
	for i, r := range rules {
		ruleIDs[i] = r.ID
	}

	// Create real automation executions (satisfies execution_id FK).
	executions, err := workflow.TestSeedAutomationExecutions(ctx, 5, ruleIDs, busDomain.Workflow)
	if err != nil {
		return fmt.Errorf("seeding automation executions for approvals: %w", err)
	}

	executionIDs := make(uuid.UUIDs, len(executions))
	for i, e := range executions {
		executionIDs[i] = e.ID
	}

	adminIDs := make(uuid.UUIDs, len(foundation.Admins))
	for i, a := range foundation.Admins {
		adminIDs[i] = a.ID
	}

	// Seed 5 pending approval requests for supervisor inbox.
	_, err = approvalrequestbus.TestSeedApprovalRequests(ctx, 5, executionIDs, ruleIDs, adminIDs, busDomain.ApprovalRequest)
	if err != nil {
		return fmt.Errorf("seeding approval requests: %w", err)
	}

	return nil
}
