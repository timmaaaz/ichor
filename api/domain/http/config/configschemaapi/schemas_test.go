package configschemaapi

import (
	"encoding/json"
	"testing"
)

func TestAllConfigSchemasValid(t *testing.T) {
	// Verify table_config schema loaded and is valid JSON.
	tc, ok := configSchemas["table_config"]
	if !ok {
		t.Fatal("table_config schema not loaded")
	}
	var dummy any
	if err := json.Unmarshal(tc, &dummy); err != nil {
		t.Fatalf("table_config schema is not valid JSON: %v", err)
	}

	// Verify layout schema loaded and is valid JSON.
	lc, ok := configSchemas["layout"]
	if !ok {
		t.Fatal("layout schema not loaded")
	}
	if err := json.Unmarshal(lc, &dummy); err != nil {
		t.Fatalf("layout schema is not valid JSON: %v", err)
	}

	// Verify content types loaded.
	if len(contentTypes) != 6 {
		t.Fatalf("expected 6 content types, got %d", len(contentTypes))
	}

	expectedTypes := map[string]bool{
		"table": false, "form": false, "chart": false,
		"tabs": false, "container": false, "text": false,
	}
	for _, ct := range contentTypes {
		if _, ok := expectedTypes[ct.Type]; !ok {
			t.Errorf("unexpected content type: %s", ct.Type)
		}
		expectedTypes[ct.Type] = true
		if ct.Name == "" {
			t.Errorf("content type %s has empty name", ct.Type)
		}
		if ct.Description == "" {
			t.Errorf("content type %s has empty description", ct.Type)
		}
	}
	for typ, found := range expectedTypes {
		if !found {
			t.Errorf("missing expected content type: %s", typ)
		}
	}
}
