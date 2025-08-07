package costhistoryapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus"
	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus/types"
	"github.com/timmaaaz/ichor/business/sdk/convert"
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
	Currency      string
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
	Currency      string `json:"currency"`
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
		Currency:      bus.Currency,
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
	ProductID     string `json:"product_id" validate:"required,min=36,max=36"`
	CostType      string `json:"cost_type" validate:"required"`
	Amount        string `json:"amount" validate:"required"`
	Currency      string `json:"currency" validate:"required"`
	EffectiveDate string `json:"effective_date" validate:"required"`
	EndDate       string `json:"end_date" validate:"required"`
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
	dest := costhistorybus.NewCostHistory{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return costhistorybus.NewCostHistory{}, fmt.Errorf("error populating cost history: %s", err)
	}

	dest.Amount, err = types.ParseMoney(app.Amount)
	if err != nil {
		return dest, err
	}

	return dest, err
}

// =========================================================================

type UpdateCostHistory struct {
	ProductID     *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	CostType      *string `json:"cost_type" validate:"omitempty"`
	Amount        *string `json:"amount" validate:"omitempty"`
	Currency      *string `json:"currency" validate:"omitempty"`
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
	dest := costhistorybus.UpdateCostHistory{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return costhistorybus.UpdateCostHistory{}, fmt.Errorf("error populating cost history: %s", err)
	}

	if app.Amount != nil {
		m, err := types.ParseMoney(*app.Amount)
		if err != nil {
			return dest, err
		}
		dest.Amount = &m
	}

	return dest, err
}
