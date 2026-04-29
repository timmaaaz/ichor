package levers_test

import (
	"sort"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/levers"
)

func Test_Defaults_HasExactlyElevenKeys(t *testing.T) {
	if got, want := len(levers.Defaults), 11; got != want {
		t.Fatalf("len(Defaults) = %d, want %d", got, want)
	}
}

func Test_Defaults_AllKnownKeysHaveDefaults(t *testing.T) {
	for _, k := range levers.KnownKeys {
		if _, ok := levers.Defaults[k]; !ok {
			t.Errorf("KnownKeys contains %q but Defaults has no entry", k)
		}
	}
}

func Test_Defaults_NoExtraKeys(t *testing.T) {
	known := make(map[string]bool, len(levers.KnownKeys))
	for _, k := range levers.KnownKeys {
		known[k] = true
	}
	for k := range levers.Defaults {
		if !known[k] {
			t.Errorf("Defaults contains %q but it is not in KnownKeys", k)
		}
	}
}

func Test_Defaults_ProductScan_AlwaysRequired(t *testing.T) {
	// Per design doc §3.3 invariant: pick.productScan is always required.
	if got := levers.Defaults["pick.productScan"]; got != "required" {
		t.Fatalf("pick.productScan default = %q, want %q (invariant)", got, "required")
	}
}

func Test_KnownKeys_Sorted(t *testing.T) {
	// Stable iteration for seeders + test output diffability.
	if !sort.StringsAreSorted(levers.KnownKeys) {
		t.Fatalf("KnownKeys is not sorted: %v", levers.KnownKeys)
	}
}

func Test_IsOverridable_ProductScanReturnsFalse(t *testing.T) {
	// Per design doc §3.3 invariant 1, pick.productScan is locked.
	if levers.IsOverridable("pick.productScan") {
		t.Fatal("IsOverridable(pick.productScan) = true, want false")
	}
}

func Test_IsOverridable_OtherKnownKeysReturnTrue(t *testing.T) {
	for _, k := range levers.KnownKeys {
		if k == "pick.productScan" {
			continue
		}
		if !levers.IsOverridable(k) {
			t.Errorf("IsOverridable(%q) = false, want true", k)
		}
	}
}

func Test_IsOverridable_UnknownKeyReturnsFalse(t *testing.T) {
	if levers.IsOverridable("pick.notALever") {
		t.Fatal("IsOverridable(pick.notALever) = true, want false")
	}
}
