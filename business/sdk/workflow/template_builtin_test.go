package workflow

import (
	"testing"
	"time"
)

func TestBuiltinVariables(t *testing.T) {
	proc := NewTemplateProcessor(DefaultTemplateProcessingOptions())
	proc.SetBuiltins(BuiltinContext{
		UserID:    "user-123",
		Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	})

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"$me resolves", "{{$me}}", "user-123"},
		{"$now resolves", "{{$now}}", "2025-01-15T10:30:00Z"},
		{"$me in text", "Created by {{$me}}", "Created by user-123"},
		{"$now in text", "At {{$now}}", "At 2025-01-15T10:30:00Z"},
		{"both builtins", "{{$me}} at {{$now}}", "user-123 at 2025-01-15T10:30:00Z"},
		{"unknown builtin uses default", "{{$unknown}}", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := proc.ProcessTemplate(tt.template, nil)
			if result.Processed != tt.expected {
				t.Errorf("got %v, want %v", result.Processed, tt.expected)
			}
		})
	}
}

func TestBuiltinVariablesWithFilters(t *testing.T) {
	proc := NewTemplateProcessor(DefaultTemplateProcessingOptions())
	proc.SetBuiltins(BuiltinContext{
		UserID:    "user-123",
		Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	})

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"$me with uppercase", "{{$me | uppercase}}", "USER-123"},
		{"$now with formatDate short", "{{$now | formatDate:short}}", "Jan 15, 2025"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := proc.ProcessTemplate(tt.template, nil)
			if result.Processed != tt.expected {
				t.Errorf("got %v, want %v", result.Processed, tt.expected)
			}
		})
	}
}

func TestBuiltinVariablesWithoutBuiltinsSet(t *testing.T) {
	proc := NewTemplateProcessor(DefaultTemplateProcessingOptions())
	// Note: SetBuiltins NOT called

	result := proc.ProcessTemplate("{{$me}}", nil)

	// Should use default value (empty string) when builtins not set
	if result.Processed != "" {
		t.Errorf("expected empty string when builtins not set, got %v", result.Processed)
	}
}

func TestBuiltinVariablesStrictMode(t *testing.T) {
	opts := DefaultTemplateProcessingOptions()
	opts.StrictMode = true
	proc := NewTemplateProcessor(opts)
	// Note: SetBuiltins NOT called

	result := proc.ProcessTemplate("{{$me}}", nil)

	// Should have error when builtins not set in strict mode
	if len(result.Errors) == 0 {
		t.Error("expected error in strict mode when builtins not set")
	}
}

func TestBuiltinVariablesMixedWithContext(t *testing.T) {
	proc := NewTemplateProcessor(DefaultTemplateProcessingOptions())
	proc.SetBuiltins(BuiltinContext{
		UserID:    "user-123",
		Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	})

	context := TemplateContext{
		"entity_name": "orders",
		"entity_id":   "order-456",
	}

	result := proc.ProcessTemplate("{{$me}} created {{entity_name}} with id {{entity_id}} at {{$now}}", context)
	expected := "user-123 created orders with id order-456 at 2025-01-15T10:30:00Z"

	if result.Processed != expected {
		t.Errorf("got %v, want %v", result.Processed, expected)
	}
}

func TestValidateVariableWithBuiltins(t *testing.T) {
	proc := NewTemplateProcessor(DefaultTemplateProcessingOptions())

	tests := []struct {
		name      string
		variable  string
		shouldErr bool
	}{
		{"valid $me", "$me", false},
		{"valid $now", "$now", false},
		{"valid $custom", "$custom_var", false},
		{"valid regular", "user.name", false},
		{"invalid $ only", "$", true},
		{"invalid $123", "$123", true},
		{"invalid empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proc.validateVariable(tt.variable)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for %s", tt.variable)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for %s: %v", tt.variable, err)
			}
		})
	}
}
