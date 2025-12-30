package workflow_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// TestDelegateHandler_OrdersCreated tests that the delegate handler correctly
// bridges ordersbus delegate events to workflow events. This validates the
// Phase 2 delegate pattern integration.
func TestDelegateHandler_OrdersCreated(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	// Real RabbitMQ container
	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing queue: %s", err)
	}

	// Real database
	db := dbtest.NewDatabase(t, "Test_DelegateHandler_OrdersCreated")
	ctx := context.Background()

	// Real workflow business layer
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))

	// Real workflow engine (no rules needed - just testing event queuing)
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	// Real queue manager
	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}
	if err := qm.ClearQueue(ctx); err != nil {
		t.Logf("Warning: could not clear queue: %v", err)
	}
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Create EventPublisher
	publisher := workflow.NewEventPublisher(log, qm)

	// Create Delegate and register handler
	del := delegate.New(log)
	delegateHandler := workflow.NewDelegateHandler(log, publisher)
	delegateHandler.RegisterDomain(del, ordersbus.DomainName, ordersbus.EntityName)

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Create a test order
	testOrder := ordersbus.Order{
		ID:                  uuid.New(),
		Number:              "ORD-TEST-001",
		CustomerID:          uuid.New(),
		DueDate:             time.Now().Add(24 * time.Hour),
		FulfillmentStatusID: uuid.New(),
		CreatedBy:           uuid.New(),
		UpdatedBy:           uuid.New(),
		CreatedDate:         time.Now(),
		UpdatedDate:         time.Now(),
	}

	// Simulate calling delegate.Call with order created event
	// This is what ordersbus.Create() does internally
	eventData := ordersbus.ActionCreatedData(testOrder)
	if err := del.Call(ctx, eventData); err != nil {
		t.Fatalf("delegate call failed: %s", err)
	}

	// Wait for async event to be queued
	time.Sleep(300 * time.Millisecond)

	// Verify event was enqueued
	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
		t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
			initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
	}

	t.Log("SUCCESS: DelegateHandler correctly bridged ordersbus created event to workflow queue")
}

// TestDelegateHandler_OrdersUpdated tests delegate handler for update events.
func TestDelegateHandler_OrdersUpdated(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	// Real RabbitMQ container
	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing queue: %s", err)
	}

	// Real database
	db := dbtest.NewDatabase(t, "Test_DelegateHandler_OrdersUpdated")
	ctx := context.Background()

	// Setup workflow infrastructure
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}
	if err := qm.ClearQueue(ctx); err != nil {
		t.Logf("Warning: could not clear queue: %v", err)
	}
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Create EventPublisher and DelegateHandler
	publisher := workflow.NewEventPublisher(log, qm)
	del := delegate.New(log)
	delegateHandler := workflow.NewDelegateHandler(log, publisher)
	delegateHandler.RegisterDomain(del, ordersbus.DomainName, ordersbus.EntityName)

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Create a test order (as if it was updated)
	testOrder := ordersbus.Order{
		ID:                  uuid.New(),
		Number:              "ORD-TEST-002",
		CustomerID:          uuid.New(),
		DueDate:             time.Now().Add(48 * time.Hour),
		FulfillmentStatusID: uuid.New(),
		CreatedBy:           uuid.New(),
		UpdatedBy:           uuid.New(),
		CreatedDate:         time.Now().Add(-1 * time.Hour),
		UpdatedDate:         time.Now(),
	}

	// Simulate update event
	eventData := ordersbus.ActionUpdatedData(testOrder)
	if err := del.Call(ctx, eventData); err != nil {
		t.Fatalf("delegate call failed: %s", err)
	}

	// Wait for async event to be queued
	time.Sleep(300 * time.Millisecond)

	// Verify event was enqueued
	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
		t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
			initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
	}

	t.Log("SUCCESS: DelegateHandler correctly bridged ordersbus updated event to workflow queue")
}

// TestDelegateHandler_OrdersDeleted tests delegate handler for delete events.
func TestDelegateHandler_OrdersDeleted(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	// Real RabbitMQ container
	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing queue: %s", err)
	}

	// Real database
	db := dbtest.NewDatabase(t, "Test_DelegateHandler_OrdersDeleted")
	ctx := context.Background()

	// Setup workflow infrastructure
	workflowBus := workflow.NewBusiness(log, workflowdb.NewStore(log, db.DB))
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, workflowBus)
	if err := engine.Initialize(ctx, workflowBus); err != nil {
		t.Fatalf("initializing engine: %s", err)
	}

	qm, err := workflow.NewQueueManager(log, db.DB, engine, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}
	if err := qm.ClearQueue(ctx); err != nil {
		t.Logf("Warning: could not clear queue: %v", err)
	}
	qm.ResetCircuitBreaker() // Reset circuit breaker state for test isolation
	qm.ResetMetrics()        // Reset metrics for clean assertions
	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Create EventPublisher and DelegateHandler
	publisher := workflow.NewEventPublisher(log, qm)
	del := delegate.New(log)
	delegateHandler := workflow.NewDelegateHandler(log, publisher)
	delegateHandler.RegisterDomain(del, ordersbus.DomainName, ordersbus.EntityName)

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Create a test order (as if it was deleted)
	testOrder := ordersbus.Order{
		ID:                  uuid.New(),
		Number:              "ORD-TEST-003",
		CustomerID:          uuid.New(),
		DueDate:             time.Now(),
		FulfillmentStatusID: uuid.New(),
		CreatedBy:           uuid.New(),
		UpdatedBy:           uuid.New(),
		CreatedDate:         time.Now().Add(-2 * time.Hour),
		UpdatedDate:         time.Now(),
	}

	// Simulate delete event
	eventData := ordersbus.ActionDeletedData(testOrder)
	if err := del.Call(ctx, eventData); err != nil {
		t.Fatalf("delegate call failed: %s", err)
	}

	// Wait for async event to be queued
	time.Sleep(300 * time.Millisecond)

	// Verify event was enqueued
	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
		t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
			initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
	}

	t.Log("SUCCESS: DelegateHandler correctly bridged ordersbus deleted event to workflow queue")
}
