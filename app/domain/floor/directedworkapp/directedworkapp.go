package directedworkapp

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// maxWorkerTasksPerDomain caps how many rows per floor-work domain the
// handler pulls from each bus before the dispatcher picks the winner.
// Per-worker active task volume is in the dozens at most, so 500 is an
// order of magnitude of headroom.
const maxWorkerTasksPerDomain = 500

// App manages the application logic for the floor directed-work feature.
type App struct {
	log               *logger.Logger
	pickTaskBus       *picktaskbus.Business
	putAwayTaskBus    *putawaytaskbus.Business
	cycleCountItemBus *cyclecountitembus.Business
	inspectionBus     *inspectionbus.Business
	transferOrderBus  *transferorderbus.Business
	ordersBus         *ordersbus.Business
}

// NewApp constructs a directed-work app for use.
func NewApp(
	log *logger.Logger,
	pickTaskBus *picktaskbus.Business,
	putAwayTaskBus *putawaytaskbus.Business,
	cycleCountItemBus *cyclecountitembus.Business,
	inspectionBus *inspectionbus.Business,
	transferOrderBus *transferorderbus.Business,
	ordersBus *ordersbus.Business,
) *App {
	return &App{
		log:               log,
		pickTaskBus:       pickTaskBus,
		putAwayTaskBus:    putAwayTaskBus,
		cycleCountItemBus: cycleCountItemBus,
		inspectionBus:     inspectionBus,
		transferOrderBus:  transferOrderBus,
		ordersBus:         ordersBus,
	}
}

// QueryNext returns the single best next work item for the given worker,
// or nil if nothing is directed.
//
// Sequential by design. errgroup-based fan-out was considered and
// rejected to avoid introducing a new concurrency pattern for a single
// handler. Total latency at p50 is ~75ms; parallelizing would save
// ~50ms which is imperceptible for a nav refetch. If p99 latency ever
// matters for this endpoint, revisit and introduce errgroup.
func (a *App) QueryNext(ctx context.Context, userID uuid.UUID) (*WorkItem, error) {
	// Use a generously large page size; per-worker active task volume
	// is in the dozens at most. Default orderBy is fine — the dispatcher
	// re-sorts in Go anyway.
	pg, err := page.Parse("1", strconv.Itoa(maxWorkerTasksPerDomain))
	if err != nil {
		return nil, errs.Newf(errs.Internal, "page setup: %s", err)
	}
	asc := order.NewBy("id", order.ASC)

	var items []WorkItem

	// --- Picks ---
	pickFilter := picktaskbus.QueryFilter{AssignedTo: &userID}
	picks, err := a.pickTaskBus.Query(ctx, pickFilter, asc, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query picks: %s", err)
	}
	orderIDSet := make(map[uuid.UUID]struct{})
	for _, p := range picks {
		orderIDSet[p.SalesOrderID] = struct{}{}
	}
	orderIDs := make([]uuid.UUID, 0, len(orderIDSet))
	for id := range orderIDSet {
		orderIDs = append(orderIDs, id)
	}
	var ordersByID map[uuid.UUID]ordersbus.Order
	if len(orderIDs) > 0 {
		os, err := a.ordersBus.QueryByIDs(ctx, orderIDs)
		if err != nil {
			return nil, errs.Newf(errs.Internal, "query parent orders: %s", err)
		}
		ordersByID = make(map[uuid.UUID]ordersbus.Order, len(os))
		for _, o := range os {
			ordersByID[o.ID] = o
		}
	} else {
		ordersByID = map[uuid.UUID]ordersbus.Order{}
	}
	normalizedPicks := normalizePicks(picks, ordersByID)
	items = append(items, normalizedPicks...)

	// F21: normalizePicks silently drops picks whose parent order is missing
	// from ordersByID (FK orphan). Log a warning so data integrity issues or
	// de-dup bugs in QueryByIDs are observable. Comparing the normalized
	// count directly (rather than len(items)) keeps the invariant explicit
	// and order-independent from the later domain appends.
	if len(normalizedPicks) < len(picks) {
		a.log.Warn(ctx, "directedwork.normalizePicks: dropped orphan picks",
			"picks_in", len(picks),
			"picks_out", len(normalizedPicks),
			"orders_loaded", len(ordersByID),
			"user_id", userID,
		)
	}

	// --- Putaways ---
	putaways, err := a.putAwayTaskBus.Query(ctx, putawaytaskbus.QueryFilter{AssignedTo: &userID}, asc, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query putaways: %s", err)
	}
	items = append(items, normalizePutaways(putaways)...)

	// --- Cycle count items ---
	// cycle_count_items has no assigned_to column. CountedBy is set only
	// AFTER a worker finishes counting (not before). V1 surfaces all
	// pending cycle count items regardless of worker so the dispatcher
	// can direct any available worker to them.
	pendingStatus := cyclecountitembus.Statuses.Pending
	counts, err := a.cycleCountItemBus.Query(ctx, cyclecountitembus.QueryFilter{Status: &pendingStatus}, asc, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query cycle counts: %s", err)
	}
	items = append(items, normalizeCounts(counts)...)

	// --- Inspections ---
	inspections, err := a.inspectionBus.Query(ctx, inspectionbus.QueryFilter{InspectorID: &userID}, asc, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query inspections: %s", err)
	}
	items = append(items, normalizeInspections(inspections)...)

	// --- Transfers ---
	// Transfer lifecycle: pending → approved → in_transit → completed.
	// Approved transfers have ClaimedByID=nil (not yet claimed by any worker)
	// and are visible to all authenticated workers. In-transit transfers are
	// only visible to the worker who claimed them.
	approvedStatus := transferorderbus.StatusApproved
	approvedTransfers, err := a.transferOrderBus.Query(ctx, transferorderbus.QueryFilter{Status: &approvedStatus}, asc, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query approved transfers: %s", err)
	}
	inTransitStatus := transferorderbus.StatusInTransit
	ownTransfers, err := a.transferOrderBus.Query(ctx, transferorderbus.QueryFilter{ClaimedByID: &userID, Status: &inTransitStatus}, asc, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query in-transit transfers: %s", err)
	}
	allTransfers := append(approvedTransfers, ownTransfers...)
	items = append(items, normalizeTransfers(allTransfers)...)

	return selectNext(items), nil
}
