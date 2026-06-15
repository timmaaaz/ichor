package inventory_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
)

func TestClaimTransferOrder_Validate(t *testing.T) {
	handler := inventory.NewClaimTransferOrderHandler(nil, nil)

	tests := []struct {
		name      string
		id        string
		raw       json.RawMessage
		wantErr   bool
		errSubstr string
	}{
		{name: "missing id", id: "", wantErr: true, errSubstr: "transfer_order_id is required"},
		{name: "bad uuid", id: "nope", wantErr: true, errSubstr: "invalid transfer_order_id"},
		{name: "good uuid", id: uuid.New().String(), wantErr: false},
		{name: "invalid json", raw: json.RawMessage(`{bad`), wantErr: true, errSubstr: "invalid config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.raw
			if config == nil {
				data, _ := json.Marshal(inventory.ClaimTransferOrderConfig{TransferOrderID: tt.id})
				config = data
			}
			err := handler.Validate(config)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.errSubstr)
			}
		})
	}
}

func TestClaimTransferOrder_Metadata(t *testing.T) {
	handler := inventory.NewClaimTransferOrderHandler(nil, nil)

	if got := handler.GetType(); got != "claim_transfer_order" {
		t.Fatalf("expected claim_transfer_order, got %s", got)
	}
	if !handler.SupportsManualExecution() {
		t.Fatal("expected SupportsManualExecution true")
	}
	if handler.IsAsync() {
		t.Fatal("expected IsAsync false")
	}

	var defaults []workflow.OutputPort
	for _, p := range handler.GetOutputPorts() {
		if p.IsDefault {
			defaults = append(defaults, p)
		}
	}
	if len(defaults) != 1 || defaults[0].Name != "claimed" {
		t.Fatalf("expected single default port 'claimed', got %+v", defaults)
	}

	mods := handler.GetEntityModifications(nil)
	if len(mods) != 1 {
		t.Fatalf("expected 1 entity modification, got %d", len(mods))
	}
	if mods[0].EntityName != "inventory.transfer_orders" || mods[0].EventType != "on_update" {
		t.Fatalf("unexpected modification target: %+v", mods[0])
	}
	if len(mods[0].Changes) != 1 || mods[0].Changes[0].Value != transferorderbus.StatusInTransit {
		t.Fatalf("expected status changed_to in_transit, got %+v", mods[0].Changes)
	}
}
