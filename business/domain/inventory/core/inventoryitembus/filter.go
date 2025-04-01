package inventoryitembus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ItemID                *uuid.UUID
	ProductID             *uuid.UUID
	LocationID            *uuid.UUID
	Quantity              *int
	ReservedQuantity      *int
	AllocatedQuantity     *int
	MinimumStock          *int
	MaximumStock          *int
	ReorderPoint          *int
	EconomicOrderQuantity *int
	SafetyStock           *int
	AvgDailyUsage         *int
	CreatedDate           *time.Time
	UpdatedDate           *time.Time
}
