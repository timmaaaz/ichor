package physicalattributedb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
)

func applyFilter(filter physicalattributebus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["attribute_id"] = *filter.ID
		wc = append(wc, "attribute_id = :attribute_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.Length != nil {
		data["length"] = *filter.Length
		wc = append(wc, "length = :length")
	}

	if filter.Width != nil {
		data["width"] = *filter.Width
		wc = append(wc, "width = :width")
	}

	if filter.Height != nil {
		data["height"] = *filter.Height
		wc = append(wc, "height = :height")
	}

	if filter.Weight != nil {
		data["weight"] = *filter.Weight
		wc = append(wc, "weight = :weight")
	}
	if filter.WeightUnit != nil {
		data["weight_unit"] = *filter.WeightUnit
		wc = append(wc, "weight_unit = :weight_unit")
	}
	if filter.Color != nil {
		data["color"] = *filter.Color
		wc = append(wc, "color = :color")
	}
	if filter.Size != nil {
		data["size"] = *filter.Size
		wc = append(wc, "size = :size")
	}
	if filter.Material != nil {
		data["material"] = *filter.Material
		wc = append(wc, "material = :material")
	}
	if filter.StorageRequirements != nil {
		data["storage_requirements"] = *filter.StorageRequirements
		wc = append(wc, "storage_requirements = :storage_requirements")
	}
	if filter.HazmatClass != nil {
		data["hazmat_class"] = *filter.HazmatClass
		wc = append(wc, "hazmat_class = :hazmat_class")
	}
	if filter.ShelfLifeDays != nil {
		data["shelf_life_days"] = *filter.ShelfLifeDays
		wc = append(wc, "shelf_life_days = :shelf_life_days")
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
