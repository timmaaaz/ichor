package directedworkapi

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
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
// be dropped.
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
// referenced by tasks.
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
		due := order.DueDate
		out = append(out, WorkItem{
			ID:         t.ID.String(),
			Type:       WorkItemTypePick,
			Status:     status,
			Title:      "Pick " + order.Number,
			DetailPath: "/floor/pick/" + order.ID.String(),
			UpdatedAt:  t.UpdatedDate,
			Priority:   parsePriority(order.Priority),
			DueAt:      &due,
			LocationID: &locID,
		})
	}
	return out
}

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
func normalizePutaways(tasks []putawaytaskbus.PutAwayTask) []WorkItem {
	out := make([]WorkItem, 0, len(tasks))
	for _, t := range tasks {
		status, ok := mapPutawayStatus(t.Status)
		if !ok {
			continue
		}
		locID := t.LocationID.String()
		out = append(out, WorkItem{
			ID:         t.ID.String(),
			Type:       WorkItemTypePutaway,
			Status:     status,
			Title:      "Putaway " + t.ReferenceNumber,
			DetailPath: "/floor/putaway/" + t.ID.String(),
			UpdatedAt:  t.UpdatedDate,
			Priority:   WorkItemPriorityMedium,
			DueAt:      nil,
			LocationID: &locID,
		})
	}
	return out
}
