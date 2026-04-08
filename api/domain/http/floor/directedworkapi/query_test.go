package directedworkapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
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

func TestNormalizePicks(t *testing.T) {
	orderID := uuid.New()
	locID := uuid.New()
	userID := uuid.New()
	due := time.Now().Add(2 * time.Hour)

	ordersByID := map[uuid.UUID]ordersbus.Order{
		orderID: {
			ID:       orderID,
			Number:   "SO-2025-0891",
			DueDate:  due,
			Priority: "high",
		},
	}

	tasks := []picktaskbus.PickTask{
		{
			ID:           uuid.New(),
			SalesOrderID: orderID,
			LocationID:   locID,
			Status:       picktaskbus.Statuses.InProgress,
			AssignedTo:   userID,
			UpdatedDate:  time.Now(),
		},
		{
			ID:           uuid.New(),
			SalesOrderID: orderID,
			LocationID:   locID,
			Status:       picktaskbus.Statuses.Completed, // should be filtered out
			AssignedTo:   userID,
			UpdatedDate:  time.Now(),
		},
		{
			ID:           uuid.New(),
			SalesOrderID: orderID,
			LocationID:   locID,
			Status:       picktaskbus.Statuses.Pending,
			AssignedTo:   userID,
			UpdatedDate:  time.Now(),
		},
	}

	got := normalizePicks(tasks, ordersByID)
	if len(got) != 2 {
		t.Fatalf("expected 2 items (1 completed filtered), got %d", len(got))
	}
	for _, w := range got {
		if w.Type != WorkItemTypePick {
			t.Errorf("expected type=pick, got %s", w.Type)
		}
		if w.Title != "Pick SO-2025-0891" {
			t.Errorf("expected title 'Pick SO-2025-0891', got %q", w.Title)
		}
		if w.Priority != WorkItemPriorityHigh {
			t.Errorf("expected priority=high, got %s", w.Priority)
		}
		if w.DueAt == nil || !w.DueAt.Equal(due) {
			t.Errorf("expected DueAt=%v, got %v", due, w.DueAt)
		}
		if w.LocationID == nil || *w.LocationID != locID.String() {
			t.Errorf("expected LocationID=%s, got %v", locID, w.LocationID)
		}
	}
}

func TestNormalizePicks_UnknownPriorityFallsBackToMedium(t *testing.T) {
	orderID := uuid.New()
	ordersByID := map[uuid.UUID]ordersbus.Order{
		orderID: {ID: orderID, Number: "SO-X", Priority: "" /* defaulted */},
	}
	tasks := []picktaskbus.PickTask{
		{ID: uuid.New(), SalesOrderID: orderID, Status: picktaskbus.Statuses.Pending, UpdatedDate: time.Now()},
	}
	got := normalizePicks(tasks, ordersByID)
	if len(got) != 1 || got[0].Priority != WorkItemPriorityMedium {
		t.Fatalf("expected medium fallback, got %+v", got)
	}
}

func TestNormalizePicks_MissingParentOrderSkips(t *testing.T) {
	tasks := []picktaskbus.PickTask{
		{ID: uuid.New(), SalesOrderID: uuid.New(), Status: picktaskbus.Statuses.Pending, UpdatedDate: time.Now()},
	}
	got := normalizePicks(tasks, map[uuid.UUID]ordersbus.Order{})
	if len(got) != 0 {
		t.Fatalf("expected 0 items when parent order is missing (FK orphan), got %d", len(got))
	}
}

func TestNormalizePutaways(t *testing.T) {
	userID := uuid.New()
	locID := uuid.New()
	tasks := []putawaytaskbus.PutAwayTask{
		{
			ID:              uuid.New(),
			LocationID:      locID,
			ReferenceNumber: "PO-2025-0042",
			Status:          putawaytaskbus.Statuses.Pending,
			AssignedTo:      userID,
			UpdatedDate:     time.Now(),
		},
		{
			ID:              uuid.New(),
			LocationID:      locID,
			ReferenceNumber: "PO-2025-0043",
			Status:          putawaytaskbus.Statuses.Completed, // filtered
			AssignedTo:      userID,
			UpdatedDate:     time.Now(),
		},
		{
			ID:              uuid.New(),
			LocationID:      locID,
			ReferenceNumber: "PO-2025-0044",
			Status:          putawaytaskbus.Statuses.InProgress,
			AssignedTo:      userID,
			UpdatedDate:     time.Now(),
		},
	}

	got := normalizePutaways(tasks)
	if len(got) != 2 {
		t.Fatalf("expected 2 items (1 completed filtered), got %d", len(got))
	}
	if got[0].Title != "Putaway PO-2025-0042" {
		t.Errorf("expected title 'Putaway PO-2025-0042', got %q", got[0].Title)
	}
	if got[0].Priority != WorkItemPriorityMedium {
		t.Errorf("expected fixed medium priority, got %s", got[0].Priority)
	}
	if got[0].DueAt != nil {
		t.Errorf("expected nil DueAt for putaway V1, got %v", got[0].DueAt)
	}
	if got[0].LocationID == nil || *got[0].LocationID != locID.String() {
		t.Errorf("expected LocationID=%s, got %v", locID, got[0].LocationID)
	}
}

func TestNormalizeCounts(t *testing.T) {
	locID := uuid.New()
	items := []cyclecountitembus.CycleCountItem{
		{
			ID:          uuid.New(),
			LocationID:  locID,
			Status:      cyclecountitembus.Statuses.Pending,
			UpdatedDate: time.Now(),
		},
		{
			ID:          uuid.New(),
			LocationID:  locID,
			Status:      cyclecountitembus.Statuses.Counted, // terminal → filtered
			UpdatedDate: time.Now(),
		},
	}

	got := normalizeCounts(items)
	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got))
	}
	w := got[0]
	if w.Type != WorkItemTypeCount {
		t.Errorf("expected type=count, got %s", w.Type)
	}
	wantTitlePrefix := "Cycle Count "
	if len(w.Title) < len(wantTitlePrefix) || w.Title[:len(wantTitlePrefix)] != wantTitlePrefix {
		t.Errorf("expected title to start with %q, got %q", wantTitlePrefix, w.Title)
	}
	if w.Priority != WorkItemPriorityMedium {
		t.Errorf("expected medium priority, got %s", w.Priority)
	}
	if w.DueAt != nil {
		t.Errorf("expected nil DueAt for counts, got %v", w.DueAt)
	}
	if w.Status != WorkItemStatusPending {
		t.Errorf("expected Pending (no in_progress concept), got %s", w.Status)
	}
}
