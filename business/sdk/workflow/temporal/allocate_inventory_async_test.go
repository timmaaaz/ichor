package temporal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"go.temporal.io/sdk/testsuite"
)

// TestAllActionsRouteThroughSyncActivity verifies that all action types
// now route through ExecuteActionActivity (synchronous Temporal activity).
//
// This replaced the previous async completion pattern that used RabbitMQ.
// Temporal handles retries, timeouts, and failure recovery natively.
func TestNonHumanActionsRouteThroughSyncActivity(t *testing.T) {
	actionTypes := []string{
		"allocate_inventory",
		"send_email",
		"credit_check",
		"fraud_detection",
		"third_party_api_call",
		"reserve_shipping",
		"evaluate_condition",
		"update_field",
	}

	for _, actionType := range actionTypes {
		t.Run(actionType, func(t *testing.T) {
			activityFunc := selectActivityFunc(actionType)
			require.Equal(t, "ExecuteActionActivity", activityFunc,
				"%s should route to ExecuteActionActivity (sync)", actionType)
		})
	}
}

func TestHumanActionsRouteThroughAsyncActivity(t *testing.T) {
	humanTypes := []string{
		"seek_approval",
		"manager_approval",
		"manual_review",
		"human_verification",
		"approval_request",
	}

	for _, actionType := range humanTypes {
		t.Run(actionType, func(t *testing.T) {
			activityFunc := selectActivityFunc(actionType)
			require.Equal(t, "ExecuteAsyncActionActivity", activityFunc,
				"%s should route to ExecuteAsyncActionActivity (async)", actionType)
		})
	}
}

// TestLongRunningActionsGetExtendedTimeouts verifies timeout configuration.
func TestLongRunningActionsGetExtendedTimeouts(t *testing.T) {
	longRunningTypes := []string{
		"allocate_inventory",
		"send_email",
		"credit_check",
		"fraud_detection",
		"third_party_api_call",
		"reserve_shipping",
	}

	for _, actionType := range longRunningTypes {
		t.Run(actionType, func(t *testing.T) {
			require.True(t, isLongRunningAction(actionType),
				"%s should be classified as long-running", actionType)

			opts := activityOptions(actionType)
			require.Equal(t, 30*time.Minute, opts.StartToCloseTimeout,
				"%s should have 30-minute timeout", actionType)
		})
	}
}

// TestAllocateInventory_RequiresHandlerRegistration verifies that
// allocate_inventory needs to be registered in the sync ActionRegistry.
func TestAllocateInventory_RequiresHandlerRegistration(t *testing.T) {
	t.Run("fails_when_handler_not_registered", func(t *testing.T) {
		// Empty registry - no handlers
		actionRegistry := workflow.NewActionRegistry()

		activities := &Activities{
			Registry:      actionRegistry,
			AsyncRegistry: NewAsyncRegistry(),
		}

		input := ActionActivityInput{
			ActionID:    uuid.MustParse("5f6808a1-89f5-49e6-b971-b71c9b267a8d"),
			ActionName:  "Allocate Inventory for Line Item",
			ActionType:  "allocate_inventory",
			Config:      json.RawMessage(`{"priority":"high","allow_partial":false,"reference_type":"order","allocation_mode":"reserve","allocation_strategy":"fifo","source_from_line_item":true}`),
			Context:     map[string]any{"entity_id": "534ebc4e-f3ae-4fb3-b513-93757ddc7a93", "product_id": "ac4448b0-db6f-494b-8429-0e654840ab4e", "quantity": float64(123)},
			RuleID:      uuid.MustParse("1862112b-42fc-44ce-9f57-768e4a558efe"),
			ExecutionID: uuid.MustParse("b1ca1d57-1fc9-43a2-86c7-76c99b03de20"),
			RuleName:    "Line Item Created - Allocate Inventory",
		}

		suite := &testsuite.WorkflowTestSuite{}
		env := suite.NewTestActivityEnvironment()
		env.RegisterActivity(activities)

		// Now routes through sync activity, but handler is not registered
		_, err := env.ExecuteActivity(activities.ExecuteActionActivity, input)

		require.Error(t, err)
		require.Contains(t, err.Error(), "no handler registered for type allocate_inventory",
			"Handler must be registered via RegisterInventoryActions in workflow-worker")
	})
}

// TestHumanActionsGetMultiDayTimeouts verifies approval actions get extended timeouts.
func TestHumanActionsGetMultiDayTimeouts(t *testing.T) {
	humanTypes := []string{
		"manager_approval",
		"manual_review",
		"human_verification",
		"approval_request",
		"seek_approval",
	}

	for _, actionType := range humanTypes {
		t.Run(actionType, func(t *testing.T) {
			require.True(t, isHumanAction(actionType),
				"%s should be classified as human action", actionType)

			opts := activityOptions(actionType)
			require.Equal(t, 7*24*time.Hour, opts.StartToCloseTimeout,
				"%s should have 7-day timeout", actionType)
		})
	}
}
