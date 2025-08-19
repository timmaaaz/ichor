// foundation/rabbitmq/rabbitmq.go
package rabbitmq

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/timmaaaz/ichor/foundation/docker"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Container represents a running RabbitMQ container for testing
type Container struct {
	docker.Container
	URL string
}

// StartRabbitMQ starts a RabbitMQ container for running tests
func StartRabbitMQ() (Container, error) {
	const (
		image = "rabbitmq:3-management"
		name  = "test-rabbitmq"
		port  = "5672"
	)

	// Docker arguments for RabbitMQ
	dockerArgs := []string{
		"-e", "RABBITMQ_DEFAULT_USER=guest",
		"-e", "RABBITMQ_DEFAULT_PASS=guest",
		"-p", "5672:5672", // AMQP port
		"-p", "15672:15672", // Management UI port
	}

	// No additional app arguments needed for RabbitMQ
	appArgs := []string{}

	c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
	if err != nil {
		return Container{}, fmt.Errorf("starting rabbitmq container: %w", err)
	}

	container := Container{
		Container: c,
		URL:       fmt.Sprintf("amqp://guest:guest@%s/", c.HostPort),
	}

	// Wait for RabbitMQ to be ready
	if err := waitForReady(container.URL); err != nil {
		docker.StopContainer(c.Name)
		return Container{}, fmt.Errorf("waiting for rabbitmq to be ready: %w", err)
	}

	return container, nil
}

// StopRabbitMQ stops and removes the RabbitMQ container
func StopRabbitMQ(c Container) error {
	return docker.StopContainer(c.Name)
}

// DumpLogs outputs logs from the RabbitMQ container
func DumpLogs(c Container) []byte {
	return docker.DumpContainerLogs(c.Name)
}

// waitForReady waits for RabbitMQ to be ready to accept connections
func waitForReady(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	config := Config{
		URL:                url,
		MaxRetries:         1,
		RetryDelay:         100 * time.Millisecond,
		PrefetchCount:      10,
		PrefetchSize:       0,
		PublisherConfirms:  true,
		ExchangeName:       "test",
		ExchangeType:       "topic",
		DeadLetterExchange: "test.dlx",
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for rabbitmq")
		case <-ticker.C:
			// Create a new client for each attempt to avoid singleton issues
			client := &Client{
				url:    config.URL,
				log:    log,
				config: config,
			}
			if err := client.Connect(); err == nil {
				client.Close()
				return nil
			}
		}
	}
}

// NewTestConfig returns a RabbitMQ configuration suitable for testing
func NewTestConfig(url string) Config {
	return Config{
		URL:                url,
		MaxRetries:         3,
		RetryDelay:         100 * time.Millisecond,
		PrefetchCount:      10,
		PrefetchSize:       0,
		PublisherConfirms:  true,
		ExchangeName:       "test_workflow",
		ExchangeType:       "topic",
		DeadLetterExchange: "test_workflow.dlx",
	}
}

// NewTestClient creates a new RabbitMQ client configured for testing
// This bypasses the singleton pattern for test isolation
func NewTestClient(url string) *Client {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	config := NewTestConfig(url)

	return &Client{
		url:    config.URL,
		log:    log,
		config: config,
	}
}
