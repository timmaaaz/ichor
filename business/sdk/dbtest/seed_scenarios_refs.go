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
	"github.com/timmaaaz/ichor/business/sdk/seedid"
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
			if rows[0].Code != code {
				return uuid.Nil, fmt.Errorf("supplier not found for code=%s (closest match: %s)", code, rows[0].Code)
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
			filter := userbus.QueryFilter{UsernameExact: &name}
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
			filter := purchaseorderstatusbus.QueryFilter{NameExact: &name}
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
	"from_location_ref":         {},
	"to_location_ref":           {},
	"tote_ref":                  {},
	"supplier_ref":              {},
	"warehouse_ref":             {},
	"currency_ref":              {},
	"user_ref":                  {},
	"purchase_order_status_ref": {},
	// Non-standard column mappings: target column name does not follow the
	// "<prefix>_id" convention so explicit entries are required.
	"requested_by_ref": {},
	"approved_by_ref":  {},
}

// buildRowIndex performs a pre-pass over the entire scenario state and
// returns a map of label → deterministic UUID for every row that contains
// a "_label" key. The UUID is derived via seedid.Stable so it is stable
// across reseeds and matches the id that resolveRefs auto-injects.
//
// Fails if any label is non-string, empty, or duplicated within the scenario.
func buildRowIndex(scenarioName string, state map[string][]map[string]any) (map[string]uuid.UUID, error) {
	index := make(map[string]uuid.UUID)
	for tableSuffix, rows := range state {
		for i, row := range rows {
			raw, ok := row["_label"]
			if !ok {
				continue
			}
			label, ok := raw.(string)
			if !ok || label == "" {
				return nil, fmt.Errorf("scenario %s: %s[%d]: _label must be non-empty string", scenarioName, tableSuffix, i)
			}
			if _, dup := index[label]; dup {
				return nil, fmt.Errorf("scenario %s: duplicate _label %q", scenarioName, label)
			}
			index[label] = seedid.Stable(fmt.Sprintf("scenario:%s:label:%s", scenarioName, label))
		}
	}
	return index, nil
}

// resolveRefs rewrites a single state.yaml row in place:
//   - strips "_label" authoring directives and auto-injects a deterministic
//     "id" field on labelled rows
//   - resolves "<prefix>_row_ref" keys to "<prefix>_id" via the row index
//     (check _row_ref before _ref — longer suffix wins)
//   - replaces "<prefix>_ref" string values with "<prefix>_id" UUID values
//     using the appropriate resolver
//   - injects scenario_id if not already present
//
// Keys that don't end in "_ref" pass through untouched. Unknown "_ref" keys
// (anything not in knownRefSuffixes above) fail-hard so silent mis-seeding
// can't happen.
//
// rowIndex may be nil when no cross-row references are present.
func resolveRefs(ctx context.Context, row map[string]any, scenarioID uuid.UUID, lookups refLookups, rowIndex map[string]uuid.UUID) (map[string]any, error) {
	out := make(map[string]any, len(row)+1)

	// --- pass 1: handle _label (strip + auto-inject id) ---
	if raw, ok := row["_label"]; ok {
		label, ok := raw.(string)
		if !ok || label == "" {
			return nil, fmt.Errorf("_label must be non-empty string")
		}
		id := seedid.Stable(fmt.Sprintf("scenario:%s:label:%s", scenarioID.String(), label))
		// rowIndex was built from the scenario Name, but resolveRefs receives
		// scenarioID. Recover the id directly from rowIndex if available
		// (preferred — same derivation key), otherwise derive from scenarioID.
		if rowIndex != nil {
			if indexed, found := rowIndex[label]; found {
				id = indexed
			}
		}
		// If an explicit id is present it must match.
		if explicitID, hasID := row["id"]; hasID {
			if explicitID.(string) != id.String() {
				return nil, fmt.Errorf("row label %q has explicit id %s but expected %s", label, explicitID, id.String())
			}
		}
		out["id"] = id.String()
		// _label is an authoring directive; do not include in output.
	}

	for k, v := range row {
		switch {
		case k == "_label":
			// Already handled above — skip.
			continue

		case strings.HasSuffix(k, "_row_ref"):
			// Cross-row reference: <prefix>_row_ref → <prefix>_id via rowIndex.
			// Check _row_ref BEFORE _ref to avoid false matches (longer suffix wins).
			label, ok := v.(string)
			if !ok || label == "" {
				return nil, fmt.Errorf("row_ref key %q must be a non-empty string, got %T", k, v)
			}
			if rowIndex == nil {
				return nil, fmt.Errorf("row_ref %q not found (key: %s) — no row index available", label, k)
			}
			id, found := rowIndex[label]
			if !found {
				return nil, fmt.Errorf("row_ref %q not found (key: %s)", label, k)
			}
			targetKey := strings.TrimSuffix(k, "_row_ref") + "_id"
			out[targetKey] = id.String()

		case strings.HasSuffix(k, "_ref"):
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
			case "from_location_ref":
				targetKey = "from_location_id"
				id, err = lookups.locationIDByCode(ctx, code)
			case "to_location_ref":
				targetKey = "to_location_id"
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
			// Non-standard column mappings: target column name does not follow
			// the "<prefix>_id" convention (inventory.transfer_orders uses
			// "requested_by" / "approved_by" instead of "*_id").
			case "requested_by_ref":
				targetKey = "requested_by"
				id, err = lookups.userIDByUsername(ctx, code)
			case "approved_by_ref":
				targetKey = "approved_by"
				id, err = lookups.userIDByUsername(ctx, code)
			}
			if err != nil {
				return nil, fmt.Errorf("resolve %s=%q: %w", k, code, err)
			}
			out[targetKey] = id.String()

		default:
			out[k] = v
		}
	}

	if _, ok := out["scenario_id"]; !ok {
		out["scenario_id"] = scenarioID.String()
	}
	return out, nil
}
