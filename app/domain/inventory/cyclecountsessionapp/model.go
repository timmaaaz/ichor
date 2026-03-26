package cyclecountsessionapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the raw query parameters from the HTTP request.
type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Status      string
	CreatedBy   string
	CreatedDate string
}

// =============================================================================
// Response model
// =============================================================================

// CycleCountSession is the app-layer response model. All fields are strings for JSON serialization.
type CycleCountSession struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	CreatedBy     string `json:"createdBy"`
	CreatedDate   string `json:"createdDate"`
	UpdatedDate   string `json:"updatedDate"`
	CompletedDate string `json:"completedDate"`
}

func (app CycleCountSession) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppCycleCountSession converts a bus model to an app-layer response model.
func ToAppCycleCountSession(bus cyclecountsessionbus.CycleCountSession) CycleCountSession {
	completedDate := ""
	if bus.CompletedDate != nil {
		completedDate = bus.CompletedDate.Format(timeutil.FORMAT)
	}

	return CycleCountSession{
		ID:            bus.ID.String(),
		Name:          bus.Name,
		Status:        bus.Status.String(),
		CreatedBy:     bus.CreatedBy.String(),
		CreatedDate:   bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:   bus.UpdatedDate.Format(timeutil.FORMAT),
		CompletedDate: completedDate,
	}
}

// ToAppCycleCountSessions converts a slice of bus models to app-layer response models.
func ToAppCycleCountSessions(bus []cyclecountsessionbus.CycleCountSession) []CycleCountSession {
	app := make([]CycleCountSession, len(bus))
	for i, v := range bus {
		app[i] = ToAppCycleCountSession(v)
	}
	return app
}

// =============================================================================
// Create model
// =============================================================================

// NewCycleCountSession is the app-layer create request model.
// CreatedBy is injected from the authenticated user — not accepted from the client.
type NewCycleCountSession struct {
	Name string `json:"name" validate:"required"`
}

func (app *NewCycleCountSession) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app NewCycleCountSession) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewCycleCountSession(app NewCycleCountSession, createdBy uuid.UUID) cyclecountsessionbus.NewCycleCountSession {
	return cyclecountsessionbus.NewCycleCountSession{
		Name:      app.Name,
		CreatedBy: createdBy,
	}
}

// =============================================================================
// Update model
// =============================================================================

// UpdateCycleCountSession is the app-layer update request model.
type UpdateCycleCountSession struct {
	Name   *string `json:"name" validate:"omitempty"`
	Status *string `json:"status" validate:"omitempty"`
}

func (app *UpdateCycleCountSession) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app UpdateCycleCountSession) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateCycleCountSession(app UpdateCycleCountSession) (cyclecountsessionbus.UpdateCycleCountSession, error) {
	bus := cyclecountsessionbus.UpdateCycleCountSession{}

	if app.Name != nil {
		bus.Name = app.Name
	}

	if app.Status != nil {
		st, err := cyclecountsessionbus.ParseStatus(*app.Status)
		if err != nil {
			return cyclecountsessionbus.UpdateCycleCountSession{}, errs.Newf(errs.InvalidArgument, "parse status: %s", err)
		}
		bus.Status = &st
	}

	return bus, nil
}
