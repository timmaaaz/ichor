// Package temporal provides support for starting and stopping Temporal
// containers for running tests.
package temporal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/foundation/docker"
	"go.temporal.io/sdk/client"
)

// Container represents a running Temporal container for testing.
type Container struct {
	docker.Container
	HostPort string
}

// StartTemporal starts a Temporal dev server container for running tests.
// Uses the temporalio/temporal CLI image with embedded SQLite for test
// isolation (no external database needed). The K8s dev cluster uses
// PostgreSQL via auto-setup, but tests use SQLite for simplicity.
func StartTemporal() (Container, error) {
	const (
		image = "temporalio/temporal:latest"
		name  = "test-temporal"
		port  = "7233"
	)

	// The temporalio/temporal image does not EXPOSE ports, so docker -P
	// won't publish anything. We explicitly map port 7233 to a random
	// host port with -p 0:7233 to allow parallel test runs.
	dockerArgs := []string{"-p", "0:7233"}

	// The temporalio/temporal image requires explicit command args
	// to run in dev server mode with all interfaces bound.
	appArgs := []string{"server", "start-dev", "--ip", "0.0.0.0"}

	c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
	if err != nil {
		return Container{}, fmt.Errorf("starting temporal container: %w", err)
	}

	// Fix the host address if it's 0.0.0.0
	hostPort := c.HostPort
	if strings.HasPrefix(hostPort, "0.0.0.0:") {
		hostPort = strings.Replace(hostPort, "0.0.0.0:", "localhost:", 1)
	}

	container := Container{
		Container: c,
		HostPort:  hostPort,
	}

	// Wait for Temporal to be ready
	if err := waitForReady(hostPort); err != nil {
		docker.StopContainer(c.Name)
		return Container{}, fmt.Errorf("waiting for temporal to be ready: %w", err)
	}

	return container, nil
}

// StopTemporal stops and removes the Temporal container.
func StopTemporal(c Container) error {
	return docker.StopContainer(c.Name)
}

// DumpLogs outputs logs from the Temporal container.
func DumpLogs(c Container) []byte {
	return docker.DumpContainerLogs(c.Name)
}

// waitForReady waits for Temporal to accept gRPC connections using the
// Temporal Go SDK client.Dial + CheckHealth pattern.
func waitForReady(hostPort string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for temporal at %s", hostPort)
		case <-ticker.C:
			c, err := client.Dial(client.Options{
				HostPort: hostPort,
			})
			if err != nil {
				continue
			}
			// Verify the server is actually healthy, not just accepting TCP
			_, err = c.CheckHealth(ctx, &client.CheckHealthRequest{})
			c.Close()
			if err == nil {
				return nil
			}
		}
	}
}

// NewTestClient creates a new Temporal client connected to the given host.
// This bypasses the singleton pattern for test isolation.
func NewTestClient(hostPort string) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort: hostPort,
	})
}

var (
	testContainer *Container
	testMu        sync.Mutex
	testStarted   bool
)

// GetTestContainer returns a shared Temporal container for tests.
func GetTestContainer(t *testing.T) Container {
	t.Helper()

	testMu.Lock()
	defer testMu.Unlock()

	if !testStarted {
		const image = "temporalio/temporal:latest"
		const name = "servicetest-temporal"
		const port = "7233"

		// Clean up any existing container
		docker.StopContainer(name)

		// The temporalio/temporal image does not EXPOSE ports, so docker -P
		// won't publish anything. We explicitly map port 7233 to a random
		// host port with -p 0:7233.
		dockerArgs := []string{"-p", "0:7233"}

		appArgs := []string{"server", "start-dev", "--ip", "0.0.0.0"}

		c, err := docker.StartContainer(image, name, port, dockerArgs, appArgs)
		if err != nil {
			t.Fatalf("starting temporal container: %s", err)
		}

		// Fix the host address if it's 0.0.0.0
		hostPort := c.HostPort
		if strings.HasPrefix(hostPort, "0.0.0.0:") {
			hostPort = strings.Replace(hostPort, "0.0.0.0:", "localhost:", 1)
		}

		container := Container{
			Container: c,
			HostPort:  hostPort,
		}

		if err := waitForReady(hostPort); err != nil {
			docker.StopContainer(c.Name)
			t.Fatalf("waiting for temporal: %s", err)
		}

		testContainer = &container
		testStarted = true

		t.Logf("Temporal Started: %s at %s", c.Name, hostPort)
	}

	return *testContainer
}
