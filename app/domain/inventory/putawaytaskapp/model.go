package putawaytaskapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the raw query parameters from the HTTP request.
type QueryParams struct {
	Page            string
	Rows            string
	OrderBy         string
	ID              string
	ProductID       string
	LocationID      string
	Status          string
	AssignedTo      string
	CreatedBy       string
	ReferenceNumber string
}

// =============================================================================
// Response model
// =============================================================================

// PutAwayTask is the app-layer response model. All fields are strings for JSON serialization.
type PutAwayTask struct {
	ID              string `json:"id"`
	ProductID       string `json:"product_id"`
	LocationID      string `json:"location_id"`
	Quantity        string `json:"quantity"`
	ReferenceNumber string `json:"reference_number"`
	Status          string `json:"status"`
	AssignedTo      string `json:"assigned_to"`
	AssignedAt      string `json:"assigned_at"`
	CompletedBy     string `json:"completed_by"`
	CompletedAt     string `json:"completed_at"`
	CreatedBy       string `json:"created_by"`
	CreatedDate     string `json:"created_date"`
	UpdatedDate     string `json:"updated_date"`
}

func (app PutAwayTask) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppPutAwayTask converts a bus model to an app-layer response model.
func ToAppPutAwayTask(bus putawaytaskbus.PutAwayTask) PutAwayTask {
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

	return PutAwayTask{
		ID:              bus.ID.String(),
		ProductID:       bus.ProductID.String(),
		LocationID:      bus.LocationID.String(),
		Quantity:        fmt.Sprintf("%d", bus.Quantity),
		ReferenceNumber: bus.ReferenceNumber,
		Status:          bus.Status.String(),
		AssignedTo:      assignedTo,
		AssignedAt:      assignedAt,
		CompletedBy:     completedBy,
		CompletedAt:     completedAt,
		CreatedBy:       bus.CreatedBy.String(),
		CreatedDate:     bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:     bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

// ToAppPutAwayTasks converts a slice of bus models to app-layer response models.
func ToAppPutAwayTasks(bus []putawaytaskbus.PutAwayTask) []PutAwayTask {
	app := make([]PutAwayTask, len(bus))
	for i, v := range bus {
		app[i] = ToAppPutAwayTask(v)
	}
	return app
}

// =============================================================================
// Create model
// =============================================================================

// NewPutAwayTask is the app-layer create request model.
// CreatedBy is injected from the authenticated user — not accepted from the client.
type NewPutAwayTask struct {
	ProductID       string `json:"product_id" validate:"required,min=36,max=36"`
	LocationID      string `json:"location_id" validate:"required,min=36,max=36"`
	Quantity        string `json:"quantity" validate:"required"`
	ReferenceNumber string `json:"reference_number"`
}

func (app *NewPutAwayTask) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app NewPutAwayTask) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewPutAwayTask(app NewPutAwayTask, createdBy uuid.UUID) (putawaytaskbus.NewPutAwayTask, error) {
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return putawaytaskbus.NewPutAwayTask{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return putawaytaskbus.NewPutAwayTask{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return putawaytaskbus.NewPutAwayTask{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
	}

	return putawaytaskbus.NewPutAwayTask{
		ProductID:       productID,
		LocationID:      locationID,
		Quantity:        quantity,
		ReferenceNumber: app.ReferenceNumber,
		CreatedBy:       createdBy,
	}, nil
}

// =============================================================================
// Update model
// =============================================================================

// UpdatePutAwayTask is the app-layer update request model.
// Only Status is expected in most cases (claim / complete / cancel).
// Other fields allow manual supervisor corrections.
type UpdatePutAwayTask struct {
	ProductID       *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	LocationID      *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	Quantity        *string `json:"quantity" validate:"omitempty"`
	ReferenceNumber *string `json:"reference_number" validate:"omitempty"`
	Status          *string `json:"status" validate:"omitempty"`
}

func (app *UpdatePutAwayTask) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app UpdatePutAwayTask) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdatePutAwayTask(app UpdatePutAwayTask) (putawaytaskbus.UpdatePutAwayTask, error) {
	bus := putawaytaskbus.UpdatePutAwayTask{}

	if app.ProductID != nil {
		id, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return putawaytaskbus.UpdatePutAwayTask{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		bus.ProductID = &id
	}

	if app.LocationID != nil {
		id, err := uuid.Parse(*app.LocationID)
		if err != nil {
			return putawaytaskbus.UpdatePutAwayTask{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
		}
		bus.LocationID = &id
	}

	if app.Quantity != nil {
		q, err := strconv.Atoi(*app.Quantity)
		if err != nil {
			return putawaytaskbus.UpdatePutAwayTask{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
		}
		bus.Quantity = &q
	}

	if app.ReferenceNumber != nil {
		bus.ReferenceNumber = app.ReferenceNumber
	}

	if app.Status != nil {
		st, err := putawaytaskbus.ParseStatus(*app.Status)
		if err != nil {
			return putawaytaskbus.UpdatePutAwayTask{}, errs.Newf(errs.InvalidArgument, "parse status: %s", err)
		}
		bus.Status = &st
	}

	return bus, nil
}
