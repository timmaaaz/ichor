package supplierbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Supplier struct {
	SupplierID     uuid.UUID          `json:"supplier_id"`
	ContactInfosID uuid.UUID          `json:"contact_infos_id"`
	Name           string             `json:"name"`
	PaymentTermID  *uuid.UUID         `json:"payment_term_id,omitempty"`
	LeadTimeDays   int                `json:"lead_time_days"`
	Rating         types.RoundedFloat `json:"rating"`
	IsActive       bool               `json:"is_active"`
	CreatedDate    time.Time          `json:"created_date"`
	UpdatedDate    time.Time          `json:"updated_date"`
}

type NewSupplier struct {
	ContactInfosID uuid.UUID          `json:"contact_infos_id"`
	Name           string             `json:"name"`
	PaymentTermID  *uuid.UUID         `json:"payment_term_id,omitempty"`
	LeadTimeDays   int                `json:"lead_time_days"`
	Rating         types.RoundedFloat `json:"rating"`
	IsActive       bool               `json:"is_active"`
}

type UpdateSupplier struct {
	ContactInfosID *uuid.UUID          `json:"contact_infos_id,omitempty"`
	Name           *string             `json:"name,omitempty"`
	PaymentTermID  *uuid.UUID          `json:"payment_term_id,omitempty"`
	LeadTimeDays   *int                `json:"lead_time_days,omitempty"`
	Rating         *types.RoundedFloat `json:"rating,omitempty"`
	IsActive       *bool               `json:"is_active,omitempty"`
}
