// Package seedid provides deterministic UUID v5 derivation for test
// seed data. Using a fixed namespace and a stable key string guarantees
// that re-running `make reseed-frontend` (or any other seeding flow)
// produces byte-identical UUIDs across builds and developer machines.
//
// This is the single source of truth for the deterministic-seed
// namespace; both business/sdk/dbtest seed helpers and per-domain
// testutil helpers must import this package rather than defining their
// own copy.
package seedid

import "github.com/google/uuid"

// Namespace is the UUID v5 namespace that anchors all deterministic
// seed UUIDs. It matches the Manitowoc generator's value; changing it
// invalidates every deterministic seed UUID in the codebase, including
// inventory.label_catalog rows.
var Namespace = uuid.MustParse("deadbeef-dead-beef-dead-beefdeadbeef")

// Stable returns a UUID v5 derived from key under Namespace. The same
// key always produces the same UUID, on any machine, in any process.
func Stable(key string) uuid.UUID {
	return uuid.NewSHA1(Namespace, []byte(key))
}
