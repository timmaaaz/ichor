package validassetdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus/types"
)

type validAsset struct {
	ID                  uuid.UUID      `db:"valid_asset_id"`
	TypeID              uuid.UUID      `db:"type_id"`
	Name                string         `db:"name"`
	EstPrice            sql.NullString `db:"est_price"`
	Price               sql.NullString `db:"price"`
	MaintenanceInterval sql.NullString `db:"maintenance_interval"`
	LifeExpectancy      sql.NullString `db:"life_expectancy"`
	SerialNumber        string         `db:"serial_number"`
	ModelNumber         string         `db:"model_number"`
	IsEnabled           bool           `db:"is_enabled"`
	DateCreated         time.Time      `db:"date_created"`
	DateUpdated         time.Time      `db:"date_updated"`
	CreatedBy           uuid.UUID      `db:"created_by"`
	UpdatedBy           uuid.UUID      `db:"updated_by"`
}

func toDBAsset(bus validassetbus.ValidAsset) validAsset {
	return validAsset{
		ID:                  bus.ID,
		TypeID:              bus.TypeID,
		Name:                bus.Name,
		EstPrice:            bus.EstPrice.DBValue(),
		Price:               bus.Price.DBValue(),
		MaintenanceInterval: bus.MaintenanceInterval.DBValue(),
		LifeExpectancy:      bus.LifeExpectancy.DBValue(),
		SerialNumber:        bus.SerialNumber,
		ModelNumber:         bus.ModelNumber,
		IsEnabled:           bus.IsEnabled,
		DateCreated:         bus.DateCreated.UTC(),
		DateUpdated:         bus.DateUpdated.UTC(),
		CreatedBy:           bus.CreatedBy,
		UpdatedBy:           bus.UpdatedBy,
	}
}

func toBusAsset(db validAsset) (validassetbus.ValidAsset, error) {
	estPrice, err := types.ParseMoney(db.EstPrice.String)
	if err != nil {
		return validassetbus.ValidAsset{}, fmt.Errorf("tobusasset: %w", err)
	}

	price, err := types.ParseMoney(db.Price.String)
	if err != nil {
		return validassetbus.ValidAsset{}, fmt.Errorf("tobusasset: %w", err)
	}

	maintenanceInterval, err := types.ParseInterval(db.MaintenanceInterval.String)
	if err != nil {
		return validassetbus.ValidAsset{}, fmt.Errorf("tobusasset: %w", err)
	}

	lifeExpectancy, err := types.ParseInterval(db.LifeExpectancy.String)
	if err != nil {
		return validassetbus.ValidAsset{}, fmt.Errorf("tobusasset: %w", err)
	}

	return validassetbus.ValidAsset{
		ID:                  db.ID,
		TypeID:              db.TypeID,
		Name:                db.Name,
		EstPrice:            estPrice,
		Price:               price,
		MaintenanceInterval: maintenanceInterval,
		LifeExpectancy:      lifeExpectancy,
		SerialNumber:        db.SerialNumber,
		ModelNumber:         db.ModelNumber,
		IsEnabled:           db.IsEnabled,
		DateCreated:         db.DateCreated.In(time.Local),
		DateUpdated:         db.DateUpdated.In(time.Local),
		CreatedBy:           db.CreatedBy,
		UpdatedBy:           db.UpdatedBy,
	}, nil
}

func toBusAssets(assets []validAsset) ([]validassetbus.ValidAsset, error) {
	busAssets := make([]validassetbus.ValidAsset, len(assets))
	for i, a := range assets {
		busAsset, err := toBusAsset(a)
		if err != nil {
			return nil, fmt.Errorf("tobusassets: %w", err)
		}
		busAssets[i] = busAsset
	}

	return busAssets, nil
}
