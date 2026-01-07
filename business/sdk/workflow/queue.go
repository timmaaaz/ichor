package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/websocket"
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

	// Circuit breaker manager (per-queue-type + global)
	circuitBreakerManager *CircuitBreakerManager

	// Handler registry for real-time message delivery (e.g., WebSocket alerts)
	handlerRegistry *websocket.HandlerRegistry
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

	// Global circuit breaker settings (triggers when aggregate failures across all queues exceed threshold)
	GlobalCircuitBreakerThreshold int           `json:"global_circuit_breaker_threshold"`
	GlobalCircuitBreakerTimeout   time.Duration `json:"global_circuit_breaker_timeout"`
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
	IsRunning            bool                     `json:"is_running"`
	QueueDepth           int                      `json:"queue_depth"`
	ActiveWorkers        int                      `json:"active_workers"`
	CircuitBreakerOn     bool                     `json:"circuit_breaker_on"`      // True if ANY breaker is open (backwards compat)
	CircuitBreakerStatus map[string]BreakerStatus `json:"circuit_breaker_status"` // Per-queue breaker status
	Metrics              QueueMetrics             `json:"metrics"`
	ConsumerStatus       map[string]bool          `json:"consumer_status"`
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

// CircuitBreakerManager manages multiple circuit breakers for different queue types
type CircuitBreakerManager struct {
	mu       sync.RWMutex
	breakers map[rabbitmq.QueueType]*CircuitBreaker
	global   *CircuitBreaker // Fallback for catastrophic failures
	config   CircuitBreakerConfig
}

// CircuitBreakerConfig holds configuration for circuit breakers
type CircuitBreakerConfig struct {
	// Per-queue-type settings (default for all queues)
	DefaultThreshold int           // Default: 50
	DefaultTimeout   time.Duration // Default: 60s

	// Global fallback settings
	GlobalThreshold int           // Default: 100
	GlobalTimeout   time.Duration // Default: 120s

	// Optional per-queue overrides
	QueueOverrides map[rabbitmq.QueueType]CircuitBreakerSettings
}

// CircuitBreakerSettings holds settings for a specific circuit breaker
type CircuitBreakerSettings struct {
	Threshold int
	Timeout   time.Duration
}

// BreakerStatus represents the status of a single circuit breaker
type BreakerStatus struct {
	State        string     `json:"state"` // "closed", "open", "half-open"
	FailureCount int32      `json:"failure_count"`
	LastFailure  *time.Time `json:"last_failure,omitempty"`
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
		BatchSize:                     10,
		MaxConcurrentWorkers:          5,
		ProcessingTimeout:             5 * time.Minute,
		RetryDelay:                    1 * time.Second,
		MaxRetryDelay:                 30 * time.Second,
		CircuitBreakerThreshold:       50,              // Per-queue default (was 5)
		CircuitBreakerTimeout:         60 * time.Second, // Per-queue default (was 30s)
		GlobalCircuitBreakerThreshold: 100,              // Global fallback
		GlobalCircuitBreakerTimeout:   120 * time.Second, // Global fallback
	}
}

// NewQueueManager creates a new queue manager with the given WorkflowQueue.
// For production, use rabbitmq.NewWorkflowQueue().
// For testing, use rabbitmq.NewTestWorkflowQueue() to get unique queue names.
func NewQueueManager(log *logger.Logger, db *sqlx.DB, engine *Engine, client *rabbitmq.Client, queue *rabbitmq.WorkflowQueue) (*QueueManager, error) {
	cfg := DefaultQueueConfig()

	// Create circuit breaker manager with config values
	cbConfig := CircuitBreakerConfig{
		DefaultThreshold: cfg.CircuitBreakerThreshold,
		DefaultTimeout:   cfg.CircuitBreakerTimeout,
		GlobalThreshold:  cfg.GlobalCircuitBreakerThreshold,
		GlobalTimeout:    cfg.GlobalCircuitBreakerTimeout,
	}

	qm := &QueueManager{
		log:                   log,
		db:                    db,
		engine:                engine,
		client:                client,
		queue:                 queue,
		config:                cfg,
		consumers:             make(map[string]*rabbitmq.Consumer),
		stopChan:              make(chan struct{}),
		circuitBreakerManager: NewCircuitBreakerManager(cbConfig),
	}

	return qm, nil
}

// SetHandlerRegistry registers a handler registry for real-time message delivery.
// Handlers in the registry are checked before processing messages as workflow events.
// This allows message types like "alert" to be routed to WebSocket delivery.
func (qm *QueueManager) SetHandlerRegistry(registry *websocket.HandlerRegistry) {
	qm.handlerRegistry = registry
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
	// Determine queue type based on event
	queueType := qm.determineQueueType(event)

	// Check circuit breaker for this queue type
	if qm.circuitBreakerManager.IsOpen(queueType) {
		return fmt.Errorf("circuit breaker is open for queue type %s, refusing new events", queueType)
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

	// Publish to RabbitMQ
	if err := qm.queue.Publish(ctx, queueType, msg); err != nil {
		qm.recordFailure(queueType)
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
	// All queues use the same generic handler that routes based on message type
	handler := func(ctx context.Context, msg *rabbitmq.Message) error {
		return qm.processMessage(ctx, msg)
	}

	consumer, err := qm.queue.Consume(ctx, queueType, handler)

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
	// Check handler registry first (for real-time handlers like alerts)
	if qm.handlerRegistry != nil {
		if handler, ok := qm.handlerRegistry.Get(msg.Type); ok {
			return handler.HandleMessage(ctx, msg)
		}
	}

	// Route async_action messages to the generic async handler
	if msg.Type == "async_action" {
		return qm.processAsyncAction(ctx, msg)
	}

	// Otherwise process as workflow trigger event
	return qm.processWorkflowEvent(ctx, msg)
}

// processAsyncAction processes async action messages generically.
// The handler is retrieved from the registry based on the request_type in the payload.
func (qm *QueueManager) processAsyncAction(ctx context.Context, msg *rabbitmq.Message) error {
	startTime := time.Now()

	// Check if this is a final failure callback from RabbitMQ client
	// (called when max retries exceeded to allow final recording)
	if _, isFinalFailure := msg.Payload["_final_failure"]; isFinalFailure {
		// Already recorded failure on first attempt, just acknowledge
		return nil
	}

	qm.log.Info(ctx, "Processing async action message",
		"messageID", msg.ID,
		"eventType", msg.EventType)

	// Serialize payload to JSON for the handler
	payloadData, err := json.Marshal(msg.Payload)
	if err != nil {
		// Use workflow queue type as fallback when we can't determine the action type
		qm.recordFailure(rabbitmq.QueueTypeWorkflow)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Deserialize just enough to get the request type
	var queuedPayload QueuedPayload
	if err := json.Unmarshal(payloadData, &queuedPayload); err != nil {
		qm.recordFailure(rabbitmq.QueueTypeWorkflow)
		qm.updateMetric(func(m *QueueMetrics) {
			m.TotalFailed++
		})
		return fmt.Errorf("failed to unmarshal queued payload: %w", err)
	}

	// Determine queue type based on the action's request type
	queueType := qm.getQueueTypeForAction(queuedPayload.RequestType)

	// Check circuit breaker for this action's queue type
	if qm.circuitBreakerManager.IsOpen(queueType) {
		return fmt.Errorf("circuit breaker is open for queue type %s", queueType)
	}

	// Get the handler from the registry
	registry := qm.engine.GetRegistry()
	if registry == nil {
		return fmt.Errorf("action registry not available")
	}

	handler, exists := registry.Get(queuedPayload.RequestType)
	if !exists {
		qm.recordFailure(queueType)
		qm.updateMetric(func(m *QueueMetrics) {
			m.TotalFailed++
		})
		return fmt.Errorf("handler not registered for request type: %s", queuedPayload.RequestType)
	}

	// Type assert to AsyncActionHandler
	asyncHandler, ok := handler.(AsyncActionHandler)
	if !ok {
		qm.recordFailure(queueType)
		qm.updateMetric(func(m *QueueMetrics) {
			m.TotalFailed++
		})
		return fmt.Errorf("handler %s does not implement AsyncActionHandler interface", queuedPayload.RequestType)
	}

	// Create processing context with timeout
	processCtx, cancel := context.WithTimeout(ctx, qm.config.ProcessingTimeout)
	defer cancel()

	// Create publisher for the handler to fire result events
	publisher := NewEventPublisher(qm.log, qm)

	// Call the handler's ProcessQueued method
	err = asyncHandler.ProcessQueued(processCtx, payloadData, publisher)

	processingTime := time.Since(startTime)

	if err != nil {
		qm.recordFailure(queueType)
		qm.updateMetric(func(m *QueueMetrics) {
			m.TotalFailed++
		})

		qm.log.Error(ctx, "Failed to process async action",
			"messageID", msg.ID,
			"requestType", queuedPayload.RequestType,
			"queueType", queueType,
			"error", err,
			"processingTime", processingTime)

		return err
	}

	// Success
	qm.recordSuccess(queueType)
	qm.updateMetric(func(m *QueueMetrics) {
		m.TotalProcessed++
		now := time.Now()
		m.LastProcessedAt = &now

		// Update average processing time
		oldAvg := m.AverageProcessTimeMs
		totalTime := oldAvg * (m.TotalProcessed - 1)
		m.AverageProcessTimeMs = (totalTime + processingTime.Milliseconds()) / m.TotalProcessed
	})

	qm.log.Info(ctx, "Processed async action",
		"messageID", msg.ID,
		"requestType", queuedPayload.RequestType,
		"queueType", queueType,
		"processingTime", processingTime)

	return nil
}

// processWorkflowEvent processes workflow trigger events
func (qm *QueueManager) processWorkflowEvent(ctx context.Context, msg *rabbitmq.Message) error {
	startTime := time.Now()

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
		uid, err := uuid.Parse(payload.(string))
		if err == nil {
			event.UserID = uid
		}
	}

	// Determine queue type from reconstructed event
	queueType := qm.determineQueueType(event)

	// Check circuit breaker for this queue type
	if qm.circuitBreakerManager.IsOpen(queueType) {
		return fmt.Errorf("circuit breaker is open for queue type %s", queueType)
	}

	// Create processing context with timeout
	processCtx, cancel := context.WithTimeout(ctx, qm.config.ProcessingTimeout)
	defer cancel()

	// Execute workflow
	execution, err := qm.engine.ExecuteWorkflow(processCtx, event)

	processingTime := time.Since(startTime)

	if err != nil {
		qm.recordFailure(queueType)
		qm.updateMetric(func(m *QueueMetrics) {
			m.TotalFailed++
		})

		qm.log.Error(ctx, "Failed to process workflow event",
			"messageID", msg.ID,
			"queueType", queueType,
			"error", err,
			"processingTime", processingTime)

		// Let RabbitMQ handle retries through its built-in mechanism
		return err
	}

	// Check if the workflow execution itself failed
	if execution.Status == StatusFailed {
		qm.recordFailure(queueType)
		qm.updateMetric(func(m *QueueMetrics) {
			m.TotalFailed++
		})

		// Log detailed error information
		qm.logExecutionErrors(ctx, execution)

		// Return an error so RabbitMQ knows the message processing failed
		return fmt.Errorf("workflow execution failed: %d errors", len(execution.Errors))
	}

	// Success
	qm.recordSuccess(queueType)
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
		"queueType", queueType,
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

	// Get per-queue circuit breaker status
	cbStatus := qm.circuitBreakerManager.GetStatus()

	return &QueueStatus{
		IsRunning:            qm.isRunning.Load(),
		QueueDepth:           stats.Messages,
		ActiveWorkers:        len(qm.consumers),
		CircuitBreakerOn:     qm.circuitBreakerManager.IsAnyOpen(), // Backwards compat: true if ANY breaker is open
		CircuitBreakerStatus: cbStatus,
		Metrics:              metrics,
		ConsumerStatus:       consumerStatus,
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

// getQueueTypeForAction maps action request types to appropriate queue types.
// This enables per-action-type circuit breaker isolation.
func (qm *QueueManager) getQueueTypeForAction(requestType string) rabbitmq.QueueType {
	switch requestType {
	case "send_email":
		return rabbitmq.QueueTypeEmail
	case "send_notification":
		return rabbitmq.QueueTypeNotification
	case "create_alert":
		return rabbitmq.QueueTypeAlert
	case "allocate_inventory":
		return rabbitmq.QueueTypeInventory
	case "seek_approval":
		return rabbitmq.QueueTypeApproval
	case "update_field":
		// Data operations go to general workflow queue
		return rabbitmq.QueueTypeWorkflow
	default:
		return rabbitmq.QueueTypeWorkflow
	}
}

func (qm *QueueManager) updateMetric(fn func(*QueueMetrics)) {
	qm.metricsLock.Lock()
	defer qm.metricsLock.Unlock()
	fn(&qm.metrics)
}

// logExecutionErrors logs detailed information about workflow execution failures
func (qm *QueueManager) logExecutionErrors(ctx context.Context, execution *WorkflowExecution) {
	// Log top-level errors
	for i, errMsg := range execution.Errors {
		qm.log.Error(ctx, "workflow execution error",
			"execution_id", execution.ExecutionID,
			"error_index", i,
			"error", errMsg)
	}

	// Log action-level errors from batch results
	for batchIdx, batch := range execution.BatchResults {
		for _, ruleResult := range batch.RuleResults {
			if ruleResult.Status == "failed" {
				qm.log.Error(ctx, "rule execution failed",
					"execution_id", execution.ExecutionID,
					"batch", batchIdx,
					"rule_id", ruleResult.RuleID,
					"rule_name", ruleResult.RuleName,
					"error", ruleResult.ErrorMessage)
			}
			for _, actionResult := range ruleResult.ActionResults {
				if actionResult.Status == "failed" {
					qm.log.Error(ctx, "action execution failed",
						"execution_id", execution.ExecutionID,
						"batch", batchIdx,
						"rule_name", ruleResult.RuleName,
						"action_id", actionResult.ActionID,
						"action_name", actionResult.ActionName,
						"action_type", actionResult.ActionType,
						"error", actionResult.ErrorMessage)
				}
			}
		}
	}
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

// ============================================================================
// CircuitBreakerManager methods
// ============================================================================

// NewCircuitBreakerManager creates a new circuit breaker manager with the given config.
func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager {
	cbm := &CircuitBreakerManager{
		breakers: make(map[rabbitmq.QueueType]*CircuitBreaker),
		config:   config,
	}

	// Initialize per-queue breakers
	queueTypes := []rabbitmq.QueueType{
		rabbitmq.QueueTypeWorkflow,
		rabbitmq.QueueTypeApproval,
		rabbitmq.QueueTypeNotification,
		rabbitmq.QueueTypeInventory,
		rabbitmq.QueueTypeEmail,
		rabbitmq.QueueTypeAlert,
	}

	for _, qt := range queueTypes {
		threshold := config.DefaultThreshold
		timeout := config.DefaultTimeout

		// Apply overrides if present
		if override, ok := config.QueueOverrides[qt]; ok {
			if override.Threshold > 0 {
				threshold = override.Threshold
			}
			if override.Timeout > 0 {
				timeout = override.Timeout
			}
		}

		cb := &CircuitBreaker{
			failureThreshold: threshold,
			resetTimeout:     timeout,
		}
		cb.state.Store("closed")
		cb.lastFailureTime.Store(time.Now())
		cbm.breakers[qt] = cb
	}

	// Initialize global breaker
	cbm.global = &CircuitBreaker{
		failureThreshold: config.GlobalThreshold,
		resetTimeout:     config.GlobalTimeout,
	}
	cbm.global.state.Store("closed")
	cbm.global.lastFailureTime.Store(time.Now())

	return cbm
}

// IsOpen checks if the circuit breaker for the given queue type is open.
// Returns true if either the queue-specific breaker OR the global breaker is open.
func (cbm *CircuitBreakerManager) IsOpen(queueType rabbitmq.QueueType) bool {
	// Check global breaker first
	if cbm.global.IsOpen() {
		return true
	}

	// Check queue-specific breaker
	cbm.mu.RLock()
	cb, exists := cbm.breakers[queueType]
	cbm.mu.RUnlock()

	if !exists {
		return false
	}

	return cb.IsOpen()
}

// IsAnyOpen returns true if any circuit breaker (including global) is open.
func (cbm *CircuitBreakerManager) IsAnyOpen() bool {
	if cbm.global.IsOpen() {
		return true
	}

	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	for _, cb := range cbm.breakers {
		if cb.IsOpen() {
			return true
		}
	}
	return false
}

// RecordFailure records a failure for the given queue type.
// Also increments the global failure counter.
func (cbm *CircuitBreakerManager) RecordFailure(queueType rabbitmq.QueueType) {
	// Record in queue-specific breaker
	cbm.mu.RLock()
	cb, exists := cbm.breakers[queueType]
	cbm.mu.RUnlock()

	if exists {
		cb.mu.Lock()
		count := cb.failureCount.Add(1)
		cb.lastFailureTime.Store(time.Now())
		if count >= int32(cb.failureThreshold) {
			cb.state.Store("open")
		}
		cb.mu.Unlock()
	}

	// Also record in global breaker
	cbm.global.mu.Lock()
	globalCount := cbm.global.failureCount.Add(1)
	cbm.global.lastFailureTime.Store(time.Now())
	if globalCount >= int32(cbm.global.failureThreshold) {
		cbm.global.state.Store("open")
	}
	cbm.global.mu.Unlock()
}

// RecordSuccess records a success for the given queue type.
// May transition breakers from half-open to closed.
func (cbm *CircuitBreakerManager) RecordSuccess(queueType rabbitmq.QueueType) {
	// Record in queue-specific breaker
	cbm.mu.RLock()
	cb, exists := cbm.breakers[queueType]
	cbm.mu.RUnlock()

	if exists {
		cb.mu.Lock()
		state := cb.state.Load().(string)
		if state == "half-open" {
			cb.failureCount.Store(0)
			cb.state.Store("closed")
		}
		cb.mu.Unlock()
	}

	// Also check global breaker
	cbm.global.mu.Lock()
	globalState := cbm.global.state.Load().(string)
	if globalState == "half-open" {
		cbm.global.failureCount.Store(0)
		cbm.global.state.Store("closed")
	}
	cbm.global.mu.Unlock()
}

// GetStatus returns the status of all circuit breakers.
func (cbm *CircuitBreakerManager) GetStatus() map[string]BreakerStatus {
	status := make(map[string]BreakerStatus)

	cbm.mu.RLock()
	for qt, cb := range cbm.breakers {
		status[string(qt)] = cb.GetStatus()
	}
	cbm.mu.RUnlock()

	// Add global breaker status
	status["global"] = cbm.global.GetStatus()

	return status
}

// Reset resets all circuit breakers to closed state.
func (cbm *CircuitBreakerManager) Reset() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	for _, cb := range cbm.breakers {
		cb.mu.Lock()
		cb.failureCount.Store(0)
		cb.state.Store("closed")
		cb.lastFailureTime.Store(time.Now())
		cb.mu.Unlock()
	}

	cbm.global.mu.Lock()
	cbm.global.failureCount.Store(0)
	cbm.global.state.Store("closed")
	cbm.global.lastFailureTime.Store(time.Now())
	cbm.global.mu.Unlock()
}

// SetThreshold sets the threshold for a specific queue type.
func (cbm *CircuitBreakerManager) SetThreshold(queueType rabbitmq.QueueType, threshold int) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if cb, exists := cbm.breakers[queueType]; exists {
		cb.mu.Lock()
		cb.failureThreshold = threshold
		cb.mu.Unlock()
	}
}

// SetTimeout sets the timeout for a specific queue type.
func (cbm *CircuitBreakerManager) SetTimeout(queueType rabbitmq.QueueType, timeout time.Duration) {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	if cb, exists := cbm.breakers[queueType]; exists {
		cb.mu.Lock()
		cb.resetTimeout = timeout
		cb.mu.Unlock()
	}
}

// SetGlobalThreshold sets the threshold for the global breaker.
func (cbm *CircuitBreakerManager) SetGlobalThreshold(threshold int) {
	cbm.global.mu.Lock()
	defer cbm.global.mu.Unlock()
	cbm.global.failureThreshold = threshold
}

// SetGlobalTimeout sets the timeout for the global breaker.
func (cbm *CircuitBreakerManager) SetGlobalTimeout(timeout time.Duration) {
	cbm.global.mu.Lock()
	defer cbm.global.mu.Unlock()
	cbm.global.resetTimeout = timeout
}

// ============================================================================
// CircuitBreaker methods
// ============================================================================

// IsOpen checks if this circuit breaker is open.
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

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

// GetStatus returns the current status of this circuit breaker.
func (cb *CircuitBreaker) GetStatus() BreakerStatus {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	lastFailure := cb.lastFailureTime.Load().(time.Time)
	return BreakerStatus{
		State:        cb.state.Load().(string),
		FailureCount: cb.failureCount.Load(),
		LastFailure:  &lastFailure,
	}
}

// ============================================================================
// QueueManager circuit breaker methods
// ============================================================================

// recordFailure records a failure for the given queue type.
func (qm *QueueManager) recordFailure(queueType rabbitmq.QueueType) {
	qm.circuitBreakerManager.RecordFailure(queueType)

	// Log when breaker opens
	if qm.circuitBreakerManager.IsOpen(queueType) {
		status := qm.circuitBreakerManager.GetStatus()
		if bs, ok := status[string(queueType)]; ok && bs.State == "open" {
			qm.log.Warn(context.Background(), "Circuit breaker opened",
				"queueType", queueType,
				"failures", bs.FailureCount)
		}
	}
}

// recordSuccess records a success for the given queue type.
func (qm *QueueManager) recordSuccess(queueType rabbitmq.QueueType) {
	wasOpen := qm.circuitBreakerManager.IsOpen(queueType)
	qm.circuitBreakerManager.RecordSuccess(queueType)

	// Log when breaker closes
	if wasOpen && !qm.circuitBreakerManager.IsOpen(queueType) {
		qm.log.Info(context.Background(), "Circuit breaker closed",
			"queueType", queueType)
	}
}

// ResetCircuitBreaker resets all circuit breakers to closed state (for testing)
func (qm *QueueManager) ResetCircuitBreaker() {
	qm.circuitBreakerManager.Reset()
}

// SetCircuitBreakerThreshold allows tests to set custom thresholds per queue.
func (qm *QueueManager) SetCircuitBreakerThreshold(queueType rabbitmq.QueueType, threshold int) {
	qm.circuitBreakerManager.SetThreshold(queueType, threshold)
}

// SetCircuitBreakerTimeout allows tests to set custom timeouts per queue.
func (qm *QueueManager) SetCircuitBreakerTimeout(queueType rabbitmq.QueueType, timeout time.Duration) {
	qm.circuitBreakerManager.SetTimeout(queueType, timeout)
}

// SetGlobalCircuitBreakerThreshold allows tests to set the global breaker threshold.
func (qm *QueueManager) SetGlobalCircuitBreakerThreshold(threshold int) {
	qm.circuitBreakerManager.SetGlobalThreshold(threshold)
}

// RecordFailureForTesting allows tests to simulate failures for a specific queue type.
func (qm *QueueManager) RecordFailureForTesting(queueType rabbitmq.QueueType) {
	qm.recordFailure(queueType)
}

// ResetMetrics resets all queue metrics to zero (for testing)
func (qm *QueueManager) ResetMetrics() {
	qm.metricsLock.Lock()
	defer qm.metricsLock.Unlock()
	qm.metrics = QueueMetrics{}
}
