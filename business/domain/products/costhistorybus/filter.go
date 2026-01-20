package costhistorybus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/types"
)

type QueryFilter struct {
	CostHistoryID *uuid.UUID
	ProductID     *uuid.UUID
	CostType      *string
	Amount        *types.Money
	CurrencyID    *uuid.UUID
	EffectiveDate *time.Time
	EndDate       *time.Time
	CreatedDate   *time.Time
	UpdatedDate   *time.Time
}
