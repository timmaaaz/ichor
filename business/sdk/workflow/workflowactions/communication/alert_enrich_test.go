package communication

import (
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
