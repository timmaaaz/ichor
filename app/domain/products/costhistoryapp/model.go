package costhistoryapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/products/costhistorybus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	CostHistoryID string
	ProductID     string
	CostType      string
	Amount        string
	CurrencyID    string
	EffectiveDate string
	EndDate       string
	CreatedDate   string
	UpdatedDate   string
}

type CostHistory struct {
	CostHistoryID string `json:"id"`
	ProductID     string `json:"product_id"`
	CostType      string `json:"cost_type"`
	Amount        string `json:"amount"`
	CurrencyID    string `json:"currency_id"`
	EffectiveDate string `json:"effective_date"`
	EndDate       string `json:"end_date"`
	CreatedDate   string `json:"created_date"`
	UpdatedDate   string `json:"updated_date"`
}

func (app CostHistory) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppCostHistory(bus costhistorybus.CostHistory) CostHistory {
	return CostHistory{
		CostHistoryID: bus.CostHistoryID.String(),
		ProductID:     bus.ProductID.String(),
		CostType:      bus.CostType,
		Amount:        bus.Amount.Value(),
		CurrencyID:    bus.CurrencyID.String(),
		EffectiveDate: bus.EffectiveDate.Format(timeutil.FORMAT),
		EndDate:       bus.EndDate.Format(timeutil.FORMAT),
		CreatedDate:   bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:   bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppCostHistories(bus []costhistorybus.CostHistory) []CostHistory {
	app := make([]CostHistory, len(bus))
	for i, v := range bus {
		app[i] = ToAppCostHistory(v)
	}
	return app
}

// =========================================================================

type NewCostHistory struct {
	ProductID     string  `json:"product_id" validate:"required,min=36,max=36"`
	CostType      string  `json:"cost_type" validate:"required"`
	Amount        string  `json:"amount" validate:"required"`
	CurrencyID    string  `json:"currency_id" validate:"required,min=36,max=36"`
	EffectiveDate string  `json:"effective_date" validate:"required"`
	EndDate       string  `json:"end_date" validate:"required"`
	CreatedDate   *string `json:"created_date"` // Optional: for seeding/import
}

func (app *NewCostHistory) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewCostHistory) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewCostHistory(app NewCostHistory) (costhistorybus.NewCostHistory, error) {
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return costhistorybus.NewCostHistory{}, errs.Newf(errs.InvalidArgument, "parse product_id: %s", err)
	}

	amount, err := types.ParseMoney(app.Amount)
	if err != nil {
		return costhistorybus.NewCostHistory{}, errs.Newf(errs.InvalidArgument, "parse amount: %s", err)
	}

	effectiveDate, err := time.Parse(timeutil.FORMAT, app.EffectiveDate)
	if err != nil {
		return costhistorybus.NewCostHistory{}, errs.Newf(errs.InvalidArgument, "parse effective_date: %s", err)
	}

	endDate, err := time.Parse(timeutil.FORMAT, app.EndDate)
	if err != nil {
		return costhistorybus.NewCostHistory{}, errs.Newf(errs.InvalidArgument, "parse end_date: %s", err)
	}

	currencyID, err := uuid.Parse(app.CurrencyID)
	if err != nil {
		return costhistorybus.NewCostHistory{}, errs.Newf(errs.InvalidArgument, "parse currency_id: %s", err)
	}

	bus := costhistorybus.NewCostHistory{
		ProductID:     productID,
		CostType:      app.CostType,
		Amount:        amount,
		CurrencyID:    currencyID,
		EffectiveDate: effectiveDate,
		EndDate:       endDate,
		// CreatedDate: nil by default - API always uses server time
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return costhistorybus.NewCostHistory{}, errs.Newf(errs.InvalidArgument, "parse createdDate: %s", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
}

// =========================================================================

type UpdateCostHistory struct {
	ProductID     *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	CostType      *string `json:"cost_type" validate:"omitempty"`
	Amount        *string `json:"amount" validate:"omitempty"`
	CurrencyID    *string `json:"currency_id" validate:"omitempty,min=36,max=36"`
	EffectiveDate *string `json:"effective_date" validate:"omitempty"`
	EndDate       *string `json:"end_date" validate:"omitempty"`
}

func (app *UpdateCostHistory) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateCostHistory) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateCostHistory(app UpdateCostHistory) (costhistorybus.UpdateCostHistory, error) {
	bus := costhistorybus.UpdateCostHistory{}

	if app.ProductID != nil {
		id, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return costhistorybus.UpdateCostHistory{}, errs.Newf(errs.InvalidArgument, "parse product_id: %s", err)
		}
		bus.ProductID = &id
	}

	if app.Amount != nil {
		amount, err := types.ParseMoney(*app.Amount)
		if err != nil {
			return costhistorybus.UpdateCostHistory{}, errs.Newf(errs.InvalidArgument, "parse amount: %s", err)
		}
		bus.Amount = &amount
	}

	if app.EffectiveDate != nil {
		date, err := time.Parse(timeutil.FORMAT, *app.EffectiveDate)
		if err != nil {
			return costhistorybus.UpdateCostHistory{}, errs.Newf(errs.InvalidArgument, "parse effective_date: %s", err)
		}
		bus.EffectiveDate = &date
	}

	if app.EndDate != nil {
		date, err := time.Parse(timeutil.FORMAT, *app.EndDate)
		if err != nil {
			return costhistorybus.UpdateCostHistory{}, errs.Newf(errs.InvalidArgument, "parse end_date: %s", err)
		}
		bus.EndDate = &date
	}

	bus.CostType = app.CostType

	if app.CurrencyID != nil {
		currencyID, err := uuid.Parse(*app.CurrencyID)
		if err != nil {
			return costhistorybus.UpdateCostHistory{}, errs.Newf(errs.InvalidArgument, "parse currency_id: %s", err)
		}
		bus.CurrencyID = &currencyID
	}

	return bus, nil
}
