package scenariobus

import (
	"context"
	"fmt"
	"sort"
)

// TestNewScenarios generates n NewScenario values with deterministic
// names so tests can assert against stable ordering.
func TestNewScenarios(n int) []NewScenario {
	scenarios := make([]NewScenario, n)
	for i := range n {
		scenarios[i] = NewScenario{
			Name:        fmt.Sprintf("test-scenario-%04d", i+1),
			Description: fmt.Sprintf("Test scenario %d", i+1),
		}
	}
	return scenarios
}

// TestSeedScenarios inserts n scenarios via the bus and returns them sorted
// by ID for stable comparison. Intended for integration tests that need
// scenario rows present without going through the YAML fixture loader.
func TestSeedScenarios(ctx context.Context, n int, api *Business) ([]Scenario, error) {
	newScenarios := TestNewScenarios(n)

	scenarios := make([]Scenario, len(newScenarios))
	for i, ns := range newScenarios {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("seeding scenario %d: %w", i, err)
		}
		scenarios[i] = s
	}

	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].ID.String() < scenarios[j].ID.String()
	})

	return scenarios, nil
}
