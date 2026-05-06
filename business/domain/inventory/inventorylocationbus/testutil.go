package inventorylocationbus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
)

// specCode describes one of the 19 hardcoded warehouse locations from
// Phase 1 spec §3.4 (physical warehouse layout). Each code corresponds to a
// physical sticker on the warehouse floor that the floor app resolves via
// LocationCodeExact lookup. See seed_scenarios_refs.go for the scan-key map.
type specCode struct {
	code    string // e.g. "STG-A01"
	zonePfx string // owning zone's ZoneCode (RCV/QA/STG/PCK/PKG/SHP)
	aisle   string // "A"/"B"/"C" for STG, "" for non-aisled zones
	rack    string // 2-digit zero-padded numeric segment
}

// specCodes is the canonical, alphabetically-sorted catalogue of the 19
// physical location codes. Order is fixed so callers indexing the returned
// slice (e.g. sd.InventoryLocations[0]) get deterministic results across runs.
var specCodes = []specCode{
	{code: "PCK-01", zonePfx: "PCK", aisle: "", rack: "01"},
	{code: "PCK-02", zonePfx: "PCK", aisle: "", rack: "02"},
	{code: "PCK-03", zonePfx: "PCK", aisle: "", rack: "03"},
	{code: "PKG-01", zonePfx: "PKG", aisle: "", rack: "01"},
	{code: "PKG-02", zonePfx: "PKG", aisle: "", rack: "02"},
	{code: "QA-01", zonePfx: "QA", aisle: "", rack: "01"},
	{code: "RCV-01", zonePfx: "RCV", aisle: "", rack: "01"},
	{code: "RCV-02", zonePfx: "RCV", aisle: "", rack: "02"},
	{code: "SHP-01", zonePfx: "SHP", aisle: "", rack: "01"},
	{code: "SHP-02", zonePfx: "SHP", aisle: "", rack: "02"},
	{code: "STG-A01", zonePfx: "STG", aisle: "A", rack: "01"},
	{code: "STG-A02", zonePfx: "STG", aisle: "A", rack: "02"},
	{code: "STG-A03", zonePfx: "STG", aisle: "A", rack: "03"},
	{code: "STG-B01", zonePfx: "STG", aisle: "B", rack: "01"},
	{code: "STG-B02", zonePfx: "STG", aisle: "B", rack: "02"},
	{code: "STG-B03", zonePfx: "STG", aisle: "B", rack: "03"},
	{code: "STG-C01", zonePfx: "STG", aisle: "C", rack: "01"},
	{code: "STG-C02", zonePfx: "STG", aisle: "C", rack: "02"},
	{code: "STG-C03", zonePfx: "STG", aisle: "C", rack: "03"},
}

// specCatalogueSize is the number of codes in specCodes. Exposed via the
// returned error from TestNewInventoryLocation when n > catalogue size so
// stale callers (passing the legacy n=25) fail loudly instead of silently
// truncating.
const specCatalogueSize = 19

// TestNewInventoryLocation produces up to n NewInventoryLocation values
// drawn from the Phase 1 spec §3.4 catalogue (specCodes). The function
// matches each spec code against the first input zone whose ZoneCode equals
// the code's zone prefix; codes whose prefix is absent from zones[] are
// skipped, so the returned slice may be shorter than n.
//
// Aisle/Shelf/Bin: STG codes carry a real Aisle ("A"/"B"/"C") and Rack;
// non-aisled zones (RCV/QA/PCK/PKG/SHP) use Aisle="" and only Rack.
// Shelf and Bin are always "" — the spec catalogue does not currently
// distinguish bin-level positions for these locations. The
// inventory_locations schema (migrate.sql v1.52) declares aisle/rack/shelf/bin
// as VARCHAR(20) NOT NULL with no CHECK constraint; empty strings are
// schema-valid.
//
// Returns an error if n > specCatalogueSize (19) — this catches callers
// still passing the legacy n=25 after the spec re-seed. Pass n=19 (or fewer
// if the test only needs a subset).
func TestNewInventoryLocation(n int, warehouseIDs []uuid.UUID, zones []zonebus.Zone) ([]NewInventoryLocation, error) {
	if n > specCatalogueSize {
		return nil, fmt.Errorf("n exceeds spec catalogue size (%d)", specCatalogueSize)
	}
	if n < 0 {
		return nil, fmt.Errorf("n must be non-negative, got %d", n)
	}
	if len(warehouseIDs) == 0 {
		return nil, fmt.Errorf("warehouseIDs must contain at least one warehouse")
	}

	// Index zones by ZoneCode for first-match lookup.
	zoneByCode := make(map[string]zonebus.Zone, len(zones))
	for _, z := range zones {
		if z.ZoneCode == nil {
			continue
		}
		// Preserve first-occurrence semantics: if multiple zones share a
		// code (zoneCodeCycle re-uses STG/PCK/PKG/RCV), only the first wins.
		if _, ok := zoneByCode[*z.ZoneCode]; !ok {
			zoneByCode[*z.ZoneCode] = z
		}
	}

	out := make([]NewInventoryLocation, 0, n)
	for i, sc := range specCodes {
		if len(out) >= n {
			break
		}
		zone, ok := zoneByCode[sc.zonePfx]
		if !ok {
			continue
		}
		code := sc.code
		// Alternate IsPickLocation/IsReserveLocation by spec-catalogue index
		// (preserves the prior testutil behavior of producing both kinds).
		isPick := i%2 == 0
		out = append(out, NewInventoryLocation{
			WarehouseID:        warehouseIDs[0],
			ZoneID:             zone.ZoneID,
			Aisle:              sc.aisle,
			Rack:               sc.rack,
			Shelf:              "",
			Bin:                "",
			LocationCode:       &code,
			IsPickLocation:     isPick,
			IsReserveLocation:  !isPick,
			MaxCapacity:        100,
			CurrentUtilization: types.RoundedFloat{Value: 0},
		})
	}

	return out, nil
}

// TestSeedInventoryLocations creates up to n inventory locations from the
// Phase 1 spec §3.4 catalogue. See TestNewInventoryLocation for the
// availability/zone-matching rules. The returned slice is sorted by
// LocationID.String() for parity with the bus Query default order
// (ORDER BY id ASC) — paged-query tests can compare positional indices
// directly against this slice.
func TestSeedInventoryLocations(ctx context.Context, n int, warehouseIDs []uuid.UUID, zones []zonebus.Zone, api *Business) ([]InventoryLocation, error) {
	newInventoryLocations, err := TestNewInventoryLocation(n, warehouseIDs, zones)
	if err != nil {
		return nil, err
	}

	inventoryLocations := make([]InventoryLocation, len(newInventoryLocations))

	for i, newInvLoc := range newInventoryLocations {
		il, err := api.Create(ctx, newInvLoc)
		if err != nil {
			return nil, fmt.Errorf("seeding error: %v", err)
		}
		inventoryLocations[i] = il
	}

	sort.Slice(inventoryLocations, func(i, j int) bool {
		return inventoryLocations[i].LocationID.String() < inventoryLocations[j].LocationID.String()
	})

	return inventoryLocations, nil
}
