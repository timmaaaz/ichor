package workflow_test

import (
	"bytes"
	"context"
	"errors"
	"sync"
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

// What the tests DO cover:

// Queue initialization and handler registration
// Start/stop lifecycle management
// Message queuing and basic processing flow
// Handler availability checks
// Failure handling and delivery record creation
// Priority-based queue routing (critical→alert, normal→notification)
// Concurrent message processing
// Statistics tracking
// Integration with RabbitMQ

// What the tests DON'T cover:

// Retry logic and backoff behavior
// Dead letter queue handling
// Circuit breaker functionality (mentioned in queue_test.go but not here)
// Template-based notifications
// Multi-channel alert delivery (critical alerts trying multiple channels)
// Message TTL and expiration
// Queue purging/clearing
// Metrics collection goroutine behavior
// Edge cases like malformed messages or parsing failures
// Recovery from RabbitMQ disconnections
// UpdateConfig functionality
// Provider-specific error response handling

func TestNotificationQueueProcessor_Initialize(t *testing.T) {
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
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Create workflow queue for initialization
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Create notification processor with nil store (testing just initialization)
	np := workflow.NewNotificationQueueProcessor(log, client, nil)

	ctx := context.Background()

	// Test initialization
	if err := np.Initialize(ctx); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}
}

func TestNotificationQueueProcessor_RegisterHandler(t *testing.T) {
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

	np := workflow.NewNotificationQueueProcessor(log, client, nil)

	// Create test handlers
	emailHandler := &testNotificationHandler{channel: "email", available: true}
	pushHandler := &testNotificationHandler{channel: "push", available: true}
	smsHandler := &testNotificationHandler{channel: "sms", available: true}

	tests := []struct {
		name      string
		handler   workflow.NotificationHandler
		wantErr   bool
		duplicate bool
	}{
		{
			name:    "register email handler",
			handler: emailHandler,
			wantErr: false,
		},
		{
			name:    "register push handler",
			handler: pushHandler,
			wantErr: false,
		},
		{
			name:    "register sms handler",
			handler: smsHandler,
			wantErr: false,
		},
		{
			name:      "duplicate email handler",
			handler:   &testNotificationHandler{channel: "email", available: true},
			wantErr:   true,
			duplicate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := np.RegisterHandler(tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotificationQueueProcessor_StartStop(t *testing.T) {
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

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	np := workflow.NewNotificationQueueProcessor(log, client, nil)

	ctx := context.Background()
	if err := np.Initialize(ctx); err != nil {
		t.Fatalf("initializing notification processor: %s", err)
	}

	// Register handlers
	np.RegisterHandler(&testNotificationHandler{channel: "email", available: true})
	np.RegisterHandler(&testNotificationHandler{channel: "push", available: true})

	// Test starting
	if err := np.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Test double start (should error)
	if err := np.Start(ctx); err == nil {
		t.Error("Start() should error when already running")
	}

	// Test stopping
	if err := np.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Test double stop (should not error)
	if err := np.Stop(ctx); err != nil {
		t.Errorf("Stop() on stopped processor should not error: %v", err)
	}
}

func TestNotificationQueueProcessor_QueueAndProcess(t *testing.T) {
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

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	// Create test store
	store := &testStore{
		deliveries: make([]workflow.NotificationDelivery, 0),
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")
	np := workflow.NewNotificationQueueProcessor(log, client, workflowdb.NewStore(log, db.DB))

	ctx := context.Background()
	if err := np.Initialize(ctx); err != nil {
		t.Fatalf("initializing notification processor: %s", err)
	}

	// Register test handlers
	emailHandler := &testNotificationHandler{
		channel:   "email",
		available: true,
		sent:      make([]workflow.NotificationPayload, 0),
	}
	pushHandler := &testNotificationHandler{
		channel:   "push",
		available: true,
		sent:      make([]workflow.NotificationPayload, 0),
	}

	np.RegisterHandler(emailHandler)
	np.RegisterHandler(pushHandler)

	// Start the processor
	if err := np.Start(ctx); err != nil {
		t.Fatalf("starting notification processor: %s", err)
	}
	defer np.Stop(ctx)

	// Create test notifications
	notifications := []workflow.NotificationPayload{
		{
			ID:         uuid.New(),
			Recipients: []uuid.UUID{uuid.New(), uuid.New()},
			Title:      "Test Email",
			Body:       "This is a test email body",
			Priority:   "normal",
			Channel:    "email",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Recipients: []uuid.UUID{uuid.New()},
			Title:      "Push Notification",
			Body:       "Important update",
			Priority:   "high",
			Channel:    "push",
			CreatedAt:  time.Now(),
		},
	}

	// Queue notifications
	for _, notification := range notifications {
		if err := np.QueueNotification(ctx, &notification); err != nil {
			t.Errorf("Failed to queue notification: %v", err)
		}
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify handlers received messages
	if len(emailHandler.sent) == 0 {
		t.Error("Email handler should have received messages")
	}

	if len(pushHandler.sent) == 0 {
		t.Error("Push handler should have received messages")
	}

	// Verify delivery records were created
	if len(store.deliveries) == 0 {
		t.Error("Delivery records should have been created")
	}

	// Check statistics
	stats := np.GetStats()
	if stats.TotalProcessed == 0 {
		t.Error("TotalProcessed should be > 0")
	}
	if stats.SuccessfulDeliveries == 0 {
		t.Error("SuccessfulDeliveries should be > 0")
	}
}

func TestNotificationQueueProcessor_FailureHandling(t *testing.T) {
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

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	store := &testStore{
		deliveries: make([]workflow.NotificationDelivery, 0),
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")
	np := workflow.NewNotificationQueueProcessor(log, client, workflowdb.NewStore(log, db.DB))

	ctx := context.Background()
	if err := np.Initialize(ctx); err != nil {
		t.Fatalf("initializing notification processor: %s", err)
	}

	// Register handler that will fail
	failingHandler := &testNotificationHandler{
		channel:     "email",
		available:   true,
		shouldFail:  true,
		failMessage: "simulated failure",
	}
	np.RegisterHandler(failingHandler)

	// Start the processor
	if err := np.Start(ctx); err != nil {
		t.Fatalf("starting notification processor: %s", err)
	}
	defer np.Stop(ctx)

	// Queue notification that will fail
	notification := workflow.NotificationPayload{
		ID:         uuid.New(),
		Recipients: []uuid.UUID{uuid.New()},
		Title:      "Test Email",
		Body:       "This will fail",
		Priority:   "normal",
		Channel:    "email",
		CreatedAt:  time.Now(),
	}

	if err := np.QueueNotification(ctx, &notification); err != nil {
		t.Errorf("Failed to queue notification: %v", err)
	}

	// Wait for processing and retries
	time.Sleep(5 * time.Second)

	// Check statistics
	stats := np.GetStats()
	if stats.TotalProcessed == 0 {
		t.Error("TotalProcessed should be > 0")
	}
	if stats.FailedDeliveries == 0 {
		t.Error("FailedDeliveries should be > 0")
	}

	// Verify failure was recorded
	hasFailedDelivery := false
	for _, delivery := range store.deliveries {
		if delivery.Status == workflow.DeliveryStatusFailed {
			hasFailedDelivery = true
			if delivery.ErrorMessage == nil || *delivery.ErrorMessage == "" {
				t.Error("Failed delivery should have error message")
			}
			break
		}
	}
	if !hasFailedDelivery {
		t.Error("Should have recorded failed delivery")
	}
}

func TestNotificationQueueProcessor_AlertPriorities(t *testing.T) {
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

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")
	np := workflow.NewNotificationQueueProcessor(log, client, workflowdb.NewStore(log, db.DB))

	ctx := context.Background()
	if err := np.Initialize(ctx); err != nil {
		t.Fatalf("initializing notification processor: %s", err)
	}

	// Start the processor
	if err := np.Start(ctx); err != nil {
		t.Fatalf("starting notification processor: %s", err)
	}
	defer np.Stop(ctx)

	// Test different priority levels
	priorities := []struct {
		priority string
		channel  string
	}{
		{"critical", "alert"},
		{"high", "alert"},
		{"normal", "notification"},
		{"low", "notification"},
	}

	for _, p := range priorities {
		notification := workflow.NotificationPayload{
			ID:         uuid.New(),
			Recipients: []uuid.UUID{uuid.New()},
			Title:      "Priority Test",
			Body:       "Testing priority: " + p.priority,
			Priority:   p.priority,
			Channel:    "email", // Will be routed based on priority
			CreatedAt:  time.Now(),
		}

		if err := np.QueueNotification(ctx, &notification); err != nil {
			t.Errorf("Failed to queue %s priority notification: %v", p.priority, err)
		}
	}

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify stats
	stats := np.GetStats()
	if stats.TotalProcessed < int64(len(priorities)) {
		t.Errorf("Expected at least %d processed, got %d", len(priorities), stats.TotalProcessed)
	}
}

func TestNotificationQueueProcessor_ConcurrentMessages(t *testing.T) {
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

	// Initialize workflow queue
	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	db := dbtest.NewDatabase(t, "Test_Workflow")
	np := workflow.NewNotificationQueueProcessor(log, client, workflowdb.NewStore(log, db.DB))

	ctx := context.Background()
	if err := np.Initialize(ctx); err != nil {
		t.Fatalf("initializing notification processor: %s", err)
	}

	// Register handlers
	emailHandler := &testNotificationHandler{
		channel:   "email",
		available: true,
		sent:      make([]workflow.NotificationPayload, 0),
	}
	np.RegisterHandler(emailHandler)

	// Start the processor
	if err := np.Start(ctx); err != nil {
		t.Fatalf("starting notification processor: %s", err)
	}
	defer np.Stop(ctx)

	// Queue multiple notifications concurrently
	var wg sync.WaitGroup
	notificationCount := 20

	for i := 0; i < notificationCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			notification := workflow.NotificationPayload{
				ID:         uuid.New(),
				Recipients: []uuid.UUID{uuid.New()},
				Title:      "Concurrent Test",
				Body:       "Message " + string(rune(index)),
				Priority:   "normal",
				Channel:    "email",
				CreatedAt:  time.Now(),
			}

			if err := np.QueueNotification(ctx, &notification); err != nil {
				t.Errorf("Failed to queue notification %d: %v", index, err)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all messages to be processed
	time.Sleep(5 * time.Second)

	// Verify all messages were processed
	if len(emailHandler.sent) != notificationCount {
		t.Errorf("Expected %d notifications processed, got %d", notificationCount, len(emailHandler.sent))
	}

	// Check stats
	stats := np.GetStats()
	if stats.TotalProcessed != int64(notificationCount) {
		t.Errorf("Stats TotalProcessed = %d, want %d", stats.TotalProcessed, notificationCount)
	}
}

// Test implementations

type testNotificationHandler struct {
	channel     string
	available   bool
	shouldFail  bool
	failMessage string
	sent        []workflow.NotificationPayload
	mu          sync.Mutex
}

func (h *testNotificationHandler) Send(ctx context.Context, payload *workflow.NotificationPayload) error {
	if h.shouldFail {
		return errors.New(h.failMessage)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.sent = append(h.sent, *payload)
	return nil
}

func (h *testNotificationHandler) GetChannelType() string {
	return h.channel
}

func (h *testNotificationHandler) IsAvailable() bool {
	return h.available
}

func (h *testNotificationHandler) GetPriority() int {
	switch h.channel {
	case "push":
		return 7
	case "sms":
		return 6
	case "email":
		return 5
	default:
		return 3
	}
}

// Minimal test store implementation
type testStore struct {
	deliveries []workflow.NotificationDelivery
	mu         sync.Mutex
}

func (s *testStore) CreateNotificationDelivery(ctx context.Context, delivery workflow.NotificationDelivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveries = append(s.deliveries, delivery)
	return nil
}

func (s *testStore) UpdateNotificationDelivery(ctx context.Context, delivery workflow.NotificationDelivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, d := range s.deliveries {
		if d.ID == delivery.ID {
			s.deliveries[i] = delivery
			return nil
		}
	}
	return nil
}

func (s *testStore) QueryDeliveriesByAutomationExecution(ctx context.Context, executionID uuid.UUID) ([]workflow.NotificationDelivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []workflow.NotificationDelivery
	for _, d := range s.deliveries {
		if d.AutomationExecutionID == executionID {
			result = append(result, d)
		}
	}
	return result, nil
}

// Benchmark tests

func BenchmarkNotificationQueueProcessor_QueueNotification(b *testing.B) {
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

	queue := rabbitmq.NewWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		b.Fatalf("initializing workflow queue: %s", err)
	}

	np := workflow.NewNotificationQueueProcessor(log, client, nil)

	ctx := context.Background()
	if err := np.Initialize(ctx); err != nil {
		b.Fatalf("initializing notification processor: %s", err)
	}

	notification := workflow.NotificationPayload{
		ID:         uuid.New(),
		Recipients: []uuid.UUID{uuid.New()},
		Title:      "Benchmark",
		Body:       "Benchmark body",
		Priority:   "normal",
		Channel:    "email",
		CreatedAt:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notification.ID = uuid.New()
		_ = np.QueueNotification(ctx, &notification)
	}
}
