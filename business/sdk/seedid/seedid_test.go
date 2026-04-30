package seedid_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
)

func TestStable_Deterministic(t *testing.T) {
	a := seedid.Stable("label:STG-A01")
	b := seedid.Stable("label:STG-A01")
	if a != b {
		t.Fatalf("Stable(%q) is not deterministic: %s vs %s", "label:STG-A01", a, b)
	}
}

func TestStable_DistinctKeysProduceDistinctUUIDs(t *testing.T) {
	a := seedid.Stable("label:STG-A01")
	b := seedid.Stable("label:STG-A02")
	if a == b {
		t.Fatalf("Stable produced collision for distinct keys: %s", a)
	}
}

func TestStable_KnownVector(t *testing.T) {
	want := uuid.NewSHA1(uuid.MustParse("deadbeef-dead-beef-dead-beefdeadbeef"), []byte("label:STG-A01"))
	got := seedid.Stable("label:STG-A01")
	if got != want {
		t.Fatalf("Stable drift: got %s want %s", got, want)
	}
}
