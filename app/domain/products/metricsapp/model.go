package metricsapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	MetricID          string
	ProductID         string
	ReturnRate        string
	DefectRate        string
	MeasurementPeriod string
	CreatedDate       string
	UpdatedDate       string
}

type Metric struct {
	MetricID          string `json:"quality_metric_id"`
	ProductID         string `json:"product_id"`
	ReturnRate        string `json:"return_rate"`
	DefectRate        string `json:"defect_rate"`
	MeasurementPeriod string `json:"measurement_period"`
	CreatedDate       string `json:"created_date"`
	UpdatedDate       string `json:"updated_date"`
}

func (app Metric) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppMetric(bus metricsbus.Metric) Metric {
	return Metric{
		MetricID:          bus.MetricID.String(),
		ProductID:         bus.ProductID.String(),
		ReturnRate:        bus.ReturnRate.String(),
		DefectRate:        bus.DefectRate.String(),
		MeasurementPeriod: bus.MeasurementPeriod.Value(),
		CreatedDate:       bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:       bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppMetrics(bus []metricsbus.Metric) []Metric {
	app := make([]Metric, len(bus))
	for i, v := range bus {
		app[i] = ToAppMetric(v)
	}
	return app
}

type NewMetric struct {
	ProductID         string `json:"product_id" validate:"required,min=36,max=36"`
	ReturnRate        string `json:"return_rate" validate:"required"`
	DefectRate        string `json:"defect_rate" validate:"required"`
	MeasurementPeriod string `json:"measurement_period" validate:"required"`
}

func (app *NewMetric) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewMetric) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewMetric(app NewMetric) (metricsbus.NewMetric, error) {

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return metricsbus.NewMetric{}, errs.NewFieldsError("productID", err)
	}

	returnRate, err := types.ParseRoundedFloat(app.ReturnRate)
	if err != nil {
		return metricsbus.NewMetric{}, errs.NewFieldsError("returnRate", err)
	}

	defectRate, err := types.ParseRoundedFloat(app.DefectRate)
	if err != nil {
		return metricsbus.NewMetric{}, errs.NewFieldsError("defectRate", err)
	}

	measurementPeriod, err := types.ParseInterval(app.MeasurementPeriod)
	if err != nil {
		return metricsbus.NewMetric{}, errs.NewFieldsError("measurementPeriod", err)
	}

	return metricsbus.NewMetric{
		ProductID:         productID,
		ReturnRate:        returnRate,
		DefectRate:        defectRate,
		MeasurementPeriod: measurementPeriod,
	}, nil
}

type UpdateMetric struct {
	ProductID         *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	ReturnRate        *string `json:"return_rate"`
	DefectRate        *string `json:"defect_rate"`
	MeasurementPeriod *string `json:"measurement_period"`
}

func (app *UpdateMetric) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateMetric) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateMetric(app UpdateMetric) (metricsbus.UpdateMetric, error) {

	dest := metricsbus.UpdateMetric{}

	if app.ProductID != nil {
		productID, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return metricsbus.UpdateMetric{}, errs.NewFieldsError("productID", err)
		}

		dest.ProductID = &productID
	}

	if app.ReturnRate != nil {
		returnRate, err := types.ParseRoundedFloat(*app.ReturnRate)
		if err != nil {
			return metricsbus.UpdateMetric{}, errs.NewFieldsError("returnRate", err)
		}

		dest.ReturnRate = &returnRate

	}

	if app.DefectRate != nil {
		defectRate, err := types.ParseRoundedFloat(*app.DefectRate)
		if err != nil {
			return metricsbus.UpdateMetric{}, errs.NewFieldsError("defectRate", err)
		}

		dest.DefectRate = &defectRate

	}

	if app.MeasurementPeriod != nil {
		measurementPeriod, err := types.ParseInterval(*app.MeasurementPeriod)
		if err != nil {
			return metricsbus.UpdateMetric{}, errs.NewFieldsError("measurementPeriod", err)
		}
		dest.MeasurementPeriod = &measurementPeriod

	}

	return dest, nil

}
