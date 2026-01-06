package metricsbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/metricsbus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Metric struct {
	MetricID          uuid.UUID          `json:"metric_id"`
	ProductID         uuid.UUID          `json:"product_id"`
	ReturnRate        types.RoundedFloat `json:"return_rate"`
	DefectRate        types.RoundedFloat `json:"defect_rate"`
	MeasurementPeriod types.Interval     `json:"measurement_period"`
	CreatedDate       time.Time          `json:"created_date"`
	UpdatedDate       time.Time          `json:"updated_date"`
}

type NewMetric struct {
	ProductID         uuid.UUID          `json:"product_id"`
	ReturnRate        types.RoundedFloat `json:"return_rate"`
	DefectRate        types.RoundedFloat `json:"defect_rate"`
	MeasurementPeriod types.Interval     `json:"measurement_period"`
}

type UpdateMetric struct {
	ProductID         *uuid.UUID          `json:"product_id,omitempty"`
	ReturnRate        *types.RoundedFloat `json:"return_rate,omitempty"`
	DefectRate        *types.RoundedFloat `json:"defect_rate,omitempty"`
	MeasurementPeriod *types.Interval     `json:"measurement_period,omitempty"`
}
