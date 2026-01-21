package currencydb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
)

type currency struct {
	ID            uuid.UUID  `db:"id"`
	Code          string     `db:"code"`
	Name          string     `db:"name"`
	Symbol        string     `db:"symbol"`
	Locale        string     `db:"locale"`
	DecimalPlaces int        `db:"decimal_places"`
	IsActive      bool       `db:"is_active"`
	SortOrder     int        `db:"sort_order"`
	CreatedBy     *uuid.UUID `db:"created_by"`
	CreatedDate   time.Time  `db:"created_date"`
	UpdatedBy     *uuid.UUID `db:"updated_by"`
	UpdatedDate   time.Time  `db:"updated_date"`
}

func toDBCurrency(bus currencybus.Currency) currency {
	return currency{
		ID:            bus.ID,
		Code:          bus.Code,
		Name:          bus.Name,
		Symbol:        bus.Symbol,
		Locale:        bus.Locale,
		DecimalPlaces: bus.DecimalPlaces,
		IsActive:      bus.IsActive,
		SortOrder:     bus.SortOrder,
		CreatedBy:     bus.CreatedBy,
		CreatedDate:   bus.CreatedDate,
		UpdatedBy:     bus.UpdatedBy,
		UpdatedDate:   bus.UpdatedDate,
	}
}

func toBusCurrency(db currency) currencybus.Currency {
	return currencybus.Currency{
		ID:            db.ID,
		Code:          db.Code,
		Name:          db.Name,
		Symbol:        db.Symbol,
		Locale:        db.Locale,
		DecimalPlaces: db.DecimalPlaces,
		IsActive:      db.IsActive,
		SortOrder:     db.SortOrder,
		CreatedBy:     db.CreatedBy,
		CreatedDate:   db.CreatedDate,
		UpdatedBy:     db.UpdatedBy,
		UpdatedDate:   db.UpdatedDate,
	}
}

func toBusCurrencies(dbs []currency) []currencybus.Currency {
	currencies := make([]currencybus.Currency, len(dbs))
	for i, db := range dbs {
		currencies[i] = toBusCurrency(db)
	}
	return currencies
}
