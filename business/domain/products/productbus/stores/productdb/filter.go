package productdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/products/productbus"
)

func applyFilter(filter productbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.SKU != nil {
		data["sku"] = *filter.SKU
		wc = append(wc, "sku = :sku")
	}

	if filter.BrandID != nil {
		data["brand_id"] = *filter.BrandID
		wc = append(wc, "brand_id = :brand_id")
	}

	if filter.ProductCategoryID != nil {
		data["product_category_id"] = *filter.ProductCategoryID
		wc = append(wc, "product_category_id = :product_category_id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name LIKE :name")
	}

	if filter.Description != nil {
		data["description"] = "%" + *filter.Description + "%"
		wc = append(wc, "description LIKE :description")
	}

	if filter.ModelNumber != nil {
		data["model_number"] = *filter.ModelNumber
		wc = append(wc, "model_number = :model_number")
	}

	if filter.UpcCode != nil {
		data["upc_code"] = *filter.UpcCode
		wc = append(wc, "upc_code = :upc_code")
	}

	if filter.Status != nil {
		data["status"] = *filter.Status
		wc = append(wc, "status = :status")
	}

	if filter.IsActive != nil {
		data["is_active"] = *filter.IsActive
		wc = append(wc, "is_active = :is_active")
	}

	if filter.IsPerishable != nil {
		data["is_perishable"] = *filter.IsPerishable
		wc = append(wc, "is_perishable = :is_perishable")
	}

	if filter.HandlingInstructions != nil {
		data["handling_instructions"] = "%" + *filter.HandlingInstructions + "%"
		wc = append(wc, "handling_instructions LIKE :handling_instructions")
	}

	if filter.UnitsPerCase != nil {
		data["units_per_case"] = *filter.UnitsPerCase
		wc = append(wc, "units_per_case = :units_per_case")
	}

	if filter.TrackingType != nil {
		data["tracking_type"] = *filter.TrackingType
		wc = append(wc, "tracking_type = :tracking_type")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
