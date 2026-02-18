package apitest

import (
	"context"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// WorkflowInfra holds the Temporal-based workflow infrastructure for tests.
type WorkflowInfra struct {
	WorkflowBus      *workflow.Business
	TemporalClient   temporalclient.Client
	WorkflowTrigger  *temporal.WorkflowTrigger
	DelegateHandler  *temporal.DelegateHandler
	TriggerProcessor *workflow.TriggerProcessor
	Worker           worker.Worker
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

	// 2. Create workflow business layer.
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowdb.NewStore(db.Log, db.DB))

	// 3. Build action registry (same 4 handlers as before).
	registry := workflow.NewActionRegistry()
	registry.Register(communication.NewSendEmailHandler(db.Log, db.DB, nil, ""))
	registry.Register(communication.NewSendNotificationHandler(db.Log, nil))
	alertBus := alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))
	registry.Register(communication.NewCreateAlertHandler(db.Log, alertBus, nil))
	registry.Register(control.NewEvaluateConditionHandler(db.Log))

	// 4. Create and start test worker with unique task queue per test.
	taskQueue := "test-workflow-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(temporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(temporal.ExecuteBranchUntilConvergence)
	activities := &temporal.Activities{
		Registry:      registry,
		AsyncRegistry: temporal.NewAsyncRegistry(),
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
		db.Log, tc, triggerProcessor, edgeStore,
	)

	// 6. Create delegate handler.
	delegateHandler := temporal.NewDelegateHandler(db.Log, workflowTrigger)

	// 7. Register cleanup.
	t.Cleanup(func() {
		w.Stop()
		tc.Close()
	})

	t.Log("Temporal workflow infrastructure initialized")

	return &WorkflowInfra{
		WorkflowBus:      workflowBus,
		TemporalClient:   tc,
		WorkflowTrigger:  workflowTrigger,
		DelegateHandler:  delegateHandler,
		TriggerProcessor: triggerProcessor,
		Worker:           w,
	}
}
