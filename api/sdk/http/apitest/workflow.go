package apitest

import (
	"context"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/integration"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// WorkflowInfra holds the Temporal-based workflow infrastructure for tests.
type WorkflowInfra struct {
	WorkflowBus        *workflow.Business
	TemporalClient     temporalclient.Client
	WorkflowTrigger    *temporal.WorkflowTrigger
	DelegateHandler    *temporal.DelegateHandler
	TriggerProcessor   *workflow.TriggerProcessor
	Worker             worker.Worker
	ApprovalRequestBus *approvalrequestbus.Business
	AlertBus           *alertbus.Business
}

// InitWorkflowInfra sets up Temporal workflow infrastructure for testing.
func InitWorkflowInfra(t *testing.T, db *dbtest.Database) *WorkflowInfra {
	t.Helper()
	ctx := context.Background()

	// 1. Get shared Temporal test container.
	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{
		HostPort: container.HostPort,
	})
	if err != nil {
		t.Fatalf("connecting to temporal: %s", err)
	}

	// 2. Create workflow business layer (share one store instance with the trigger).
	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)

	// 3. Build action registry (same 4 handlers as before).
	registry := workflow.NewActionRegistry()
	registry.Register(communication.NewSendEmailHandler(db.Log, db.DB, nil, ""))
	registry.Register(communication.NewSendNotificationHandler(db.Log, nil))
	// alertBus and approvalRequestBus are constructed independently (not from
	// db.BusDomain) so each test gets its own bus instance. This ensures
	// approval request and alert state is not shared across test runs that
	// reuse the same db.BusDomain reference (e.g., db.BusDomain.Alert could
	// accumulate state from prior tests in the same DB). Seek_approval
	// activities also require a live approvalRequestBus to create DB records.
	alertBus := alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))
	registry.Register(communication.NewCreateAlertHandler(db.Log, alertBus, nil))
	registry.Register(control.NewEvaluateConditionHandler(db.Log))
	registry.Register(integration.NewCallWebhookHandler(db.Log))

	// Build approval request bus so seek_approval can create real DB records.
	approvalRequestStore := approvalrequestdb.NewStore(db.Log, db.DB)
	approvalBus := approvalrequestbus.NewBusiness(db.Log, db.BusDomain.Delegate, approvalRequestStore)

	// 4. Create and start test worker with unique task queue per test.
	taskQueue := "test-workflow-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)
	asyncRegistry := temporal.NewAsyncRegistry()
	asyncRegistry.Register("seek_approval", approval.NewSeekApprovalHandler(db.Log, db.DB, approvalBus, alertBus))

	activities := &temporal.Activities{
		Registry:      registry,
		AsyncRegistry: asyncRegistry,
	}
	w.RegisterActivity(activities)

	if err := w.Start(); err != nil {
		tc.Close()
		t.Fatalf("starting temporal worker: %s", err)
	}

	// 5. Create trigger infrastructure.
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		w.Stop()
		tc.Close()
		t.Fatalf("initializing trigger processor: %s", err)
	}

	workflowTrigger := temporal.NewWorkflowTrigger(
		db.Log, tc, triggerProcessor, edgeStore, workflowStore,
	).WithTaskQueue(taskQueue)

	// 6. Create delegate handler.
	delegateHandler := temporal.NewDelegateHandler(db.Log, workflowTrigger)

	// 7. Register cleanup.
	t.Cleanup(func() {
		w.Stop()
		tc.Close()
	})

	t.Log("Temporal workflow infrastructure initialized")

	return &WorkflowInfra{
		WorkflowBus:        workflowBus,
		TemporalClient:     tc,
		WorkflowTrigger:    workflowTrigger,
		DelegateHandler:    delegateHandler,
		TriggerProcessor:   triggerProcessor,
		Worker:             w,
		ApprovalRequestBus: approvalBus,
		AlertBus:           alertBus,
	}
}
