package approvalrequestbus_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// =============================================================================
// Mock storer for unit-level delegate test (no DB required)

type mockApprovalStorer struct {
	resolveResult approvalrequestbus.ApprovalRequest
}

func (m *mockApprovalStorer) NewWithTx(_ sqldb.CommitRollbacker) (approvalrequestbus.Storer, error) {
	return m, nil
}
func (m *mockApprovalStorer) Create(_ context.Context, _ approvalrequestbus.ApprovalRequest) error {
	return nil
}
func (m *mockApprovalStorer) QueryByID(_ context.Context, id uuid.UUID) (approvalrequestbus.ApprovalRequest, error) {
	return approvalrequestbus.ApprovalRequest{ID: id}, nil
}
func (m *mockApprovalStorer) Resolve(_ context.Context, id, _ uuid.UUID, status, reason string) (approvalrequestbus.ApprovalRequest, error) {
	result := m.resolveResult
	result.ID = id
	result.Status = status
	result.ResolutionReason = reason
	return result, nil
}
func (m *mockApprovalStorer) Query(_ context.Context, _ approvalrequestbus.QueryFilter, _ order.By, _ page.Page) ([]approvalrequestbus.ApprovalRequest, error) {
	return nil, nil
}
func (m *mockApprovalStorer) Count(_ context.Context, _ approvalrequestbus.QueryFilter) (int, error) {
	return 0, nil
}
func (m *mockApprovalStorer) IsApprover(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockApprovalStorer) ClearTaskToken(_ context.Context, _ uuid.UUID) error {
	return nil
}

// =============================================================================

func Test_Resolve_FiresDelegateEvent(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "" })
	del := delegate.New(log)

	var capturedDomain, capturedAction string
	del.Register(approvalrequestbus.DomainName, approvalrequestbus.ActionUpdated, func(_ context.Context, data delegate.Data) error {
		capturedDomain = data.Domain
		capturedAction = data.Action
		return nil
	})

	now := time.Now()
	resolvedBy := uuid.New()
	storer := &mockApprovalStorer{
		resolveResult: approvalrequestbus.ApprovalRequest{
			ResolvedBy:   &resolvedBy,
			ResolvedDate: &now,
		},
	}

	bus := approvalrequestbus.NewBusiness(log, del, storer)

	id := uuid.New()
	_, err := bus.Resolve(context.Background(), id, resolvedBy, approvalrequestbus.StatusApproved, "approved via test")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if capturedDomain != approvalrequestbus.DomainName {
		t.Errorf("expected delegate domain %q, got %q", approvalrequestbus.DomainName, capturedDomain)
	}
	if capturedAction != approvalrequestbus.ActionUpdated {
		t.Errorf("expected delegate action %q, got %q", approvalrequestbus.ActionUpdated, capturedAction)
	}
}
