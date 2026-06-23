// Package executionapp provides the application layer for workflow execution operations.
package executionapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// Reranner re-fires the rule behind an execution with a fresh execution id.
// *temporal.WorkflowTrigger satisfies this interface.
type Reranner interface {
	RerunExecution(ctx context.Context, executionID uuid.UUID) (uuid.UUID, error)
}

// App is the application layer for execution operations.
type App struct {
	rerunner Reranner
}

// NewApp constructs an App. rerunner may be nil when the workflow engine
// (Temporal) is disabled; Rerun then returns an Internal error.
func NewApp(rerunner Reranner) *App {
	return &App{rerunner: rerunner}
}

// Rerun re-runs the given execution and returns the original + new execution ids.
func (a *App) Rerun(ctx context.Context, executionID uuid.UUID) (RerunResponse, error) {
	if a.rerunner == nil {
		return RerunResponse{}, errs.Newf(errs.Internal, "workflow engine is not enabled")
	}

	newID, err := a.rerunner.RerunExecution(ctx, executionID)
	if err != nil {
		switch {
		case errors.Is(err, workflow.ErrNotFound):
			return RerunResponse{}, errs.New(errs.NotFound, err)
		case errors.Is(err, temporal.ErrExecutionNotRerunnable):
			return RerunResponse{}, errs.New(errs.FailedPrecondition, err)
		default:
			return RerunResponse{}, errs.Newf(errs.Internal, "rerun execution: %s", err)
		}
	}

	return RerunResponse{OriginalExecutionID: executionID, NewExecutionID: newID}, nil
}
