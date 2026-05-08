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
	purchaseorderstatusbus "github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	supplierbus "github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	supplierproductbus "github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	customersbus "github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	lineitemfulfillmentstatusbus "github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	orderfulfillmentstatusbus "github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
)

// stableRowID returns the deterministic UUID for a fixture row's primary key
// when the row has no _label. Labelled rows get their UUID from buildRowIndex;
// this helper handles the no-label path so every row ends up with a PK value
// without authors having to label rows that aren't cross-row-referenced.
//
// Key prefix "scenario-row:" is distinct from the "scenario:%s:label:%s"
// prefix used by buildRowIndex (labelled-row UUIDs) and the "fixture:%s:%s:%d"
// prefix used by SeedScenariosFromRoot for inventory.scenario_fixtures rows.
//
// The actual PK column written is determined by pkColumnFor(targetTable); see
// resolveRefs.
func stableRowID(scenarioName, targetTable string, rowIndex int) uuid.UUID {
	return seedid.Stable(fmt.Sprintf("scenario-row:%s:%s:%d", scenarioName, targetTable, rowIndex))
}

// pkColumnFor returns the primary-key column name for a scenario-scoped table.
// 18 of 19 tables in resolveTargetTable follow the "id UUID NOT NULL" convention;
// workflow.approval_requests uses approval_request_id. Centralizing the
// special case here means both the _label path and the auto-inject path in
// resolveRefs stay column-name-aware without duplicating table knowledge.
func pkColumnFor(targetTable string) string {
	if targetTable == "workflow.approval_requests" {
		return "approval_request_id"
	}
	return "id"
}

// refResolver resolves a single stable human-readable code to a UUID.
// Each ref suffix has its own resolver so the dispatch in resolveRefs
// can stay a flat switch.
type refResolver func(ctx context.Context, value string) (uuid.UUID, error)

// refLookups bundles the resolver functions the seeder uses at
// fixture-materialization time. Exposed as an interface (via fields, not
// methods) so the unit test can pass fakes without touching a live DB.
type refLookups struct {
	productIDBySKU                    refResolver
	locationIDByCode                  refResolver
	labelIDByCode                     refResolver
	supplierIDByCode                  refResolver
	warehouseIDByCode                 refResolver
	currencyIDByCode                  refResolver
	userIDByUsername                  refResolver
	purchaseOrderStatusIDByName       refResolver
	orderFulfillmentStatusIDByName    refResolver
	lineItemFulfillmentStatusIDByName refResolver
	customerIDByName                  refResolver
	supplierProductIDByPartNumber     refResolver
}

// newRefLookups wires resolvers against real bus instances. Seeder path.
// Takes the full BusDomain rather than individual *Business pointers so
// adding new resolvers doesn't expand a positional argument list.
func newRefLookups(bd BusDomain) refLookups {
	return refLookups{
		productIDBySKU: func(ctx context.Context, sku string) (uuid.UUID, error) {
			filter := productbus.QueryFilter{SKU: &sku}
			orderBy := productbus.DefaultOrderBy
			pg := page.MustParse("1", "1")
			rows, err := bd.Product.Query(ctx, filter, orderBy, pg)
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
			orderBy := inventorylocationbus.DefaultOrderBy
			pg := page.MustParse("1", "1")
			rows, err := bd.InventoryLocation.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("location query code=%s: %w", code, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("location not found for code=%s", code)
			}
			return rows[0].LocationID, nil
		},
		labelIDByCode: func(ctx context.Context, code string) (uuid.UUID, error) {
			lc, err := bd.Label.QueryByCode(ctx, code)
			if err != nil {
				return uuid.Nil, fmt.Errorf("label queryByCode=%s: %w", code, err)
			}
			return lc.ID, nil
		},
		// Note: each resolver below uses an exact-match filter that returns at most
		// one row, so orderBy is decorative. Where a bus has a non-default sort key
		// (e.g. purchaseorderstatusbus.DefaultOrderBy = "sort_order"), we name the
		// orderBy explicitly to keep the resolver readable; otherwise we use
		// <bus>.DefaultOrderBy. Raw string literals like order.NewBy("code", ...)
		// match the bus's whitelist by convention; if a bus's OrderByFields whitelist
		// changes, the failing test will be Test_Seed_Integration.
		supplierIDByCode: func(ctx context.Context, code string) (uuid.UUID, error) {
			filter := supplierbus.QueryFilter{Code: &code}
			orderBy := order.NewBy("code", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := bd.Supplier.Query(ctx, filter, orderBy, pg)
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
			rows, err := bd.Warehouse.Query(ctx, filter, orderBy, pg)
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
			rows, err := bd.Currency.Query(ctx, filter, orderBy, pg)
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
			rows, err := bd.User.Query(ctx, filter, orderBy, pg)
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
			rows, err := bd.PurchaseOrderStatus.Query(ctx, filter, orderBy, pg)
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
		orderFulfillmentStatusIDByName: func(ctx context.Context, name string) (uuid.UUID, error) {
			filter := orderfulfillmentstatusbus.QueryFilter{Name: &name}
			orderBy := orderfulfillmentstatusbus.DefaultOrderBy
			pg := page.MustParse("1", "1")
			rows, err := bd.OrderFulfillmentStatus.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("order_fulfillment_status query name=%s: %w", name, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("order_fulfillment_status not found for name=%s", name)
			}
			return rows[0].ID, nil
		},
		lineItemFulfillmentStatusIDByName: func(ctx context.Context, name string) (uuid.UUID, error) {
			filter := lineitemfulfillmentstatusbus.QueryFilter{Name: &name}
			orderBy := lineitemfulfillmentstatusbus.DefaultOrderBy
			pg := page.MustParse("1", "1")
			rows, err := bd.LineItemFulfillmentStatus.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("line_item_fulfillment_status query name=%s: %w", name, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("line_item_fulfillment_status not found for name=%s", name)
			}
			return rows[0].ID, nil
		},
		customerIDByName: func(ctx context.Context, name string) (uuid.UUID, error) {
			filter := customersbus.QueryFilter{Name: &name}
			orderBy := order.NewBy("name", order.ASC)
			pg := page.MustParse("1", "1")
			rows, err := bd.Customers.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("customer query name=%s: %w", name, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("customer not found for name=%s", name)
			}
			return rows[0].ID, nil
		},
		// supplier_product_ref uses supplier_part_number as the lookup key.
		// supplier_products is a junction table (supplier_id + product_id) with no
		// stable natural key; scenario authors that self-author a supplier_products
		// row in state.yaml choose a unique part number that they also reference here.
		// The SQL filter is an exact equality match (= :supplier_part_number) so no
		// post-filter guard is needed.
		supplierProductIDByPartNumber: func(ctx context.Context, partNumber string) (uuid.UUID, error) {
			filter := supplierproductbus.QueryFilter{SupplierPartNumber: &partNumber}
			orderBy := supplierproductbus.DefaultOrderBy
			pg := page.MustParse("1", "1")
			rows, err := bd.SupplierProduct.Query(ctx, filter, orderBy, pg)
			if err != nil {
				return uuid.Nil, fmt.Errorf("supplier_product query supplier_part_number=%s: %w", partNumber, err)
			}
			if len(rows) == 0 {
				return uuid.Nil, fmt.Errorf("supplier_product not found for supplier_part_number=%s", partNumber)
			}
			return rows[0].SupplierProductID, nil
		},
	}
}

// refKeySuffix identifies a key that needs ref→id resolution. Kept as a
// constant set so unknown suffixes (e.g. vendor_ref) fail loudly rather
// than being silently passed through as strings into payload_json.
var knownRefSuffixes = map[string]struct{}{
	"product_ref":                      {},
	"location_ref":                     {},
	"from_location_ref":                {},
	"to_location_ref":                  {},
	"tote_ref":                         {},
	"supplier_ref":                     {},
	"supplier_product_ref":             {},
	"warehouse_ref":                    {},
	"currency_ref":                     {},
	"user_ref":                         {},
	"purchase_order_status_ref":        {},
	"order_fulfillment_status_ref":     {},
	"line_item_fulfillment_status_ref": {},
	"customer_ref":                     {},
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
//   - strips "_label" authoring directives and auto-injects the row's PK on
//     labelled rows (UUID derived from buildRowIndex)
//   - auto-injects the row's PK on UNLABELLED rows from defaultID (when
//     defaultID != uuid.Nil and no explicit PK is set); this lets shipped
//     scenarios author plain rows on tables with NOT NULL PKs without forcing
//     every row to be labelled. defaultID == uuid.Nil opts out (used by unit
//     tests that don't exercise the auto-inject path).
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
// PK contract priority (column name from pkColumnFor(targetTable) — usually
// "id", but "approval_request_id" for workflow.approval_requests):
//  1. Explicit PK key in row → respected (must match _label UUID if labelled).
//  2. _label present → UUID from rowIndex[label].
//  3. Otherwise → defaultID (when non-Nil); zero-value Nil opts out entirely.
//
// rowIndex may be nil when no cross-row references are present. targetTable
// may be "" (treated as the standard "id" PK convention) for unit tests that
// don't exercise non-standard PKs.
func resolveRefs(ctx context.Context, row map[string]any, scenarioID uuid.UUID, defaultID uuid.UUID, targetTable string, lookups refLookups, rowIndex map[string]uuid.UUID) (map[string]any, error) {
	out := make(map[string]any, len(row)+1)
	pkColumn := pkColumnFor(targetTable)

	// --- pass 1: handle _label (strip + auto-inject PK) ---
	if raw, ok := row["_label"]; ok {
		label, ok := raw.(string)
		if !ok || label == "" {
			return nil, fmt.Errorf("_label must be non-empty string")
		}
		if rowIndex == nil {
			return nil, fmt.Errorf("row label %q present but rowIndex is nil — caller must build row index via buildRowIndex before calling resolveRefs", label)
		}
		id, found := rowIndex[label]
		if !found {
			return nil, fmt.Errorf("row label %q not found in rowIndex", label)
		}
		// If an explicit PK is present it must match.
		if explicitID, hasID := row[pkColumn]; hasID {
			explicitStr, ok := explicitID.(string)
			if !ok {
				return nil, fmt.Errorf("row label %q has non-string explicit %s (type %T)", label, pkColumn, explicitID)
			}
			if explicitStr != id.String() {
				return nil, fmt.Errorf("row label %q has explicit %s %s but expected %s", label, pkColumn, explicitStr, id.String())
			}
		}
		out[pkColumn] = id.String()
		// _label is an authoring directive; do not include in output.
	}

	// If neither _label nor an explicit PK is set, auto-inject the
	// deterministic position-based UUID. defaultID == uuid.Nil opts out
	// (used by unit tests that don't exercise the auto-inject path).
	if _, hasLabel := row["_label"]; !hasLabel {
		if _, hasExplicitID := row[pkColumn]; !hasExplicitID && defaultID != uuid.Nil {
			out[pkColumn] = defaultID.String()
		}
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
			case "supplier_product_ref":
				targetKey = "supplier_product_id"
				id, err = lookups.supplierProductIDByPartNumber(ctx, code)
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
			case "order_fulfillment_status_ref":
				targetKey = "order_fulfillment_status_id"
				id, err = lookups.orderFulfillmentStatusIDByName(ctx, code)
			case "line_item_fulfillment_status_ref":
				// Column name on order_line_items is line_item_fulfillment_statuses_id
				// (plural "statuses"), not "line_item_fulfillment_status_id".
				targetKey = "line_item_fulfillment_statuses_id"
				id, err = lookups.lineItemFulfillmentStatusIDByName(ctx, code)
			case "customer_ref":
				targetKey = "customer_id"
				id, err = lookups.customerIDByName(ctx, code)
			// Non-standard column mappings: target column name does not follow
			// the "<prefix>_id" convention (inventory.transfer_orders uses
			// "requested_by" / "approved_by" instead of "*_id").
			case "requested_by_ref":
				targetKey = "requested_by"
				id, err = lookups.userIDByUsername(ctx, code)
			case "approved_by_ref":
				targetKey = "approved_by"
				id, err = lookups.userIDByUsername(ctx, code)
			default:
				// Defensive: knownRefSuffixes (above) and the switch (here) must
				// stay in lock-step. If a contributor adds an entry to the map
				// but forgets the case, fail loudly instead of silently writing
				// out[""] = uuid.Nil.String() which would drop the FK.
				return nil, fmt.Errorf("knownRefSuffixes/switch drift: %q has no resolver case", k)
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
