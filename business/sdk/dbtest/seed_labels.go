package dbtest

import (
	"context"
	"encoding/json"
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

// seedLabels inserts the 79-label Phase 0g.B4 catalog (19 locations + 20
// containers + 40 product labels) with deterministic UUIDs. Matches spec §3.3.
func seedLabels(ctx context.Context, bus *labelbus.Business, products ProductsSeed) error {
	type entry struct {
		code        string
		typ         string
		entityRef   string
		payloadJSON string
	}

	entries := []entry{
		{code: "RCV-01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "RCV-02", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "QA-01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-A01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-A02", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-A03", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-B01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-B02", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-B03", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-C01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-C02", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "STG-C03", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "PCK-01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "PCK-02", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "PCK-03", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "PKG-01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "PKG-02", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "SHP-01", typ: labelbus.TypeLocation, payloadJSON: "{}"},
		{code: "SHP-02", typ: labelbus.TypeLocation, payloadJSON: "{}"},
	}

	// 20 deterministic container/tote labels.
	for i := 1; i <= 20; i++ {
		entries = append(entries, entry{
			code:        fmt.Sprintf("TOTE-%03d", i),
			typ:         labelbus.TypeContainer,
			payloadJSON: "{}",
		})
	}

	// 40 product labels — one per seeded product, with the product UUID as
	// entity_ref and a JSON payload carrying SKU/UPC/name for label rendering.
	for _, p := range products.Products {
		payload, err := json.Marshal(struct {
			SKU         string `json:"sku"`
			UPC         string `json:"upc"`
			ProductName string `json:"productName"`
		}{SKU: p.SKU, UPC: p.UpcCode, ProductName: p.Name})
		if err != nil {
			return fmt.Errorf("marshal product label payload for %s: %w", p.SKU, err)
		}
		entries = append(entries, entry{
			code:        fmt.Sprintf("PRD-%s", p.SKU),
			typ:         labelbus.TypeProduct,
			entityRef:   p.ProductID.String(),
			payloadJSON: string(payload),
		})
	}

	const expectedTotal = 79
	if len(entries) != expectedTotal {
		return fmt.Errorf("expected %d seed entries, got %d", expectedTotal, len(entries))
	}

	for _, e := range entries {
		lc := labelbus.LabelCatalog{
			ID:          detUUID("label:" + e.code),
			Code:        e.code,
			Type:        e.typ,
			EntityRef:   e.entityRef,
			PayloadJSON: e.payloadJSON,
		}
		if err := bus.SeedCreate(ctx, lc); err != nil {
			return fmt.Errorf("seedcreate %s: %w", e.code, err)
		}
	}
	return nil
}
