package communication

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestBuildAlertTemplateData_AddsIDsAndCopiesRawData(t *testing.T) {
	execID := uuid.New()
	ruleID := uuid.New()
	raw := map[string]any{"product_id": "abc", "quantity": float64(5)}

	out := buildAlertTemplateData(raw, execID, &ruleID)

	if out["execution_id"] != execID.String() {
		t.Fatalf("execution_id = %v, want %s", out["execution_id"], execID)
	}
	if out["rule_id"] != ruleID.String() {
		t.Fatalf("rule_id = %v, want %s", out["rule_id"], ruleID)
	}
	if out["product_id"] != "abc" || out["quantity"] != float64(5) {
		t.Fatalf("raw data not copied: %+v", out)
	}
	// Must NOT mutate the caller's map.
	if _, ok := raw["execution_id"]; ok {
		t.Fatal("buildAlertTemplateData mutated the input RawData map")
	}
}

func TestBuildAlertTemplateData_NilRuleAndRawData(t *testing.T) {
	execID := uuid.New()
	out := buildAlertTemplateData(nil, execID, nil)
	if out["execution_id"] != execID.String() {
		t.Fatalf("execution_id missing")
	}
	if _, ok := out["rule_id"]; ok {
		t.Fatal("rule_id should be absent when ruleID is nil")
	}
}

func TestEnrichAlertContext_MergesIntoEmptyObject(t *testing.T) {
	execID := uuid.New()
	ruleID := uuid.New()

	out, err := enrichAlertContext(json.RawMessage(`{}`), execID, &ruleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if m["execution_id"] != execID.String() || m["rule_id"] != ruleID.String() {
		t.Fatalf("ids not merged: %+v", m)
	}
}

func TestEnrichAlertContext_PreservesExistingKeys(t *testing.T) {
	execID := uuid.New()
	out, err := enrichAlertContext(json.RawMessage(`{"foo":"bar"}`), execID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]any
	_ = json.Unmarshal(out, &m)
	if m["foo"] != "bar" {
		t.Fatalf("existing key dropped: %+v", m)
	}
	if m["execution_id"] != execID.String() {
		t.Fatalf("execution_id not added: %+v", m)
	}
}

func TestUUIDFromData(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name   string
		data   map[string]any
		key    string
		want   uuid.UUID
		wantOK bool
	}{
		{"valid", map[string]any{"order_id": id.String()}, "order_id", id, true},
		{"missing key", map[string]any{}, "order_id", uuid.Nil, false},
		{"non-string", map[string]any{"order_id": 123}, "order_id", uuid.Nil, false},
		{"invalid uuid", map[string]any{"order_id": "not-a-uuid"}, "order_id", uuid.Nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := uuidFromData(tt.data, tt.key)
			if ok != tt.wantOK || got != tt.want {
				t.Fatalf("uuidFromData = (%v, %v), want (%v, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// TestResolveEntityLabels_NilBuses_NoOp verifies that with no ordersBus/productBus
// wired, FK-label resolution is skipped entirely: the data map is untouched (the
// literal {{order_number}}/{{product_name}} placeholders survive) and it never
// panics. The actual DB resolution is covered by the over-order integration tests.
func TestResolveEntityLabels_NilBuses_NoOp(t *testing.T) {
	h := NewCreateAlertHandler(nil, nil, nil, nil, nil)
	data := map[string]any{
		"order_id":   uuid.New().String(),
		"product_id": uuid.New().String(),
	}

	h.resolveEntityLabels(context.Background(), data,
		"Over-order on {{order_number}} of {{product_name}}")

	if _, ok := data["order_number"]; ok {
		t.Error("order_number resolved despite nil ordersBus")
	}
	if _, ok := data["product_name"]; ok {
		t.Error("product_name resolved despite nil productBus")
	}
}
