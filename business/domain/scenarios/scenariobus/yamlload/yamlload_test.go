package yamlload_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestLoad_LeverOverridesParsed(t *testing.T) {
	dir := t.TempDir()
	scenarioDir := filepath.Join(dir, "with-overrides")
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scenarioDir, "scenario.yaml"),
		[]byte("name: with-overrides\nlever_overrides:\n  pick.lotScan: required-if-lot-tracked\n  pick.destinationScan: required\n"),
		0o644); err != nil {
		t.Fatalf("write scenario.yaml: %v", err)
	}

	scenarios, err := yamlload.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(scenarios) != 1 {
		t.Fatalf("got %d scenarios, want 1", len(scenarios))
	}
	got := scenarios[0].LeverOverrides
	if want := map[string]string{
		"pick.lotScan":         "required-if-lot-tracked",
		"pick.destinationScan": "required",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LeverOverrides = %v, want %v", got, want)
	}
}

func TestLoad_WorkersParsed(t *testing.T) {
	dir := t.TempDir()
	scenarioDir := filepath.Join(dir, "with-workers")
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scenarioDir, "scenario.yaml"),
		[]byte("name: with-workers\n"), 0o644); err != nil {
		t.Fatalf("write scenario.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scenarioDir, "bindings.yaml"),
		[]byte("workers:\n  - username: alice@example.com\n    zones: [STG-A, STG-B]\n  - username: bob@example.com\n    zones: [PCK]\n"),
		0o644); err != nil {
		t.Fatalf("write bindings.yaml: %v", err)
	}

	scenarios, err := yamlload.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := scenarios[0].Bindings.Workers
	if len(got) != 2 {
		t.Fatalf("Workers len = %d, want 2", len(got))
	}
	if got[0].Username != "alice@example.com" || len(got[0].Zones) != 2 {
		t.Fatalf("Workers[0] = %+v, want alice with 2 zones", got[0])
	}
	if got[1].Username != "bob@example.com" || len(got[1].Zones) != 1 {
		t.Fatalf("Workers[1] = %+v, want bob with 1 zone", got[1])
	}
}

func TestValidate_RejectsUnknownLeverKey(t *testing.T) {
	s := yamlload.Scenario{
		Name:           "bad-lever",
		LeverOverrides: map[string]string{"pick.notALever": "anything"},
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil; want error for unknown lever key")
	}
	if !strings.Contains(err.Error(), "unknown lever key") {
		t.Fatalf("error %q does not mention 'unknown lever key'", err.Error())
	}
}

func TestValidate_RejectsEmptyWorkerUsername(t *testing.T) {
	s := yamlload.Scenario{
		Name: "bad-worker",
		Bindings: yamlload.Bindings{
			Workers: []yamlload.WorkerBinding{{Username: "", Zones: []string{"STG-A"}}},
		},
	}
	if err := s.Validate(); err == nil {
		t.Fatal("Validate() returned nil; want error for empty worker username")
	}
}

func TestValidate_RejectsEmptyWorkerZones(t *testing.T) {
	s := yamlload.Scenario{
		Name: "bad-zones",
		Bindings: yamlload.Bindings{
			Workers: []yamlload.WorkerBinding{{Username: "alice", Zones: nil}},
		},
	}
	if err := s.Validate(); err == nil {
		t.Fatal("Validate() returned nil; want error for empty zones list")
	}
}

func TestValidate_RejectsDuplicateWorkerUsername(t *testing.T) {
	s := yamlload.Scenario{
		Name: "dup-worker",
		Bindings: yamlload.Bindings{
			Workers: []yamlload.WorkerBinding{
				{Username: "alice@example.com", Zones: []string{"STG-A"}},
				{Username: "alice@example.com", Zones: []string{"PCK"}},
			},
		},
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil; want error for duplicate worker username")
	}
	if !strings.Contains(err.Error(), "duplicate worker username") {
		t.Fatalf("error %q does not mention 'duplicate worker username'", err.Error())
	}
}

func TestValidate_AcceptsKnownLeverKey(t *testing.T) {
	s := yamlload.Scenario{
		Name:           "ok",
		LeverOverrides: map[string]string{"pick.lotScan": "required-if-lot-tracked"},
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestValidate_RejectsNonOverridableLeverKey(t *testing.T) {
	// pick.productScan is in KnownKeys/Defaults but locked per design doc
	// §3.3 invariant 1 — must be rejected as a scenario override key.
	s := yamlload.Scenario{
		Name:           "locked-lever",
		LeverOverrides: map[string]string{"pick.productScan": "disabled"},
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil; want error for non-overridable lever key")
	}
	if !strings.Contains(err.Error(), "not overridable") {
		t.Fatalf("error %q does not mention 'not overridable'", err.Error())
	}
}
