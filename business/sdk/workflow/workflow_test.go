package workflow_test

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// mockStorer is a minimal mock that only implements QueryAllTemplates and
// QueryActiveTemplates. All other Storer methods panic if called.
type mockStorer struct {
	workflow.Storer
	allTemplates    []workflow.ActionTemplate
	activeTemplates []workflow.ActionTemplate
	allErr          error
	activeErr       error
}

func (m *mockStorer) QueryAllTemplates(_ context.Context) ([]workflow.ActionTemplate, error) {
	return m.allTemplates, m.allErr
}

func (m *mockStorer) QueryActiveTemplates(_ context.Context) ([]workflow.ActionTemplate, error) {
	return m.activeTemplates, m.activeErr
}

func TestQueryAllTemplates(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	userID := uuid.New()
	config := json.RawMessage(`{"key":"value"}`)

	templates := []workflow.ActionTemplate{
		{
			ID:            uuid.New(),
			Name:          "template_a",
			Description:   "First template",
			ActionType:    "webhook",
			Icon:          "icon-a",
			DefaultConfig: config,
			CreatedDate:   now,
			CreatedBy:     userID,
			IsActive:      true,
		},
		{
			ID:            uuid.New(),
			Name:          "template_b",
			Description:   "Second template",
			ActionType:    "email",
			Icon:          "icon-b",
			DefaultConfig: config,
			CreatedDate:   now,
			CreatedBy:     userID,
			IsActive:      false,
		},
	}

	log := logger.New(io.Discard, logger.LevelInfo, "test", func(context.Context) string { return "" })

	mock := &mockStorer{
		allTemplates:    templates,
		activeTemplates: templates[:1],
	}

	bus := workflow.NewBusiness(log, nil, mock)

	t.Run("QueryAllTemplates returns all templates", func(t *testing.T) {
		got, err := bus.QueryAllTemplates(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if diff := cmp.Diff(got, templates); diff != "" {
			t.Fatalf("mismatch (-got +want):\n%s", diff)
		}
	})

	t.Run("QueryActiveTemplates returns only active templates", func(t *testing.T) {
		got, err := bus.QueryActiveTemplates(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if diff := cmp.Diff(got, templates[:1]); diff != "" {
			t.Fatalf("mismatch (-got +want):\n%s", diff)
		}
	})
}
