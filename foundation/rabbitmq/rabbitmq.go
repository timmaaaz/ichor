// foundation/rabbitmq/rabbitmq.go
package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
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
	// Note: We do NOT specify -p port mappings here.
	// The docker.StartContainer uses -P to assign random host ports,
	// which allows tests to run alongside a dev server using port 5672.
	dockerArgs := []string{
		"-e", "RABBITMQ_DEFAULT_USER=guest",
		"-e", "RABBITMQ_DEFAULT_PASS=guest",
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

	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

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
	log := logger.New(os.Stdout, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })

	config := NewTestConfig(url)

	return &Client{
		url:    config.URL,
		log:    log,
		config: config,
	}
}

var (
	testContainer *Container
	testMu        sync.Mutex
	testStarted   bool
)

// GetTestContainer returns a shared RabbitMQ container for tests
func GetTestContainer(t *testing.T) Container {
	t.Helper()

	testMu.Lock()
	defer testMu.Unlock()

	if !testStarted {
		const image = "rabbitmq:3-management"
		const name = "servicetest-rabbit"
		const port = "5672"

		// Clean up any existing container
		docker.StopContainer(name)

		dockerArgs := []string{
			"-e", "RABBITMQ_DEFAULT_USER=guest",
			"-e", "RABBITMQ_DEFAULT_PASS=guest",
			// Note: We do NOT specify -p port mappings here.
			// The docker.StartContainer uses -P to assign random host ports,
			// which allows tests to run alongside a dev server using port 5672.
		}

		c, err := docker.StartContainer(image, name, port, dockerArgs, []string{})
		if err != nil {
			t.Fatalf("starting rabbitmq container: %s", err)
		}

		// Fix the host address if it's 0.0.0.0
		hostPort := c.HostPort
		if strings.HasPrefix(hostPort, "0.0.0.0:") {
			hostPort = strings.Replace(hostPort, "0.0.0.0:", "localhost:", 1)
		}

		container := Container{
			Container: c,
			URL:       fmt.Sprintf("amqp://guest:guest@%s/", hostPort),
		}

		if err := waitForReady(container.URL); err != nil {
			docker.StopContainer(c.Name)
			t.Fatalf("waiting for rabbitmq: %s", err)
		}

		testContainer = &container
		testStarted = true

		t.Logf("RabbitMQ Started: %s at %s (URL: %s)", c.Name, hostPort, container.URL)
	}

	return *testContainer
}

// NewTestWorkflowQueue creates a WorkflowQueue with unique queue names for test isolation.
// Each call generates a unique prefix based on timestamp, ensuring tests don't interfere
// with each other even when running in parallel or alongside a dev server.
func NewTestWorkflowQueue(client *Client, log *logger.Logger) *WorkflowQueue {
	prefix := fmt.Sprintf("test_%d_", time.Now().UnixNano())
	return NewWorkflowQueueWithPrefix(client, log, prefix)
}
