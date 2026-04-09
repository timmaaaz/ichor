// Package directedworkapp provides the application logic for the floor
// directed-work feature. GET /v1/floor/work/next returns the single best
// next task for the authenticated worker, unified across picks, putaways,
// cycle counts, inspections, and transfers.
package directedworkapp

import (
	"time"
)

// WorkItemType enumerates the kinds of floor tasks directed-work can surface.
type WorkItemType string

const (
	WorkItemTypePick     WorkItemType = "pick"
	WorkItemTypePutaway  WorkItemType = "putaway"
	WorkItemTypeCount    WorkItemType = "count"
	WorkItemTypeInspect  WorkItemType = "inspect"
	WorkItemTypeTransfer WorkItemType = "transfer"
)

// WorkItemStatus is the unified (non-terminal-only) status surface for
// directed work. Each bus-specific status maps to one of these.
type WorkItemStatus string

const (
	WorkItemStatusPending    WorkItemStatus = "pending"
	WorkItemStatusInProgress WorkItemStatus = "in_progress"
)

// WorkItemPriority mirrors the four-level priority scheme used by
// workflow.notifications, sales.orders, procurement.purchase_orders.
type WorkItemPriority string

const (
	WorkItemPriorityLow      WorkItemPriority = "low"
	WorkItemPriorityMedium   WorkItemPriority = "medium"
	WorkItemPriorityHigh     WorkItemPriority = "high"
	WorkItemPriorityCritical WorkItemPriority = "critical"
)

// priorityRank returns a numeric ordering (higher = more urgent) so the
// dispatcher can compare priorities without a switch ladder at every call site.
func priorityRank(p WorkItemPriority) int {
	switch p {
	case WorkItemPriorityCritical:
		return 4
	case WorkItemPriorityHigh:
		return 3
	case WorkItemPriorityMedium:
		return 2
	case WorkItemPriorityLow:
		return 1
	default:
		return 0 // unknown → lowest
	}
}

// WorkItem is the unified shape returned by GET /v1/floor/work/next.
// Fields mirror the Phase 1 Vue FloorTask interface (after the Phase 3
// interface extension adds priority, due_at, location_id).
type WorkItem struct {
	ID         string           `json:"id"`
	Type       WorkItemType     `json:"type"`
	Status     WorkItemStatus   `json:"status"`
	Title      string           `json:"title"`
	DetailPath string           `json:"detail_path"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Priority   WorkItemPriority `json:"priority"`
	DueAt      *time.Time       `json:"due_at,omitempty"`
	LocationID *string          `json:"location_id,omitempty"`
}
