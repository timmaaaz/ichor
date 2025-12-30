package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// EventPublisher provides non-blocking workflow event publishing.
// Events are queued asynchronously - failures are logged but never block
// the primary operation.
type EventPublisher struct {
	log          *logger.Logger
	queueManager *QueueManager
}

// NewEventPublisher creates a new event publisher.
func NewEventPublisher(log *logger.Logger, qm *QueueManager) *EventPublisher {
	return &EventPublisher{
		log:          log,
		queueManager: qm,
	}
}

// PublishCreateEvent fires an on_create event for the given entity.
func (ep *EventPublisher) PublishCreateEvent(ctx context.Context, entityName string, result any, userID uuid.UUID) {
	ep.publishEvent(ctx, "on_create", entityName, result, nil, userID)
}

// PublishUpdateEvent fires an on_update event with optional field changes.
func (ep *EventPublisher) PublishUpdateEvent(ctx context.Context, entityName string, result any, fieldChanges map[string]FieldChange, userID uuid.UUID) {
	ep.publishEvent(ctx, "on_update", entityName, result, fieldChanges, userID)
}

// PublishDeleteEvent fires an on_delete event.
func (ep *EventPublisher) PublishDeleteEvent(ctx context.Context, entityName string, entityID uuid.UUID, userID uuid.UUID) {
	event := TriggerEvent{
		EventType:  "on_delete",
		EntityName: entityName,
		EntityID:   entityID,
		Timestamp:  time.Now().UTC(),
		UserID:     userID,
	}
	ep.queueEventNonBlocking(ctx, event)
}

func (ep *EventPublisher) publishEvent(ctx context.Context, eventType, entityName string, result any, fieldChanges map[string]FieldChange, userID uuid.UUID) {
	entityID, rawData, err := ep.extractEntityData(result)
	if err != nil {
		ep.log.Error(ctx, "workflow event: extract entity data failed",
			"entityName", entityName,
			"eventType", eventType,
			"error", err)
		return
	}

	event := TriggerEvent{
		EventType:    eventType,
		EntityName:   entityName,
		EntityID:     entityID,
		FieldChanges: fieldChanges,
		Timestamp:    time.Now().UTC(),
		RawData:      rawData,
		UserID:       userID,
	}

	ep.queueEventNonBlocking(ctx, event)
}

func (ep *EventPublisher) queueEventNonBlocking(ctx context.Context, event TriggerEvent) {
	// Fire in goroutine to avoid blocking the primary operation
	go func() {
		queueCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := ep.queueManager.QueueEvent(queueCtx, event); err != nil {
			ep.log.Error(queueCtx, "workflow event: queue failed",
				"entityName", event.EntityName,
				"entityID", event.EntityID,
				"eventType", event.EventType,
				"error", err)
			// Future: fire notification action for admin alerting
		} else {
			ep.log.Info(queueCtx, "workflow event queued",
				"entityName", event.EntityName,
				"entityID", event.EntityID,
				"eventType", event.EventType)
		}
	}()
}

// extractEntityData extracts ID and raw data from entity result.
func (ep *EventPublisher) extractEntityData(result any) (uuid.UUID, map[string]any, error) {
	if result == nil {
		return uuid.Nil, nil, fmt.Errorf("nil result")
	}

	// JSON marshal/unmarshal to get map representation
	data, err := json.Marshal(result)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("marshal result: %w", err)
	}

	var rawData map[string]any
	if err := json.Unmarshal(data, &rawData); err != nil {
		return uuid.Nil, nil, fmt.Errorf("unmarshal to map: %w", err)
	}

	// Extract ID from JSON (app layer uses string IDs)
	var entityID uuid.UUID
	if id, ok := rawData["id"].(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			entityID = parsed
		}
	}

	// Fallback: reflection for struct field ID
	if entityID == uuid.Nil {
		entityID = ep.extractIDViaReflection(result)
	}

	return entityID, rawData, nil
}

func (ep *EventPublisher) extractIDViaReflection(result any) uuid.UUID {
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
