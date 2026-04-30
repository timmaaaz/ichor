// Package yamlload parses and validates scenario fixture YAML files.
//
// Directory layout:
//
//	deployments/scenarios/<name>/
//	  scenario.yaml   metadata: name, description, optional id
//	  bindings.yaml   references to baseline entities by human-readable code
//	  state.yaml      row blueprints to INSERT at Load time
//
// scenario.yaml is required; bindings.yaml and state.yaml are optional.
//
// Validation is Go-struct-driven (no JSON schema lib). Corrupt YAML aborts
// seedScenarios() so a broken fixture never produces a half-seeded dev DB.
package yamlload

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/levers"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
	"gopkg.in/yaml.v3"
)

// ErrNotFound is returned when scenariosRoot does not exist or contains no
// loadable scenario directories. Callers use IsNotFoundErr to skip cleanly.
var ErrNotFound = errors.New("no scenarios found")

func IsNotFoundErr(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// Scenario is the parsed + validated representation of one on-disk scenario
// directory. Returned by Load.
type Scenario struct {
	ID             uuid.UUID         `yaml:"id,omitempty"`
	Name           string            `yaml:"name"`
	Description    string            `yaml:"description"`
	LeverOverrides map[string]string `yaml:"lever_overrides,omitempty"`
	Bindings       Bindings          `yaml:"-"` // loaded from bindings.yaml separately
	State          State             `yaml:"-"` // loaded from state.yaml separately
}

// Bindings references baseline entities by stable identifier. Resolved
// against the DB at seed time.
type Bindings struct {
	Totes     []ToteBinding     `yaml:"totes"`
	Locations []LocationBinding `yaml:"locations"`
	Lots      []LotBinding      `yaml:"lots"`
	Serials   []SerialBinding   `yaml:"serials"`
	Workers   []WorkerBinding   `yaml:"workers"`
}

type ToteBinding struct {
	Code string `yaml:"code"`
}

type LocationBinding struct {
	Code string `yaml:"code"`
	Role string `yaml:"role,omitempty"`
}

type LotBinding struct {
	Ref         string `yaml:"ref"`
	ProductCode string `yaml:"product_code"`
}

type SerialBinding struct {
	Ref         string `yaml:"ref"`
	ProductCode string `yaml:"product_code"`
}

type WorkerBinding struct {
	Username string   `yaml:"username"`
	Zones    []string `yaml:"zones"`
}

// State is an open-shape map keyed by target-table suffix (e.g. "purchase_orders"
// for procurement.purchase_orders). Values are slices of raw YAML rows the
// seeder marshals to JSONB payloads. The seeder owns the suffix→schema.table
// mapping (see seed_scenarios.go:resolveTargetTable).
type State map[string][]map[string]any

// Load reads every subdirectory of scenariosRoot as a scenario, returning
// validated Scenario values. Subdirectories starting with '_' or '.' are
// skipped; plain files (e.g. _schema.yaml documentation) are also skipped.
func Load(scenariosRoot string) ([]Scenario, error) {
	entries, err := os.ReadDir(scenariosRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("readdir: %w", err)
	}

	var out []Scenario
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "" || name[0] == '_' || name[0] == '.' {
			continue
		}
		s, err := loadOne(filepath.Join(scenariosRoot, name))
		if err != nil {
			return nil, fmt.Errorf("scenario %s: %w", name, err)
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil, ErrNotFound
	}
	return out, nil
}

func loadOne(dir string) (Scenario, error) {
	s := Scenario{}

	if err := readYAML(filepath.Join(dir, "scenario.yaml"), &s); err != nil {
		return Scenario{}, fmt.Errorf("scenario.yaml: %w", err)
	}

	bindingsPath := filepath.Join(dir, "bindings.yaml")
	if _, err := os.Stat(bindingsPath); err == nil {
		if err := readYAML(bindingsPath, &s.Bindings); err != nil {
			return Scenario{}, fmt.Errorf("bindings.yaml: %w", err)
		}
	}

	statePath := filepath.Join(dir, "state.yaml")
	if _, err := os.Stat(statePath); err == nil {
		if err := readYAML(statePath, &s.State); err != nil {
			return Scenario{}, fmt.Errorf("state.yaml: %w", err)
		}
	}

	if s.ID == (uuid.UUID{}) {
		s.ID = seedid.Stable("scenario:" + s.Name)
	}

	if err := s.Validate(); err != nil {
		return Scenario{}, err
	}
	return s, nil
}

func readYAML(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, out)
}

// Validate runs fail-hard validation. Called automatically by loadOne;
// exported so tests and seeders can validate synthetic Scenario values.
func (s Scenario) Validate() error {
	if s.Name == "" {
		return errors.New("scenario: name is required")
	}

	// Lot/serial refs are scenario-scoped FKs the seeder resolves when
	// building state.yaml rows. Duplicate refs are ambiguous.
	seenLot := map[string]bool{}
	for _, l := range s.Bindings.Lots {
		if l.Ref == "" {
			return fmt.Errorf("scenario %s: lot with empty ref", s.Name)
		}
		if seenLot[l.Ref] {
			return fmt.Errorf("scenario %s: duplicate lot ref %q", s.Name, l.Ref)
		}
		seenLot[l.Ref] = true
	}

	seenSer := map[string]bool{}
	for _, sr := range s.Bindings.Serials {
		if sr.Ref == "" {
			return fmt.Errorf("scenario %s: serial with empty ref", s.Name)
		}
		if seenSer[sr.Ref] {
			return fmt.Errorf("scenario %s: duplicate serial ref %q", s.Name, sr.Ref)
		}
		seenSer[sr.Ref] = true
	}

	for k := range s.LeverOverrides {
		if !levers.IsKnown(k) {
			return fmt.Errorf("scenario %s: unknown lever key %q (see business/domain/config/settingsbus/levers)", s.Name, k)
		}
		if !levers.IsOverridable(k) {
			return fmt.Errorf("scenario %s: lever key %q is not overridable (see business/domain/config/settingsbus/levers)", s.Name, k)
		}
	}

	seenWorker := map[string]bool{}
	for _, w := range s.Bindings.Workers {
		if w.Username == "" {
			return fmt.Errorf("scenario %s: worker with empty username", s.Name)
		}
		if len(w.Zones) == 0 {
			return fmt.Errorf("scenario %s: worker %q has empty zones list", s.Name, w.Username)
		}
		if seenWorker[w.Username] {
			return fmt.Errorf("scenario %s: duplicate worker username %q", s.Name, w.Username)
		}
		seenWorker[w.Username] = true
	}

	return nil
}

// PayloadJSON serializes a state.yaml row into the JSONB payload format
// expected by scenario_fixtures.payload_json. YAML parses into
// map[string]any which json.Marshal handles natively.
func PayloadJSON(row map[string]any) ([]byte, error) {
	return json.Marshal(row)
}
