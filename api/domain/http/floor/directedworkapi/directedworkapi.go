package directedworkapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// maxWorkerTasksPerDomain caps how many rows per floor-work domain the
// handler pulls from each bus before the dispatcher picks the winner.
// Per-worker active task volume is in the dozens at most, so 500 is an
// order of magnitude of headroom.
const maxWorkerTasksPerDomain = 500

type api struct {
	log               *logger.Logger
	pickTaskBus       *picktaskbus.Business
	putAwayTaskBus    *putawaytaskbus.Business
	cycleCountItemBus *cyclecountitembus.Business
	inspectionBus     *inspectionbus.Business
	transferOrderBus  *transferorderbus.Business
	ordersBus         *ordersbus.Business
}

func newAPI(cfg Config) *api {
	return &api{
		log:               cfg.Log,
		pickTaskBus:       cfg.PickTaskBus,
		putAwayTaskBus:    cfg.PutAwayTaskBus,
		cycleCountItemBus: cfg.CycleCountItemBus,
		inspectionBus:     cfg.InspectionBus,
		transferOrderBus:  cfg.TransferOrderBus,
		ordersBus:         cfg.OrdersBus,
	}
}

// queryNext returns the single best next work item for the authenticated
// worker, or {"work_item": null} if nothing is directed. See the spec
// for the full dispatcher policy.
//
// Sequential by design. errgroup-based fan-out was considered and
// rejected to avoid introducing a new concurrency pattern for a single
// handler. Total latency at p50 is ~75ms; parallelizing would save
// ~50ms which is imperceptible for a nav refetch. If p99 latency ever
// matters for this endpoint, revisit and introduce errgroup.
func (a *api) queryNext(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// Use a generously large page size; per-worker active task volume
	// is in the dozens at most. Default orderBy is fine — the dispatcher
	// re-sorts in Go anyway.
	pg, err := page.Parse("1", strconv.Itoa(maxWorkerTasksPerDomain))
	if err != nil {
		return errs.Newf(errs.Internal, "page setup: %s", err)
	}
	asc := order.NewBy("id", order.ASC)

	var items []WorkItem

	// --- Picks ---
	pickFilter := picktaskbus.QueryFilter{AssignedTo: &userID}
	picks, err := a.pickTaskBus.Query(ctx, pickFilter, asc, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query picks: %s", err)
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
			return errs.Newf(errs.Internal, "query parent orders: %s", err)
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
		return errs.Newf(errs.Internal, "query putaways: %s", err)
	}
	items = append(items, normalizePutaways(putaways)...)

	// --- Cycle count items ---
	counts, err := a.cycleCountItemBus.Query(ctx, cyclecountitembus.QueryFilter{CountedBy: &userID}, asc, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query cycle counts: %s", err)
	}
	items = append(items, normalizeCounts(counts)...)

	// --- Inspections ---
	inspections, err := a.inspectionBus.Query(ctx, inspectionbus.QueryFilter{InspectorID: &userID}, asc, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query inspections: %s", err)
	}
	items = append(items, normalizeInspections(inspections)...)

	// --- Transfers ---
	transfers, err := a.transferOrderBus.Query(ctx, transferorderbus.QueryFilter{ClaimedByID: &userID}, asc, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query transfers: %s", err)
	}
	items = append(items, normalizeTransfers(transfers)...)

	return Response{WorkItem: selectNext(items)}
}
