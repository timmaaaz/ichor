package workflow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// QueueManager manages workflow event queuing and processing
type QueueManager struct {
	log    *logger.Logger
	db     *sqlx.DB
	engine *Engine
	queue  *rabbitmq.WorkflowQueue
	client *rabbitmq.Client

	// Configuration
	config QueueConfig

	// State management
	mu           sync.RWMutex
	isRunning    atomic.Bool
	consumers    map[string]*rabbitmq.Consumer
	stopChan     chan struct{}
	processingWG sync.WaitGroup

	// Metrics
	metrics     QueueMetrics
	metricsLock sync.RWMutex

	// Circuit breaker
	circuitBreaker *CircuitBreaker
}

// QueueConfig holds configuration for the queue manager
type QueueConfig struct {
	BatchSize               int           `json:"batch_size"`
	MaxConcurrentWorkers    int           `json:"max_concurrent_workers"`
	ProcessingTimeout       time.Duration `json:"processing_timeout"`
	RetryDelay              time.Duration `json:"retry_delay"`
	MaxRetryDelay           time.Duration `json:"max_retry_delay"`
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
}

// QueueMetrics tracks queue performance metrics
type QueueMetrics struct {
	TotalEnqueued        int64      `json:"total_enqueued"`
	TotalProcessed       int64      `json:"total_processed"`
	TotalFailed          int64      `json:"total_failed"`
	TotalRetried         int64      `json:"total_retried"`
	TotalDeadLettered    int64      `json:"total_dead_lettered"`
	AverageProcessTimeMs int64      `json:"average_process_time_ms"`
	LastProcessedAt      *time.Time `json:"last_processed_at"`
	CurrentQueueDepth    int        `json:"current_queue_depth"`
	ActiveWorkers        int        `json:"active_workers"`
}

// QueueStatus represents the current queue status
type QueueStatus struct {
	IsRunning        bool            `json:"is_running"`
	QueueDepth       int             `json:"queue_depth"`
	ActiveWorkers    int             `json:"active_workers"`
	CircuitBreakerOn bool            `json:"circuit_breaker_on"`
	Metrics          QueueMetrics    `json:"metrics"`
	ConsumerStatus   map[string]bool `json:"consumer_status"`
}

// CircuitBreaker manages downstream service failures
type CircuitBreaker struct {
	mu               sync.RWMutex
	failureThreshold int
	resetTimeout     time.Duration
	failureCount     atomic.Int32
	lastFailureTime  atomic.Value // stores time.Time
	state            atomic.Value // stores string: "closed", "open", "half-open"
}

// QueueProcessingResult represents the result of batch processing
type QueueProcessingResult struct {
	ProcessedCount int
	FailedCount    int
	StartTime      time.Time
	EndTime        time.Time
}

// DefaultQueueConfig returns default configuration
func DefaultQueueConfig() QueueConfig {
	return QueueConfig{
		BatchSize:               10,
		MaxConcurrentWorkers:    5,
		ProcessingTimeout:       5 * time.Minute,
		RetryDelay:              1 * time.Second,
		MaxRetryDelay:           30 * time.Second,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
	}
}

// NewQueueManager creates a new queue manager
func NewQueueManager(log *logger.Logger, db *sqlx.DB, engine *Engine, client *rabbitmq.Client) (*QueueManager, error) {
	qm := &QueueManager{
		log:       log,
		db:        db,
		engine:    engine,
		client:    client,
		queue:     rabbitmq.NewWorkflowQueue(client, log),
		config:    DefaultQueueConfig(),
		consumers: make(map[string]*rabbitmq.Consumer),
		stopChan:  make(chan struct{}),
		circuitBreaker: &CircuitBreaker{
			failureThreshold: 5,
			resetTimeout:     30 * time.Second,
		},
	}

	qm.circuitBreaker.state.Store("closed")
	qm.circuitBreaker.lastFailureTime.Store(time.Now())

	return qm, nil
}

// Initialize sets up the queue infrastructure
func (qm *QueueManager) Initialize(ctx context.Context) error {
	qm.log.Info(ctx, "Initializing queue manager...")

	// Initialize workflow queues
	if err := qm.queue.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize queues: %w", err)
	}

	qm.log.Info(ctx, "Queue manager initialized successfully")
	return nil
}

// QueueEvent adds a trigger event to the queue
func (qm *QueueManager) QueueEvent(ctx context.Context, event TriggerEvent) error {
	// Check circuit breaker
	if qm.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker is open, refusing new events")
	}

	// Convert TriggerEvent to RabbitMQ Message
	msg := &rabbitmq.Message{
		ID:         uuid.New(),
		Type:       "workflow_trigger",
		EntityName: event.EntityName,
		EntityID:   event.EntityID,
		EventType:  event.EventType,
		Payload: map[string]interface{}{
			"field_changes": event.FieldChanges,
			"raw_data":      event.RawData,
			"timestamp":     event.Timestamp,
			"user_id":       event.UserID,
		},
		Priority:     uint8(rabbitmq.PriorityNormal),
		MaxAttempts:  3,
		CreatedAt:    time.Now(),
		ScheduledFor: time.Now(),
	}

	// Determine queue type based on event
	queueType := qm.determineQueueType(event)

	// Publish to RabbitMQ
	if err := qm.queue.Publish(ctx, queueType, msg); err != nil {
		qm.recordFailure()
		return fmt.Errorf("failed to queue event: %w", err)
	}

	// Update metrics
	qm.updateMetric(func(m *QueueMetrics) {
		m.TotalEnqueued++
	})

	qm.log.Info(ctx, "Queued workflow event",
		"eventID", msg.ID,
		"entityName", event.EntityName,
		"eventType", event.EventType,
		"queueType", queueType)

	return nil
}

// ProcessEvents processes queued events (batch processing for compatibility)
func (qm *QueueManager) ProcessEvents(ctx context.Context, batchSize int) (*QueueProcessingResult, error) {
	if !qm.isRunning.Load() {
		return nil, fmt.Errorf("queue manager is not running")
	}

	if batchSize <= 0 {
		batchSize = qm.config.BatchSize
	}

	result := &QueueProcessingResult{
		ProcessedCount: 0,
		FailedCount:    0,
		StartTime:      time.Now(),
	}

	// Since RabbitMQ handles consumption differently, we'll track processing
	// through our consumers which are already running
	stats := qm.GetMetrics()
	result.ProcessedCount = int(stats.TotalProcessed)
	result.FailedCount = int(stats.TotalFailed)
	result.EndTime = time.Now()

	return result, nil
}

// Start begins processing events
func (qm *QueueManager) Start(ctx context.Context) error {
	if qm.isRunning.Load() {
		return fmt.Errorf("queue manager already running")
	}

	qm.isRunning.Store(true)
	qm.log.Info(ctx, "Starting queue manager")

	// Start consumers for each queue type
	queueTypes := []rabbitmq.QueueType{
		rabbitmq.QueueTypeWorkflow,
		rabbitmq.QueueTypeApproval,
		rabbitmq.QueueTypeNotification,
		rabbitmq.QueueTypeInventory,
		rabbitmq.QueueTypeEmail,
		rabbitmq.QueueTypeAlert,
	}

	for _, qt := range queueTypes {
		if err := qm.startConsumer(ctx, qt); err != nil {
			qm.log.Error(ctx, "Failed to start consumer",
				"queueType", qt,
				"error", err)
		}
	}

	// Start metrics collector
	qm.processingWG.Add(1)
	go qm.metricsCollector(ctx)

	return nil
}

// startConsumer starts a consumer for a specific queue type
func (qm *QueueManager) startConsumer(ctx context.Context, queueType rabbitmq.QueueType) error {
	consumer, err := qm.queue.Consume(ctx, queueType, func(ctx context.Context, msg *rabbitmq.Message) error {
		return qm.processMessage(ctx, msg)
	})

	if err != nil {
		return fmt.Errorf("failed to start consumer for %s: %w", queueType, err)
	}

	qm.mu.Lock()
	qm.consumers[string(queueType)] = consumer
	qm.mu.Unlock()

	return nil
}

// processMessage processes a single message from the queue
func (qm *QueueManager) processMessage(ctx context.Context, msg *rabbitmq.Message) error {
	startTime := time.Now()

	// Check circuit breaker
	if qm.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker is open")
	}

	// Convert RabbitMQ message back to TriggerEvent
	event := TriggerEvent{
		EventType:  msg.EventType,
		EntityName: msg.EntityName,
		EntityID:   msg.EntityID,
	}

	// Extract payload data with proper type assertions
	if payload, ok := msg.Payload["field_changes"]; ok {
		// field_changes might be a map[string]interface{} that needs conversion
		if fcMap, ok := payload.(map[string]interface{}); ok {
			fieldChanges := make(map[string]FieldChange)
			for key, value := range fcMap {
				if changeData, ok := value.(map[string]interface{}); ok {
					fieldChanges[key] = FieldChange{
						OldValue: changeData["old_value"],
						NewValue: changeData["new_value"],
					}
				}
			}
			event.FieldChanges = fieldChanges
		}
	}

	if payload, ok := msg.Payload["raw_data"]; ok {
		// raw_data needs to be asserted as map[string]interface{}
		if rawData, ok := payload.(map[string]interface{}); ok {
			event.RawData = rawData
		}
	}

	if payload, ok := msg.Payload["timestamp"]; ok {
		switch v := payload.(type) {
		case time.Time:
			event.Timestamp = v
		case string:
			// Try to parse string timestamp
			if ts, err := time.Parse(time.RFC3339, v); err == nil {
				event.Timestamp = ts
			}
		}
	}

	if payload, ok := msg.Payload["user_id"]; ok {
		// if uid, ok := payload.(string); ok {
		// 	event.UserID = uid
		// }
		uid, err := uuid.Parse(payload.(string))
		if err == nil {
			event.UserID = uid
		}
	}

	// Create processing context with timeout
	processCtx, cancel := context.WithTimeout(ctx, qm.config.ProcessingTimeout)
	defer cancel()

	// Execute workflow
	_, err := qm.engine.ExecuteWorkflow(processCtx, event)

	processingTime := time.Since(startTime)

	if err != nil {
		qm.recordFailure()
		qm.updateMetric(func(m *QueueMetrics) {
			m.TotalFailed++
		})

		qm.log.Error(ctx, "Failed to process workflow event",
			"messageID", msg.ID,
			"error", err,
			"processingTime", processingTime)

		// Let RabbitMQ handle retries through its built-in mechanism
		return err
	}

	// Success
	qm.recordSuccess()
	qm.updateMetric(func(m *QueueMetrics) {
		m.TotalProcessed++
		now := time.Now()
		m.LastProcessedAt = &now

		// Update average processing time
		oldAvg := m.AverageProcessTimeMs
		totalTime := oldAvg * (m.TotalProcessed - 1)
		m.AverageProcessTimeMs = (totalTime + processingTime.Milliseconds()) / m.TotalProcessed
	})

	qm.log.Info(ctx, "Processed workflow event",
		"messageID", msg.ID,
		"processingTime", processingTime)

	return nil
}

// Stop gracefully stops the queue manager
func (qm *QueueManager) Stop(ctx context.Context) error {
	if !qm.isRunning.Load() {
		return nil
	}

	qm.isRunning.Store(false)
	qm.log.Info(ctx, "Stopping queue manager")

	// Stop all consumers
	qm.mu.Lock()
	for name, consumer := range qm.consumers {
		if err := consumer.Stop(); err != nil {
			qm.log.Error(ctx, "Failed to stop consumer",
				"name", name,
				"error", err)
		}
	}
	qm.consumers = make(map[string]*rabbitmq.Consumer)
	qm.mu.Unlock()

	// Signal stop
	close(qm.stopChan)

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		qm.processingWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		qm.log.Info(ctx, "Queue manager stopped gracefully")
	case <-time.After(30 * time.Second):
		qm.log.Warn(ctx, "Queue manager stop timeout")
	}

	return nil
}

// GetQueueStatus returns current queue status
func (qm *QueueManager) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	consumerStatus := make(map[string]bool)
	for name := range qm.consumers {
		consumerStatus[name] = true
	}

	// Get queue depth from RabbitMQ
	stats, err := qm.queue.GetQueueStats(ctx, rabbitmq.QueueTypeWorkflow)
	if err != nil {
		return nil, err
	}

	metrics := qm.GetMetrics()
	metrics.CurrentQueueDepth = stats.Messages

	return &QueueStatus{
		IsRunning:        qm.isRunning.Load(),
		QueueDepth:       stats.Messages,
		ActiveWorkers:    len(qm.consumers),
		CircuitBreakerOn: qm.circuitBreaker.IsOpen(),
		Metrics:          metrics,
		ConsumerStatus:   consumerStatus,
	}, nil
}

// ClearQueue removes all messages from the queue (for testing/maintenance)
func (qm *QueueManager) ClearQueue(ctx context.Context) error {
	queueTypes := []rabbitmq.QueueType{
		rabbitmq.QueueTypeWorkflow,
		rabbitmq.QueueTypeApproval,
		rabbitmq.QueueTypeNotification,
		rabbitmq.QueueTypeInventory,
		rabbitmq.QueueTypeEmail,
		rabbitmq.QueueTypeAlert,
	}

	for _, qt := range queueTypes {
		if err := qm.queue.PurgeQueue(ctx, qt); err != nil {
			return fmt.Errorf("failed to purge queue %s: %w", qt, err)
		}
	}

	qm.log.Info(ctx, "All queues cleared")
	return nil
}

// GetMetrics returns current queue metrics
func (qm *QueueManager) GetMetrics() QueueMetrics {
	qm.metricsLock.RLock()
	defer qm.metricsLock.RUnlock()
	return qm.metrics
}

// Helper methods

func (qm *QueueManager) determineQueueType(event TriggerEvent) rabbitmq.QueueType {
	// Determine queue type based on event characteristics
	// This could be enhanced with more sophisticated routing logic

	// Check if it's an approval-related event
	if event.EntityName == "approvals" || event.EntityName == "approval_requests" {
		return rabbitmq.QueueTypeApproval
	}

	// Check if it's an inventory-related event
	if event.EntityName == "inventory" || event.EntityName == "stock" {
		return rabbitmq.QueueTypeInventory
	}

	// Check if it's a notification event
	if event.EntityName == "notifications" {
		return rabbitmq.QueueTypeNotification
	}

	// Default to general workflow queue
	return rabbitmq.QueueTypeWorkflow
}

func (qm *QueueManager) updateMetric(fn func(*QueueMetrics)) {
	qm.metricsLock.Lock()
	defer qm.metricsLock.Unlock()
	fn(&qm.metrics)
}

func (qm *QueueManager) metricsCollector(ctx context.Context) {
	defer qm.processingWG.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-qm.stopChan:
			return
		case <-ticker.C:
			// Collect queue stats
			if stats, err := qm.queue.GetQueueStats(ctx, rabbitmq.QueueTypeWorkflow); err == nil {
				qm.updateMetric(func(m *QueueMetrics) {
					m.CurrentQueueDepth = stats.Messages
					m.ActiveWorkers = len(qm.consumers)
				})
			}
		}
	}
}

// Circuit breaker methods

func (cb *CircuitBreaker) IsOpen() bool {
	state := cb.state.Load().(string)
	if state == "closed" {
		return false
	}

	if state == "open" {
		lastFailure := cb.lastFailureTime.Load().(time.Time)
		if time.Since(lastFailure) > cb.resetTimeout {
			cb.state.Store("half-open")
			return false
		}
		return true
	}

	// half-open
	return false
}

func (qm *QueueManager) recordFailure() {
	count := qm.circuitBreaker.failureCount.Add(1)
	qm.circuitBreaker.lastFailureTime.Store(time.Now())

	if count >= int32(qm.circuitBreaker.failureThreshold) {
		qm.circuitBreaker.state.Store("open")
		qm.log.Warn(context.Background(), "Circuit breaker opened",
			"failures", count)
	}
}

func (qm *QueueManager) recordSuccess() {
	state := qm.circuitBreaker.state.Load().(string)
	if state == "half-open" {
		qm.circuitBreaker.failureCount.Store(0)
		qm.circuitBreaker.state.Store("closed")
		qm.log.Info(context.Background(), "Circuit breaker closed")
	}
}
