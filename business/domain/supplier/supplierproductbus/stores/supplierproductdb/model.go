package supplierproductdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierproductbus/types"
)

type supplierProduct struct {
	SupplierProductID  uuid.UUID      `db:"supplier_product_id"`
	SupplierID         uuid.UUID      `db:"supplier_id"`
	ProductID          uuid.UUID      `db:"product_id"`
	SupplierPartNumber string         `db:"supplier_part_number"`
	MinOrderQuantity   int            `db:"min_order_quantity"`
	MaxOrderQuantity   int            `db:"max_order_quantity"`
	LeadTimeDays       int            `db:"lead_time_days"`
	UnitCost           sql.NullString `db:"unit_cost"`
	IsPrimarySupplier  bool           `db:"is_primary_supplier"`
	CreatedDate        time.Time      `db:"created_date"`
	UpdatedDate        time.Time      `db:"updated_date"`
}

func toDBSupplierProduct(bus supplierproductbus.SupplierProduct) supplierProduct {
	return supplierProduct{
		SupplierProductID:  bus.SupplierProductID,
		SupplierID:         bus.SupplierID,
		ProductID:          bus.ProductID,
		SupplierPartNumber: bus.SupplierPartNumber,
		MinOrderQuantity:   bus.MinOrderQuantity,
		MaxOrderQuantity:   bus.MaxOrderQuantity,
		LeadTimeDays:       bus.LeadTimeDays,
		UnitCost:           bus.UnitCost.DBValue(),
		IsPrimarySupplier:  bus.IsPrimarySupplier,
		CreatedDate:        bus.CreatedDate.UTC(),
		UpdatedDate:        bus.UpdatedDate.UTC(),
	}
}

func toBusSupplierProduct(db supplierProduct) (supplierproductbus.SupplierProduct, error) {
	unitCost, err := types.ParseMoney(db.UnitCost.String)
	if err != nil {
		return supplierproductbus.SupplierProduct{}, err
	}

	return supplierproductbus.SupplierProduct{
		SupplierProductID:  db.SupplierProductID,
		SupplierID:         db.SupplierID,
		ProductID:          db.ProductID,
		SupplierPartNumber: db.SupplierPartNumber,
		MinOrderQuantity:   db.MinOrderQuantity,
		MaxOrderQuantity:   db.MaxOrderQuantity,
		LeadTimeDays:       db.LeadTimeDays,
		UnitCost:           unitCost,
		IsPrimarySupplier:  db.IsPrimarySupplier,
		CreatedDate:        db.CreatedDate.UTC(),
		UpdatedDate:        db.UpdatedDate.UTC(),
	}, nil
}

func toBusSupplierProducts(db []supplierProduct) ([]supplierproductbus.SupplierProduct, error) {
	busSupplierProducts := make([]supplierproductbus.SupplierProduct, len(db))

	for i, dbSupplierProduct := range db {
		busSupplierProduct, err := toBusSupplierProduct(dbSupplierProduct)
		if err != nil {
			return nil, fmt.Errorf("tobussupplierproducts %v", err)
		}
		busSupplierProducts[i] = busSupplierProduct
	}

	return busSupplierProducts, nil
}
