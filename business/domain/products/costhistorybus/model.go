package costhistorybus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type CostHistory struct {
	CostHistoryID uuid.UUID   `json:"cost_history_id"`
	ProductID     uuid.UUID   `json:"product_id"`
	CostType      string      `json:"cost_type"`
	Amount        types.Money `json:"amount"`
	Currency      string      `json:"currency"`
	EffectiveDate time.Time   `json:"effective_date"`
	EndDate       time.Time   `json:"end_date"`
	CreatedDate   time.Time   `json:"created_date"`
	UpdatedDate   time.Time   `json:"updated_date"`
}

type NewCostHistory struct {
	ProductID     uuid.UUID   `json:"product_id"`
	CostType      string      `json:"cost_type"`
	Amount        types.Money `json:"amount"`
	Currency      string      `json:"currency"`
	EffectiveDate time.Time   `json:"effective_date"`
	EndDate       time.Time   `json:"end_date"`
	CreatedDate   *time.Time  `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateCostHistory struct {
	ProductID     *uuid.UUID   `json:"product_id,omitempty"`
	CostType      *string      `json:"cost_type,omitempty"`
	Amount        *types.Money `json:"amount,omitempty"`
	Currency      *string      `json:"currency,omitempty"`
	EffectiveDate *time.Time   `json:"effective_date,omitempty"`
	EndDate       *time.Time   `json:"end_date,omitempty"`
}
