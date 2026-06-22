package workflow

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// fakeRuleActionStorer embeds Storer (so the ~28 unused methods are inherited
// as nil implementations) and records whether the write methods were reached.
// It lets us prove the guard rejects an unexecutable action BEFORE the store is
// ever touched, with no database.
type fakeRuleActionStorer struct {
	Storer
	createCalled bool
	updateCalled bool
}

func (f *fakeRuleActionStorer) CreateRuleAction(ctx context.Context, action RuleAction) error {
	f.createCalled = true
	return nil
}

func (f *fakeRuleActionStorer) UpdateRuleAction(ctx context.Context, action RuleAction) error {
	f.updateCalled = true
	return nil
}

func Test_CreateRuleAction_rejectsUnexecutable(t *testing.T) {
	t.Run("no template and no inline action_type is rejected before the store", func(t *testing.T) {
		fs := &fakeRuleActionStorer{}
		b := NewBusiness(nil, nil, fs)

		_, err := b.CreateRuleAction(context.Background(), NewRuleAction{
			AutomationRuleID: uuid.New(),
			Name:             "broken",
			ActionConfig:     json.RawMessage(`{"to":"x@y.z"}`),
			TemplateID:       nil,
		})

		if err == nil {
			t.Fatal("expected rejection for template-less, type-less action")
		}
		if fs.createCalled {
			t.Error("store must not be called when validation fails")
		}
	})

	t.Run("inline action_type is accepted", func(t *testing.T) {
		fs := &fakeRuleActionStorer{}
		b := NewBusiness(nil, nil, fs)

		_, err := b.CreateRuleAction(context.Background(), NewRuleAction{
			AutomationRuleID: uuid.New(),
			Name:             "ok",
			ActionConfig:     json.RawMessage(`{"action_type":"send_email"}`),
			TemplateID:       nil,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fs.createCalled {
			t.Error("store should be called for a valid action")
		}
	})
}

func Test_UpdateRuleAction_rejectsUnexecutable(t *testing.T) {
	t.Run("overwriting config to strip the inline type on a template-less action is rejected", func(t *testing.T) {
		fs := &fakeRuleActionStorer{}
		b := NewBusiness(nil, nil, fs)

		existing := RuleAction{
			ID:           uuid.New(),
			Name:         "welcome email",
			TemplateID:   nil,
			ActionConfig: json.RawMessage(`{"action_type":"send_email"}`),
		}
		stripped := json.RawMessage(`{"to":"x@y.z"}`)

		_, err := b.UpdateRuleAction(context.Background(), existing, UpdateRuleAction{ActionConfig: &stripped})

		if err == nil {
			t.Fatal("expected rejection: update strips inline action_type and leaves no template")
		}
		if fs.updateCalled {
			t.Error("store must not be called when the merged state is unexecutable")
		}
	})

	t.Run("updating an unrelated field on a valid action is accepted", func(t *testing.T) {
		fs := &fakeRuleActionStorer{}
		b := NewBusiness(nil, nil, fs)

		existing := RuleAction{
			ID:           uuid.New(),
			Name:         "welcome email",
			TemplateID:   nil,
			ActionConfig: json.RawMessage(`{"action_type":"send_email"}`),
		}
		active := false

		_, err := b.UpdateRuleAction(context.Background(), existing, UpdateRuleAction{IsActive: &active})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fs.updateCalled {
			t.Error("store should be called for a valid update")
		}
	})
}
