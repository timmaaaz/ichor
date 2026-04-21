package picktaskapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the raw query parameters from the HTTP request.
type QueryParams struct {
	Page                 string
	Rows                 string
	OrderBy              string
	ID                   string
	SalesOrderID         string
	SalesOrderLineItemID string
	ProductID            string
	LocationID           string
	Status               string
	AssignedTo           string
	CreatedBy            string
}

// =============================================================================
// Response model
// =============================================================================

// PickTask is the app-layer response model. All fields are strings for JSON serialization.
type PickTask struct {
	ID                   string `json:"id"`
	TaskNumber           string `json:"task_number"`
	SalesOrderID         string `json:"sales_order_id"`
	SalesOrderLineItemID string `json:"sales_order_line_item_id"`
	ProductID            string `json:"product_id"`
	LotID                string `json:"lot_id"`
	SerialID             string `json:"serial_id"`
	LocationID           string `json:"location_id"`
	QuantityToPick       string `json:"quantity_to_pick"`
	QuantityPicked       string `json:"quantity_picked"`
	Status               string `json:"status"`
	AssignedTo           string `json:"assigned_to"`
	AssignedAt           string `json:"assigned_at"`
	CompletedBy          string `json:"completed_by"`
	CompletedAt          string `json:"completed_at"`
	ShortPickReason      string `json:"short_pick_reason"`
	CreatedBy            string `json:"created_by"`
	CreatedDate          string `json:"created_date"`
	UpdatedDate          string `json:"updated_date"`
	ScenarioID           string `json:"scenario_id,omitempty"`
}

func (app PickTask) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppPickTask converts a bus model to an app-layer response model.
func ToAppPickTask(bus picktaskbus.PickTask) PickTask {
	taskNumber := ""
	if bus.TaskNumber != nil {
		taskNumber = *bus.TaskNumber
	}

	lotID := ""
	if bus.LotID != nil {
		lotID = bus.LotID.String()
	}

	serialID := ""
	if bus.SerialID != nil {
		serialID = bus.SerialID.String()
	}

	assignedTo := ""
	if bus.AssignedTo != uuid.Nil {
		assignedTo = bus.AssignedTo.String()
	}

	assignedAt := ""
	if !bus.AssignedAt.IsZero() {
		assignedAt = bus.AssignedAt.Format(timeutil.FORMAT)
	}

	completedBy := ""
	if bus.CompletedBy != uuid.Nil {
		completedBy = bus.CompletedBy.String()
	}

	completedAt := ""
	if !bus.CompletedAt.IsZero() {
		completedAt = bus.CompletedAt.Format(timeutil.FORMAT)
	}

	scenarioID := ""
	if bus.ScenarioID != nil {
		scenarioID = bus.ScenarioID.String()
	}

	return PickTask{
		ID:                   bus.ID.String(),
		TaskNumber:           taskNumber,
		SalesOrderID:         bus.SalesOrderID.String(),
		SalesOrderLineItemID: bus.SalesOrderLineItemID.String(),
		ProductID:            bus.ProductID.String(),
		LotID:                lotID,
		SerialID:             serialID,
		LocationID:           bus.LocationID.String(),
		QuantityToPick:       fmt.Sprintf("%d", bus.QuantityToPick),
		QuantityPicked:       fmt.Sprintf("%d", bus.QuantityPicked),
		Status:               bus.Status.String(),
		AssignedTo:           assignedTo,
		AssignedAt:           assignedAt,
		CompletedBy:          completedBy,
		CompletedAt:          completedAt,
		ShortPickReason:      bus.ShortPickReason,
		CreatedBy:            bus.CreatedBy.String(),
		CreatedDate:          bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:          bus.UpdatedDate.Format(timeutil.FORMAT),
		ScenarioID:           scenarioID,
	}
}

// ToAppPickTasks converts a slice of bus models to app-layer response models.
func ToAppPickTasks(bus []picktaskbus.PickTask) []PickTask {
	app := make([]PickTask, len(bus))
	for i, v := range bus {
		app[i] = ToAppPickTask(v)
	}
	return app
}

// =============================================================================
// Create model
// =============================================================================

// NewPickTask is the app-layer create request model.
// CreatedBy is injected from the authenticated user — not accepted from the client.
type NewPickTask struct {
	TaskNumber           *string `json:"task_number" validate:"omitempty,min=1,max=32"`
	SalesOrderID         string  `json:"sales_order_id" validate:"required,min=36,max=36"`
	SalesOrderLineItemID string  `json:"sales_order_line_item_id" validate:"required,min=36,max=36"`
	ProductID            string  `json:"product_id" validate:"required,min=36,max=36"`
	LotID                *string `json:"lot_id" validate:"omitempty,min=36,max=36"`
	SerialID             *string `json:"serial_id" validate:"omitempty,min=36,max=36"`
	LocationID           string  `json:"location_id" validate:"required,min=36,max=36"`
	QuantityToPick       string  `json:"quantity_to_pick" validate:"required"`
}

func (app *NewPickTask) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app NewPickTask) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewPickTask(app NewPickTask, createdBy uuid.UUID) (picktaskbus.NewPickTask, error) {
	salesOrderID, err := uuid.Parse(app.SalesOrderID)
	if err != nil {
		return picktaskbus.NewPickTask{}, errs.Newf(errs.InvalidArgument, "parse salesOrderID: %s", err)
	}

	salesOrderLineItemID, err := uuid.Parse(app.SalesOrderLineItemID)
	if err != nil {
		return picktaskbus.NewPickTask{}, errs.Newf(errs.InvalidArgument, "parse salesOrderLineItemID: %s", err)
	}

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return picktaskbus.NewPickTask{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return picktaskbus.NewPickTask{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
	}

	quantityToPick, err := strconv.Atoi(app.QuantityToPick)
	if err != nil {
		return picktaskbus.NewPickTask{}, errs.Newf(errs.InvalidArgument, "parse quantityToPick: %s", err)
	}

	var lotID *uuid.UUID
	if app.LotID != nil {
		id, err := uuid.Parse(*app.LotID)
		if err != nil {
			return picktaskbus.NewPickTask{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
		}
		lotID = &id
	}

	var serialID *uuid.UUID
	if app.SerialID != nil {
		id, err := uuid.Parse(*app.SerialID)
		if err != nil {
			return picktaskbus.NewPickTask{}, errs.Newf(errs.InvalidArgument, "parse serialID: %s", err)
		}
		serialID = &id
	}

	bus := picktaskbus.NewPickTask{
		SalesOrderID:         salesOrderID,
		SalesOrderLineItemID: salesOrderLineItemID,
		ProductID:            productID,
		LotID:                lotID,
		SerialID:             serialID,
		LocationID:           locationID,
		QuantityToPick:       quantityToPick,
		CreatedBy:            createdBy,
	}

	if app.TaskNumber != nil {
		bus.TaskNumber = app.TaskNumber
	}

	return bus, nil
}

// =============================================================================
// Update model
// =============================================================================

// UpdatePickTask is the app-layer update request model.
type UpdatePickTask struct {
	TaskNumber      *string `json:"task_number" validate:"omitempty,min=1,max=32"`
	LotID           *string `json:"lot_id" validate:"omitempty,min=36,max=36"`
	SerialID        *string `json:"serial_id" validate:"omitempty,min=36,max=36"`
	LocationID      *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	QuantityToPick  *string `json:"quantity_to_pick" validate:"omitempty"`
	QuantityPicked  *string `json:"quantity_picked" validate:"omitempty"`
	Status          *string `json:"status" validate:"omitempty"`
	ShortPickReason *string `json:"short_pick_reason" validate:"omitempty"`
}

func (app *UpdatePickTask) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app UpdatePickTask) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdatePickTask(app UpdatePickTask) (picktaskbus.UpdatePickTask, error) {
	bus := picktaskbus.UpdatePickTask{}

	if app.LotID != nil {
		id, err := uuid.Parse(*app.LotID)
		if err != nil {
			return picktaskbus.UpdatePickTask{}, errs.Newf(errs.InvalidArgument, "parse lotID: %s", err)
		}
		bus.LotID = &id
	}

	if app.SerialID != nil {
		id, err := uuid.Parse(*app.SerialID)
		if err != nil {
			return picktaskbus.UpdatePickTask{}, errs.Newf(errs.InvalidArgument, "parse serialID: %s", err)
		}
		bus.SerialID = &id
	}

	if app.LocationID != nil {
		id, err := uuid.Parse(*app.LocationID)
		if err != nil {
			return picktaskbus.UpdatePickTask{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
		}
		bus.LocationID = &id
	}

	if app.QuantityToPick != nil {
		q, err := strconv.Atoi(*app.QuantityToPick)
		if err != nil {
			return picktaskbus.UpdatePickTask{}, errs.Newf(errs.InvalidArgument, "parse quantityToPick: %s", err)
		}
		bus.QuantityToPick = &q
	}

	if app.QuantityPicked != nil {
		q, err := strconv.Atoi(*app.QuantityPicked)
		if err != nil {
			return picktaskbus.UpdatePickTask{}, errs.Newf(errs.InvalidArgument, "parse quantityPicked: %s", err)
		}
		bus.QuantityPicked = &q
	}

	if app.Status != nil {
		st, err := picktaskbus.ParseStatus(*app.Status)
		if err != nil {
			return picktaskbus.UpdatePickTask{}, errs.Newf(errs.InvalidArgument, "parse status: %s", err)
		}
		bus.Status = &st
	}

	if app.ShortPickReason != nil {
		bus.ShortPickReason = app.ShortPickReason
	}

	if app.TaskNumber != nil {
		bus.TaskNumber = app.TaskNumber
	}

	return bus, nil
}
