package costhistorybus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/types"
)

type CostHistory struct {
	CostHistoryID uuid.UUID
	ProductID     uuid.UUID
	CostType      string
	Amount        types.Money
	Currency      string
	EffectiveDate time.Time
	EndDate       time.Time
	CreatedDate   time.Time
	UpdatedDate   time.Time
}

type NewCostHistory struct {
	ProductID     uuid.UUID
	CostType      string
	Amount        types.Money
	Currency      string
	EffectiveDate time.Time
	EndDate       time.Time
}

type UpdateCostHistory struct {
	ProductID     *uuid.UUID
	CostType      *string
	Amount        *types.Money
	Currency      *string
	EffectiveDate *time.Time
	EndDate       *time.Time
}
