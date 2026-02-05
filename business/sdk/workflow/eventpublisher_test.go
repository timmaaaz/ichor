package workflow_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// TestEventPublisher_PublishCreateEvent tests that EventPublisher correctly
// queues create events to RabbitMQ. This is a unit test for the EventPublisher
// using real infrastructure (no mocks).
func TestEventPublisher_PublishCreateEvent(t *testing.T) {
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
	db := dbtest.NewDatabase(t, "Test_EventPublisher_Create")
	ctx := context.Background()

	// Real workflow business layer
	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))

	// Real workflow engine (no rules needed - just testing event queuing)
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
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

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Test entity result (simulates app layer response)
	orderResult := struct {
		ID         string `json:"id"`
		Number     string `json:"number"`
		CustomerID string `json:"customer_id"`
	}{
		ID:         uuid.New().String(),
		Number:     "ORD-001",
		CustomerID: uuid.New().String(),
	}

	// Publish create event
	publisher.PublishCreateEvent(ctx, "sales.orders", orderResult, uuid.New())

	// Wait for async event to be queued
	time.Sleep(300 * time.Millisecond)

	// Verify event was enqueued
	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
		t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
			initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
	}

	t.Log("SUCCESS: PublishCreateEvent queued event correctly")
}

// TestEventPublisher_PublishUpdateEvent tests that EventPublisher correctly
// queues update events with field changes to RabbitMQ.
func TestEventPublisher_PublishUpdateEvent(t *testing.T) {
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
	db := dbtest.NewDatabase(t, "Test_EventPublisher_Update")
	ctx := context.Background()

	// Real workflow business layer
	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))

	// Real workflow engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
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

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Test entity result
	orderResult := map[string]interface{}{
		"id":     uuid.New().String(),
		"number": "ORD-002",
		"status": "completed",
	}

	// Field changes for update event
	fieldChanges := map[string]workflow.FieldChange{
		"status": {
			OldValue: "pending",
			NewValue: "completed",
		},
	}

	// Publish update event with field changes
	publisher.PublishUpdateEvent(ctx, "sales.orders", orderResult, fieldChanges, uuid.New())

	// Wait for async event to be queued
	time.Sleep(300 * time.Millisecond)

	// Verify event was enqueued
	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
		t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
			initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
	}

	t.Log("SUCCESS: PublishUpdateEvent queued event correctly")
}

// TestEventPublisher_PublishDeleteEvent tests that EventPublisher correctly
// queues delete events to RabbitMQ.
func TestEventPublisher_PublishDeleteEvent(t *testing.T) {
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
	db := dbtest.NewDatabase(t, "Test_EventPublisher_Delete")
	ctx := context.Background()

	// Real workflow business layer
	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))

	// Real workflow engine
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
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

	// Get initial metrics
	initialMetrics := qm.GetMetrics()

	// Publish delete event (only needs entity ID)
	entityID := uuid.New()
	publisher.PublishDeleteEvent(ctx, "sales.orders", entityID, uuid.New())

	// Wait for async event to be queued
	time.Sleep(300 * time.Millisecond)

	// Verify event was enqueued
	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
		t.Errorf("Expected TotalEnqueued to increase by 1, got %d -> %d",
			initialMetrics.TotalEnqueued, finalMetrics.TotalEnqueued)
	}

	t.Log("SUCCESS: PublishDeleteEvent queued event correctly")
}

// TestEventPublisher_ExtractEntityID tests ID extraction from various result formats.
func TestEventPublisher_ExtractEntityID(t *testing.T) {
	// This test validates the EventPublisher correctly extracts IDs from different
	// result formats. Since extractEntityData is private, we test through the
	// public interface by verifying events are queued with correct data.

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

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

	db := dbtest.NewDatabase(t, "Test_EventPublisher_ExtractID")
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
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

	publisher := workflow.NewEventPublisher(log, qm)

	tests := []struct {
		name   string
		result interface{}
	}{
		{
			name: "struct with string ID",
			result: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   uuid.New().String(),
				Name: "Test Entity",
			},
		},
		{
			name: "struct with uuid.UUID ID",
			result: struct {
				ID   uuid.UUID `json:"id"`
				Name string    `json:"name"`
			}{
				ID:   uuid.New(),
				Name: "Test Entity",
			},
		},
		{
			name: "map with string ID",
			result: map[string]interface{}{
				"id":   uuid.New().String(),
				"name": "Test Entity",
			},
		},
		{
			name: "pointer to struct",
			result: &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   uuid.New().String(),
				Name: "Test Entity",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialMetrics := qm.GetMetrics()

			publisher.PublishCreateEvent(ctx, "test_entity", tt.result, uuid.New())

			time.Sleep(300 * time.Millisecond)

			finalMetrics := qm.GetMetrics()
			if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued+1 {
				t.Errorf("Expected event to be enqueued for %s", tt.name)
			}
		})
	}

	t.Log("SUCCESS: All ID extraction formats work correctly")
}

// TestEventPublisher_NilResult tests that EventPublisher handles nil results gracefully.
func TestEventPublisher_NilResult(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

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

	db := dbtest.NewDatabase(t, "Test_EventPublisher_NilResult")
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
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

	publisher := workflow.NewEventPublisher(log, qm)

	initialMetrics := qm.GetMetrics()

	// Publish with nil result - should log error but not panic
	publisher.PublishCreateEvent(ctx, "test_entity", nil, uuid.New())

	time.Sleep(300 * time.Millisecond)

	// Event should NOT be queued for nil result
	finalMetrics := qm.GetMetrics()
	if finalMetrics.TotalEnqueued != initialMetrics.TotalEnqueued {
		t.Errorf("Expected no events queued for nil result, got %d new events",
			finalMetrics.TotalEnqueued-initialMetrics.TotalEnqueued)
	}

	t.Log("SUCCESS: Nil result handled gracefully (logged error, no panic)")
}

// TestEventPublisher_NonBlocking tests that PublishCreateEvent is non-blocking.
func TestEventPublisher_NonBlocking(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

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

	db := dbtest.NewDatabase(t, "Test_EventPublisher_NonBlocking")
	ctx := context.Background()

	workflowBus := workflow.NewBusiness(log, nil, workflowdb.NewStore(log, db.DB))
	workflow.ResetEngineForTesting()
	engine := workflow.NewEngine(log, db.DB, nil, workflowBus)
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

	publisher := workflow.NewEventPublisher(log, qm)

	// Measure time to publish - should be nearly instant since it's non-blocking
	start := time.Now()

	for i := 0; i < 100; i++ {
		result := map[string]interface{}{
			"id":   uuid.New().String(),
			"name": "Test Entity",
		}
		publisher.PublishCreateEvent(ctx, "test_entity", result, uuid.New())
	}

	elapsed := time.Since(start)

	// 100 publishes should complete in well under 100ms since they're non-blocking
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected non-blocking behavior, but 100 publishes took %v", elapsed)
	}

	t.Logf("SUCCESS: 100 non-blocking publishes completed in %v", elapsed)

	// Wait for events to actually be queued
	time.Sleep(500 * time.Millisecond)
}
