package approval

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuildApprovalRequestMessages_OnePerApprover(t *testing.T) {
	approvalID := uuid.New()
	ruleID := uuid.New()
	a1, a2 := uuid.New(), uuid.New()

	msgs := buildApprovalRequestMessages(approvalID, ruleID, "approval_hold", []uuid.UUID{a1, a2})

	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2 (one per approver)", len(msgs))
	}
	for i, m := range msgs {
		if m.Type != "approval_request" {
			t.Fatalf("msg[%d].Type = %q, want approval_request", i, m.Type)
		}
		if m.EntityID != approvalID {
			t.Fatalf("msg[%d].EntityID = %v, want approval id", i, m.EntityID)
		}
		if m.Payload["approvalId"] != approvalID.String() {
			t.Fatalf("msg[%d] payload approvalId = %v", i, m.Payload["approvalId"])
		}
	}
	if msgs[0].UserID != a1 || msgs[1].UserID != a2 {
		t.Fatalf("per-approver UserID targeting wrong: %v, %v", msgs[0].UserID, msgs[1].UserID)
	}
}
