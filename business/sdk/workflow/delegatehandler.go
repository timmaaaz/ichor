// Package workflow provides workflow automation infrastructure including
// event publishing, queue management, and delegate integration.
package workflow

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// DelegateHandler bridges domain events from the delegate pattern to the
// workflow event publisher. It registers handlers for domain events and
// publishes them as workflow trigger events.
type DelegateHandler struct {
	log       *logger.Logger
	publisher *EventPublisher
}

// NewDelegateHandler creates a new handler that bridges delegate events
// to workflow events.
func NewDelegateHandler(log *logger.Logger, publisher *EventPublisher) *DelegateHandler {
	return &DelegateHandler{
		log:       log,
		publisher: publisher,
	}
}

// RegisterDomain registers delegate handlers for a domain's CRUD events.
// This maps domain actions (created, updated, deleted) to workflow events
// (on_create, on_update, on_delete).
//
// entityName should match the workflow entity name (e.g., "sales.orders").
func (h *DelegateHandler) RegisterDomain(del *delegate.Delegate, domainName, entityName string) {
	// Register created action -> on_create event
	del.Register(domainName, ActionCreated, func(ctx context.Context, data delegate.Data) error {
		return h.handleCreated(ctx, entityName, data)
	})

	// Register updated action -> on_update event
	del.Register(domainName, ActionUpdated, func(ctx context.Context, data delegate.Data) error {
		return h.handleUpdated(ctx, entityName, data)
	})

	// Register deleted action -> on_delete event
	del.Register(domainName, ActionDeleted, func(ctx context.Context, data delegate.Data) error {
		return h.handleDeleted(ctx, entityName, data)
	})

	h.log.Info(context.Background(), "workflow delegate handler registered",
		"domain", domainName,
		"entity", entityName)
}

// Standard action names used by domain event.go files
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// DelegateEventParams is the standard structure for delegate event parameters.
// Domain event.go files should use this structure for their ActionXxxParms types.
// The UserID field is extracted from the entity's CreatedBy/UpdatedBy fields
// to identify who triggered the action.
type DelegateEventParams struct {
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"` // Who triggered this action
	Entity   any       `json:"entity,omitempty"`
}

func (h *DelegateHandler) handleCreated(ctx context.Context, entityName string, data delegate.Data) error {
	var params DelegateEventParams
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		h.log.Error(ctx, "workflow delegate: unmarshal created params failed",
			"entity", entityName,
			"error", err)
		return nil // Don't fail the delegate chain
	}

	h.publisher.PublishCreateEvent(ctx, entityName, params.Entity, params.UserID)
	return nil
}

func (h *DelegateHandler) handleUpdated(ctx context.Context, entityName string, data delegate.Data) error {
	var params DelegateEventParams
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		h.log.Error(ctx, "workflow delegate: unmarshal updated params failed",
			"entity", entityName,
			"error", err)
		return nil // Don't fail the delegate chain
	}

	// Note: FieldChanges is nil in Phase 2. Full change tracking requires
	// passing old vs new values through the delegate, which is Phase 3 work.
	h.publisher.PublishUpdateEvent(ctx, entityName, params.Entity, nil, params.UserID)
	return nil
}

func (h *DelegateHandler) handleDeleted(ctx context.Context, entityName string, data delegate.Data) error {
	var params DelegateEventParams
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		h.log.Error(ctx, "workflow delegate: unmarshal deleted params failed",
			"entity", entityName,
			"error", err)
		return nil // Don't fail the delegate chain
	}

	h.publisher.PublishDeleteEvent(ctx, entityName, params.EntityID, params.UserID)
	return nil
}
