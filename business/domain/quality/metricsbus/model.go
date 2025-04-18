package metricsbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus/types"
)

type Metric struct {
	MetricID          uuid.UUID
	ProductID         uuid.UUID
	ReturnRate        types.RoundedFloat
	DefectRate        types.RoundedFloat
	MeasurementPeriod types.Interval
	CreatedDate       time.Time
	UpdatedDate       time.Time
}

type NewMetric struct {
	ProductID         uuid.UUID
	ReturnRate        types.RoundedFloat
	DefectRate        types.RoundedFloat
	MeasurementPeriod types.Interval
}

type UpdateMetric struct {
	ProductID         *uuid.UUID
	ReturnRate        *types.RoundedFloat
	DefectRate        *types.RoundedFloat
	MeasurementPeriod *types.Interval
}
