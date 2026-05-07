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

// TestLoad_AllScenarios is a golden-test loop that walks every subdirectory
// under deployments/scenarios/ and asserts each one parses + validates without
// error. Complements Test_Seed_Integration (full seed chain, ~20s, needs DB) by
// providing <1s CI coverage for YAML shape regressions.
//
// The count assertion uses >= so future scenario additions don't break the test.
// Path depth: test file lives 5 dirs deep (business/domain/scenarios/scenariobus/yamlload/)
// so 5x ".." reaches repo root.
func TestLoad_AllScenarios(t *testing.T) {
	scenariosRoot := filepath.Join("..", "..", "..", "..", "..", "deployments", "scenarios")

	// Per-scenario sub-tests so failures name the offending scenario.
	entries, err := os.ReadDir(scenariosRoot)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("scenarios root not found at %s", scenariosRoot)
		}
		t.Fatalf("readdir %s: %v", scenariosRoot, err)
	}

	var loadable []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "" || name[0] == '_' || name[0] == '.' {
			continue
		}
		loadable = append(loadable, name)
	}

	const wantMin = 19
	if len(loadable) < wantMin {
		t.Fatalf("found %d scenario directories under %s, want >= %d", len(loadable), scenariosRoot, wantMin)
	}

	// Load the whole directory at once (exercises Load's error-abort path).
	all, err := yamlload.Load(scenariosRoot)
	if err != nil {
		t.Fatalf("yamlload.Load(%s): %v", scenariosRoot, err)
	}
	if len(all) < wantMin {
		t.Fatalf("Load returned %d scenarios, want >= %d", len(all), wantMin)
	}

	// Sub-tests per scenario name for diagnosable failure output.
	loaded := make(map[string]yamlload.Scenario, len(all))
	for _, s := range all {
		loaded[s.Name] = s
	}
	for _, name := range loadable {
		name := name // capture
		t.Run(name, func(t *testing.T) {
			s, ok := loaded[name]
			if !ok {
				t.Fatalf("scenario directory %q not present in Load output", name)
			}
			if err := s.Validate(); err != nil {
				t.Fatalf("Validate(): %v", err)
			}
		})
	}
}

// TestLoad_AllScenarios_BrokenFixture proves the golden test catches regressions:
// a scenario directory with an empty name must cause Load to return an error.
func TestLoad_AllScenarios_BrokenFixture(t *testing.T) {
	dir := t.TempDir()
	brokenDir := filepath.Join(dir, "broken-scenario")
	if err := os.MkdirAll(brokenDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Validates that a scenario with empty name fails Load + Validate.
	if err := os.WriteFile(filepath.Join(brokenDir, "scenario.yaml"),
		[]byte("name: \"\"\n"), // empty name fails Validate
		0o644); err != nil {
		t.Fatalf("write scenario.yaml: %v", err)
	}

	_, err := yamlload.Load(dir)
	if err == nil {
		t.Fatal("Load() returned nil error for scenario with empty name; want error")
	}
}
