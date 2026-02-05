package apitest

import (
	"context"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus/stores/alertdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// WorkflowInfra holds the workflow infrastructure components for tests.
// This is optional infrastructure that can be initialized by tests that need
// to test workflow event processing.
type WorkflowInfra struct {
	QueueManager *workflow.QueueManager
	Engine       *workflow.Engine
	WorkflowBus  *workflow.Business
	Client       *rabbitmq.Client
}

// InitWorkflowInfra sets up the workflow infrastructure for testing.
// This is a standalone function that tests can call when they need workflow
// capabilities. It uses the shared RabbitMQ test container.
//
// Usage:
//
//	db := dbtest.NewDatabase(t, "Test_Name")
//	wf := apitest.InitWorkflowInfra(t, db)
//	defer wf.Cleanup()
//
//	// Use wf.QueueManager, wf.Engine, wf.WorkflowBus as needed
func InitWorkflowInfra(t *testing.T, db *dbtest.Database) *WorkflowInfra {
	t.Helper()
	ctx := context.Background()

	// Get shared RabbitMQ container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}

	// Initialize workflow queue with test-prefixed names for isolation
	queue := rabbitmq.NewTestWorkflowQueue(client, db.Log)
	if err := queue.Initialize(ctx); err != nil {
		client.Close()
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Create workflow business layer
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowdb.NewStore(db.Log, db.DB))

	// Create and initialize engine
	engine := workflow.NewEngine(db.Log, db.DB, db.BusDomain.Delegate, workflowBus)
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		client.Close()
		t.Fatalf("initializing workflow engine: %s", err)
	}

	// Register action handlers
	registry := engine.GetRegistry()
	registry.Register(communication.NewSendEmailHandler(db.Log, db.DB))
	registry.Register(communication.NewSendNotificationHandler(db.Log, db.DB))

	// Create alertbus for CreateAlertHandler
	alertBus := alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))
	registry.Register(communication.NewCreateAlertHandler(db.Log, alertBus, nil))

	// Register control flow handlers
	registry.Register(control.NewEvaluateConditionHandler(db.Log))

	// Create queue manager
	qm, err := workflow.NewQueueManager(db.Log, db.DB, engine, client, queue)
	if err != nil {
		client.Close()
		t.Fatalf("creating queue manager: %s", err)
	}

	if err := qm.Initialize(ctx); err != nil {
		client.Close()
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Clear any lingering messages
	if err := qm.ClearQueue(ctx); err != nil {
		t.Logf("Warning: could not clear queue: %v", err)
	}

	// Start consumers
	if err := qm.Start(ctx); err != nil {
		client.Close()
		t.Fatalf("starting queue manager: %s", err)
	}

	// Register cleanup via t.Cleanup for automatic resource release
	t.Cleanup(func() {
		qm.Stop(context.Background())
		client.Close()
	})

	t.Log("Workflow infrastructure initialized")

	return &WorkflowInfra{
		QueueManager: qm,
		Engine:       engine,
		WorkflowBus:  workflowBus,
		Client:       client,
	}
}
