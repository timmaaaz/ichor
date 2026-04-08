package directedworkapi

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
