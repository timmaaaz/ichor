// foundation/rabbitmq/client.go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Connection Management
// =============================================================================

// Client represents a RabbitMQ client with automatic reconnection
type Client struct {
	url     string
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *logger.Logger
	mu      sync.RWMutex

	// Reconnection handling
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool

	// Configuration
	config Config
}

// Config holds RabbitMQ configuration
type Config struct {
	URL                string
	MaxRetries         int
	RetryDelay         time.Duration
	PrefetchCount      int
	PrefetchSize       int
	PublisherConfirms  bool
	ExchangeName       string
	ExchangeType       string
	DeadLetterExchange string
}

// DefaultConfig returns default RabbitMQ configuration
func DefaultConfig() Config {
	return Config{
		URL:                "amqp://guest:guest@localhost:5672/",
		MaxRetries:         5,
		RetryDelay:         5 * time.Second,
		PrefetchCount:      10,
		PrefetchSize:       0,
		PublisherConfirms:  true,
		ExchangeName:       "workflow",
		ExchangeType:       "topic",
		DeadLetterExchange: "workflow.dlx",
	}
}

var (
	clientInstance *Client
	clientOnce     sync.Once
)

// NewClient creates or returns the singleton RabbitMQ client
func NewClient(log *logger.Logger, config Config) *Client {
	clientOnce.Do(func() {
		clientInstance = &Client{
			url:    config.URL,
			log:    log,
			config: config,
		}

		// Start connection in background
		go clientInstance.handleReconnect()
	})

	return clientInstance
}

// Connect establishes connection to RabbitMQ
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return nil
	}

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	err = ch.Qos(
		c.config.PrefetchCount,
		c.config.PrefetchSize,
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Enable publisher confirms if configured
	if c.config.PublisherConfirms {
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			conn.Close()
			return fmt.Errorf("failed to enable publisher confirms: %w", err)
		}
		c.notifyConfirm = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	}

	c.conn = conn
	c.channel = ch
	c.isConnected = true

	// Setup close notification
	c.notifyClose = make(chan *amqp.Error)
	c.conn.NotifyClose(c.notifyClose)

	// Setup exchanges and queues
	if err := c.setupTopology(); err != nil {
		c.Close()
		return fmt.Errorf("failed to setup topology: %w", err)
	}

	c.log.Info(context.Background(), "Connected to RabbitMQ")
	return nil
}

// handleReconnect manages automatic reconnection
func (c *Client) handleReconnect() {
	for {
		if !c.isConnected {
			c.log.Info(context.Background(), "Attempting to connect to RabbitMQ...")

			for i := 0; i < c.config.MaxRetries; i++ {
				if err := c.Connect(); err != nil {
					c.log.Error(context.Background(), "Failed to connect to RabbitMQ",
						"attempt", i+1,
						"error", err)
					time.Sleep(c.config.RetryDelay)
					continue
				}
				break
			}
		}

		// Wait for connection to close
		if c.isConnected {
			select {
			case err := <-c.notifyClose:
				if err != nil {
					c.log.Error(context.Background(), "RabbitMQ connection closed",
						"error", err)
				}
				c.mu.Lock()
				c.isConnected = false
				c.mu.Unlock()
			}
		} else {
			time.Sleep(c.config.RetryDelay)
		}
	}
}

// setupTopology creates exchanges and queues
func (c *Client) setupTopology() error {
	// Main exchange
	if err := c.channel.ExchangeDeclare(
		c.config.ExchangeName,
		c.config.ExchangeType,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // args
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Dead letter exchange
	if err := c.channel.ExchangeDeclare(
		c.config.DeadLetterExchange,
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // args
	); err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return nil
	}

	if c.channel != nil {
		c.channel.Close()
	}

	if c.conn != nil {
		c.conn.Close()
	}

	c.isConnected = false
	return nil
}

// IsConnected returns connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// GetChannel returns the current channel (use with caution)
func (c *Client) GetChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	return c.channel, nil
}

// GetConnection returns the current connection (use with caution)
func (c *Client) GetConnection() (*amqp.Connection, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	return c.conn, nil
}

// WaitForConnection waits until connected or timeout
func (c *Client) WaitForConnection(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if c.IsConnected() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for connection")
}

// =============================================================================
// Message Types
// =============================================================================

// Message represents a workflow queue message
type Message struct {
	ID            uuid.UUID              `json:"id"`
	Type          string                 `json:"type"`
	EntityName    string                 `json:"entity_name"`
	EntityID      uuid.UUID              `json:"entity_id,omitempty"`
	EventType     string                 `json:"event_type"`
	Payload       map[string]interface{} `json:"payload"`
	Priority      uint8                  `json:"priority"`
	Attempts      int                    `json:"attempts"`
	MaxAttempts   int                    `json:"max_attempts"`
	CreatedAt     time.Time              `json:"created_at"`
	ScheduledFor  time.Time              `json:"scheduled_for"`
	CorrelationID uuid.UUID              `json:"correlation_id,omitempty"`
	UserID        uuid.UUID              `json:"user_id,omitempty"`
}

// QueuePriority defines message priority levels
type QueuePriority uint8

const (
	PriorityLow      QueuePriority = 1
	PriorityNormal   QueuePriority = 5
	PriorityHigh     QueuePriority = 8
	PriorityCritical QueuePriority = 10
)

// QueueType defines different queue types for workflow
type QueueType string

const (
	QueueTypeWorkflow     QueueType = "workflow"
	QueueTypeApproval     QueueType = "approval"
	QueueTypeNotification QueueType = "notification"
	QueueTypeInventory    QueueType = "inventory"
	QueueTypeEmail        QueueType = "email"
	QueueTypeAlert        QueueType = "alert"
)

// =============================================================================
// WorkflowQueue
// =============================================================================

// WorkflowQueue manages workflow-specific queuing operations
type WorkflowQueue struct {
	client *Client
	log    *logger.Logger

	// Queue configurations
	queues map[QueueType]QueueConfig
}

// QueueConfig defines configuration for a specific queue
type QueueConfig struct {
	Name              string
	RoutingKey        string
	MaxPriority       uint8
	MessageTTL        time.Duration
	MaxRetries        int
	DeadLetterEnabled bool
	Durable           bool
	Arguments         amqp.Table
}

// NewWorkflowQueue creates a new workflow queue manager
func NewWorkflowQueue(client *Client, log *logger.Logger) *WorkflowQueue {
	wq := &WorkflowQueue{
		client: client,
		log:    log,
		queues: make(map[QueueType]QueueConfig),
	}

	// Configure default queues
	wq.setupDefaultQueues()

	return wq
}

// setupDefaultQueues configures the default queue types
func (wq *WorkflowQueue) setupDefaultQueues() {
	wq.queues[QueueTypeWorkflow] = QueueConfig{
		Name:              "workflow.general",
		RoutingKey:        "workflow.*",
		MaxPriority:       10,
		MessageTTL:        24 * time.Hour,
		MaxRetries:        3,
		DeadLetterEnabled: true,
		Durable:           true,
	}

	wq.queues[QueueTypeApproval] = QueueConfig{
		Name:              "workflow.approval",
		RoutingKey:        "approval.*",
		MaxPriority:       10,
		MessageTTL:        72 * time.Hour, // 3 days for approvals
		MaxRetries:        5,
		DeadLetterEnabled: true,
		Durable:           true,
	}

	wq.queues[QueueTypeNotification] = QueueConfig{
		Name:              "workflow.notification",
		RoutingKey:        "notification.*",
		MaxPriority:       5,
		MessageTTL:        1 * time.Hour,
		MaxRetries:        3,
		DeadLetterEnabled: true,
		Durable:           true,
	}

	wq.queues[QueueTypeInventory] = QueueConfig{
		Name:              "workflow.inventory",
		RoutingKey:        "inventory.*",
		MaxPriority:       8,
		MessageTTL:        2 * time.Hour,
		MaxRetries:        3,
		DeadLetterEnabled: true,
		Durable:           true,
	}

	wq.queues[QueueTypeEmail] = QueueConfig{
		Name:              "workflow.email",
		RoutingKey:        "email.*",
		MaxPriority:       3,
		MessageTTL:        24 * time.Hour,
		MaxRetries:        5,
		DeadLetterEnabled: true,
		Durable:           true,
	}

	wq.queues[QueueTypeAlert] = QueueConfig{
		Name:              "workflow.alert",
		RoutingKey:        "alert.*",
		MaxPriority:       10,
		MessageTTL:        30 * time.Minute,
		MaxRetries:        1,
		DeadLetterEnabled: true,
		Durable:           true,
	}
}

// In WorkflowQueue.Initialize method, after declaring the main queue:
func (wq *WorkflowQueue) Initialize(ctx context.Context) error {
	ch, err := wq.client.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Create retry exchange first
	retryExchangeName := wq.client.config.ExchangeName + ".retry"
	if err := ch.ExchangeDeclare(
		retryExchangeName,
		"topic", // Change to topic to handle patterns
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare retry exchange: %w", err)
	}

	for _, config := range wq.queues {
		// Main queue args - dead-letters to retry exchange
		args := amqp.Table{
			"x-max-priority": config.MaxPriority,
			"x-message-ttl":  int64(config.MessageTTL.Milliseconds()),
		}

		if config.DeadLetterEnabled {
			// Main queue dead-letters to retry exchange, preserving the original routing key
			args["x-dead-letter-exchange"] = retryExchangeName
			// Don't specify x-dead-letter-routing-key - it will preserve the original
		}

		// Declare main queue
		_, err := ch.QueueDeclare(
			config.Name,
			config.Durable,
			false, false, false,
			args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", config.Name, err)
		}

		// Create retry queue
		if config.DeadLetterEnabled {
			retryQueueName := config.Name + ".retry"
			retryArgs := amqp.Table{
				"x-message-ttl":          5000,
				"x-dead-letter-exchange": wq.client.config.ExchangeName,
				// Don't specify routing key - preserve original
			}

			_, err = ch.QueueDeclare(
				retryQueueName,
				config.Durable,
				false, false, false,
				retryArgs,
			)
			if err != nil {
				return fmt.Errorf("failed to declare retry queue %s: %w", retryQueueName, err)
			}

			// Bind retry queue to retry exchange with the same pattern
			err = ch.QueueBind(
				retryQueueName,
				config.RoutingKey, // Use the same pattern (e.g., "email.*")
				retryExchangeName,
				false, nil,
			)
			if err != nil {
				return fmt.Errorf("failed to bind retry queue %s: %w", retryQueueName, err)
			}
		}

		// Bind main queue to main exchange
		err = ch.QueueBind(
			config.Name,
			config.RoutingKey,
			wq.client.config.ExchangeName,
			false, nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", config.Name, err)
		}
	}

	return wq.createDeadLetterQueue(ch)
}

// createDeadLetterQueue creates the dead letter queue
func (wq *WorkflowQueue) createDeadLetterQueue(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(
		"workflow.dead_letter",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-message-ttl": int64(30 * 24 * time.Hour / time.Millisecond), // 30 days
		},
	)
	if err != nil {
		return err
	}

	// Bind to DLX
	return ch.QueueBind(
		"workflow.dead_letter",
		"#", // receive all dead letters
		wq.client.config.DeadLetterExchange,
		false,
		nil,
	)
}

// =============================================================================
// Publishing Methods
// =============================================================================

// Publish publishes a message to the appropriate queue
func (wq *WorkflowQueue) Publish(ctx context.Context, queueType QueueType, msg *Message) error {
	return wq.PublishWithDelay(ctx, queueType, msg, 0)
}

// PublishWithDelay publishes a message with a delay
func (wq *WorkflowQueue) PublishWithDelay(ctx context.Context, queueType QueueType, msg *Message, delay time.Duration) error {

	config, exists := wq.queues[queueType]
	if !exists {
		return fmt.Errorf("unknown queue type: %s", queueType)
	}

	ch, err := wq.client.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Set message defaults
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.ScheduledFor.IsZero() {
		msg.ScheduledFor = time.Now().Add(delay)
	}
	if msg.MaxAttempts == 0 {
		msg.MaxAttempts = config.MaxRetries
	}

	// Marshal message
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Build routing key
	routingKey := fmt.Sprintf("%s.%s", queueType, msg.EventType)

	// Prepare headers
	headers := amqp.Table{
		"x-retry-count": msg.Attempts,
		"x-max-retries": msg.MaxAttempts,
		"entity-name":   msg.EntityName,
		"entity-id":     msg.EntityID.String(),
		"event-type":    msg.EventType,
	}

	if delay > 0 {
		headers["x-delay"] = int64(delay.Milliseconds())
	}

	// Publish message
	err = ch.PublishWithContext(
		ctx,
		wq.client.config.ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			DeliveryMode:  amqp.Persistent,
			Priority:      msg.Priority,
			MessageId:     msg.ID.String(),
			Timestamp:     msg.CreatedAt,
			Headers:       headers,
			CorrelationId: msg.CorrelationID.String(),
			// UserId:        msg.UserID.String(), // Needs rabbitmq permissions setup, also this is for request based users not background processes
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation if enabled
	if wq.client.config.PublisherConfirms {
		select {
		case confirm := <-wq.client.notifyConfirm:
			if !confirm.Ack {
				return fmt.Errorf("message not confirmed by broker")
			}
		case <-time.After(5 * time.Second):
			return fmt.Errorf("publish confirmation timeout")
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	wq.log.Info(ctx, "Message published",
		"queue_type", queueType,
		"message_id", msg.ID,
		"routing_key", routingKey)

	return nil
}

// PublishBatch publishes multiple messages
func (wq *WorkflowQueue) PublishBatch(ctx context.Context, queueType QueueType, messages []*Message) error {
	for _, msg := range messages {
		if err := wq.Publish(ctx, queueType, msg); err != nil {
			return fmt.Errorf("failed to publish message %s: %w", msg.ID, err)
		}
	}
	return nil
}

// =============================================================================
// Consuming Methods
// =============================================================================

// Consumer represents a message consumer
type Consumer struct {
	queue    QueueConfig
	handler  MessageHandler
	channel  *amqp.Channel
	delivery <-chan amqp.Delivery
	done     chan bool
	tag      string
}

// MessageHandler processes messages
type MessageHandler func(ctx context.Context, msg *Message) error

// Consume starts consuming messages from a queue
func (wq *WorkflowQueue) Consume(ctx context.Context, queueType QueueType, handler MessageHandler) (*Consumer, error) {
	config, exists := wq.queues[queueType]
	if !exists {
		return nil, fmt.Errorf("unknown queue type: %s", queueType)
	}

	// Get connection to create a new channel
	conn, err := wq.client.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Create a new channel for this consumer
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Set QoS
	err = ch.Qos(
		wq.client.config.PrefetchCount,
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming
	deliveries, err := ch.Consume(
		config.Name,
		"",    // consumer tag (auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	consumer := &Consumer{
		queue:    config,
		handler:  handler,
		channel:  ch,
		delivery: deliveries,
		done:     make(chan bool),
		tag:      fmt.Sprintf("consumer-%s-%d", queueType, time.Now().Unix()),
	}

	// Start processing in background
	go consumer.process(ctx, wq.log)

	wq.log.Info(ctx, "Consumer started",
		"queue_type", queueType,
		"queue_name", config.Name,
		"tag", consumer.tag)

	return consumer, nil
}

// process handles incoming messages
// In consumer's process method in client.go
// In consumer's process method in client.go
func (c *Consumer) process(ctx context.Context, log *logger.Logger) {
	for {
		select {
		case delivery, ok := <-c.delivery:
			if !ok {
				log.Info(ctx, "Delivery channel closed", "tag", c.tag)
				return
			}

			var msg Message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				log.Error(ctx, "Failed to unmarshal message",
					"error", err,
					"body", string(delivery.Body))
				delivery.Nack(false, false)
				continue
			}

			// Calculate actual attempt number based on rejection count
			actualAttempts := 1

			if xDeath, ok := delivery.Headers["x-death"].([]interface{}); ok {
				for _, death := range xDeath {
					if deathTable, ok := death.(amqp.Table); ok {
						reason, _ := deathTable["reason"].(string)
						count, _ := deathTable["count"].(int64)
						queue, _ := deathTable["queue"].(string)

						// Count rejections from the main queue
						if reason == "rejected" && queue == c.queue.Name {
							actualAttempts = int(count) + 1
						}
					}
				}
			}

			msg.Attempts = actualAttempts

			startTime := time.Now()
			err := c.handler(ctx, &msg)
			duration := time.Since(startTime)

			if err != nil {
				log.Error(ctx, "Failed to process message",
					"error", err,
					"message_id", msg.ID,
					"attempts", msg.Attempts,
					"max_attempts", msg.MaxAttempts,
					"duration", duration)

				if msg.Attempts < msg.MaxAttempts {
					log.Info(ctx, "Retrying message",
						"message_id", msg.ID,
						"attempt", msg.Attempts,
						"max_attempts", msg.MaxAttempts)
					delivery.Nack(false, false) // Send to DLX/retry queue
				} else {
					log.Error(ctx, "Message max retries exceeded",
						"message_id", msg.ID,
						"attempts", msg.Attempts,
						"max_attempts", msg.MaxAttempts)

					// Store the error for final failure recording
					msg.Payload["_final_failure_error"] = err.Error()
					msg.Payload["_final_failure"] = true

					// Call handler one more time to record the failure
					_ = c.handler(ctx, &msg)

					delivery.Ack(false) // Remove from queue
				}
			} else {
				delivery.Ack(false)
				log.Info(ctx, "Message processed successfully",
					"message_id", msg.ID,
					"attempts", msg.Attempts,
					"duration", duration)
			}

		case <-c.done:
			log.Info(ctx, "Consumer stopped", "tag", c.tag)
			return

		case <-ctx.Done():
			log.Info(ctx, "Consumer context cancelled", "tag", c.tag)
			return
		}
	}
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	close(c.done)
	return c.channel.Close()
}

// =============================================================================
// Management Methods
// =============================================================================

// GetQueueStats returns statistics for a queue
func (wq *WorkflowQueue) GetQueueStats(ctx context.Context, queueType QueueType) (*QueueStats, error) {
	config, exists := wq.queues[queueType]
	if !exists {
		return nil, fmt.Errorf("unknown queue type: %s", queueType)
	}

	ch, err := wq.client.GetChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	queue, err := ch.QueueInspect(config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return &QueueStats{
		Name:      config.Name,
		Messages:  queue.Messages,
		Consumers: queue.Consumers,
		Type:      string(queueType),
	}, nil
}

// QueueStats contains queue statistics
type QueueStats struct {
	Name      string `json:"name"`
	Messages  int    `json:"messages"`
	Consumers int    `json:"consumers"`
	Type      string `json:"type"`
}

// PurgeQueue removes all messages from a queue
func (wq *WorkflowQueue) PurgeQueue(ctx context.Context, queueType QueueType) error {
	config, exists := wq.queues[queueType]
	if !exists {
		return fmt.Errorf("unknown queue type: %s", queueType)
	}

	ch, err := wq.client.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	_, err = ch.QueuePurge(config.Name, false)
	if err != nil {
		return fmt.Errorf("failed to purge queue: %w", err)
	}

	wq.log.Info(ctx, "Queue purged", "queue_type", queueType, "name", config.Name)
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func generateMessageID() string {
	return fmt.Sprintf("msg_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
