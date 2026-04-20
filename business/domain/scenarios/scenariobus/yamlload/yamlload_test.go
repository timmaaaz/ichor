package yamlload_test

import (
	"path/filepath"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus/yamlload"
)

// TestLoad_RushReceiving reads the real in-tree fixture if one exists.
// Task 0d.5 runs BEFORE Task 0d.9 adds the rush-receiving fixture, so the
// directory may not exist yet. A missing directory surfaces as ErrNotFound
// via IsNotFoundErr, which skips — 0d.10 re-runs this test after the
// fixture lands.
//
// Path depth: this test file lives 5 directories deep
// (business/domain/scenarios/scenariobus/yamlload/) so 5x ".." reaches repo root.
func TestLoad_RushReceiving(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..", "deployments", "scenarios")
	scenarios, err := yamlload.Load(root)
	if err != nil {
		if yamlload.IsNotFoundErr(err) {
			t.Skip("no scenarios found at", root)
		}
		t.Fatal(err)
	}
	if len(scenarios) == 0 {
		t.Fatal("expected at least one scenario loaded")
	}
}

func TestValidate_MissingName(t *testing.T) {
	s := yamlload.Scenario{} // empty name
	err := s.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}
}

func TestValidate_DuplicateBindingRef(t *testing.T) {
	s := yamlload.Scenario{
		Name: "test",
		Bindings: yamlload.Bindings{
			Lots: []yamlload.LotBinding{
				{Ref: "dup", ProductCode: "SKU-1"},
				{Ref: "dup", ProductCode: "SKU-2"},
			},
		},
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("expected validation error for duplicate lot ref")
	}
}
