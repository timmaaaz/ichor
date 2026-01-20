package costhistorydb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/types"
)

type costHistory struct {
	CostHistoryID uuid.UUID      `db:"id"`
	ProductID     uuid.UUID      `db:"product_id"`
	CostType      string         `db:"cost_type"`
	Amount        sql.NullString `db:"amount"`
	CurrencyID    uuid.UUID      `db:"currency_id"`
	EffectiveDate time.Time      `db:"effective_date"`
	EndDate       time.Time      `db:"end_date"`
	CreatedDate   time.Time      `db:"created_date"`
	UpdatedDate   time.Time      `db:"updated_date"`
}

func toDBCostHistory(bus costhistorybus.CostHistory) costHistory {
	return costHistory{
		CostHistoryID: bus.CostHistoryID,
		ProductID:     bus.ProductID,
		CostType:      bus.CostType,
		Amount:        bus.Amount.DBValue(),
		CurrencyID:    bus.CurrencyID,
		EffectiveDate: bus.EffectiveDate.UTC(),
		EndDate:       bus.EndDate.UTC(),
		CreatedDate:   bus.CreatedDate.UTC(),
		UpdatedDate:   bus.UpdatedDate.UTC(),
	}
}

func toBusCostHistory(db costHistory) (costhistorybus.CostHistory, error) {
	amt, err := types.ParseMoney(db.Amount.String)
	if err != nil {
		return costhistorybus.CostHistory{}, fmt.Errorf("tobuscosthistory: %v", err)
	}

	return costhistorybus.CostHistory{
		CostHistoryID: db.CostHistoryID,
		ProductID:     db.ProductID,
		CostType:      db.CostType,
		Amount:        amt,
		CurrencyID:    db.CurrencyID,
		EffectiveDate: db.EffectiveDate.Local(),
		EndDate:       db.EndDate.Local(),
		CreatedDate:   db.CreatedDate.Local(),
		UpdatedDate:   db.UpdatedDate.Local(),
	}, nil
}

func toBusCostHistories(db []costHistory) ([]costhistorybus.CostHistory, error) {
	bus := make([]costhistorybus.CostHistory, len(db))

	for i, dbCostHistory := range db {
		busCostHistory, err := toBusCostHistory(dbCostHistory)
		if err != nil {
			return nil, fmt.Errorf("tobuscosthistories: %v", err)
		}
		bus[i] = busCostHistory
	}

	return bus, nil
}
