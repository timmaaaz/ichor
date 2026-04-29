// Package levers owns the canonical list of scan-discipline lever keys
// and their SMB defaults. Seeders, validators, and tests import this
// package as the single source of truth so the three never drift.
//
// See docs/superpowers/specs/2026-04-24-label-scan-workflow-redesign.md §3.3
// for the per-key semantics and the SMB-vs-strict-regulated contrast.
package levers

// KnownKeys is the complete, sorted list of lever keys. Sorted to
// stabilize seed insert order and test diff output.
var KnownKeys = []string{
	"pick.assignmentGranularity",
	"pick.destinationMode",
	"pick.destinationScan",
	"pick.lotScan",
	"pick.productScan",
	"pick.sourceLocationScan",
	"receive.expiryCapture",
	"receive.lotCapture",
	"receive.poScan",
	"transfer.destinationScan",
	"transfer.sourceLocationScan",
}

// Defaults is the SMB-default value for every lever key. Per design doc
// §3.3 invariant 1, pick.productScan is always "required" and intentionally
// not exposed as a configurable lever — included here for resolver
// completeness, NOT for customer override.
var Defaults = map[string]string{
	"pick.assignmentGranularity":  "whole-order",
	"pick.destinationMode":        "direct-stage",
	"pick.destinationScan":        "button-confirm",
	"pick.lotScan":                "disabled",
	"pick.productScan":            "required",
	"pick.sourceLocationScan":     "button-confirm",
	"receive.expiryCapture":       "required-if-lot-tracked",
	"receive.lotCapture":          "required-if-lot-tracked",
	"receive.poScan":              "required",
	"transfer.destinationScan":    "button-confirm",
	"transfer.sourceLocationScan": "button-confirm",
}

// IsKnown reports whether key is one of the canonical lever keys. Used by
// scenario YAML validation to reject typos in lever_overrides.
func IsKnown(key string) bool {
	_, ok := Defaults[key]
	return ok
}

// nonOverridableKeys holds keys that exist in Defaults for resolver
// completeness but must never be changed by a scenario or customer override.
// Per design doc §3.3 invariant 1, pick.productScan is locked to "required".
var nonOverridableKeys = map[string]bool{
	"pick.productScan": true,
}

// IsOverridable reports whether key may appear in scenario lever_overrides.
// All known keys are overridable except those locked by design invariants.
func IsOverridable(key string) bool {
	return IsKnown(key) && !nonOverridableKeys[key]
}
