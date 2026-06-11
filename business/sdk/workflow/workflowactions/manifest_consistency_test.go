package workflowactions_test

// Manifest consistency — structural half (no DB).
//
// DESIGN §6: the cascade scheme's soundness reduces to "the runtime fires a delegate for
// exactly the mutations the manifest (GetEntityModifications) declares." This file guards
// the DECLARED side: every EntityModifier handler's manifest must be well-formed, and the
// value-aware extension (P0.2 — ProducedChange) must carry the right value/operator/
// indeterminate flags. The "== fired" half (does the runtime actually emit these events)
// is the integration concern; see the actionhandlers integration test and the knownSilent
// set below, which P4 will convert into asserted firing as it enables cascades.
//
// This test iterates the FULL registry (RegisterAll) so a newly-added EntityModifier is
// covered automatically. Buses are zero-value pointers: GetEntityModifications never
// dereferences its handler's dependencies (it only reads config), the same property
// RegisterCoreActions already relies on by constructing handlers with nil buses.

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// validOperators is the shared trigger/produced-change operator vocabulary.
var validOperators = map[string]bool{
	workflow.OperatorEquals:      true,
	workflow.OperatorNotEquals:   true,
	workflow.OperatorChangedFrom: true,
	workflow.OperatorChangedTo:   true,
	workflow.OperatorGreaterThan: true,
	workflow.OperatorLessThan:    true,
	workflow.OperatorContains:    true,
	workflow.OperatorIn:          true,
}

var validEventTypes = map[string]bool{
	"on_create": true,
	"on_update": true,
	"on_delete": true,
}

// sampleConfigs supplies a representative config for the handlers whose
// GetEntityModifications reads config (it returns nil on an unparseable config).
// Handlers that ignore config are absent and receive nil.
func sampleConfigs() map[string]json.RawMessage {
	return map[string]json.RawMessage{
		"update_field":             json.RawMessage(`{"target_entity":"sales.orders","target_field":"priority","new_value":"high"}`),
		"create_entity":            json.RawMessage(`{"target_entity":"sales.orders","fields":{"priority":"high"}}`),
		"transition_status":        json.RawMessage(`{"target_entity":"sales.orders","status_field":"status","to_status":"shipped"}`),
		"resolve_approval_request": json.RawMessage(`{"approval_request_id":"` + uuid.NewString() + `","resolution":"approved"}`),
	}
}

// buildFullRegistry registers every standard action with zero-value bus dependencies.
func buildFullRegistry(t *testing.T) *workflow.ActionRegistry {
	t.Helper()
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelError, "TEST", func(context.Context) string { return "" })

	reg := workflow.NewActionRegistry()
	workflowactions.RegisterAll(reg, workflowactions.ActionConfig{
		Log: log,
		DB:  nil,
		Buses: workflowactions.BusDependencies{
			// Non-nil so the guarded approve/reject/create handlers register. Zero-value
			// is safe: GetEntityModifications never dereferences these.
			InventoryAdjustment: &inventoryadjustmentbus.Business{},
			TransferOrder:       &transferorderbus.Business{},
			PutAwayTask:         &putawaytaskbus.Business{},
			PurchaseOrder:       &purchaseorderbus.Business{},
		},
	})
	return reg
}

// entityModifiers returns every registered handler implementing EntityModifier.
func entityModifiers(t *testing.T, reg *workflow.ActionRegistry) map[string]workflow.EntityModifier {
	t.Helper()
	out := make(map[string]workflow.EntityModifier)
	for _, at := range reg.GetAll() {
		h, ok := reg.Get(at)
		if !ok {
			continue
		}
		if em, ok := h.(workflow.EntityModifier); ok {
			out[at] = em
		}
	}
	return out
}

// Test_Manifest_WellFormed asserts every EntityModifier produces a structurally valid
// manifest: non-empty entity, valid event type, and every ProducedChange names a field
// the modification also lists in Fields, uses a known operator, and (when indeterminate)
// carries no value.
func Test_Manifest_WellFormed(t *testing.T) {
	reg := buildFullRegistry(t)
	cfgs := sampleConfigs()

	mods := entityModifiers(t, reg)
	if len(mods) < 15 {
		t.Fatalf("expected the full EntityModifier set (>=15), got %d — registry build likely incomplete", len(mods))
	}

	// Read-only handlers implement EntityModifier but modify nothing, so nil is correct.
	readOnly := map[string]bool{"check_inventory": true, "check_reorder_point": true}

	for actionType, em := range mods {
		t.Run(actionType, func(t *testing.T) {
			result := em.GetEntityModifications(cfgs[actionType])
			if readOnly[actionType] {
				if result != nil {
					t.Errorf("%s: read-only handler should declare no modifications, got %v", actionType, result)
				}
				return
			}
			if result == nil {
				t.Fatalf("%s: GetEntityModifications returned nil for its representative config", actionType)
			}
			for i, mod := range result {
				if mod.EntityName == "" {
					t.Errorf("%s mod[%d]: empty EntityName", actionType, i)
				}
				if !validEventTypes[mod.EventType] {
					t.Errorf("%s mod[%d]: invalid EventType %q", actionType, i, mod.EventType)
				}
				fieldSet := make(map[string]bool, len(mod.Fields))
				for _, f := range mod.Fields {
					fieldSet[f] = true
				}
				for j, ch := range mod.Changes {
					if !validOperators[ch.Operator] {
						t.Errorf("%s mod[%d].Changes[%d]: invalid operator %q", actionType, i, j, ch.Operator)
					}
					if ch.FieldName == "" {
						t.Errorf("%s mod[%d].Changes[%d]: empty FieldName", actionType, i, j)
					} else if !fieldSet[ch.FieldName] {
						t.Errorf("%s mod[%d].Changes[%d]: FieldName %q not present in Fields %v", actionType, i, j, ch.FieldName, mod.Fields)
					}
					if ch.Indeterminate && ch.Value != nil {
						t.Errorf("%s mod[%d].Changes[%d]: indeterminate change must not carry a Value (got %v)", actionType, i, j, ch.Value)
					}
				}
			}
		})
	}
}

// Test_Manifest_ValueExtension pins the P0.2 value-aware classification: the enum-const
// approve/reject handlers must declare their fixed produced status, and the config-literal
// handlers must surface a concrete value for a literal config and fall back to indeterminate
// for a templated one.
func Test_Manifest_ValueExtension(t *testing.T) {
	reg := buildFullRegistry(t)
	mods := entityModifiers(t, reg)

	// findChange returns the ProducedChange for (field) across a handler's modifications.
	findChange := func(t *testing.T, actionType, field string, cfg json.RawMessage) workflow.ProducedChange {
		t.Helper()
		em, ok := mods[actionType]
		if !ok {
			t.Fatalf("%s not registered as EntityModifier", actionType)
		}
		for _, mod := range em.GetEntityModifications(cfg) {
			for _, ch := range mod.Changes {
				if ch.FieldName == field {
					return ch
				}
			}
		}
		t.Fatalf("%s: no ProducedChange for field %q", actionType, field)
		return workflow.ProducedChange{}
	}

	// Enum-const: value baked into the handler/bus, no config dependency.
	enumCases := []struct {
		actionType string
		field      string
		want       string
	}{
		{"approve_inventory_adjustment", "approval_status", inventoryadjustmentbus.ApprovalStatusApproved},
		{"reject_inventory_adjustment", "approval_status", inventoryadjustmentbus.ApprovalStatusRejected},
		{"approve_transfer_order", "status", transferorderbus.StatusApproved},
		{"reject_transfer_order", "status", transferorderbus.StatusRejected},
	}
	for _, tc := range enumCases {
		t.Run("enum/"+tc.actionType, func(t *testing.T) {
			ch := findChange(t, tc.actionType, tc.field, nil)
			if ch.Indeterminate {
				t.Fatalf("%s: %s should be statically known, got indeterminate", tc.actionType, tc.field)
			}
			if ch.Operator != workflow.OperatorChangedTo {
				t.Errorf("%s: operator = %q, want changed_to", tc.actionType, ch.Operator)
			}
			if ch.Value != tc.want {
				t.Errorf("%s: value = %v, want %q", tc.actionType, ch.Value, tc.want)
			}
		})
	}

	// Config-literal: concrete value for a literal config, indeterminate when templated.
	literalCases := []struct {
		actionType  string
		field       string
		literalCfg  json.RawMessage
		wantValue   string
		templateCfg json.RawMessage
	}{
		{
			"transition_status", "status",
			json.RawMessage(`{"target_entity":"sales.orders","status_field":"status","to_status":"shipped"}`),
			"shipped",
			json.RawMessage(`{"target_entity":"sales.orders","status_field":"status","to_status":"{{trigger.status}}"}`),
		},
		{
			"update_field", "priority",
			json.RawMessage(`{"target_entity":"sales.orders","target_field":"priority","new_value":"high"}`),
			"high",
			json.RawMessage(`{"target_entity":"sales.orders","target_field":"priority","new_value":"{{trigger.priority}}"}`),
		},
		{
			"resolve_approval_request", "status",
			json.RawMessage(`{"approval_request_id":"` + uuid.NewString() + `","resolution":"approved"}`),
			"approved",
			json.RawMessage(`{"approval_request_id":"` + uuid.NewString() + `","resolution":"{{trigger.resolution}}"}`),
		},
	}
	for _, tc := range literalCases {
		t.Run("literal/"+tc.actionType, func(t *testing.T) {
			ch := findChange(t, tc.actionType, tc.field, tc.literalCfg)
			if ch.Indeterminate {
				t.Fatalf("%s: literal config should yield a known value, got indeterminate", tc.actionType)
			}
			if ch.Value != tc.wantValue {
				t.Errorf("%s: value = %v, want %q", tc.actionType, ch.Value, tc.wantValue)
			}
		})
		t.Run("templated/"+tc.actionType, func(t *testing.T) {
			ch := findChange(t, tc.actionType, tc.field, tc.templateCfg)
			if !ch.Indeterminate {
				t.Errorf("%s: templated config should be indeterminate, got value %v", tc.actionType, ch.Value)
			}
			if ch.Value != nil {
				t.Errorf("%s: indeterminate change carried Value %v", tc.actionType, ch.Value)
			}
		})
	}
}
