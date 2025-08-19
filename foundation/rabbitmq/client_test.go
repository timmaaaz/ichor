// foundation/rabbitmq/client_test.go
package rabbitmq_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

var testContainer rabbitmq.Container

// TestMain manages the RabbitMQ container lifecycle for all tests
func TestMain(m *testing.M) {
	// Parse flags first - this is required before calling testing.Short()
	flag.Parse()

	// Skip container setup in short mode
	if testing.Short() {
		os.Exit(m.Run())
	}

	// Start RabbitMQ container
	var err error
	testContainer, err = rabbitmq.StartRabbitMQ()
	if err != nil {
		fmt.Printf("Failed to start RabbitMQ container: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := rabbitmq.StopRabbitMQ(testContainer); err != nil {
		fmt.Printf("Failed to stop RabbitMQ container: %v\n", err)
	}

	os.Exit(code)
}

func TestClient_Connect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RabbitMQ integration test in short mode")
	}

	client := rabbitmq.NewTestClient(testContainer.URL)

	// Test connection
	if err := client.Connect(); err != nil {
		t.Fatalf("Should be able to connect: %s", err)
	}

	// Check connection status
	if !client.IsConnected() {
		t.Error("Client should be connected after successful Connect()")
	}

	// Test getting channel
	ch, err := client.GetChannel()
	if err != nil {
		t.Fatalf("Should be able to get channel: %s", err)
	}
	if ch == nil {
		t.Error("Channel should not be nil")
	}

	// Clean up
	if err := client.Close(); err != nil {
		t.Fatalf("Should be able to close connection: %s", err)
	}

	// Check connection status after close
	if client.IsConnected() {
		t.Error("Client should not be connected after Close()")
	}
}

func TestWorkflowQueue_PublishAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RabbitMQ integration test in short mode")
	}

	client := rabbitmq.NewTestClient(testContainer.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Purge queue to start fresh
	wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow)

	// Create test message
	msg := &rabbitmq.Message{
		Type:       "test_workflow",
		EntityName: "test_entity",
		EntityID:   "test_123",
		EventType:  "on_create",
		Priority:   uint8(rabbitmq.PriorityNormal),
		Payload: map[string]interface{}{
			"test_field": "test_value",
		},
	}

	// Publish message
	if err := wq.Publish(ctx, rabbitmq.QueueTypeWorkflow, msg); err != nil {
		t.Fatalf("Should be able to publish message: %s", err)
	}

	// Set up consumer
	received := make(chan *rabbitmq.Message, 1)
	consumer, err := wq.Consume(ctx, rabbitmq.QueueTypeWorkflow, func(ctx context.Context, msg *rabbitmq.Message) error {
		received <- msg
		return nil
	})
	if err != nil {
		t.Fatalf("Should be able to start consumer: %s", err)
	}
	defer consumer.Stop()

	// Wait for message
	select {
	case receivedMsg := <-received:
		if receivedMsg.Type != "test_workflow" {
			t.Errorf("Expected message type 'test_workflow', got '%s'", receivedMsg.Type)
		}
		if receivedMsg.EntityID != "test_123" {
			t.Errorf("Expected entity ID 'test_123', got '%s'", receivedMsg.EntityID)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestWorkflowQueue_BatchPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RabbitMQ integration test in short mode")
	}

	client := rabbitmq.NewTestClient(testContainer.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Purge queue to start fresh
	wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow)

	// Create batch of messages
	messages := make([]*rabbitmq.Message, 5)
	for i := 0; i < 5; i++ {
		messages[i] = &rabbitmq.Message{
			Type:       "batch_test",
			EntityName: "test_entity",
			EntityID:   fmt.Sprintf("batch_%d", i),
			EventType:  "on_create",
		}
	}

	// Publish batch
	if err := wq.PublishBatch(ctx, rabbitmq.QueueTypeWorkflow, messages); err != nil {
		t.Fatalf("Should be able to publish batch: %s", err)
	}

	// Check queue stats
	stats, err := wq.GetQueueStats(ctx, rabbitmq.QueueTypeWorkflow)
	if err != nil {
		t.Fatalf("Should be able to get queue stats: %s", err)
	}

	if stats.Messages != 5 {
		t.Errorf("Expected 5 messages in queue, got %d", stats.Messages)
	}
}

func TestWorkflowQueue_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RabbitMQ integration test in short mode")
	}

	client := rabbitmq.NewTestClient(testContainer.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Purge queue to start fresh
	wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow)

	// Create message that will fail processing
	msg := &rabbitmq.Message{
		Type:        "error_workflow",
		EntityName:  "test_entity",
		EntityID:    "error_123",
		EventType:   "on_error",
		MaxAttempts: 2,
	}

	// Publish message
	if err := wq.Publish(ctx, rabbitmq.QueueTypeWorkflow, msg); err != nil {
		t.Fatalf("Should be able to publish message: %s", err)
	}

	// Set up consumer that always fails
	attempts := 0
	var mu sync.Mutex
	consumer, err := wq.Consume(ctx, rabbitmq.QueueTypeWorkflow, func(ctx context.Context, msg *rabbitmq.Message) error {
		if msg.EntityID == "error_123" {
			mu.Lock()
			attempts++
			mu.Unlock()
			return fmt.Errorf("intentional error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Should be able to start consumer: %s", err)
	}
	defer consumer.Stop()

	// Wait for retries
	time.Sleep(3 * time.Second)

	// Should have attempted at least twice
	mu.Lock()
	finalAttempts := attempts
	mu.Unlock()

	if finalAttempts < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", finalAttempts)
	}
}

func TestWorkflowQueue_PurgeQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RabbitMQ integration test in short mode")
	}

	client := rabbitmq.NewTestClient(testContainer.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// IMPORTANT: Purge queue first to ensure clean state
	if err := wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow); err != nil {
		t.Fatalf("Failed to purge queue before test: %s", err)
	}

	// Publish a message
	msg := &rabbitmq.Message{
		Type:       "purge_test",
		EntityName: "test_entity",
		EntityID:   "purge_123",
		EventType:  "on_create",
	}

	if err := wq.Publish(ctx, rabbitmq.QueueTypeWorkflow, msg); err != nil {
		t.Fatalf("Should be able to publish message: %s", err)
	}

	// Verify message exists
	stats, err := wq.GetQueueStats(ctx, rabbitmq.QueueTypeWorkflow)
	if err != nil {
		t.Fatalf("Should be able to get queue stats: %s", err)
	}

	if stats.Messages != 1 {
		t.Errorf("Expected 1 message in queue before purge, got %d", stats.Messages)
	}

	// Purge queue
	if err := wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow); err != nil {
		t.Fatalf("Should be able to purge queue: %s", err)
	}

	// Verify queue is empty
	stats, err = wq.GetQueueStats(ctx, rabbitmq.QueueTypeWorkflow)
	if err != nil {
		t.Fatalf("Should be able to get queue stats after purge: %s", err)
	}

	if stats.Messages != 0 {
		t.Errorf("Expected 0 messages in queue after purge, got %d", stats.Messages)
	}
}

func TestWorkflowQueue_MultipleQueueTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RabbitMQ integration test in short mode")
	}

	client := rabbitmq.NewTestClient(testContainer.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Test different queue types
	queueTypes := []rabbitmq.QueueType{
		rabbitmq.QueueTypeWorkflow,
		rabbitmq.QueueTypeApproval,
		rabbitmq.QueueTypeNotification,
	}

	for _, qt := range queueTypes {
		// Purge first
		wq.PurgeQueue(ctx, qt)

		// Publish to each queue type
		msg := &rabbitmq.Message{
			Type:       string(qt),
			EntityName: "test_entity",
			EntityID:   fmt.Sprintf("%s_123", qt),
			EventType:  "on_create",
		}

		if err := wq.Publish(ctx, qt, msg); err != nil {
			t.Errorf("Should be able to publish to queue type %s: %s", qt, err)
		}

		// Verify stats for each queue
		stats, err := wq.GetQueueStats(ctx, qt)
		if err != nil {
			t.Errorf("Should be able to get stats for queue type %s: %s", qt, err)
		}

		if stats.Type != string(qt) {
			t.Errorf("Expected queue type %s, got %s", qt, stats.Type)
		}
	}
}
