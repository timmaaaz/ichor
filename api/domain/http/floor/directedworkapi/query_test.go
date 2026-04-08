package directedworkapi

import (
	"testing"
	"time"
)

func ptrTime(t time.Time) *time.Time { return &t }

func TestSelectNext_EmptyInputReturnsNil(t *testing.T) {
	got := selectNext(nil)
	if got != nil {
		t.Fatalf("expected nil for empty input, got %+v", got)
	}
	got = selectNext([]WorkItem{})
	if got != nil {
		t.Fatalf("expected nil for empty slice, got %+v", got)
	}
}

func TestSelectNext_InProgressBeatsPending(t *testing.T) {
	now := time.Now()
	items := []WorkItem{
		{ID: "pending-critical", Status: WorkItemStatusPending, Priority: WorkItemPriorityCritical, UpdatedAt: now},
		{ID: "inprog-low", Status: WorkItemStatusInProgress, Priority: WorkItemPriorityLow, UpdatedAt: now.Add(-time.Hour)},
	}
	got := selectNext(items)
	if got == nil || got.ID != "inprog-low" {
		t.Fatalf("expected in-progress to win over pending-critical, got %+v", got)
	}
}

func TestSelectNext_MostRecentInProgressWins(t *testing.T) {
	now := time.Now()
	items := []WorkItem{
		{ID: "older", Status: WorkItemStatusInProgress, Priority: WorkItemPriorityMedium, UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "newer", Status: WorkItemStatusInProgress, Priority: WorkItemPriorityMedium, UpdatedAt: now.Add(-30 * time.Minute)},
		{ID: "oldest", Status: WorkItemStatusInProgress, Priority: WorkItemPriorityMedium, UpdatedAt: now.Add(-5 * time.Hour)},
	}
	got := selectNext(items)
	if got == nil || got.ID != "newer" {
		t.Fatalf("expected most-recent updated_at, got %+v", got)
	}
}

func TestSelectNext_HighestPriorityPendingWins(t *testing.T) {
	now := time.Now()
	items := []WorkItem{
		{ID: "low", Status: WorkItemStatusPending, Priority: WorkItemPriorityLow, UpdatedAt: now},
		{ID: "high", Status: WorkItemStatusPending, Priority: WorkItemPriorityHigh, UpdatedAt: now},
		{ID: "medium", Status: WorkItemStatusPending, Priority: WorkItemPriorityMedium, UpdatedAt: now},
		{ID: "critical", Status: WorkItemStatusPending, Priority: WorkItemPriorityCritical, UpdatedAt: now},
	}
	got := selectNext(items)
	if got == nil || got.ID != "critical" {
		t.Fatalf("expected critical to win, got %+v", got)
	}
}

func TestSelectNext_PriorityTieBrokenByEarliestDueAt(t *testing.T) {
	now := time.Now()
	items := []WorkItem{
		{ID: "due-later", Status: WorkItemStatusPending, Priority: WorkItemPriorityHigh, UpdatedAt: now, DueAt: ptrTime(now.Add(5 * time.Hour))},
		{ID: "due-sooner", Status: WorkItemStatusPending, Priority: WorkItemPriorityHigh, UpdatedAt: now, DueAt: ptrTime(now.Add(1 * time.Hour))},
	}
	got := selectNext(items)
	if got == nil || got.ID != "due-sooner" {
		t.Fatalf("expected earliest due_at to win priority tie, got %+v", got)
	}
}

func TestSelectNext_PriorityTieNilDueAtLosesToSetDueAt(t *testing.T) {
	now := time.Now()
	items := []WorkItem{
		{ID: "no-due", Status: WorkItemStatusPending, Priority: WorkItemPriorityHigh, UpdatedAt: now, DueAt: nil},
		{ID: "has-due", Status: WorkItemStatusPending, Priority: WorkItemPriorityHigh, UpdatedAt: now, DueAt: ptrTime(now.Add(1 * time.Hour))},
	}
	got := selectNext(items)
	if got == nil || got.ID != "has-due" {
		t.Fatalf("expected items with a due date to beat nil-due on priority tie, got %+v", got)
	}
}

func TestSelectNext_DueAtTieBrokenByEarliestUpdatedAt(t *testing.T) {
	now := time.Now()
	due := now.Add(3 * time.Hour)
	items := []WorkItem{
		{ID: "touched-now", Status: WorkItemStatusPending, Priority: WorkItemPriorityMedium, UpdatedAt: now, DueAt: &due},
		{ID: "touched-earlier", Status: WorkItemStatusPending, Priority: WorkItemPriorityMedium, UpdatedAt: now.Add(-2 * time.Hour), DueAt: &due},
	}
	got := selectNext(items)
	if got == nil || got.ID != "touched-earlier" {
		t.Fatalf("expected earliest updated_at to win due_at tie, got %+v", got)
	}
}
