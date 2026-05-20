package scenarios_test

import (
	"testing"
)

func TestDeriveFamily(t *testing.T) {
	cases := []struct {
		name string
		want family
	}{
		{"transfer-intra-zone", familyTransfer},
		{"transfer-cross-zone", familyTransfer},
		{"pick-whole-order", familyPick},
		{"pick-short-pick", familyPick},
		{"receive-lot-tracking", familyReceive},
		{"cycle-count-variance-over", familyCycleCount},
		{"profile-medical-device-rental", familyProfile},
		{"profile-strict-regulated", familyProfile},
		{"rush-receiving", familyReceive},   // override
		{"e2e-pick-strict", familyPick},     // override
		{"e2e-baseline", ""},                // unset — falls through to Custom
		{"unknown-prefix", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveFamily(tc.name); got != tc.want {
				t.Errorf("deriveFamily(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestDiscoverScenarios_Smoke(t *testing.T) {
	rows, err := discoverScenarios(scenarioRoots())
	if err != nil {
		t.Fatalf("discoverScenarios: %v", err)
	}
	// We expect exactly 21 scenarios in deployments/scenarios/ as of 2026-05-20.
	// If this count drifts, either a scenario was added/removed or the
	// discovery glob broke — investigate, do not blindly update the number.
	const wantCount = 21
	if len(rows) != wantCount {
		names := make([]string, 0, len(rows))
		for _, r := range rows {
			names = append(names, r.Name)
		}
		t.Errorf("discoverScenarios returned %d rows, want %d. Got: %v", len(rows), wantCount, names)
	}
}
