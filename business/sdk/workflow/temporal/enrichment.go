package temporal

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// Enrichment helpers shared by the cascade relay (relay.go) to rebuild a
// workflow.TriggerEvent from a persisted outbox row's delegate.Data payload.
//
// This is the exact entity-data / field-change logic the retired DelegateHandler
// did inline (handleEvent) before F2 moved cascade dispatch off the best-effort
// delegate and onto the durable transactional outbox + polling relay. The relay
// is now the only caller; the functions live here (not relay.go) so the
// rebuild-from-payload concern is testable on its own (enrichment_test.go).

// computeFieldChanges builds a FieldChange map by diffing before and after snapshots.
// Both maps are expected to come from extractEntityData (JSON round-tripped).
func computeFieldChanges(before, after map[string]any) map[string]workflow.FieldChange {
	changes := make(map[string]workflow.FieldChange)
	for key, afterVal := range after {
		beforeVal := before[key]
		if fmt.Sprintf("%v", beforeVal) != fmt.Sprintf("%v", afterVal) {
			changes[key] = workflow.FieldChange{
				OldValue: beforeVal,
				NewValue: afterVal,
			}
		}
	}
	if len(changes) == 0 {
		return nil
	}
	return changes
}

// extractEntityData extracts ID and raw data from an entity result. The app layer
// uses string IDs, so the entity is JSON round-tripped into a generic map and the
// id read from there, falling back to reflection on the struct's ID field.
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
