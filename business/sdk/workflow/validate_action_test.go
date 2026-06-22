package workflow

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// Test_validateActionExecutable covers the guard that rejects a rule action
// which has neither a template nor an inline "action_type" — the state that
// saves cleanly but fails every Temporal dispatch ("action_type is required").
func Test_validateActionExecutable(t *testing.T) {
	tmpl := uuid.New()

	tests := []struct {
		name       string
		templateID *uuid.UUID
		config     json.RawMessage
		wantErr    bool
	}{
		{"template present, no inline type → ok", &tmpl, json.RawMessage(`{}`), false},
		{"no template, inline action_type → ok", nil, json.RawMessage(`{"action_type":"send_email"}`), false},
		{"no template, legacy inline type → ok", nil, json.RawMessage(`{"type":"send_notification"}`), false},
		{"no template, empty action_type but legacy type → ok", nil, json.RawMessage(`{"action_type":"","type":"send_email"}`), false},
		{"template present AND inline type → ok", &tmpl, json.RawMessage(`{"action_type":"send_email"}`), false},
		{"no template, no inline type → rejected", nil, json.RawMessage(`{"to":"x"}`), true},
		{"no template, empty inline type → rejected", nil, json.RawMessage(`{"action_type":""}`), true},
		{"no template, invalid json → rejected", nil, json.RawMessage(`not json`), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateActionExecutable(tc.templateID, tc.config)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errs.IsFieldErrors(err) {
					t.Errorf("expected errs.FieldErrors, got %T (%v)", err, err)
				}
				return
			}
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}
