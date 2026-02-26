package inventoryitemapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID                    string
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
	IncludeLocationDetails string
}

type InventoryItem struct {
	ID                    string `json:"id"`
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
		ID:                    bus.ID.String(),
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

// InventoryItemWithLocation adds location context fields to InventoryItem.
type InventoryItemWithLocation struct {
	InventoryItem
	LocationCode  string `json:"location_code"`
	Aisle         string `json:"aisle"`
	Rack          string `json:"rack"`
	Shelf         string `json:"shelf"`
	Bin           string `json:"bin"`
	ZoneName      string `json:"zone_name"`
	WarehouseName string `json:"warehouse_name"`
}

func (app InventoryItemWithLocation) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppInventoryItemWithLocation(bus inventoryitembus.InventoryItemWithLocation) InventoryItemWithLocation {
	return InventoryItemWithLocation{
		InventoryItem: ToAppInventoryItem(bus.InventoryItem),
		LocationCode:  bus.LocationCode,
		Aisle:         bus.Aisle,
		Rack:          bus.Rack,
		Shelf:         bus.Shelf,
		Bin:           bus.Bin,
		ZoneName:      bus.ZoneName,
		WarehouseName: bus.WarehouseName,
	}
}

func toAppInventoryItemsWithLocation(bus []inventoryitembus.InventoryItemWithLocation) []InventoryItemWithLocation {
	app := make([]InventoryItemWithLocation, len(bus))
	for i, v := range bus {
		app[i] = toAppInventoryItemWithLocation(v)
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
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
	}

	reservedQuantity, err := strconv.Atoi(app.ReservedQuantity)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse reservedQuantity: %s", err)
	}

	allocatedQuantity, err := strconv.Atoi(app.AllocatedQuantity)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse allocatedQuantity: %s", err)
	}

	minimumStock, err := strconv.Atoi(app.MinimumStock)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse minimumStock: %s", err)
	}

	maximumStock, err := strconv.Atoi(app.MaximumStock)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse maximumStock: %s", err)
	}

	reorderPoint, err := strconv.Atoi(app.ReorderPoint)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse reorderPoint: %s", err)
	}

	economicOrderQuantity, err := strconv.Atoi(app.EconomicOrderQuantity)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse economicOrderQuantity: %s", err)
	}

	safetyStock, err := strconv.Atoi(app.SafetyStock)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse safetyStock: %s", err)
	}

	avgDailyUsage, err := strconv.Atoi(app.AvgDailyUsage)
	if err != nil {
		return inventoryitembus.NewInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse avgDailyUsage: %s", err)
	}

	bus := inventoryitembus.NewInventoryItem{
		ProductID:             productID,
		LocationID:            locationID,
		Quantity:              quantity,
		ReservedQuantity:      reservedQuantity,
		AllocatedQuantity:     allocatedQuantity,
		MinimumStock:          minimumStock,
		MaximumStock:          maximumStock,
		ReorderPoint:          reorderPoint,
		EconomicOrderQuantity: economicOrderQuantity,
		SafetyStock:           safetyStock,
		AvgDailyUsage:         avgDailyUsage,
	}
	return bus, nil
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
	bus := inventoryitembus.UpdateInventoryItem{}

	if app.ProductID != nil {
		productID, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		bus.ProductID = &productID
	}

	if app.LocationID != nil {
		locationID, err := uuid.Parse(*app.LocationID)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
		}
		bus.LocationID = &locationID
	}

	if app.Quantity != nil {
		quantity, err := strconv.Atoi(*app.Quantity)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
		}
		bus.Quantity = &quantity
	}

	if app.ReservedQuantity != nil {
		reservedQuantity, err := strconv.Atoi(*app.ReservedQuantity)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse reservedQuantity: %s", err)
		}
		bus.ReservedQuantity = &reservedQuantity
	}

	if app.AllocatedQuantity != nil {
		allocatedQuantity, err := strconv.Atoi(*app.AllocatedQuantity)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse allocatedQuantity: %s", err)
		}
		bus.AllocatedQuantity = &allocatedQuantity
	}

	if app.MinimumStock != nil {
		minimumStock, err := strconv.Atoi(*app.MinimumStock)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse minimumStock: %s", err)
		}
		bus.MinimumStock = &minimumStock
	}

	if app.MaximumStock != nil {
		maximumStock, err := strconv.Atoi(*app.MaximumStock)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse maximumStock: %s", err)
		}
		bus.MaximumStock = &maximumStock
	}

	if app.ReorderPoint != nil {
		reorderPoint, err := strconv.Atoi(*app.ReorderPoint)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse reorderPoint: %s", err)
		}
		bus.ReorderPoint = &reorderPoint
	}

	if app.EconomicOrderQuantity != nil {
		economicOrderQuantity, err := strconv.Atoi(*app.EconomicOrderQuantity)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse economicOrderQuantity: %s", err)
		}
		bus.EconomicOrderQuantity = &economicOrderQuantity
	}

	if app.SafetyStock != nil {
		safetyStock, err := strconv.Atoi(*app.SafetyStock)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse safetyStock: %s", err)
		}
		bus.SafetyStock = &safetyStock
	}

	if app.AvgDailyUsage != nil {
		avgDailyUsage, err := strconv.Atoi(*app.AvgDailyUsage)
		if err != nil {
			return inventoryitembus.UpdateInventoryItem{}, errs.Newf(errs.InvalidArgument, "parse avgDailyUsage: %s", err)
		}
		bus.AvgDailyUsage = &avgDailyUsage
	}

	return bus, nil
}
