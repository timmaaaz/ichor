package directedworkapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
)

// selectNext applies the Phase 3 dispatcher policy to a list of
// normalized work items and returns the single best next task, or
// nil if none is directed.
//
// Policy (from spec):
//  1. If any items are in_progress → return the most recently
//     updated one. Done.
//  2. Else filter to pending items → return the highest-priority one.
//     Tiebreak: earliest DueAt (nil DueAt loses to a set DueAt).
//     Final tiebreak: earliest UpdatedAt.
//  3. Else return nil.
//
// This function is pure (no DB, no context) so unit tests drive it
// table-style with plain slices.
func selectNext(items []WorkItem) *WorkItem {
	if len(items) == 0 {
		return nil
	}

	// Step 1: in-progress wins.
	var bestInProgress *WorkItem
	for i := range items {
		if items[i].Status != WorkItemStatusInProgress {
			continue
		}
		if bestInProgress == nil || items[i].UpdatedAt.After(bestInProgress.UpdatedAt) {
			bestInProgress = &items[i]
		}
	}
	if bestInProgress != nil {
		return bestInProgress
	}

	// Step 2: highest-priority pending.
	var best *WorkItem
	for i := range items {
		if items[i].Status != WorkItemStatusPending {
			continue
		}
		if best == nil {
			best = &items[i]
			continue
		}
		if pendingBeats(items[i], *best) {
			best = &items[i]
		}
	}
	return best
}

// pendingBeats reports whether candidate outranks current under the
// pending-ordering rules: priority desc, then DueAt asc (nil loses),
// then UpdatedAt asc.
func pendingBeats(candidate, current WorkItem) bool {
	cr := priorityRank(candidate.Priority)
	rr := priorityRank(current.Priority)
	if cr != rr {
		return cr > rr
	}

	// Priority tie → earliest DueAt. nil DueAt loses to a set one;
	// two nils are equal on this axis and fall through to UpdatedAt.
	switch {
	case candidate.DueAt != nil && current.DueAt == nil:
		return true
	case candidate.DueAt == nil && current.DueAt != nil:
		return false
	case candidate.DueAt != nil && current.DueAt != nil:
		if candidate.DueAt.Before(*current.DueAt) {
			return true
		}
		if current.DueAt.Before(*candidate.DueAt) {
			return false
		}
	}

	// Due tie → earliest UpdatedAt.
	return candidate.UpdatedAt.Before(current.UpdatedAt)
}

// mapPickStatus converts picktaskbus.Status to the unified WorkItemStatus,
// returning (status, ok) — ok=false means the task is terminal and should
// be dropped. The default branch also drops "short_picked" and "cancelled"
// (terminal states from a worker's perspective) and any future additions
// to picktaskbus.Statuses that the directed-work surface hasn't learned
// about yet.
func mapPickStatus(s picktaskbus.Status) (WorkItemStatus, bool) {
	switch s.String() {
	case "pending":
		return WorkItemStatusPending, true
	case "in_progress":
		return WorkItemStatusInProgress, true
	default:
		return "", false
	}
}

// parsePriority coerces a string priority from a DB row into the
// WorkItemPriority enum. Empty / unknown falls back to Medium.
func parsePriority(p string) WorkItemPriority {
	switch p {
	case "low":
		return WorkItemPriorityLow
	case "medium":
		return WorkItemPriorityMedium
	case "high":
		return WorkItemPriorityHigh
	case "critical":
		return WorkItemPriorityCritical
	default:
		return WorkItemPriorityMedium
	}
}

// normalizePicks turns a slice of pick tasks into WorkItems, filtering
// out terminal statuses and FK-orphaned rows. ordersByID is the batch
// lookup result from ordersbus.QueryByIDs covering every SalesOrderID
// referenced by tasks — the caller is responsible for de-duping
// SalesOrderIDs and ensuring the map has an entry for each one. Tasks
// whose parent order is missing from the map are treated as FK orphans
// and silently dropped; callers that care should compare len(out) to
// len(tasks) and log the gap.
func normalizePicks(tasks []picktaskbus.PickTask, ordersByID map[uuid.UUID]ordersbus.Order) []WorkItem {
	out := make([]WorkItem, 0, len(tasks))
	for _, t := range tasks {
		status, ok := mapPickStatus(t.Status)
		if !ok {
			continue
		}
		order, exists := ordersByID[t.SalesOrderID]
		if !exists {
			// FK orphan: parent order was deleted. Skip this task.
			continue
		}
		locID := t.LocationID.String()
		// order.DueDate is non-pointer time.Time, so a NULL column or
		// unset field manifests as the zero time. Treat zero as "no due
		// date" so the dispatcher's nil-DueAt handling kicks in.
		var dueAt *time.Time
		if !order.DueDate.IsZero() {
			due := order.DueDate
			dueAt = &due
		}
		out = append(out, WorkItem{
			ID:         t.ID.String(),
			Type:       WorkItemTypePick,
			Status:     status,
			Title:      "Pick " + order.Number,
			DetailPath: "/floor/pick/" + t.ID.String(),
			UpdatedAt:  t.UpdatedDate,
			Priority:   parsePriority(order.Priority),
			DueAt:      dueAt,
			LocationID: &locID,
		})
	}
	return out
}

// mapPutawayStatus converts putawaytaskbus.Status to WorkItemStatus.
// The default branch drops "completed" and "cancelled" (terminal) and
// any future statuses not yet recognised by directed work.
func mapPutawayStatus(s putawaytaskbus.Status) (WorkItemStatus, bool) {
	switch s.String() {
	case "pending":
		return WorkItemStatusPending, true
	case "in_progress":
		return WorkItemStatusInProgress, true
	default:
		return "", false
	}
}

// normalizePutaways maps PutAwayTask → WorkItem with no parent lookup.
// V1 uses the reference_number string directly as the title — the field
// is free-text (not a UUID), so parent purchase-order enrichment via
// QueryByIDs doesn't apply. Priority fixed 'medium', DueAt nil.
//
// F22: putaway_tasks.reference_number is declared NOT NULL DEFAULT ''
// (migrate.sql:2117), so empty strings are possible. Fall back to an
// 8-char ID prefix so the title is never the bare "Putaway " with a
// trailing space.
func normalizePutaways(tasks []putawaytaskbus.PutAwayTask) []WorkItem {
	out := make([]WorkItem, 0, len(tasks))
	for _, t := range tasks {
		status, ok := mapPutawayStatus(t.Status)
		if !ok {
			continue
		}
		locID := t.LocationID.String()
		ref := t.ReferenceNumber
		if ref == "" {
			ref = t.ID.String()[:8]
		}
		out = append(out, WorkItem{
			ID:         t.ID.String(),
			Type:       WorkItemTypePutaway,
			Status:     status,
			Title:      "Putaway " + ref,
			DetailPath: "/floor/putaway/" + t.ID.String(),
			UpdatedAt:  t.UpdatedDate,
			Priority:   WorkItemPriorityMedium,
			DueAt:      nil,
			LocationID: &locID,
		})
	}
	return out
}

// mapCountStatus converts cyclecountitembus.Status to WorkItemStatus.
// Cycle counts have no in_progress state — the lifecycle is pending →
// counted → variance_approved / variance_rejected — so only "pending"
// survives. Everything else is terminal from a worker's perspective.
func mapCountStatus(s cyclecountitembus.Status) (WorkItemStatus, bool) {
	if s.String() == "pending" {
		return WorkItemStatusPending, true
	}
	return "", false
}

// normalizeCounts maps CycleCountItem → WorkItem. Cycle counts have no
// in_progress state (pending → counted → variance_*), so all non-pending
// items are dropped. Title uses an ID-substring fallback because no
// batch location-name lookup exists in V1 (see spec open question #2).
func normalizeCounts(items []cyclecountitembus.CycleCountItem) []WorkItem {
	out := make([]WorkItem, 0, len(items))
	for _, it := range items {
		if _, ok := mapCountStatus(it.Status); !ok {
			continue
		}
		locIDStr := it.LocationID.String()
		// Short ID prefix so the title isn't a full UUID wall.
		titleSuffix := locIDStr
		if len(titleSuffix) > 8 {
			titleSuffix = titleSuffix[:8]
		}
		out = append(out, WorkItem{
			ID:         it.ID.String(),
			Type:       WorkItemTypeCount,
			Status:     WorkItemStatusPending,
			Title:      "Cycle Count " + titleSuffix,
			DetailPath: "/floor/cycle-count/" + locIDStr,
			UpdatedAt:  it.UpdatedDate,
			Priority:   WorkItemPriorityMedium,
			DueAt:      nil,
			LocationID: &locIDStr,
		})
	}
	return out
}

func mapInspectionStatus(s string) (WorkItemStatus, bool) {
	switch s {
	case inspectionbus.StatusPending:
		return WorkItemStatusPending, true
	case inspectionbus.StatusInProgress:
		return WorkItemStatusInProgress, true
	default:
		return "", false
	}
}

// normalizeInspections maps inspectionbus.Inspection → WorkItem.
// Title uses an 8-char ID prefix; lot-number lookup is deferred (the
// lottrackingsbus has no batch QueryByIDs in V1).
func normalizeInspections(items []inspectionbus.Inspection) []WorkItem {
	out := make([]WorkItem, 0, len(items))
	for _, it := range items {
		status, ok := mapInspectionStatus(it.Status)
		if !ok {
			continue
		}
		idStr := it.InspectionID.String()
		titleSuffix := idStr
		if len(titleSuffix) > 8 {
			titleSuffix = titleSuffix[:8]
		}
		// NextInspectionDate is non-pointer time.Time, so treat the zero
		// value as "no scheduled date" to avoid polluting the dispatcher
		// with epoch-rank due dates.
		var dueAt *time.Time
		if !it.NextInspectionDate.IsZero() {
			next := it.NextInspectionDate
			dueAt = &next
		}
		out = append(out, WorkItem{
			ID:         idStr,
			Type:       WorkItemTypeInspect,
			Status:     status,
			Title:      "Inspect " + titleSuffix,
			DetailPath: "/floor/inspections/" + idStr,
			UpdatedAt:  it.UpdatedDate,
			Priority:   WorkItemPriorityMedium,
			DueAt:      dueAt,
			LocationID: nil,
		})
	}
	return out
}

// mapTransferStatus translates the free-string transferorderbus status
// into the unified surface. Only StatusApproved (ready to start) and
// StatusInTransit (in flight) are visible to directed work;
// StatusPending (awaiting supervisor), StatusRejected, and
// StatusCompleted are filtered. Uses transferorderbus constants instead
// of raw literals so a future rename of those constants breaks at
// compile time instead of silently dropping items.
func mapTransferStatus(s string) (WorkItemStatus, bool) {
	switch s {
	case transferorderbus.StatusApproved:
		return WorkItemStatusPending, true
	case transferorderbus.StatusInTransit:
		return WorkItemStatusInProgress, true
	default:
		return "", false
	}
}

func normalizeTransfers(items []transferorderbus.TransferOrder) []WorkItem {
	out := make([]WorkItem, 0, len(items))
	for _, it := range items {
		status, ok := mapTransferStatus(it.Status)
		if !ok {
			continue
		}
		idStr := it.TransferID.String()
		titleSuffix := idStr
		if len(titleSuffix) > 8 {
			titleSuffix = titleSuffix[:8]
		}
		fromLoc := it.FromLocationID.String()
		out = append(out, WorkItem{
			ID:         idStr,
			Type:       WorkItemTypeTransfer,
			Status:     status,
			Title:      "Transfer " + titleSuffix,
			DetailPath: "/floor/transfers/" + idStr,
			UpdatedAt:  it.UpdatedDate,
			Priority:   WorkItemPriorityMedium,
			DueAt:      nil,
			LocationID: &fromLoc,
		})
	}
	return out
}
