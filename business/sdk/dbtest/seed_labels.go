package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
)

// detNamespace matches the Manitowoc generator's UUID v5 namespace.
// Using the same namespace guarantees label codes produce byte-identical
// UUIDs across `make reseed-frontend` invocations and across builds.
var detNamespace = uuid.MustParse("deadbeef-dead-beef-dead-beefdeadbeef")

// detUUID returns a UUID v5 derived from a stable key string.
func detUUID(key string) uuid.UUID {
	return uuid.NewSHA1(detNamespace, []byte(key))
}

// seedLabels inserts the 39-label Phase 1 catalog (19 locations + 20 totes)
// with deterministic UUIDs. Matches spec §3.3.
func seedLabels(ctx context.Context, bus *labelbus.Business) error {
	entries := []struct {
		code, typ string
	}{
		{"RCV-01", labelbus.TypeLocation}, {"RCV-02", labelbus.TypeLocation},
		{"QA-01", labelbus.TypeLocation},
		{"STG-A01", labelbus.TypeLocation}, {"STG-A02", labelbus.TypeLocation}, {"STG-A03", labelbus.TypeLocation},
		{"STG-B01", labelbus.TypeLocation}, {"STG-B02", labelbus.TypeLocation}, {"STG-B03", labelbus.TypeLocation},
		{"STG-C01", labelbus.TypeLocation}, {"STG-C02", labelbus.TypeLocation}, {"STG-C03", labelbus.TypeLocation},
		{"PCK-01", labelbus.TypeLocation}, {"PCK-02", labelbus.TypeLocation}, {"PCK-03", labelbus.TypeLocation},
		{"PKG-01", labelbus.TypeLocation}, {"PKG-02", labelbus.TypeLocation},
		{"SHP-01", labelbus.TypeLocation}, {"SHP-02", labelbus.TypeLocation},
	}
	for i := 1; i <= 20; i++ {
		entries = append(entries, struct{ code, typ string }{
			code: fmt.Sprintf("TOTE-%03d", i),
			typ:  labelbus.TypeContainer,
		})
	}
	if len(entries) != 39 {
		return fmt.Errorf("expected 39 seed entries, got %d", len(entries))
	}
	for _, e := range entries {
		lc := labelbus.LabelCatalog{
			ID:          detUUID("label:" + e.code),
			Code:        e.code,
			Type:        e.typ,
			PayloadJSON: "{}",
		}
		if err := bus.SeedCreate(ctx, lc); err != nil {
			return fmt.Errorf("seedcreate %s: %w", e.code, err)
		}
	}
	return nil
}
