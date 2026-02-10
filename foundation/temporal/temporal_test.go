package temporal_test

import (
	"context"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/foundation/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func TestGetTestContainer(t *testing.T) {
	c := temporal.GetTestContainer(t)

	if c.HostPort == "" {
		t.Fatal("expected non-empty HostPort")
	}
	t.Logf("Container HostPort: %s", c.HostPort)

	// Verify singleton returns same container
	c2 := temporal.GetTestContainer(t)
	if c.HostPort != c2.HostPort {
		t.Fatalf("expected same container, got %s and %s", c.HostPort, c2.HostPort)
	}
}

func TestNewTestClient(t *testing.T) {
	c := temporal.GetTestContainer(t)

	tc, err := temporal.NewTestClient(c.HostPort)
	if err != nil {
		t.Fatalf("creating test client: %s", err)
	}
	defer tc.Close()

	// Verify the client can check health
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = tc.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("health check failed: %s", err)
	}
}

// helloWorkflow is a simple workflow for testing.
func helloWorkflow(ctx workflow.Context, name string) (string, error) {
	return "Hello, " + name + "!", nil
}

func TestSimpleWorkflow(t *testing.T) {
	c := temporal.GetTestContainer(t)

	tc, err := temporal.NewTestClient(c.HostPort)
	if err != nil {
		t.Fatalf("creating test client: %s", err)
	}
	defer tc.Close()

	const taskQueue = "test-simple-workflow"

	// Create a worker and register the workflow
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(helloWorkflow)

	if err := w.Start(); err != nil {
		t.Fatalf("starting worker: %s", err)
	}
	defer w.Stop()

	// Execute the workflow
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	run, err := tc.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		TaskQueue: taskQueue,
	}, helloWorkflow, "Temporal")
	if err != nil {
		t.Fatalf("executing workflow: %s", err)
	}

	// Get the result
	var result string
	if err := run.Get(ctx, &result); err != nil {
		t.Fatalf("getting workflow result: %s", err)
	}

	expected := "Hello, Temporal!"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}

	t.Logf("Workflow result: %s", result)
}
