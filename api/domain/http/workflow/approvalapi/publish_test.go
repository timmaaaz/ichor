package approvalapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
)

// TestBuildApprovalResolvedPayload asserts the shape of the WebSocket payload
// built for an approval_resolved event. If a new field is added to the shape,
// every caller is forced to supply it through the signature.
func TestBuildApprovalResolvedPayload(t *testing.T) {
	resolvedBy := uuid.New()
	resolvedDate := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)

	approval := approvalrequestbus.ApprovalRequest{
		ID:           uuid.New(),
		RuleID:       uuid.New(),
		ActionName:   "seek_approval_0",
		Status:       approvalrequestbus.StatusApproved,
		ResolvedDate: &resolvedDate,
	}

	got := buildApprovalResolvedPayload(approval, resolvedBy)

	wantKeys := []string{"approvalId", "status", "resolvedBy", "ruleId", "actionName", "resolvedDate"}
	for _, key := range wantKeys {
		if _, ok := got[key]; !ok {
			t.Errorf("payload missing key %q", key)
		}
	}

	if got["approvalId"] != approval.ID.String() {
		t.Errorf("approvalId = %v, want %s", got["approvalId"], approval.ID.String())
	}
	if got["status"] != approval.Status {
		t.Errorf("status = %v, want %s", got["status"], approval.Status)
	}
	if got["resolvedBy"] != resolvedBy.String() {
		t.Errorf("resolvedBy = %v, want %s", got["resolvedBy"], resolvedBy.String())
	}
	if got["ruleId"] != approval.RuleID.String() {
		t.Errorf("ruleId = %v, want %s", got["ruleId"], approval.RuleID.String())
	}
	if got["actionName"] != approval.ActionName {
		t.Errorf("actionName = %v, want %s", got["actionName"], approval.ActionName)
	}
	if got["resolvedDate"] != resolvedDate.Format(time.RFC3339) {
		t.Errorf("resolvedDate = %v, want %s", got["resolvedDate"], resolvedDate.Format(time.RFC3339))
	}
}

// TestBuildApprovalResolvedPayload_NoResolvedDate asserts resolvedDate is omitted
// when the approval hasn't been resolved yet (nil pointer case).
func TestBuildApprovalResolvedPayload_NoResolvedDate(t *testing.T) {
	approval := approvalrequestbus.ApprovalRequest{
		ID:         uuid.New(),
		RuleID:     uuid.New(),
		ActionName: "seek_approval_0",
		Status:     approvalrequestbus.StatusPending,
	}

	got := buildApprovalResolvedPayload(approval, uuid.New())

	if _, ok := got["resolvedDate"]; ok {
		t.Error("resolvedDate should not be present when approval.ResolvedDate is nil")
	}
}
