package ordersbus

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// BindingDomainName represents the delegate domain for OrderContainerBinding
// events. The binding is its own entity (separate table, separate ID,
// independent lifecycle), so it gets its own domain name distinct from
// DomainName ("order").
const BindingDomainName = "order_container_binding"

// BindingEntityName is the workflow entity name used for event matching.
// Matches the inventory.order_container_bindings table name.
const BindingEntityName = "order_container_bindings"

// Binding-specific delegate action constants.
const (
	ActionBindingCreated = "created"
	ActionBindingUpdated = "updated"
)

// =============================================================================
// Binding Created Event (fired by BindContainer)
// =============================================================================

type ActionBindingCreatedParms struct {
	Entity OrderContainerBinding `json:"entity"`
}

func (p *ActionBindingCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionBindingCreatedData constructs delegate data for binding creation
// events. Bindings have no UserID column (they are system-issued from
// scan/dispatch flows), so the payload omits UserID.
func ActionBindingCreatedData(binding OrderContainerBinding) delegate.Data {
	params := ActionBindingCreatedParms{Entity: binding}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    BindingDomainName,
		Action:    ActionBindingCreated,
		RawParams: rawParams,
	}
}

// =============================================================================
// Binding Updated Event (fired by UnbindContainer on state transition)
// =============================================================================

type ActionBindingUpdatedParms struct {
	Entity       OrderContainerBinding `json:"entity"`
	BeforeEntity OrderContainerBinding `json:"beforeEntity,omitempty"`
}

func (p *ActionBindingUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// ActionBindingUpdatedData constructs delegate data for binding state
// transitions (currently only Unbind sets unbound_at). Mirrors
// approvalrequestbus.Resolve's pattern of passing an ID-only "before"
// when the caller does not need the prior row.
func ActionBindingUpdatedData(before, after OrderContainerBinding) delegate.Data {
	params := ActionBindingUpdatedParms{
		Entity:       after,
		BeforeEntity: before,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    BindingDomainName,
		Action:    ActionBindingUpdated,
		RawParams: rawParams,
	}
}
