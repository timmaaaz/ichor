package inventoryitemapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ItemID                string
	ProductID             string
	LocationID            string
	Quantity              string
	ReservedQuantity      string
	AllocatedQuantity     string
	MinimumStock          string
	MaximumStock          string
	ReorderPoint          string
	EconomicOrderQuantity string
	SafetyStock           string
	AvgDailyUsage         string
	CreatedDate           string
	UpdatedDate           string
}

type InventoryItem struct {
	ItemID                string `json:"id"`
	ProductID             string `json:"product_id"`
	LocationID            string `json:"location_id"`
	Quantity              string `json:"quantity"`
	ReservedQuantity      string `json:"reserved_quantity"`
	AllocatedQuantity     string `json:"allocated_quantity"`
	MinimumStock          string `json:"minimum_stock"`
	MaximumStock          string `json:"maximum_stock"`
	ReorderPoint          string `json:"reorder_point"`
	EconomicOrderQuantity string `json:"economic_order_quantity"`
	SafetyStock           string `json:"safety_stock"`
	AvgDailyUsage         string `json:"avg_daily_usage"`
	CreatedDate           string `json:"created_date"`
	UpdatedDate           string `json:"updated_date"`
}

func (app InventoryItem) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppInventoryItem(bus inventoryitembus.InventoryItem) InventoryItem {
	return InventoryItem{
		ItemID:                bus.ItemID.String(),
		ProductID:             bus.ProductID.String(),
		LocationID:            bus.LocationID.String(),
		Quantity:              fmt.Sprintf("%d", bus.Quantity),
		ReservedQuantity:      fmt.Sprintf("%d", bus.ReservedQuantity),
		AllocatedQuantity:     fmt.Sprintf("%d", bus.AllocatedQuantity),
		MinimumStock:          fmt.Sprintf("%d", bus.MinimumStock),
		MaximumStock:          fmt.Sprintf("%d", bus.MaximumStock),
		ReorderPoint:          fmt.Sprintf("%d", bus.ReorderPoint),
		EconomicOrderQuantity: fmt.Sprintf("%d", bus.EconomicOrderQuantity),
		SafetyStock:           fmt.Sprintf("%d", bus.SafetyStock),
		AvgDailyUsage:         fmt.Sprintf("%d", bus.AvgDailyUsage),
		CreatedDate:           bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:           bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppInventoryItems(bus []inventoryitembus.InventoryItem) []InventoryItem {
	app := make([]InventoryItem, len(bus))
	for i, v := range bus {
		app[i] = ToAppInventoryItem(v)
	}
	return app
}

type NewInventoryItem struct {
	ProductID             string `json:"product_id" validate:"required,min=36,max=36"`
	LocationID            string `json:"location_id" validate:"required,min=36,max=36"`
	Quantity              string `json:"quantity" validate:"required"`
	ReservedQuantity      string `json:"reserved_quantity" validate:"required"`
	AllocatedQuantity     string `json:"allocated_quantity" validate:"required"`
	MinimumStock          string `json:"minimum_stock" validate:"required"`
	MaximumStock          string `json:"maximum_stock" validate:"required"`
	ReorderPoint          string `json:"reorder_point" validate:"required"`
	EconomicOrderQuantity string `json:"economic_order_quantity" validate:"required"`
	SafetyStock           string `json:"safety_stock" validate:"required"`
	AvgDailyUsage         string `json:"avg_daily_usage" validate:"required"`
}

func (app *NewInventoryItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewInventoryItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusNewInventoryItem(app NewInventoryItem) (inventoryitembus.NewInventoryItem, error) {
	dest := inventoryitembus.NewInventoryItem{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, fmt.Errorf("toBusNewInventoryItem: %w", err)
	}

	return dest, nil
}

type UpdateInventoryItem struct {
	ProductID             *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	LocationID            *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	Quantity              *string `json:"quantity" validate:"omitempty"`
	ReservedQuantity      *string `json:"reserved_quantity" validate:"omitempty"`
	AllocatedQuantity     *string `json:"allocated_quantity" validate:"omitempty"`
	MinimumStock          *string `json:"minimum_stock" validate:"omitempty"`
	MaximumStock          *string `json:"maximum_stock" validate:"omitempty"`
	ReorderPoint          *string `json:"reorder_point" validate:"omitempty"`
	EconomicOrderQuantity *string `json:"economic_order_quantity" validate:"omitempty"`
	SafetyStock           *string `json:"safety_stock" validate:"omitempty"`
	AvgDailyUsage         *string `json:"avg_daily_usage" validate:"omitempty"`
}

func (app *UpdateInventoryItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateInventoryItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusUpdateInventoryItem(app UpdateInventoryItem) (inventoryitembus.UpdateInventoryItem, error) {
	dest := inventoryitembus.UpdateInventoryItem{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return inventoryitembus.UpdateInventoryItem{}, fmt.Errorf("toBusUpdateInventoryItem: %w", err)
	}

	return dest, nil
}
