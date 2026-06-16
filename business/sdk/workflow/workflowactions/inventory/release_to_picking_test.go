package inventory_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
)

func TestReleaseToPicking_Validate(t *testing.T) {
	handler := inventory.NewReleaseToPickingHandler(nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name      string
		orderID   string
		raw       json.RawMessage // overrides orderID-built config when non-nil
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "missing order_id",
			orderID:   "",
			wantErr:   true,
			errSubstr: "order_id is required",
		},
		{
			name:    "templated entity_id accepted",
			orderID: "{{entity_id}}",
			wantErr: false,
		},
		{
			name:      "bad static uuid rejected",
			orderID:   "not-a-uuid",
			wantErr:   true,
			errSubstr: "invalid order_id",
		},
		{
			name:    "good static uuid accepted",
			orderID: uuid.New().String(),
			wantErr: false,
		},
		{
			name:      "invalid json",
			raw:       json.RawMessage(`{invalid`),
			wantErr:   true,
			errSubstr: "invalid config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.raw
			if config == nil {
				data, _ := json.Marshal(inventory.ReleaseToPickingConfig{OrderID: tt.orderID})
				config = data
			}

			err := handler.Validate(config)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

func TestReleaseToPicking_Metadata(t *testing.T) {
	handler := inventory.NewReleaseToPickingHandler(nil, nil, nil, nil, nil, nil, nil)

	t.Run("GetType", func(t *testing.T) {
		if got := handler.GetType(); got != "release_to_picking" {
			t.Fatalf("expected release_to_picking, got %s", got)
		}
	})

	t.Run("SupportsManualExecution", func(t *testing.T) {
		if !handler.SupportsManualExecution() {
			t.Fatal("expected true")
		}
	})

	t.Run("IsAsync", func(t *testing.T) {
		if handler.IsAsync() {
			t.Fatal("expected false")
		}
	})

	t.Run("GetOutputPorts has single default named released", func(t *testing.T) {
		ports := handler.GetOutputPorts()

		var defaults []workflow.OutputPort
		for _, p := range ports {
			if p.IsDefault {
				defaults = append(defaults, p)
			}
		}
		if len(defaults) != 1 {
			t.Fatalf("expected exactly 1 default port, got %d", len(defaults))
		}
		if defaults[0].Name != "released" {
			t.Fatalf("expected default port named 'released', got %q", defaults[0].Name)
		}
	})

	t.Run("GetEntityModifications", func(t *testing.T) {
		mods := handler.GetEntityModifications(nil)
		if len(mods) != 2 {
			t.Fatalf("expected 2 entity modifications, got %d", len(mods))
		}

		byEntity := make(map[string]workflow.EntityModification)
		for _, m := range mods {
			byEntity[m.EntityName] = m
		}

		ordersMod, ok := byEntity["sales.orders"]
		if !ok {
			t.Fatal("missing sales.orders modification")
		}
		if ordersMod.EventType != "on_update" {
			t.Fatalf("sales.orders: expected on_update, got %q", ordersMod.EventType)
		}
		if len(ordersMod.Changes) != 1 {
			t.Fatalf("sales.orders: expected 1 produced change, got %d", len(ordersMod.Changes))
		}
		ch := ordersMod.Changes[0]
		if ch.FieldName != "order_fulfillment_status_id" {
			t.Fatalf("sales.orders: expected change on order_fulfillment_status_id, got %q", ch.FieldName)
		}
		if !ch.Indeterminate {
			t.Fatal("sales.orders: expected produced change to be Indeterminate")
		}

		pickMod, ok := byEntity["inventory.pick_tasks"]
		if !ok {
			t.Fatal("missing inventory.pick_tasks modification")
		}
		if pickMod.EventType != "on_create" {
			t.Fatalf("inventory.pick_tasks: expected on_create, got %q", pickMod.EventType)
		}
	})
}
