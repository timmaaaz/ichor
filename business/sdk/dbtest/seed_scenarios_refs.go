package dbtest

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	currencybus "github.com/timmaaaz/ichor/business/domain/core/currencybus"
	userbus "github.com/timmaaaz/ichor/business/domain/core/userbus"
	inventorylocationbus "github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	warehousebus "github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	purchaseorderstatusbus "github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	supplierbus "github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// refResolver resolves a single stable human-readable code to a UUID.
// Each ref suffix has its own resolver so the dispatch in resolveRefs
// can stay a flat switch.
type refResolver func(ctx context.Context, value string) (uuid.UUID, error)

// refLookups bundles the resolver functions the seeder uses at
// fixture-materialization time. Exposed as an interface (via fields, not
// methods) so the unit test can pass fakes without touching a live DB.
type refLookups struct {
	productIDBySKU              refResolver
	locationIDByCode            refResolver
	labelIDByCode               refResolver
	supplierIDByCode            refResolver
	warehouseIDByCode           refResolver
	currencyIDByCode            refResolver
	userIDByUsername            refResolver
	purchaseOrderStatusIDByName refResolver
}

// newRefLookups wires resolvers against real bus instances. Seeder path.
func newRefLookups(
	prod *productbus.Business,
	loc *inventorylocationbus.Business,
	lbl *labelbus.Business,
	sup *supplierbus.Business,
	wh *warehousebus.Business,
	cur *currencybus.Business,
	usr *userbus.Business,
	pos *purchaseorderstatusbus.Business,
) refLookups {
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
		supplierIDByCode: func(ctx context.Context, code string) (uuid.UUID, error) {
			filter := supplierbus.QueryFilter{Code: &code}
			orderBy := order.NewBy("code", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := sup.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("supplier query code=%s: %w", code, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("supplier not found for code=%s", code)
			}
			return rows[0].SupplierID, nil
		},
		warehouseIDByCode: func(ctx context.Context, code string) (uuid.UUID, error) {
			filter := warehousebus.QueryFilter{Code: &code}
			orderBy := order.NewBy("code", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := wh.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("warehouse query code=%s: %w", code, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("warehouse not found for code=%s", code)
			}
			if rows[0].Code != code {
				return uuid.Nil, fmt.Errorf("warehouse not found for code=%s (closest match: %s)", code, rows[0].Code)
			}
			return rows[0].ID, nil
		},
		currencyIDByCode: func(ctx context.Context, code string) (uuid.UUID, error) {
			filter := currencybus.QueryFilter{Code: &code}
			orderBy := order.NewBy("code", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := cur.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("currency query code=%s: %w", code, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("currency not found for code=%s", code)
			}
			return rows[0].ID, nil
		},
		userIDByUsername: func(ctx context.Context, username string) (uuid.UUID, error) {
			name, err := userbus.ParseName(username)
			if err != nil {
				return uuid.Nil, fmt.Errorf("user parse username=%s: %w", username, err)
			}
			filter := userbus.QueryFilter{Username: &name}
			orderBy := order.NewBy("username", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := usr.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("user query username=%s: %w", username, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("user not found for username=%s", username)
			}
			if rows[0].Username.String() != username {
				return uuid.Nil, fmt.Errorf("user not found for username=%s (closest match: %s)", username, rows[0].Username.String())
			}
			return rows[0].ID, nil
		},
		purchaseOrderStatusIDByName: func(ctx context.Context, name string) (uuid.UUID, error) {
			filter := purchaseorderstatusbus.QueryFilter{Name: &name}
			orderBy := order.NewBy("name", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := pos.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("purchase_order_status query name=%s: %w", name, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("purchase_order_status not found for name=%s", name)
			}
			if rows[0].Name != name {
				return uuid.Nil, fmt.Errorf("purchase_order_status not found for name=%s (closest match: %s)", name, rows[0].Name)
			}
			return rows[0].ID, nil
		},
	}
}

// refKeySuffix identifies a key that needs ref→id resolution. Kept as a
// constant set so unknown suffixes (e.g. vendor_ref) fail loudly rather
// than being silently passed through as strings into payload_json.
var knownRefSuffixes = map[string]struct{}{
	"product_ref":               {},
	"location_ref":              {},
	"tote_ref":                  {},
	"supplier_ref":              {},
	"warehouse_ref":             {},
	"currency_ref":              {},
	"user_ref":                  {},
	"purchase_order_status_ref": {},
}

// resolveRefs rewrites a single state.yaml row in place:
//   - replaces "<prefix>_ref" string values with "<prefix>_id" UUID values
//     using the appropriate resolver
//   - injects scenario_id if not already present
//
// Keys that don't end in "_ref" pass through untouched. Unknown "_ref" keys
// (anything not in knownRefSuffixes above) fail-hard so silent mis-seeding
// can't happen.
func resolveRefs(ctx context.Context, row map[string]any, scenarioID uuid.UUID, lookups refLookups) (map[string]any, error) {
	out := make(map[string]any, len(row)+1)
	for k, v := range row {
		if !strings.HasSuffix(k, "_ref") {
			out[k] = v
			continue
		}
		if _, ok := knownRefSuffixes[k]; !ok {
			return nil, fmt.Errorf("unknown ref key %q (grep knownRefSuffixes in seed_scenarios_refs.go for the supported set)", k)
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
		case "supplier_ref":
			targetKey = "supplier_id"
			id, err = lookups.supplierIDByCode(ctx, code)
		case "warehouse_ref":
			targetKey = "warehouse_id"
			id, err = lookups.warehouseIDByCode(ctx, code)
		case "currency_ref":
			targetKey = "currency_id"
			id, err = lookups.currencyIDByCode(ctx, code)
		case "user_ref":
			targetKey = "user_id"
			id, err = lookups.userIDByUsername(ctx, code)
		case "purchase_order_status_ref":
			targetKey = "purchase_order_status_id"
			id, err = lookups.purchaseOrderStatusIDByName(ctx, code)
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
