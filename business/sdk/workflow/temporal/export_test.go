package temporal

import (
	"context"

	"github.com/timmaaaz/ichor/business/sdk/outbox"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Test-only accessors that let the EXTERNAL relay test (package temporal_test) exercise
// unexported relay internals. The relay test must be external because it imports dbtest,
// which since F8 imports this package (for MarshalLineageFromContext); an INTERNAL test
// importing dbtest would form an import cycle (temporal test -> dbtest -> temporal).

// NewRelayForTest builds a Relay with only a logger, for unit-testing buildEvent without
// a db or dispatcher.
func NewRelayForTest(log *logger.Logger) *Relay {
	return &Relay{log: log}
}

// BuildEvent exposes the unexported buildEvent for the external relay test.
func (r *Relay) BuildEvent(ctx context.Context, row outbox.Outbox) (workflow.TriggerEvent, bool) {
	return r.buildEvent(ctx, row)
}

// DecodeLineage exposes the unexported decodeLineage for the external relay test.
func DecodeLineage(b []byte) WorkflowLineage {
	return decodeLineage(b)
}
