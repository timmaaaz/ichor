package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// DelegateHandler bridges domain delegate events to Temporal workflow dispatch.
// It registers handlers for domain CRUD events and converts them to TriggerEvents
// for WorkflowTrigger.OnEntityEvent().
type DelegateHandler struct {
	log     *logger.Logger
	trigger *WorkflowTrigger
}

// NewDelegateHandler creates a handler bridging delegate events to Temporal.
func NewDelegateHandler(log *logger.Logger, trigger *WorkflowTrigger) *DelegateHandler {
	return &DelegateHandler{
		log:     log,
		trigger: trigger,
	}
}

// RegisterDomain registers delegate handlers for a domain's CRUD events.
// entityName should match the workflow entity name (e.g., "orders").
func (h *DelegateHandler) RegisterDomain(del *delegate.Delegate, domainName, entityName string) {
	del.Register(domainName, workflow.ActionCreated, func(ctx context.Context, data delegate.Data) error {
		return h.handleEvent(ctx, "on_create", entityName, data)
	})
	del.Register(domainName, workflow.ActionUpdated, func(ctx context.Context, data delegate.Data) error {
		return h.handleEvent(ctx, "on_update", entityName, data)
	})
	del.Register(domainName, workflow.ActionDeleted, func(ctx context.Context, data delegate.Data) error {
		return h.handleEvent(ctx, "on_delete", entityName, data)
	})

	h.log.Info(context.Background(), "temporal delegate handler registered",
		"domain", domainName, "entity", entityName)
}

func (h *DelegateHandler) handleEvent(ctx context.Context, eventType, entityName string, data delegate.Data) error {
	var params workflow.DelegateEventParams
	if err := json.Unmarshal(data.RawParams, &params); err != nil {
		h.log.Error(ctx, "temporal delegate: unmarshal params failed",
			"entity", entityName, "event_type", eventType, "error", err)
		return nil // Don't fail the delegate chain
	}

	// Build TriggerEvent from delegate data.
	event := workflow.TriggerEvent{
		EventType:  eventType,
		EntityName: entityName,
		EntityID:   params.EntityID,
		Timestamp:  time.Now().UTC(),
		UserID:     params.UserID,
	}

	// Extract entity raw data if present.
	if params.Entity != nil {
		entityID, rawData, err := extractEntityData(params.Entity)
		if err == nil {
			event.RawData = rawData
			if event.EntityID == uuid.Nil {
				event.EntityID = entityID
			}
		}
	}

	// Fire in goroutine to avoid blocking the delegate chain.
	go func() {
		if err := h.trigger.OnEntityEvent(context.Background(), event); err != nil {
			h.log.Error(context.Background(), "temporal delegate: dispatch failed",
				"entity", entityName, "event_type", eventType, "error", err)
		}
	}()

	return nil
}

// extractEntityData extracts ID and raw data from an entity result.
// Copied from eventpublisher.go (source file being deleted in Task 5).
func extractEntityData(result any) (uuid.UUID, map[string]any, error) {
	if result == nil {
		return uuid.Nil, nil, fmt.Errorf("nil result")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("marshal result: %w", err)
	}

	var rawData map[string]any
	if err := json.Unmarshal(data, &rawData); err != nil {
		return uuid.Nil, nil, fmt.Errorf("unmarshal to map: %w", err)
	}

	// Extract ID from JSON (app layer uses string IDs).
	var entityID uuid.UUID
	if id, ok := rawData["id"].(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			entityID = parsed
		}
	}

	// Fallback: reflection for struct field ID.
	if entityID == uuid.Nil {
		entityID = extractIDViaReflection(result)
	}

	return entityID, rawData, nil
}

func extractIDViaReflection(result any) uuid.UUID {
	val := reflect.ValueOf(result)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return uuid.Nil
	}

	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return uuid.Nil
	}

	switch id := idField.Interface().(type) {
	case uuid.UUID:
		return id
	case string:
		if parsed, err := uuid.Parse(id); err == nil {
			return parsed
		}
	}
	return uuid.Nil
}
