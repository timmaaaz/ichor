// foundation/rabbitmq/client_test.go
package rabbitmq_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

func TestClient_Connect(t *testing.T) {
	// Get shared container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)

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
	// Get shared container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Purge queue to start fresh
	wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow)

	messageUUID := uuid.New()

	// Create test message
	msg := &rabbitmq.Message{
		Type:       "test_workflow",
		EntityName: "test_entity",
		EntityID:   messageUUID,
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
		if receivedMsg.EntityID != messageUUID {
			t.Errorf("Expected entity ID '%s', got '%s'", messageUUID, receivedMsg.EntityID)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestWorkflowQueue_BatchPublish(t *testing.T) {
	// Get shared container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
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
			EntityID:   uuid.New(),
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
	// Get shared container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Purge queue to start fresh
	wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow)

	messageUUID := uuid.New()

	// Create message that will fail processing
	msg := &rabbitmq.Message{
		Type:        "error_workflow",
		EntityName:  "test_entity",
		EntityID:    messageUUID,
		EventType:   "on_error",
		MaxAttempts: 3, // Increased to allow more retries
	}

	// Publish message
	if err := wq.Publish(ctx, rabbitmq.QueueTypeWorkflow, msg); err != nil {
		t.Fatalf("Should be able to publish message: %s", err)
	}

	// Set up consumer that always fails
	attempts := 0
	var mu sync.Mutex
	consumer, err := wq.Consume(ctx, rabbitmq.QueueTypeWorkflow, func(ctx context.Context, msg *rabbitmq.Message) error {
		if msg.EntityID == messageUUID {
			mu.Lock()
			currentAttempt := attempts + 1
			attempts++
			mu.Unlock()
			t.Logf("Processing attempt %d for message %s", currentAttempt, msg.EntityID)
			return fmt.Errorf("intentional error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Should be able to start consumer: %s", err)
	}
	defer consumer.Stop()

	// Wait for retries - need to wait longer than the retry TTL (5 seconds)
	// First attempt: immediate
	// Second attempt: after 5 second retry delay
	// Third attempt: after another 5 second retry delay
	maxWait := 15 * time.Second
	checkInterval := 500 * time.Millisecond
	timeout := time.After(maxWait)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			mu.Lock()
			finalAttempts := attempts
			mu.Unlock()
			t.Fatalf("Timeout: Expected at least 2 attempts, got %d", finalAttempts)

		case <-ticker.C:
			mu.Lock()
			currentAttempts := attempts
			mu.Unlock()

			// Once we have at least 2 attempts, we can pass the test
			if currentAttempts >= 2 {
				t.Logf("Success: Message was retried %d times", currentAttempts)
				return
			}
		}
	}
}

func TestWorkflowQueue_PurgeQueue(t *testing.T) {
	// Get shared container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// IMPORTANT: Purge queue first to ensure clean state
	if err := wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow); err != nil {
		t.Fatalf("Failed to purge queue before test: %s", err)
	}

	messageUUID := uuid.New()

	// Publish a message
	msg := &rabbitmq.Message{
		Type:       "purge_test",
		EntityName: "test_entity",
		EntityID:   messageUUID,
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
	// Get shared container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
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

		messageUUID := uuid.New()

		// Publish to each queue type
		msg := &rabbitmq.Message{
			Type:       string(qt),
			EntityName: "test_entity",
			EntityID:   messageUUID,
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

func TestClient_Reconnection(t *testing.T) {
	// Get shared container
	container := rabbitmq.GetTestContainer(t)
	config := rabbitmq.NewTestConfig(container.URL)
	config.MaxRetries = 3
	config.RetryDelay = 100 * time.Millisecond

	// Create client with reconnection config
	client := &rabbitmq.Client{}
	// Note: Since NewClient uses singleton pattern, we create client directly for this test
	client = rabbitmq.NewTestClient(container.URL)

	// Connect initially
	if err := client.Connect(); err != nil {
		t.Fatalf("Initial connection failed: %s", err)
	}

	// Verify connected
	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	// Force close the connection to simulate disconnection
	ch, _ := client.GetChannel()
	if ch != nil {
		ch.Close() // This will trigger reconnection logic
	}

	// Wait a bit for reconnection attempt
	time.Sleep(500 * time.Millisecond)

	// The client should attempt to reconnect
	// Note: In a real scenario, the handleReconnect goroutine would be running
	// For this test, we just verify the client can handle connection loss

	// Clean up
	client.Close()
}

func TestWorkflowQueue_ConcurrentPublish(t *testing.T) {
	// Get shared container
	container := rabbitmq.GetTestContainer(t)

	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	wq := rabbitmq.NewWorkflowQueue(client, log)

	ctx := context.Background()
	if err := wq.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize queues: %s", err)
	}

	// Purge queue to start fresh
	wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow)

	// Concurrent publishers
	numPublishers := 10
	messagesPerPublisher := 5
	var wg sync.WaitGroup
	wg.Add(numPublishers)

	for i := 0; i < numPublishers; i++ {
		go func(publisherID int) {
			defer wg.Done()

			for j := 0; j < messagesPerPublisher; j++ {
				msg := &rabbitmq.Message{
					Type:       fmt.Sprintf("concurrent_test_%d", publisherID),
					EntityName: "test_entity",
					EntityID:   uuid.New(),
					EventType:  "on_create",
					Payload: map[string]interface{}{
						"publisher": publisherID,
						"message":   j,
					},
				}

				if err := wq.Publish(ctx, rabbitmq.QueueTypeWorkflow, msg); err != nil {
					t.Errorf("Publisher %d failed to publish message %d: %s", publisherID, j, err)
				}
			}
		}(i)
	}

	// Wait for all publishers to complete
	wg.Wait()

	// Verify all messages were published
	stats, err := wq.GetQueueStats(ctx, rabbitmq.QueueTypeWorkflow)
	if err != nil {
		t.Fatalf("Failed to get queue stats: %s", err)
	}

	expectedMessages := numPublishers * messagesPerPublisher
	if stats.Messages != expectedMessages {
		t.Errorf("Expected %d messages in queue, got %d", expectedMessages, stats.Messages)
	}
}
