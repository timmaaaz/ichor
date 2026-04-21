package scenarioapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams filters scenarios via the list endpoint.
type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	Name    string
}

// Scenario is the API-facing shape of a scenario.
type Scenario struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
}

// Encode implements the web.Encoder interface.
func (app Scenario) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Scenarios is a slice wrapper so it implements web.Encoder directly.
type Scenarios []Scenario

// Encode implements the web.Encoder interface.
func (app Scenarios) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppScenario(bus scenariobus.Scenario) Scenario {
	return Scenario{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
		CreatedDate: bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate: bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func toAppScenarios(bus []scenariobus.Scenario) Scenarios {
	out := make(Scenarios, len(bus))
	for i, v := range bus {
		out[i] = toAppScenario(v)
	}
	return out
}

// =============================================================================

// NewScenario is the input shape for POST /v1/scenarios.
type NewScenario struct {
	Name        string `json:"name" validate:"required,min=1,max=64"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1024"`
}

// Decode implements the decoder interface.
func (app *NewScenario) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks that the new scenario is well-formed.
func (app NewScenario) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewScenario(app NewScenario) scenariobus.NewScenario {
	return scenariobus.NewScenario{
		Name:        app.Name,
		Description: app.Description,
	}
}

// UpdateScenario carries optional patch fields for PUT /v1/scenarios/{id}.
type UpdateScenario struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1024"`
}

// Decode implements the decoder interface.
func (app *UpdateScenario) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateScenario) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateScenario(app UpdateScenario) scenariobus.UpdateScenario {
	return scenariobus.UpdateScenario{
		Name:        app.Name,
		Description: app.Description,
	}
}
