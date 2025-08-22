// business/sdk/workflow/notificationQueue.go
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// =============================================================================
// Notification Event Types
// =============================================================================

// NotificationEvent represents events related to notifications in the workflow
type NotificationEvent struct {
	EventType      string         `json:"event_type"` // created, updated, delivered, failed
	UserID         string         `json:"user_id"`
	NotificationID string         `json:"notification_id"`
	Timestamp      time.Time      `json:"timestamp"`
	Channel        string         `json:"channel,omitempty"` // email, push, sms, in_app
	Data           map[string]any `json:"data,omitempty"`
	WorkflowID     string         `json:"workflow_id,omitempty"` // Link to workflow execution
	RuleID         string         `json:"rule_id,omitempty"`     // Which rule triggered this
	ActionID       string         `json:"action_id,omitempty"`   // Which action created this
}

// NotificationPayload represents a notification to be processed
type NotificationPayload struct {
	ID                    uuid.UUID      `json:"id"`
	Recipients            uuid.UUIDs     `json:"recipients"`
	Title                 string         `json:"title"`
	Body                  string         `json:"body"`
	Priority              string         `json:"priority"`
	Channel               string         `json:"channel"` // email, push, sms, in_app
	TemplateID            uuid.UUID      `json:"template_id,omitempty"`
	TemplateData          map[string]any `json:"template_data,omitempty"`
	ReferenceID           uuid.UUID      `json:"reference_id,omitempty"`
	ReferenceType         string         `json:"reference_type,omitempty"`
	Config                map[string]any `json:"config,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	AutomationExecutionID uuid.UUID      `json:"automation_execution_id,omitempty"`
}

// NotificationProcessorStats tracks notification processing statistics
type NotificationProcessorStats struct {
	TotalProcessed          int64                    `json:"total_processed"`
	SuccessfulDeliveries    int64                    `json:"successful_deliveries"`
	FailedDeliveries        int64                    `json:"failed_deliveries"`
	RetryCount              int64                    `json:"retry_count"`
	AverageProcessingTimeMs int64                    `json:"average_processing_time_ms"`
	QueueDepth              int                      `json:"queue_depth"`
	LastProcessedAt         *time.Time               `json:"last_processed_at"`
	ActiveWorkers           int                      `json:"active_workers"`
	ByChannel               map[string]*ChannelStats `json:"by_channel"`
}

// ChannelStats tracks per-channel statistics
type ChannelStats struct {
	Sent   int64 `json:"sent"`
	Failed int64 `json:"failed"`
	AvgMs  int64 `json:"avg_ms"`
}

// =============================================================================
// Notification Queue Processor
// =============================================================================

// NotificationQueueProcessor handles notification processing within workflows
type NotificationQueueProcessor struct {
	log    *logger.Logger
	client *rabbitmq.Client
	queue  *rabbitmq.WorkflowQueue
	store  Storer // *comment* Need notification persistence store

	// Configuration
	config NotificationConfig

	// State management
	mu           sync.RWMutex
	isRunning    atomic.Bool
	consumers    map[string]*rabbitmq.Consumer
	stopChan     chan struct{}
	processingWG sync.WaitGroup

	// Statistics
	stats     NotificationProcessorStats
	statsLock sync.RWMutex

	// Channel handlers
	handlers map[string]NotificationHandler
}

// NotificationConfig holds notification processor configuration
type NotificationConfig struct {
	MaxWorkers        int           `json:"max_workers"`
	RetryDelayMs      time.Duration `json:"retry_delay_ms"`
	MaxRetries        int           `json:"max_retries"`
	ProcessingTimeout time.Duration `json:"processing_timeout"`
	EnabledChannels   []string      `json:"enabled_channels"`
	DefaultPriority   string        `json:"default_priority"`
}

// NotificationHandler interface for channel-specific handlers
type NotificationHandler interface {
	// Send sends a notification through this channel
	Send(ctx context.Context, payload *NotificationPayload) error
	// GetChannelType returns the channel type this handler manages
	GetChannelType() string
	// IsAvailable checks if the channel is currently available
	IsAvailable() bool
	// GetPriority returns the priority for this channel
	GetPriority() int
}

// DefaultNotificationConfig returns default configuration
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		MaxWorkers:        3,
		RetryDelayMs:      5 * time.Second,
		MaxRetries:        3,
		ProcessingTimeout: 30 * time.Second,
		EnabledChannels:   []string{"email", "in_app", "push", "sms"},
		DefaultPriority:   "normal",
	}
}

// NewNotificationQueueProcessor creates a new notification processor
func NewNotificationQueueProcessor(log *logger.Logger, client *rabbitmq.Client, store Storer) *NotificationQueueProcessor {
	np := &NotificationQueueProcessor{
		log:       log,
		client:    client,
		queue:     rabbitmq.NewWorkflowQueue(client, log),
		store:     store,
		config:    DefaultNotificationConfig(),
		consumers: make(map[string]*rabbitmq.Consumer),
		stopChan:  make(chan struct{}),
		handlers:  make(map[string]NotificationHandler),
		stats: NotificationProcessorStats{
			ByChannel: make(map[string]*ChannelStats),
		},
	}

	// Initialize channel stats
	for _, channel := range np.config.EnabledChannels {
		np.stats.ByChannel[channel] = &ChannelStats{}
	}

	return np
}

// RegisterHandler registers a notification handler for a specific channel
func (np *NotificationQueueProcessor) RegisterHandler(handler NotificationHandler) error {
	channelType := handler.GetChannelType()

	np.mu.Lock()
	defer np.mu.Unlock()

	if _, exists := np.handlers[channelType]; exists {
		return fmt.Errorf("handler already registered for channel: %s", channelType)
	}

	np.handlers[channelType] = handler
	np.log.Info(context.Background(), "Registered notification handler",
		"channel", channelType)

	// Initialize stats for this channel if not exists
	if _, exists := np.stats.ByChannel[channelType]; !exists {
		np.stats.ByChannel[channelType] = &ChannelStats{}
	}

	return nil
}

// Initialize sets up notification queues
func (np *NotificationQueueProcessor) Initialize(ctx context.Context) error {
	np.log.Info(ctx, "Initializing notification queue processor...")

	// Base queue initialization is handled by WorkflowQueue
	// We don't need to reinitialize here since the queues are shared

	// Verify required handlers are registered
	for _, channel := range np.config.EnabledChannels {
		if _, exists := np.handlers[channel]; !exists {
			np.log.Warn(ctx, "No handler registered for enabled channel",
				"channel", channel)
		}
	}

	np.log.Info(ctx, "Notification processor initialized")
	return nil
}

// Start begins processing notifications
func (np *NotificationQueueProcessor) Start(ctx context.Context) error {
	if np.isRunning.Load() {
		return fmt.Errorf("notification processor already running")
	}

	np.isRunning.Store(true)
	np.log.Info(ctx, "Starting notification processor")

	// Start consumers for notification-related queues
	queueTypes := []rabbitmq.QueueType{
		rabbitmq.QueueTypeNotification,
		rabbitmq.QueueTypeEmail,
		rabbitmq.QueueTypeAlert,
	}

	for _, qt := range queueTypes {
		if err := np.startConsumer(ctx, qt); err != nil {
			np.log.Error(ctx, "Failed to start consumer",
				"queueType", qt,
				"error", err)
		}
	}

	// Start metrics collector
	np.processingWG.Add(1)
	go np.metricsCollector(ctx)

	return nil
}

// startConsumer starts a consumer for a specific queue type
func (np *NotificationQueueProcessor) startConsumer(ctx context.Context, queueType rabbitmq.QueueType) error {
	var handler rabbitmq.MessageHandler

	switch queueType {
	case rabbitmq.QueueTypeNotification:
		handler = np.processNotificationMessage
	case rabbitmq.QueueTypeEmail:
		handler = np.processEmailMessage
	case rabbitmq.QueueTypeAlert:
		handler = np.processAlertMessage
	default:
		return fmt.Errorf("unsupported queue type: %s", queueType)
	}

	consumer, err := np.queue.Consume(ctx, queueType, handler)
	if err != nil {
		return fmt.Errorf("failed to start consumer for %s: %w", queueType, err)
	}

	np.mu.Lock()
	np.consumers[string(queueType)] = consumer
	np.mu.Unlock()

	return nil
}

// processNotificationMessage handles general notification messages
func (np *NotificationQueueProcessor) processNotificationMessage(ctx context.Context, msg *rabbitmq.Message) error {
	startTime := time.Now()

	// Create processing context with timeout
	processCtx, cancel := context.WithTimeout(ctx, np.config.ProcessingTimeout)
	defer cancel()

	// Parse notification payload
	payload, err := np.parseNotificationPayload(msg)
	if err != nil {
		np.log.Error(ctx, "Failed to parse notification payload",
			"messageID", msg.ID,
			"error", err)
		return err // Will be retried by RabbitMQ
	}

	// Route to appropriate channel handler
	handler, exists := np.handlers[payload.Channel]
	if !exists {
		np.log.Error(ctx, "No handler for channel",
			"channel", payload.Channel,
			"messageID", msg.ID)
		// Don't retry if handler doesn't exist
		return nil
	}

	// Check if handler is available
	if !handler.IsAvailable() {
		np.log.Warn(ctx, "Handler not available",
			"channel", payload.Channel,
			"messageID", msg.ID)
		return fmt.Errorf("handler not available for channel: %s", payload.Channel)
	}

	// Send notification
	err = handler.Send(processCtx, payload)
	processingTime := time.Since(startTime)

	// Update statistics
	np.updateStats(payload.Channel, err == nil, processingTime)

	if err != nil {
		np.log.Error(ctx, "Failed to send notification",
			"channel", payload.Channel,
			"messageID", msg.ID,
			"error", err,
			"processingTime", processingTime)

		// Check if we should retry
		if msg.Attempts < msg.MaxAttempts {
			return err // Let RabbitMQ handle retry
		}

		// Max retries exceeded, record failure
		if err := np.recordDeliveryFailure(ctx, payload, err); err != nil {
			np.log.Error(ctx, "Failed to record delivery failure", "error", err)
		}

		return nil // Don't retry further
	}

	// Record successful delivery
	if err := np.recordDeliverySuccess(ctx, payload); err != nil {
		np.log.Error(ctx, "Failed to record delivery success", "error", err)
	}

	np.log.Info(ctx, "Notification sent successfully",
		"channel", payload.Channel,
		"messageID", msg.ID,
		"processingTime", processingTime)

	return nil
}

// processEmailMessage handles email-specific messages
func (np *NotificationQueueProcessor) processEmailMessage(ctx context.Context, msg *rabbitmq.Message) error {
	// Email messages are routed through the email channel handler
	msg.Payload["channel"] = "email"
	return np.processNotificationMessage(ctx, msg)
}

// processAlertMessage handles high-priority alert messages
func (np *NotificationQueueProcessor) processAlertMessage(ctx context.Context, msg *rabbitmq.Message) error {
	// Alerts might go through multiple channels
	startTime := time.Now()

	payload, err := np.parseNotificationPayload(msg)
	if err != nil {
		return err
	}

	// For alerts, try multiple channels in priority order
	channels := np.getAlertChannels(payload.Priority)

	var lastErr error
	successCount := 0

	for _, channel := range channels {
		handler, exists := np.handlers[channel]
		if !exists || !handler.IsAvailable() {
			continue
		}

		// Clone payload for this channel
		channelPayload := *payload
		channelPayload.Channel = channel

		if err := handler.Send(ctx, &channelPayload); err != nil {
			lastErr = err
			np.log.Error(ctx, "Failed to send alert via channel",
				"channel", channel,
				"error", err)
		} else {
			successCount++
			np.updateStats(channel, true, time.Since(startTime))
		}
	}

	if successCount == 0 && lastErr != nil {
		return lastErr // Retry if all channels failed
	}

	return nil
}

// Helper methods

func (np *NotificationQueueProcessor) parseNotificationPayload(msg *rabbitmq.Message) (*NotificationPayload, error) {
	payload := &NotificationPayload{
		ID:                    msg.ID,
		CreatedAt:             msg.CreatedAt,
		AutomationExecutionID: msg.CorrelationID,
	}

	// Extract fields from message payload
	if recipients, ok := msg.Payload["recipients"].([]any); ok {
		for _, r := range recipients {
			if str, ok := r.(string); ok {
				payload.Recipients = append(payload.Recipients, uuid.MustParse(str))
			}
		}
	}

	if title, ok := msg.Payload["title"].(string); ok {
		payload.Title = title
	}

	if body, ok := msg.Payload["body"].(string); ok {
		payload.Body = body
	} else if message, ok := msg.Payload["message"].(string); ok {
		payload.Body = message
	}

	if priority, ok := msg.Payload["priority"].(string); ok {
		payload.Priority = priority
	} else {
		payload.Priority = np.config.DefaultPriority
	}

	if channel, ok := msg.Payload["channel"].(string); ok {
		payload.Channel = channel
	}

	if templateID, ok := msg.Payload["template_id"].(uuid.UUID); ok {
		payload.TemplateID = templateID
	}

	if templateData, ok := msg.Payload["template_data"].(map[string]any); ok {
		payload.TemplateData = templateData
	}

	if config, ok := msg.Payload["config"].(map[string]any); ok {
		payload.Config = config
	}

	return payload, nil
}

func (np *NotificationQueueProcessor) getAlertChannels(priority string) []string {
	// Return channels in priority order based on alert priority
	switch priority {
	case "critical":
		return []string{"push", "sms", "email", "in_app"}
	case "high":
		return []string{"push", "email", "in_app"}
	case "medium":
		return []string{"email", "in_app"}
	default:
		return []string{"in_app", "email"}
	}
}

func (np *NotificationQueueProcessor) recordDeliverySuccess(ctx context.Context, payload *NotificationPayload) error {
	now := time.Now()

	// Create a delivery record for each recipient
	for _, recipientID := range payload.Recipients {
		delivery := NotificationDelivery{
			ID:                    uuid.New(),
			NotificationID:        payload.ID,
			AutomationExecutionID: payload.AutomationExecutionID,
			RecipientID:           recipientID,
			Channel:               payload.Channel,
			Status:                DeliveryStatusDelivered,
			Attempts:              1, // Could track from message
			SentAt:                &payload.CreatedAt,
			DeliveredAt:           &now,
			CreatedDate:           payload.CreatedAt,
			UpdatedDate:           now,
		}

		if err := np.store.CreateNotificationDelivery(ctx, delivery); err != nil {
			np.log.Error(ctx, "Failed to record delivery success",
				"recipientID", recipientID,
				"notificationID", payload.ID,
				"error", err)
			// Don't fail the whole operation if recording fails
			continue
		}
	}

	return nil
}

func (np *NotificationQueueProcessor) recordDeliveryFailure(ctx context.Context, payload *NotificationPayload, deliveryErr error) error {
	now := time.Now()

	dErr := deliveryErr.Error()

	// Create failure records for each recipient
	for _, recipientID := range payload.Recipients {
		delivery := NotificationDelivery{
			ID:                    uuid.New(),
			NotificationID:        payload.ID,
			AutomationExecutionID: payload.AutomationExecutionID,
			RecipientID:           recipientID,
			Channel:               payload.Channel,
			Status:                DeliveryStatusFailed,
			Attempts:              1, // Could track from message
			SentAt:                &payload.CreatedAt,
			FailedAt:              &now,
			ErrorMessage:          &dErr,
			CreatedDate:           payload.CreatedAt,
			UpdatedDate:           now,
		}

		// If we have provider-specific error details, store them
		if providerErr, ok := deliveryErr.(interface{ ProviderResponse() json.RawMessage }); ok {
			delivery.ProviderResponse = providerErr.ProviderResponse()
		}

		if err := np.store.CreateNotificationDelivery(ctx, delivery); err != nil {
			np.log.Error(ctx, "Failed to record delivery failure",
				"recipientID", recipientID,
				"notificationID", payload.ID,
				"error", err)
			continue
		}
	}

	return nil
}

func (np *NotificationQueueProcessor) updateStats(channel string, success bool, duration time.Duration) {
	np.statsLock.Lock()
	defer np.statsLock.Unlock()

	np.stats.TotalProcessed++

	if success {
		np.stats.SuccessfulDeliveries++
	} else {
		np.stats.FailedDeliveries++
	}

	now := time.Now()
	np.stats.LastProcessedAt = &now

	// Update average processing time
	oldAvg := np.stats.AverageProcessingTimeMs
	totalTime := oldAvg * (np.stats.TotalProcessed - 1)
	np.stats.AverageProcessingTimeMs = (totalTime + duration.Milliseconds()) / np.stats.TotalProcessed

	// Update channel-specific stats
	if channelStats, exists := np.stats.ByChannel[channel]; exists {
		if success {
			channelStats.Sent++
		} else {
			channelStats.Failed++
		}

		// Update channel average
		oldChannelAvg := channelStats.AvgMs
		totalChannelTime := oldChannelAvg * (channelStats.Sent + channelStats.Failed - 1)
		channelStats.AvgMs = (totalChannelTime + duration.Milliseconds()) / (channelStats.Sent + channelStats.Failed)
	}
}

func (np *NotificationQueueProcessor) metricsCollector(ctx context.Context) {
	defer np.processingWG.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-np.stopChan:
			return
		case <-ticker.C:
			// Collect queue stats
			totalDepth := 0
			for _, qt := range []rabbitmq.QueueType{
				rabbitmq.QueueTypeNotification,
				rabbitmq.QueueTypeEmail,
				rabbitmq.QueueTypeAlert,
			} {
				if stats, err := np.queue.GetQueueStats(ctx, qt); err == nil {
					totalDepth += stats.Messages
				}
			}

			np.statsLock.Lock()
			np.stats.QueueDepth = totalDepth
			np.stats.ActiveWorkers = len(np.consumers)
			np.statsLock.Unlock()
		}
	}
}

// Stop gracefully stops the notification processor
func (np *NotificationQueueProcessor) Stop(ctx context.Context) error {
	if !np.isRunning.Load() {
		return nil
	}

	np.isRunning.Store(false)
	np.log.Info(ctx, "Stopping notification processor")

	// Stop all consumers
	np.mu.Lock()
	for name, consumer := range np.consumers {
		if err := consumer.Stop(); err != nil {
			np.log.Error(ctx, "Failed to stop consumer",
				"name", name,
				"error", err)
		}
	}
	np.consumers = make(map[string]*rabbitmq.Consumer)
	np.mu.Unlock()

	// Signal stop
	close(np.stopChan)

	// Wait for goroutines
	done := make(chan struct{})
	go func() {
		np.processingWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		np.log.Info(ctx, "Notification processor stopped gracefully")
	case <-time.After(30 * time.Second):
		np.log.Warn(ctx, "Notification processor stop timeout")
	}

	return nil
}

// GetStats returns current processing statistics
func (np *NotificationQueueProcessor) GetStats() NotificationProcessorStats {
	np.statsLock.RLock()
	defer np.statsLock.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := np.stats
	statsCopy.ByChannel = make(map[string]*ChannelStats)
	for k, v := range np.stats.ByChannel {
		statsCopy.ByChannel[k] = &ChannelStats{
			Sent:   v.Sent,
			Failed: v.Failed,
			AvgMs:  v.AvgMs,
		}
	}

	return statsCopy
}

// UpdateConfig updates the processor configuration
func (np *NotificationQueueProcessor) UpdateConfig(config NotificationConfig) {
	np.mu.Lock()
	defer np.mu.Unlock()
	np.config = config
	np.log.Info(context.Background(), "Notification processor configuration updated")
}

// QueueNotification queues a notification for processing
func (np *NotificationQueueProcessor) QueueNotification(ctx context.Context, notification *NotificationPayload) error {
	// Convert to RabbitMQ message
	msg := &rabbitmq.Message{
		ID:            notification.ID,
		Type:          "notification",
		EntityName:    "notifications",
		EventType:     "send",
		Payload:       np.payloadToMap(notification),
		Priority:      np.priorityToUint8(notification.Priority),
		MaxAttempts:   np.config.MaxRetries,
		CreatedAt:     time.Now(),
		ScheduledFor:  time.Now(),
		CorrelationID: notification.AutomationExecutionID,
	}

	// Determine queue type based on channel/priority
	queueType := np.determineQueueType(notification)

	// Publish to queue
	return np.queue.Publish(ctx, queueType, msg)
}

func (np *NotificationQueueProcessor) payloadToMap(notification *NotificationPayload) map[string]any {
	return map[string]any{
		"recipients":     notification.Recipients,
		"title":          notification.Title,
		"body":           notification.Body,
		"priority":       notification.Priority,
		"channel":        notification.Channel,
		"template_id":    notification.TemplateID,
		"template_data":  notification.TemplateData,
		"reference_id":   notification.ReferenceID,
		"reference_type": notification.ReferenceType,
		"config":         notification.Config,
	}
}

func (np *NotificationQueueProcessor) priorityToUint8(priority string) uint8 {
	switch priority {
	case "critical":
		return uint8(rabbitmq.PriorityCritical)
	case "high":
		return uint8(rabbitmq.PriorityHigh)
	case "low":
		return uint8(rabbitmq.PriorityLow)
	default:
		return uint8(rabbitmq.PriorityNormal)
	}
}

func (np *NotificationQueueProcessor) determineQueueType(notification *NotificationPayload) rabbitmq.QueueType {
	// Route based on priority and channel
	if notification.Priority == "critical" || notification.Priority == "high" {
		return rabbitmq.QueueTypeAlert
	}

	if notification.Channel == "email" {
		return rabbitmq.QueueTypeEmail
	}

	return rabbitmq.QueueTypeNotification
}
