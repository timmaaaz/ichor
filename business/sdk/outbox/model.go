// Package outbox implements the transactional outbox for cascade delegate events
// (F2). A cascade-relevant domain write persists one Outbox row in the SAME unit
// of work as the entity write (see Writer.Emit); a polling relay
// (business/sdk/workflow/temporal/relay.go) drains pending rows into
// WorkflowTrigger.OnEntityEvent at-least-once, keyed for dedup on the row id.
//
// This package is deliberately free of any dependency on the temporal package so
// the relay (which lives in temporal) can import it without an import cycle. The
// two facts an Outbox row needs that delegate.Data does not carry — the workflow
// entity name (resolved from the delegate domain) and the serialized cascade
// lineage (which rides an unexported temporal context key) — are supplied to the
// Writer as injected dependencies by the composition root (all.go / the worker).
package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// Outbox is one persisted cascade event (a row of workflow.cascade_outbox).
//
// Payload is the JSON-encoded delegate.Data the originating bus emitted; the relay
// re-hydrates the TriggerEvent from it (DESIGN §3). Lineage is the JSON-encoded
// cascade loop-guard visited-set (nil for a human/non-workflow write, which starts
// a fresh chain). ID doubles as the Temporal workflow-id dedup key; Seq imposes a
// total order on the relay.
type Outbox struct {
	ID          uuid.UUID
	Seq         int64
	Domain      string
	Action      string
	EventType   string
	EntityName  string
	Payload     []byte
	Lineage     []byte
	CreatedAt   time.Time
	Attempts    int
	LastError   string
	PublishedAt *time.Time
	Dead        bool
}

// eventTypeForAction maps a delegate action to the workflow trigger event type,
// mirroring the action→event_type wiring DelegateHandler.RegisterDomain hard-codes
// (delegatehandler.go). Cascade buses only ever emit the three CRUD actions; an
// unrecognized action returns ("", false) so the caller can warn rather than
// silently persist a row no rule could match.
func eventTypeForAction(action string) (string, bool) {
	switch action {
	case workflow.ActionCreated:
		return workflow.EventTypeOnCreate, true
	case workflow.ActionUpdated:
		return workflow.EventTypeOnUpdate, true
	case workflow.ActionDeleted:
		return workflow.EventTypeOnDelete, true
	default:
		return "", false
	}
}

// entityIDFromData best-effort extracts the entity id a delegate event concerns,
// used only for logging/observability at the emit seam (the authoritative id is
// re-derived by the relay from the payload, exactly as DelegateHandler does). It
// reads delegate.Data's RawParams as DelegateEventParams; a zero id is acceptable
// and never blocks emission.
func entityIDFromData(data delegate.Data) uuid.UUID {
	var params workflow.DelegateEventParams
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		return uuid.Nil
	}
	return params.EntityID
}
