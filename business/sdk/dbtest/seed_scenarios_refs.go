package dbtest

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	inventorylocationbus "github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// refResolver resolves a single stable human-readable code to a UUID.
// Each ref suffix (product_ref, location_ref, tote_ref) has its own
// resolver so the dispatch in resolveRefs can stay a flat switch.
type refResolver func(ctx context.Context, value string) (uuid.UUID, error)

// refLookups bundles the three resolver functions the seeder uses at
// fixture-materialization time. Exposed as an interface (via fields, not
// methods) so the unit test can pass fakes without touching a live DB.
type refLookups struct {
	productIDBySKU   refResolver
	locationIDByCode refResolver
	labelIDByCode    refResolver
}

// newRefLookups wires resolvers against real bus instances. Seeder path.
func newRefLookups(prod *productbus.Business, loc *inventorylocationbus.Business, lbl *labelbus.Business) refLookups {
	return refLookups{
		productIDBySKU: func(ctx context.Context, sku string) (uuid.UUID, error) {
			filter := productbus.QueryFilter{SKU: &sku}
			orderBy := productbus.DefaultOrderBy
			pg := page.MustParse("1", "1")
			rows, err := prod.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("product query sku=%s: %w", sku, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("product not found for sku=%s", sku)
			}
			return rows[0].ProductID, nil
		},
		locationIDByCode: func(ctx context.Context, code string) (uuid.UUID, error) {
			filter := inventorylocationbus.QueryFilter{LocationCodeExact: &code}
			orderBy := order.NewBy("location_code", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := loc.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("location query code=%s: %w", code, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("location not found for code=%s", code)
			}
			return rows[0].LocationID, nil
		},
		labelIDByCode: func(ctx context.Context, code string) (uuid.UUID, error) {
			lc, err := lbl.QueryByCode(ctx, code)
			if err != nil {
				return uuid.Nil, fmt.Errorf("label queryByCode=%s: %w", code, err)
			}
			return lc.ID, nil
		},
	}
}

// refKeySuffix identifies a key that needs ref→id resolution. Kept as a
// constant set so unknown suffixes (e.g. warehouse_ref) fail loudly rather
// than being silently passed through as strings into payload_json.
var knownRefSuffixes = map[string]struct{}{
	"product_ref":  {},
	"location_ref": {},
	"tote_ref":     {},
}

// resolveRefs rewrites a single state.yaml row in place:
//   - replaces "<prefix>_ref" string values with "<prefix>_id" UUID values
//     using the appropriate resolver
//   - injects scenario_id if not already present
//
// Keys that don't end in "_ref" pass through untouched. Unknown "_ref" keys
// (e.g. future "supplier_ref" before a resolver lands) fail-hard so silent
// mis-seeding can't happen.
func resolveRefs(ctx context.Context, row map[string]any, scenarioID uuid.UUID, lookups refLookups) (map[string]any, error) {
	out := make(map[string]any, len(row)+1)
	for k, v := range row {
		if !strings.HasSuffix(k, "_ref") {
			out[k] = v
			continue
		}
		if _, ok := knownRefSuffixes[k]; !ok {
			return nil, fmt.Errorf("unknown ref key %q (expected one of product_ref/location_ref/tote_ref)", k)
		}
		code, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("ref key %q must be a string, got %T", k, v)
		}

		var id uuid.UUID
		var err error
		var targetKey string
		switch k {
		case "product_ref":
			targetKey = "product_id"
			id, err = lookups.productIDBySKU(ctx, code)
		case "location_ref":
			targetKey = "location_id"
			id, err = lookups.locationIDByCode(ctx, code)
		case "tote_ref":
			targetKey = "label_catalog_id"
			id, err = lookups.labelIDByCode(ctx, code)
		}
		if err != nil {
			return nil, fmt.Errorf("resolve %s=%q: %w", k, code, err)
		}
		out[targetKey] = id.String()
	}

	if _, ok := out["scenario_id"]; !ok {
		out["scenario_id"] = scenarioID.String()
	}
	return out, nil
}
