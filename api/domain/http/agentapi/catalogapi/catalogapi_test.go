package catalogapi

import (
	"testing"
)

func TestCatalogCompleteness(t *testing.T) {
	if len(catalog) == 0 {
		t.Fatal("catalog is empty")
	}

	validCategories := map[string]bool{
		"ui":       true,
		"workflow": true,
		"system":   true,
	}

	seen := make(map[string]bool)
	for _, surface := range catalog {
		if surface.Name == "" {
			t.Error("found surface with empty name")
		}
		if surface.Description == "" {
			t.Errorf("surface %q has empty description", surface.Name)
		}
		if !validCategories[surface.Category] {
			t.Errorf("surface %q has invalid category %q", surface.Name, surface.Category)
		}
		if surface.Endpoints.List == "" {
			t.Errorf("surface %q has no list endpoint", surface.Name)
		}
		if seen[surface.Name] {
			t.Errorf("duplicate surface name: %s", surface.Name)
		}
		seen[surface.Name] = true
	}

	// Verify minimum expected surfaces.
	expectedSurfaces := []string{
		"Page Configs",
		"Page Content",
		"Page Actions",
		"Table Configs",
		"Forms",
		"Form Fields",
		"Workflow Rules",
		"Action Templates",
		"Alerts",
		"Action Permissions",
		"Enum Labels",
		"Database Introspection",
	}
	for _, name := range expectedSurfaces {
		if !seen[name] {
			t.Errorf("missing expected surface: %s", name)
		}
	}

	if len(catalog) < 12 {
		t.Errorf("expected at least 12 config surfaces, got %d", len(catalog))
	}
}
