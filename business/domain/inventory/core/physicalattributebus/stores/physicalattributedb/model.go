package physicalattributedb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
)

type physicalAttribute struct {
	AttributeID         uuid.UUID `db:"id"`
	ProductID           uuid.UUID `db:"product_id"`
	Length              float64   `db:"length"`
	Width               float64   `db:"width"`
	Height              float64   `db:"height"`
	Weight              float64   `db:"weight"`
	WeightUnit          string    `db:"weight_unit"`
	Color               string    `db:"color"`
	Size                string    `db:"size"`
	Material            string    `db:"material"`
	StorageRequirements string    `db:"storage_requirements"`
	HazmatClass         string    `db:"hazmat_class"`
	ShelfLifeDays       int       `db:"shelf_life_days"`
	CreatedDate         time.Time `db:"created_date"`
	UpdatedDate         time.Time `db:"updated_date"`
}

func toDBPhysicalAttribute(bus physicalattributebus.PhysicalAttribute) physicalAttribute {
	return physicalAttribute{
		AttributeID:         bus.AttributeID,
		ProductID:           bus.ProductID,
		Length:              bus.Length.Value(),
		Width:               bus.Width.Value(),
		Height:              bus.Height.Value(),
		Weight:              bus.Weight.Value(),
		WeightUnit:          bus.WeightUnit,
		Color:               bus.Color,
		Size:                bus.Size,
		Material:            bus.Material,
		StorageRequirements: bus.StorageRequirements,
		HazmatClass:         bus.HazmatClass,
		ShelfLifeDays:       bus.ShelfLifeDays,
		CreatedDate:         bus.CreatedDate,
		UpdatedDate:         bus.UpdatedDate,
	}
}

func toBusPhysicalAttribute(db physicalAttribute) physicalattributebus.PhysicalAttribute {
	return physicalattributebus.PhysicalAttribute{
		AttributeID:         db.AttributeID,
		ProductID:           db.ProductID,
		Length:              physicalattributebus.NewDimension(db.Length),
		Width:               physicalattributebus.NewDimension(db.Width),
		Height:              physicalattributebus.NewDimension(db.Height),
		Weight:              physicalattributebus.NewDimension(db.Weight),
		WeightUnit:          db.WeightUnit,
		Color:               db.Color,
		Size:                db.Size,
		Material:            db.Material,
		StorageRequirements: db.StorageRequirements,
		HazmatClass:         db.HazmatClass,
		ShelfLifeDays:       db.ShelfLifeDays,
		CreatedDate:         db.CreatedDate.Local(),
		UpdatedDate:         db.UpdatedDate.Local(),
	}
}

func toBusPhysicalAttributes(dbs []physicalAttribute) []physicalattributebus.PhysicalAttribute {
	bus := make([]physicalattributebus.PhysicalAttribute, len(dbs))

	for i, db := range dbs {
		bus[i] = toBusPhysicalAttribute(db)
	}

	return bus
}
