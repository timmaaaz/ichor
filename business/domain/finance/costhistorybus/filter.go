package costhistorybus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus/types"
)

type QueryFilter struct {
	CostHistoryID *uuid.UUID
	ProductID     *uuid.UUID
	CostType      *string
	Amount        *types.Money
	Currency      *string
	EffectiveDate *time.Time
	EndDate       *time.Time
	CreatedDate   *time.Time
	UpdatedDate   *time.Time
}
