package metricsbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus/types"
)

type QueryFilter struct {
	MetricID          *uuid.UUID
	ProductID         *uuid.UUID
	ReturnRate        *types.RoundedFloat
	DefectRate        *types.RoundedFloat
	MeasurementPeriod *types.Interval
	CreatedDate       *time.Time
	UpdatedDate       *time.Time
}
