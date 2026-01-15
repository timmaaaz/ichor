package formdataapp

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseIntFromAny(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  int
		expectErr error
	}{
		{"nil returns zero", nil, 0, nil},
		{"int", 42, 42, nil},
		{"int64", int64(42), 42, nil},
		{"float64 truncates", 42.7, 42, nil},
		{"json.Number valid", json.Number("42"), 42, nil},
		{"json.Number invalid", json.Number("abc"), 0, ErrParseInt},
		{"string valid", "42", 42, nil},
		{"string negative", "-5", -5, nil},
		{"string invalid", "abc", 0, ErrParseInt},
		{"string empty", "", 0, ErrParseInt},
		{"string float", "42.5", 0, ErrParseInt}, // Atoi doesn't parse floats
		{"unsupported type slice", []int{42}, 0, ErrInvalidType},
		{"unsupported type map", map[string]int{"a": 1}, 0, ErrInvalidType},
		{"unsupported type bool", true, 0, ErrInvalidType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntFromAny(tt.input)
			if tt.expectErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectErr)
				} else if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestParseDecimalFromAny(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  decimal.Decimal
		expectErr error
	}{
		{"nil returns zero", nil, decimal.Zero, nil},
		{"string valid", "10.50", decimal.NewFromFloat(10.50), nil},
		{"string integer", "42", decimal.NewFromInt(42), nil},
		{"string negative", "-5.25", decimal.NewFromFloat(-5.25), nil},
		{"string invalid", "abc", decimal.Zero, ErrParseDecimal},
		{"string empty", "", decimal.Zero, ErrParseDecimal},
		{"float64", 42.75, decimal.NewFromFloat(42.75), nil},
		{"float64 zero", 0.0, decimal.Zero, nil},
		{"int", 42, decimal.NewFromInt(42), nil},
		{"int64", int64(100), decimal.NewFromInt(100), nil},
		{"json.Number valid", json.Number("99.99"), decimal.NewFromFloat(99.99), nil},
		{"json.Number integer", json.Number("50"), decimal.NewFromInt(50), nil},
		{"json.Number invalid", json.Number("not-a-number"), decimal.Zero, ErrParseDecimal},
		{"unsupported type slice", []float64{1.5}, decimal.Zero, ErrInvalidType},
		{"unsupported type bool", true, decimal.Zero, ErrInvalidType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDecimalFromAny(tt.input)
			if tt.expectErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectErr)
				} else if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !got.Equal(tt.expected) {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestParseStringFromAny(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  string
		expectErr error
	}{
		{"nil returns empty", nil, "", nil},
		{"string value", "hello", "hello", nil},
		{"string empty", "", "", nil},
		{"string with spaces", "  spaces  ", "  spaces  ", nil},
		{"unsupported type int", 42, "", ErrInvalidType},
		{"unsupported type float", 3.14, "", ErrInvalidType},
		{"unsupported type bool", true, "", ErrInvalidType},
		{"unsupported type slice", []string{"a"}, "", ErrInvalidType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStringFromAny(tt.input)
			if tt.expectErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectErr)
				} else if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestSentinelErrorsCanBeChecked verifies that the sentinel errors
// can be properly checked using errors.Is() after wrapping.
func TestSentinelErrorsCanBeChecked(t *testing.T) {
	// Test that wrapped errors still match with errors.Is()

	// parseIntFromAny with invalid string
	_, err := parseIntFromAny("not-a-number")
	if !errors.Is(err, ErrParseInt) {
		t.Errorf("expected errors.Is(err, ErrParseInt) to be true, got false")
	}

	// parseDecimalFromAny with invalid string
	_, err = parseDecimalFromAny("not-a-decimal")
	if !errors.Is(err, ErrParseDecimal) {
		t.Errorf("expected errors.Is(err, ErrParseDecimal) to be true, got false")
	}

	// parseStringFromAny with invalid type
	_, err = parseStringFromAny(123)
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected errors.Is(err, ErrInvalidType) to be true, got false")
	}

	// parseIntFromAny with invalid type
	_, err = parseIntFromAny([]int{1, 2, 3})
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected errors.Is(err, ErrInvalidType) to be true for parseIntFromAny, got false")
	}
}
