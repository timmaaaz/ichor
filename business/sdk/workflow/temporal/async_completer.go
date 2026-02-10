package temporal

import (
	"context"
	"fmt"
)

// ActivityCompleter is a narrow interface for completing async activities.
// Extracted from client.Client (~50 methods) for testability.
// client.Client satisfies this interface (backwards-compatible).
type ActivityCompleter interface {
	CompleteActivity(ctx context.Context, taskToken []byte, result any, err error) error
}

// AsyncCompleter completes Temporal activities from external systems.
//
// Usage pattern:
//  1. Async activity starts (ExecuteAsyncActionActivity) and publishes work with task token
//  2. External system processes the work (e.g., RabbitMQ consumer, webhook)
//  3. External system calls Complete or Fail with the task token
//  4. Temporal resumes the workflow with the result
type AsyncCompleter struct {
	client ActivityCompleter
}

// NewAsyncCompleter creates a completer for async activities.
func NewAsyncCompleter(c ActivityCompleter) *AsyncCompleter {
	return &AsyncCompleter{client: c}
}

// Complete finishes an async activity with a successful result.
// taskToken is the correlation ID from activity.GetInfo(ctx).TaskToken,
// forwarded by the async handler to the external system.
func (c *AsyncCompleter) Complete(ctx context.Context, taskToken []byte, result ActionActivityOutput) error {
	if err := c.client.CompleteActivity(ctx, taskToken, result, nil); err != nil {
		return fmt.Errorf("complete async activity: %w", err)
	}
	return nil
}

// Fail finishes an async activity with an error.
// The workflow will see this as an activity failure and may retry
// depending on the retry policy configured in workflow.go.
func (c *AsyncCompleter) Fail(ctx context.Context, taskToken []byte, activityErr error) error {
	if err := c.client.CompleteActivity(ctx, taskToken, nil, activityErr); err != nil {
		return fmt.Errorf("fail async activity: %w", err)
	}
	return nil
}
