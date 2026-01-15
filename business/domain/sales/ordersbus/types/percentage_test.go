package types_test

import (
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

func TestParsePercentage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid cases
		{"zero", "0", false},
		{"whole number", "8", false},
		{"decimal one place", "8.5", false},
		{"decimal two places", "8.25", false},
		{"max value", "100", false},
		{"empty string", "", false},
		{"very small", "0.01", false},
		{"very small one decimal", "0.1", false},
		{"boundary max minus one cent", "99.99", false},

		// Invalid cases - out of range
		{"negative", "-1", true},
		{"over 100", "101", true},
		{"way over", "150", true},
		{"just over max", "100.01", true},

		// Invalid cases - format errors
		{"three decimals", "8.125", true},
		{"not a number", "abc", true},
		{"special chars", "8%", true},
		{"leading zeros", "08", true},
		{"leading zeros decimal", "08.5", true},
		{"trailing decimal", "8.", true},
		{"scientific notation", "1e2", true},
		{"whitespace leading", " 8", true},
		{"whitespace trailing", "8 ", true},
		{"whitespace both", " 8 ", true},
		{"unicode digits", "８", true},
		{"multiple dots", "8.2.5", true},
		{"comma decimal", "8,25", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := types.ParsePercentage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePercentage(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestPercentage_Value(t *testing.T) {
	p, _ := types.ParsePercentage("8.25")
	if got := p.Value(); got != "8.25" {
		t.Errorf("Value() = %q, want %q", got, "8.25")
	}
}

func TestPercentage_DBValue(t *testing.T) {
	p, _ := types.ParsePercentage("8.25")
	dbVal := p.DBValue()
	if !dbVal.Valid || dbVal.String != "8.25" {
		t.Errorf("DBValue() = {%v, %q}, want {true, \"8.25\"}", dbVal.Valid, dbVal.String)
	}

	// Empty percentage
	empty, _ := types.ParsePercentage("")
	emptyDB := empty.DBValue()
	if emptyDB.Valid {
		t.Errorf("empty.DBValue().Valid = true, want false")
	}
}

func TestPercentage_JSON(t *testing.T) {
	type wrapper struct {
		Rate types.Percentage `json:"rate"`
	}

	// Marshal
	w := wrapper{Rate: types.MustParsePercentage("8.25")}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	expected := `{"rate":"8.25"}`
	if string(data) != expected {
		t.Errorf("Marshal = %s, want %s", data, expected)
	}

	// Unmarshal
	var w2 wrapper
	if err := json.Unmarshal([]byte(`{"rate":"10.5"}`), &w2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if w2.Rate.Value() != "10.5" {
		t.Errorf("Unmarshal rate = %q, want %q", w2.Rate.Value(), "10.5")
	}
}

func TestPercentage_Equal(t *testing.T) {
	p1 := types.MustParsePercentage("8.25")
	p2 := types.MustParsePercentage("8.25")
	p3 := types.MustParsePercentage("8.5")

	if !p1.Equal(p2) {
		t.Error("Equal(same value) = false, want true")
	}
	if p1.Equal(p3) {
		t.Error("Equal(different value) = true, want false")
	}
}

func TestMustParsePercentage_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParsePercentage(invalid) did not panic")
		}
	}()
	types.MustParsePercentage("invalid")
}

func TestParsePercentagePtr(t *testing.T) {
	// Valid value returns pointer
	p, err := types.ParsePercentagePtr("8.25")
	if err != nil {
		t.Fatalf("ParsePercentagePtr(valid) error = %v", err)
	}
	if p == nil {
		t.Fatal("ParsePercentagePtr(valid) = nil, want non-nil")
	}
	if p.Value() != "8.25" {
		t.Errorf("ParsePercentagePtr(valid).Value() = %q, want %q", p.Value(), "8.25")
	}

	// Empty string returns nil pointer
	empty, err := types.ParsePercentagePtr("")
	if err != nil {
		t.Fatalf("ParsePercentagePtr(empty) error = %v", err)
	}
	if empty != nil {
		t.Error("ParsePercentagePtr(empty) = non-nil, want nil")
	}

	// Invalid value returns error
	_, err = types.ParsePercentagePtr("invalid")
	if err == nil {
		t.Error("ParsePercentagePtr(invalid) error = nil, want error")
	}
}
