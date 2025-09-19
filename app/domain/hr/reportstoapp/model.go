package reportstoapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
)

// QueryParams represents the query parameters that can be used.
type QueryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	BossID     string
	ReporterID string
}

type ReportsTo struct {
	ID         string `json:"id"`
	BossID     string `json:"boss_id"`
	ReporterID string `json:"reporter_id"`
}

// Encode implements the encoder interface.
func (app ReportsTo) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppReportsTo(bus reportstobus.ReportsTo) ReportsTo {
	return ReportsTo{
		ID:         bus.ID.String(),
		BossID:     bus.BossID.String(),
		ReporterID: bus.ReporterID.String(),
	}
}

func ToAppReportsTos(bus []reportstobus.ReportsTo) []ReportsTo {
	app := make([]ReportsTo, len(bus))
	for i, v := range bus {
		app[i] = ToAppReportsTo(v)
	}
	return app
}

// =============================================================================

type NewReportsTo struct {
	BossID     string `json:"boss_id" validate:"required,min=36,max=36"`
	ReporterID string `json:"reporter_id" validate:"required,min=36,max=36"`
}

// Decode implements the decoder interface.
func (app *NewReportsTo) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewReportsTo) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusNewReportsTo(app NewReportsTo) reportstobus.NewReportsTo {
	return reportstobus.NewReportsTo{
		BossID:     uuid.MustParse(app.BossID),
		ReporterID: uuid.MustParse(app.ReporterID),
	}
}

// =============================================================================

type UpdateReportsTo struct {
	BossID     *string `json:"boss_id" validate:"omitempty,min=36,max=36"`
	ReporterID *string `json:"reporter_id" validate:"omitempty,min=36,max=36"`
}

// Decode implements the decoder interface.
func (app *UpdateReportsTo) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateReportsTo) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusUpdateReportsTo(app UpdateReportsTo) reportstobus.UpdateReportsTo {

	bossID := uuid.MustParse(*app.BossID)
	reporterID := uuid.MustParse(*app.ReporterID)

	return reportstobus.UpdateReportsTo{
		BossID:     &bossID,
		ReporterID: &reporterID,
	}
}
