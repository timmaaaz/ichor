# RabbitMQ Foundation Package - Developer Guide

This guide covers the RabbitMQ foundation package built for the Ardan Labs architecture pattern. The package provides a production-ready RabbitMQ client with automatic reconnection, workflow queuing, and comprehensive testing utilities.

## Package Structure

```
foundation/rabbitmq/
├── client.go         # Main RabbitMQ client and workflow queue implementation
├── rabbitmq.go       # Docker container management for testing
└── client_test.go    # Integration tests
```

## Core Components

### 1. Client (`client.go`)

The main RabbitMQ client with singleton pattern, automatic reconnection, and publisher confirms.

**Key Features:**

- Singleton connection management
- Automatic reconnection with configurable retries
- Publisher confirmations for message reliability
- Thread-safe operations with mutex protection
- Dead letter queue support

### 2. WorkflowQueue (`client.go`)

High-level abstraction for workflow-specific message queuing with predefined queue types and configurations.

**Supported Queue Types:**

- `QueueTypeWorkflow` - General workflow messages
- `QueueTypeApproval` - Approval workflow messages (72h TTL)
- `QueueTypeNotification` - Notification messages (1h TTL)
- `QueueTypeInventory` - Inventory-related messages (2h TTL)
- `QueueTypeEmail` - Email messages (24h TTL, 5 retries)
- `QueueTypeAlert` - Alert messages (30min TTL, 1 retry)

### 3. Container Management (`rabbitmq.go`)

Docker-based RabbitMQ instances for development and testing.

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "log"

    "github.com/timmaaaz/ichor/foundation/logger"
    "github.com/timmaaaz/ichor/foundation/rabbitmq"
)

func main() {
    // Create logger
    var buf bytes.Buffer
    log := logger.New(&buf, logger.LevelInfo, "APP", traceIDFunc)

    // Configure RabbitMQ
    config := rabbitmq.DefaultConfig()
    config.URL = "amqp://guest:guest@localhost:5672/"

    // Create client (singleton)
    client := rabbitmq.NewClient(log, config)

    // Create workflow queue manager
    wq := rabbitmq.NewWorkflowQueue(client, log)

    // Initialize queues
    ctx := context.Background()
    if err := wq.Initialize(ctx); err != nil {
        log.Fatal("Failed to initialize queues:", err)
    }
}
```

### Publishing Messages

```go
// Create a workflow message
msg := &rabbitmq.Message{
    Type:       "user_registration",
    EntityName: "user",
    EntityID:   "user_123",
    EventType:  "on_create",
    Priority:   uint8(rabbitmq.PriorityNormal),
    Payload: map[string]interface{}{
        "email":      "user@example.com",
        "created_at": time.Now(),
    },
}

// Publish to workflow queue
err := wq.Publish(ctx, rabbitmq.QueueTypeWorkflow, msg)
if err != nil {
    log.Error("Failed to publish message:", err)
}

// Publish with delay
err = wq.PublishWithDelay(ctx, rabbitmq.QueueTypeWorkflow, msg, 5*time.Minute)

// Batch publish
messages := []*rabbitmq.Message{msg1, msg2, msg3}
err = wq.PublishBatch(ctx, rabbitmq.QueueTypeWorkflow, messages)
```

### Consuming Messages

```go
// Define message handler
handler := func(ctx context.Context, msg *rabbitmq.Message) error {
    log.Info("Processing message", "id", msg.ID, "type", msg.Type)

    // Process the message
    switch msg.EventType {
    case "on_create":
        return handleUserCreation(msg)
    case "on_update":
        return handleUserUpdate(msg)
    default:
        return fmt.Errorf("unknown event type: %s", msg.EventType)
    }
}

// Start consumer
consumer, err := wq.Consume(ctx, rabbitmq.QueueTypeWorkflow, handler)
if err != nil {
    log.Fatal("Failed to start consumer:", err)
}

// Stop consumer when done
defer consumer.Stop()
```

## Configuration

### Default Configuration

```go
config := rabbitmq.DefaultConfig()
// Results in:
// URL: "amqp://guest:guest@localhost:5672/"
// MaxRetries: 5
// RetryDelay: 5 * time.Second
// PrefetchCount: 10
// PublisherConfirms: true
// ExchangeName: "workflow"
// ExchangeType: "topic"
// DeadLetterExchange: "workflow.dlx"
```

### Custom Configuration

```go
config := rabbitmq.Config{
    URL:                "amqp://user:pass@prod-rabbitmq:5672/",
    MaxRetries:         10,
    RetryDelay:         2 * time.Second,
    PrefetchCount:      20,
    PrefetchSize:       0,
    PublisherConfirms:  true,
    ExchangeName:       "production_workflow",
    ExchangeType:       "topic",
    DeadLetterExchange: "production_workflow.dlx",
}
```

## Message Priority Levels

```go
const (
    PriorityLow      QueuePriority = 1
    PriorityNormal   QueuePriority = 5
    PriorityHigh     QueuePriority = 8
    PriorityCritical QueuePriority = 10
)

// Set message priority
msg.Priority = uint8(rabbitmq.PriorityCritical)
```

## Error Handling and Retries

The package automatically handles:

- **Connection failures** - Automatic reconnection with configurable retries
- **Message processing errors** - Configurable retry attempts per queue type
- **Dead letter queues** - Failed messages are sent to DLQ after max retries
- **Publisher confirms** - Ensures message delivery to broker

### Retry Configuration per Queue Type

```go
// Approval queue: 5 retries, 72-hour TTL
// Email queue: 5 retries, 24-hour TTL
// Alert queue: 1 retry, 30-minute TTL
// Workflow queue: 3 retries, 24-hour TTL
```

## Monitoring and Management

### Queue Statistics

```go
stats, err := wq.GetQueueStats(ctx, rabbitmq.QueueTypeWorkflow)
if err != nil {
    log.Error("Failed to get stats:", err)
    return
}

fmt.Printf("Queue: %s, Messages: %d, Consumers: %d\n",
    stats.Name, stats.Messages, stats.Consumers)
```

### Queue Management

```go
// Purge all messages from a queue
err := wq.PurgeQueue(ctx, rabbitmq.QueueTypeWorkflow)

// Check connection status
if client.IsConnected() {
    log.Info("RabbitMQ is connected")
}

// Wait for connection (useful during startup)
err := client.WaitForConnection(30 * time.Second)
```

## Testing

### Running Tests

```bash
# Run all tests (requires Docker)
go test ./foundation/rabbitmq/...

# Run without integration tests
go test -short ./foundation/rabbitmq/...
```

### Test Container Management

```go
// Start RabbitMQ container for tests
container, err := rabbitmq.StartRabbitMQ()
if err != nil {
    t.Fatal("Failed to start RabbitMQ:", err)
}
defer rabbitmq.StopRabbitMQ(container)

// Create test client
client := rabbitmq.NewTestClient(container.URL)
```

### Test Setup Example

```go
func TestMyWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping RabbitMQ integration test in short mode")
    }

    // Use global test container from TestMain
    client := rabbitmq.NewTestClient(testContainer.URL)
    if err := client.Connect(); err != nil {
        t.Fatal("Failed to connect:", err)
    }
    defer client.Close()

    // Create workflow queue
    wq := rabbitmq.NewWorkflowQueue(client, testLogger)
    if err := wq.Initialize(context.Background()); err != nil {
        t.Fatal("Failed to initialize:", err)
    }

    // Your test logic here...
}
```

## Best Practices

### 1. Connection Management

- Use the singleton client for production
- Always check connection status before critical operations
- Implement graceful shutdown procedures

### 2. Message Design

- Include correlation IDs for tracing
- Set appropriate priorities based on business requirements
- Use structured payloads with clear schemas

### 3. Error Handling

- Implement idempotent message handlers
- Log processing errors with context
- Monitor dead letter queues for failed messages

### 4. Performance

- Configure appropriate prefetch counts based on processing capacity
- Use batch publishing for high-throughput scenarios
- Monitor queue depths and consumer counts

### 5. Testing

- Use dedicated test containers for isolation
- Clean queue state between tests with `PurgeQueue`
- Test both success and failure scenarios

## Production Considerations

### Health Checks

```go
func healthCheck() error {
    if !client.IsConnected() {
        return fmt.Errorf("RabbitMQ not connected")
    }
    return nil
}
```

### Graceful Shutdown

```go
func shutdown(ctx context.Context) error {
    // Stop consumers first
    for _, consumer := range consumers {
        if err := consumer.Stop(); err != nil {
            log.Error("Failed to stop consumer:", err)
        }
    }

    // Close client connection
    return client.Close()
}
```

### Monitoring

Monitor these key metrics:

- Queue depths and growth rates
- Message processing times
- Connection stability
- Dead letter queue accumulation
- Consumer count and distribution

## Troubleshooting

### Common Issues

1. **Connection failures** - Check network connectivity and credentials
2. **Messages not processing** - Verify consumer handlers and error logs
3. **High queue depths** - Scale consumers or optimize processing logic
4. **Memory issues** - Adjust prefetch counts and message TTLs

### Debug Logging

```go
// Enable debug logging
log := logger.New(os.Stdout, logger.LevelDebug, "DEBUG", traceIDFunc)
```

### Container Logs

```go
// Dump RabbitMQ container logs for debugging
logs := rabbitmq.DumpLogs(container)
fmt.Printf("RabbitMQ logs:\n%s", logs)
```
