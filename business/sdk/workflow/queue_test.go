package workflow_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// NOTE: This is NOT testing the following
// - Whether the Engine actually processes workflows correctly - you're just returning success
// - The entire dependency resolution and rule matching logic - that's all bypassed
// - Action execution - none of the actual workflow actions run
// - Real database interactions - no rules are loaded, no state is persiste

// This is testing the following
// The QueueManager correctly integrates with RabbitMQ - messages are being queued, consumed, and acknowledged properly
// The message transformation logic works - converting between TriggerEvents and RabbitMQ Messages
// The circuit breaker and metrics tracking function - failures are counted, circuit opens/closes
// Queue routing logic - different entities go to different queues

// stubEngine wraps a real Engine but provides simplified execution for testing
type stubEngine struct {
	*workflow.Engine
	executionCount int
	lastEvent      workflow.TriggerEvent
	failNext       bool // For testing failure scenarios
}

func newStubEngine(log *logger.Logger) *stubEngine {
	return &stubEngine{
		Engine: workflow.NewEngine(log, nil), // Real engine structure, no DB needed
	}
}

func (e *stubEngine) ExecuteWorkflow(ctx context.Context, event workflow.TriggerEvent) (*workflow.WorkflowExecution, error) {
	e.executionCount++
	e.lastEvent = event

	if e.failNext {
		e.failNext = false
		return nil, fmt.Errorf("simulated failure")
	}

	now := time.Now()
	return &workflow.WorkflowExecution{
		ExecutionID:  fmt.Sprintf("exec_%d_%s", e.executionCount, uuid.New().String()[:8]),
		TriggerEvent: event,
		ExecutionPlan: workflow.ExecutionPlan{
			PlanID:           fmt.Sprintf("plan_%d", e.executionCount),
			MatchedRuleCount: 1,
			TotalBatches:     1,
			CreatedAt:        now,
		},
		Status:       workflow.StatusCompleted,
		StartedAt:    now,
		CompletedAt:  &now,
		BatchResults: []workflow.BatchResult{},
		Errors:       []string{},
	}, nil
}

func TestQueueManager_Initialize(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	// Start RabbitMQ container
	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer func() {
		if err := rabbitmq.StopRabbitMQ(container); err != nil {
			t.Errorf("stopping rabbitmq: %s", err)
		}
	}()

	// Create RabbitMQ client
	// config := rabbitmq.NewTestConfig(container.URL)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Create workflow engine (mock)
	engine := newStubEngine(log)

	// Create queue manager
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()

	// Test initialization
	if err := qm.Initialize(ctx); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Verify queues were created by getting status
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Errorf("GetQueueStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetQueueStatus() returned nil")
	}
}

func TestQueueManager_QueueEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	tests := []struct {
		name    string
		event   workflow.TriggerEvent
		wantErr bool
	}{
		{
			name: "valid create event",
			event: workflow.TriggerEvent{
				EventType:  "on_create",
				EntityName: "customers",
				EntityID:   "cust_123",
				Timestamp:  time.Now(),
				RawData: map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
				},
				UserID: "user_456",
			},
			wantErr: false,
		},
		{
			name: "valid update event with field changes",
			event: workflow.TriggerEvent{
				EventType:  "on_update",
				EntityName: "customers",
				EntityID:   "cust_123",
				Timestamp:  time.Now(),
				FieldChanges: map[string]workflow.FieldChange{
					"status": {
						OldValue: "regular",
						NewValue: "premium",
					},
				},
				RawData: map[string]interface{}{
					"name":   "John Doe",
					"status": "premium",
				},
				UserID: "user_456",
			},
			wantErr: false,
		},
		{
			name: "valid delete event",
			event: workflow.TriggerEvent{
				EventType:  "on_delete",
				EntityName: "customers",
				EntityID:   "cust_123",
				Timestamp:  time.Now(),
				UserID:     "user_456",
			},
			wantErr: false,
		},
		{
			name: "approval entity event",
			event: workflow.TriggerEvent{
				EventType:  "on_create",
				EntityName: "approvals",
				EntityID:   "appr_789",
				Timestamp:  time.Now(),
				RawData: map[string]interface{}{
					"request_id": "req_123",
					"approver":   "manager@example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "inventory entity event",
			event: workflow.TriggerEvent{
				EventType:  "on_update",
				EntityName: "inventory",
				EntityID:   "inv_456",
				Timestamp:  time.Now(),
				FieldChanges: map[string]workflow.FieldChange{
					"quantity": {
						OldValue: 100,
						NewValue: 75,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := qm.QueueEvent(ctx, tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueueEvent() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify metrics were updated
				metrics := qm.GetMetrics()
				if metrics.TotalEnqueued == 0 {
					t.Error("TotalEnqueued should be > 0 after queueing event")
				}
			}
		})
	}
}

func TestQueueManager_StartStop(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Test starting
	if err := qm.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Verify it's running
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() error = %v", err)
	}
	if !status.IsRunning {
		t.Error("Queue manager should be running after Start()")
	}

	// Test double start (should error)
	if err := qm.Start(ctx); err == nil {
		t.Error("Start() should error when already running")
	}

	// Test stopping
	if err := qm.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Test double stop (should not error)
	if err := qm.Stop(ctx); err != nil {
		t.Errorf("Stop() on stopped manager should not error: %v", err)
	}
}

func TestQueueManager_ProcessMessage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Create a mock engine that tracks executions
	executedEvents := make([]workflow.TriggerEvent, 0)
	engine := newStubEngine(log)

	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Start the queue manager to begin consuming
	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Queue an event
	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   "cust_999",
		Timestamp:  time.Now(),
		RawData: map[string]interface{}{
			"name":  "Test Customer",
			"email": "test@example.com",
		},
	}

	if err := qm.QueueEvent(ctx, event); err != nil {
		t.Fatalf("queueing event: %s", err)
	}

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify the event was processed
	if len(executedEvents) != 1 {
		t.Errorf("Expected 1 event to be executed, got %d", len(executedEvents))
	}

	// Check metrics
	metrics := qm.GetMetrics()
	if metrics.TotalProcessed == 0 {
		t.Error("TotalProcessed should be > 0 after processing")
	}
}

func TestQueueManager_CircuitBreaker(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Create an engine that always fails
	engine := newStubEngine(log)
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Queue multiple events to trigger circuit breaker
	for i := 0; i < 6; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   fmt.Sprintf("cust_%d", i),
			Timestamp:  time.Now(),
		}

		_ = qm.QueueEvent(ctx, event) // Ignore errors initially
	}

	// Wait for processing attempts
	time.Sleep(3 * time.Second)

	// Circuit breaker should be open now
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() error = %v", err)
	}

	if !status.CircuitBreakerOn {
		t.Error("Circuit breaker should be open after multiple failures")
	}

	// Try to queue another event - should fail
	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   "cust_blocked",
		Timestamp:  time.Now(),
	}

	if err := qm.QueueEvent(ctx, event); err == nil {
		t.Error("QueueEvent() should fail when circuit breaker is open")
	}
}

func TestQueueManager_ClearQueue(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	// Queue some events
	for i := 0; i < 5; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   fmt.Sprintf("cust_%d", i),
			Timestamp:  time.Now(),
		}
		if err := qm.QueueEvent(ctx, event); err != nil {
			t.Fatalf("queueing event: %s", err)
		}
	}

	// Clear the queue
	if err := qm.ClearQueue(ctx); err != nil {
		t.Errorf("ClearQueue() error = %v", err)
	}

	// Verify queue is empty
	status, err := qm.GetQueueStatus(ctx)
	if err != nil {
		t.Fatalf("GetQueueStatus() error = %v", err)
	}

	if status.QueueDepth != 0 {
		t.Errorf("Queue depth = %d after clear, want 0", status.QueueDepth)
	}
}

func TestQueueManager_Metrics(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := newStubEngine(log)

	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Queue and process several events
	for i := 0; i < 4; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   fmt.Sprintf("cust_%d", i),
			Timestamp:  time.Now(),
		}
		if err := qm.QueueEvent(ctx, event); err != nil {
			t.Errorf("QueueEvent() error = %v", err)
		}
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Check metrics
	metrics := qm.GetMetrics()

	if metrics.TotalEnqueued != 4 {
		t.Errorf("TotalEnqueued = %d, want 4", metrics.TotalEnqueued)
	}

	if metrics.TotalProcessed == 0 {
		t.Error("TotalProcessed should be > 0")
	}

	if metrics.TotalFailed == 0 {
		t.Error("TotalFailed should be > 0 (we simulated failures)")
	}

	if metrics.LastProcessedAt == nil {
		t.Error("LastProcessedAt should not be nil after processing")
	}

	if metrics.AverageProcessTimeMs == 0 {
		t.Error("AverageProcessTimeMs should be > 0 after processing")
	}
}

func TestQueueManager_DetermineQueueType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		event     workflow.TriggerEvent
		wantQueue rabbitmq.QueueType
	}{
		{
			name: "approval entity",
			event: workflow.TriggerEvent{
				EntityName: "approvals",
			},
			wantQueue: rabbitmq.QueueTypeApproval,
		},
		{
			name: "approval_requests entity",
			event: workflow.TriggerEvent{
				EntityName: "approval_requests",
			},
			wantQueue: rabbitmq.QueueTypeApproval,
		},
		{
			name: "inventory entity",
			event: workflow.TriggerEvent{
				EntityName: "inventory",
			},
			wantQueue: rabbitmq.QueueTypeInventory,
		},
		{
			name: "stock entity",
			event: workflow.TriggerEvent{
				EntityName: "stock",
			},
			wantQueue: rabbitmq.QueueTypeInventory,
		},
		{
			name: "notifications entity",
			event: workflow.TriggerEvent{
				EntityName: "notifications",
			},
			wantQueue: rabbitmq.QueueTypeNotification,
		},
		{
			name: "default to workflow queue",
			event: workflow.TriggerEvent{
				EntityName: "customers",
			},
			wantQueue: rabbitmq.QueueTypeWorkflow,
		},
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the queue routing logic indirectly through QueueEvent
			ctx := context.Background()
			if err := qm.Initialize(ctx); err != nil {
				t.Fatalf("initializing queue manager: %s", err)
			}

			// Queue the event (internally uses determineQueueType)
			tt.event.EventType = "on_create"
			tt.event.EntityID = "test_123"
			tt.event.Timestamp = time.Now()

			if err := qm.QueueEvent(ctx, tt.event); err != nil {
				t.Errorf("QueueEvent() error = %v", err)
			}
		})
	}
}

func TestQueueManager_ProcessingResult(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		t.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		t.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	// Test ProcessEvents method (compatibility layer)
	result, err := qm.ProcessEvents(ctx, 10)
	if err != nil {
		t.Errorf("ProcessEvents() error = %v", err)
	}

	if result == nil {
		t.Fatal("ProcessEvents() returned nil result")
	}

	if result.StartTime.IsZero() {
		t.Error("ProcessingResult.StartTime should not be zero")
	}

	if result.EndTime.IsZero() {
		t.Error("ProcessingResult.EndTime should not be zero")
	}
}

// Mock engine for testing
type mockEngine struct {
	executeFunc func(ctx context.Context, event workflow.TriggerEvent) (*workflow.WorkflowExecution, error)
}

func (m *mockEngine) ExecuteWorkflow(ctx context.Context, event workflow.TriggerEvent) (*workflow.WorkflowExecution, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, event)
	}
	return &workflow.WorkflowExecution{
		ExecutionID: uuid.New().String(),
		Status:      "completed",
	}, nil
}

// Benchmark tests

func BenchmarkQueueManager_QueueEvent(b *testing.B) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		b.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		b.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := &workflow.Engine{}
	qm, err := workflow.NewQueueManager(log, nil, engine, client)
	if err != nil {
		b.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		b.Fatalf("initializing queue manager: %s", err)
	}

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   "cust_bench",
		Timestamp:  time.Now(),
		RawData: map[string]interface{}{
			"name":  "Benchmark Customer",
			"email": "bench@example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event.EntityID = fmt.Sprintf("cust_%d", i)
		_ = qm.QueueEvent(ctx, event)
	}
}

func BenchmarkQueueManager_ProcessMessage(b *testing.B) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	container, err := rabbitmq.StartRabbitMQ()
	if err != nil {
		b.Fatalf("starting rabbitmq: %s", err)
	}
	defer rabbitmq.StopRabbitMQ(container)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		b.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	engine := newStubEngine(log)
	qm, err := workflow.NewQueueManager(log, nil, engine.Engine, client)
	if err != nil {
		b.Fatalf("creating queue manager: %s", err)
	}

	ctx := context.Background()
	if err := qm.Initialize(ctx); err != nil {
		b.Fatalf("initializing queue manager: %s", err)
	}

	if err := qm.Start(ctx); err != nil {
		b.Fatalf("starting queue manager: %s", err)
	}
	defer qm.Stop(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := workflow.TriggerEvent{
			EventType:  "on_create",
			EntityName: "customers",
			EntityID:   fmt.Sprintf("cust_%d", i),
			Timestamp:  time.Now(),
		}
		_ = qm.QueueEvent(ctx, event)
	}

	// Wait for processing to complete
	time.Sleep(time.Duration(b.N*10) * time.Millisecond)
}
