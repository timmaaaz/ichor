package temporal

import (
	"context"
	"encoding/json"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Async Activity Handler Interface
// =============================================================================

// AsyncActivityHandler defines the interface for actions that complete
// asynchronously via Temporal's async completion pattern.
//
// Unlike workflow.AsyncActionHandler (which uses ProcessQueued for RabbitMQ),
// this interface is specifically for the Temporal activity side:
//  1. StartAsync publishes work with the task token
//  2. Activity returns ErrResultPending
//  3. External system calls AsyncCompleter.Complete with the task token
//
// Existing async handlers (SeekApprovalHandler, AllocateInventoryHandler)
// will be adapted to implement this interface in Phase 9.
type AsyncActivityHandler interface {
	workflow.ActionHandler

	// StartAsync initiates the async operation.
	// taskToken must be forwarded to the external system for later completion.
	// The handler should publish to RabbitMQ (or other queue) with the task token
	// as a correlation identifier.
	StartAsync(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext, taskToken []byte) error
}

// =============================================================================
// Async Registry
// =============================================================================

// AsyncRegistry holds async action handlers separate from synchronous ones.
// Used by Activities.ExecuteAsyncActionActivity to dispatch async actions.
type AsyncRegistry struct {
	handlers map[string]AsyncActivityHandler
}

// NewAsyncRegistry creates a new async handler registry.
func NewAsyncRegistry() *AsyncRegistry {
	return &AsyncRegistry{
		handlers: make(map[string]AsyncActivityHandler),
	}
}

// Register adds an async handler for an action type.
func (r *AsyncRegistry) Register(actionType string, handler AsyncActivityHandler) {
	r.handlers[actionType] = handler
}

// Get retrieves an async handler by action type.
func (r *AsyncRegistry) Get(actionType string) (AsyncActivityHandler, bool) {
	h, ok := r.handlers[actionType]
	return h, ok
}
