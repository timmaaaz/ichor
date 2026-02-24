package pickingapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
)

// PickQuantityRequest is the request body for the pick-quantity endpoint.
type PickQuantityRequest struct {
	Quantity   string `json:"quantity"   validate:"required,numeric"`
	PickedBy   string `json:"picked_by"  validate:"required,uuid4"`
	LocationID string `json:"location_id" validate:"required,uuid4"`
}

func (r *PickQuantityRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r PickQuantityRequest) Validate() error {
	if err := errs.Check(r); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// PickQuantityResponse wraps the updated line item.
type PickQuantityResponse = orderlineitemsapp.OrderLineItem

// ShortPickRequest is the request body for the short-pick endpoint.
// LocationID is optional for the "backorder" type (no inventory is touched),
// but required for "partial", "substitute", and "skip" types.
type ShortPickRequest struct {
	PickedQuantity       string  `json:"picked_quantity"        validate:"required,numeric"`
	ShortPickType        string  `json:"short_pick_type"        validate:"required,oneof=partial backorder substitute skip"`
	ShortPickReason      string  `json:"short_pick_reason"      validate:"omitempty"`
	PickedBy             string  `json:"picked_by"              validate:"required,uuid4"`
	LocationID           string  `json:"location_id"            validate:"omitempty,uuid4"`
	SubstituteProductID  *string `json:"substitute_product_id"  validate:"omitempty,uuid4"`
	SubstituteQuantity   *string `json:"substitute_quantity"    validate:"omitempty,numeric"`
}

func (r *ShortPickRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r ShortPickRequest) Validate() error {
	if err := errs.Check(r); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// ShortPickResponse wraps the updated line item.
type ShortPickResponse = orderlineitemsapp.OrderLineItem

// CompletePackingRequest is the request body for the complete-packing endpoint.
type CompletePackingRequest struct {
	PackedBy string `json:"packed_by" validate:"required,uuid4"`
}

func (r *CompletePackingRequest) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r CompletePackingRequest) Validate() error {
	if err := errs.Check(r); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// CompletePackingResponse wraps the updated order.
type CompletePackingResponse = ordersapp.Order
