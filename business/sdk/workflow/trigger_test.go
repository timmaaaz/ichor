package workflow

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestEvaluateFieldCondition_ContainsOperator tests the "contains" string operator
// in TriggerProcessor.evaluateFieldCondition.
func TestEvaluateFieldCondition_ContainsOperator(t *testing.T) {
	t.Parallel()

	tp := &TriggerProcessor{}

	tests := []struct {
		name      string
		fieldVal  any
		searchVal string
		wantMatch bool
	}{
		{"match substring", "foobar", "foo", true},
		{"match end", "foobar", "bar", true},
		{"no match", "foobar", "baz", false},
		{"empty search", "foobar", "", true},
		{"empty field", "", "foo", false},
		{"exact match", "hello", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := FieldCondition{
				FieldName: "test_field",
				Operator:  "contains",
				Value:     tt.searchVal,
			}
			event := TriggerEvent{
				EventType:  "on_create",
				EntityName: "test_entity",
				EntityID:   uuid.New(),
				Timestamp:  time.Now(),
				RawData:    map[string]any{"test_field": tt.fieldVal},
			}
			result := tp.evaluateFieldCondition(cond, event)
			if result.Matched != tt.wantMatch {
				t.Errorf("evaluateFieldCondition(contains, %q, %q) = %v, want %v",
					tt.fieldVal, tt.searchVal, result.Matched, tt.wantMatch)
			}
		})
	}
}

// TestEvaluateFieldCondition_UnknownOperator verifies that unknown operators
// return an error and Matched=false (not a panic).
func TestEvaluateFieldCondition_UnknownOperator(t *testing.T) {
	t.Parallel()

	tp := &TriggerProcessor{}
	cond := FieldCondition{
		FieldName: "test_field",
		Operator:  "starts_with", // Not yet implemented — should return error
		Value:     "foo",
	}
	event := TriggerEvent{
		EventType:  "on_create",
		EntityName: "test_entity",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"test_field": "foobar"},
	}
	result := tp.evaluateFieldCondition(cond, event)
	if result.Matched {
		t.Error("expected Matched=false for unknown operator")
	}
	if result.Error == "" {
		t.Error("expected Error to be set for unknown operator")
	}
}

// TestEvaluateFieldCondition_NilRawData verifies that nil RawData is handled
// without panic (field simply has no value).
func TestEvaluateFieldCondition_NilRawData(t *testing.T) {
	t.Parallel()

	tp := &TriggerProcessor{}
	cond := FieldCondition{
		FieldName: "test_field",
		Operator:  "equals",
		Value:     "active",
	}
	event := TriggerEvent{
		EventType:  "on_create",
		EntityName: "test_entity",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    nil, // nil — should not panic
	}
	result := tp.evaluateFieldCondition(cond, event)
	// nil field value vs "active" → should not match
	if result.Matched {
		t.Error("expected Matched=false when RawData is nil")
	}
}
