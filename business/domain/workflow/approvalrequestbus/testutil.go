package approvalrequestbus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// TestNewApprovalRequests generates n new approval requests for testing.
// executionIDs and ruleIDs must reference real rows (FK constraints).
func TestNewApprovalRequests(n int, executionIDs, ruleIDs, approverIDs uuid.UUIDs) []NewApprovalRequest {
	requests := make([]NewApprovalRequest, n)

	taskNames := []string{
		"Transfer Order Approval",
		"Inventory Adjustment Review",
		"Purchase Order Approval",
		"Quality Hold Release",
		"Cycle Count Variance Review",
	}

	for i := range n {
		requests[i] = NewApprovalRequest{
			ExecutionID:     executionIDs[i%len(executionIDs)],
			RuleID:          ruleIDs[i%len(ruleIDs)],
			ActionName:      fmt.Sprintf("approve_%d", i+1),
			Approvers:       approverIDs,
			ApprovalType:    ApprovalTypeAny,
			TimeoutHours:    24,
			TaskToken:       fmt.Sprintf("task-token-%04d", i+1),
			ApprovalMessage: fmt.Sprintf("Please review: %s #%d", taskNames[i%len(taskNames)], i+1),
		}
	}

	return requests
}

// TestSeedApprovalRequests creates n approval requests in the database for testing.
// executionIDs and ruleIDs must reference existing DB rows.
func TestSeedApprovalRequests(ctx context.Context, n int, executionIDs, ruleIDs, approverIDs uuid.UUIDs, api *Business) ([]ApprovalRequest, error) {
	newRequests := TestNewApprovalRequests(n, executionIDs, ruleIDs, approverIDs)

	requests := make([]ApprovalRequest, len(newRequests))
	for i, nr := range newRequests {
		req, err := api.Create(ctx, nr)
		if err != nil {
			return nil, fmt.Errorf("seeding approval request %d: %w", i, err)
		}
		requests[i] = req
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].ID.String() < requests[j].ID.String()
	})

	return requests, nil
}
