package metricsdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus/types"
)

type metric struct {
	MetricID          uuid.UUID      `db:"quality_metric_id"`
	ProductID         uuid.UUID      `db:"product_id"`
	ReturnRate        sql.NullString `db:"return_rate"`
	DefectRate        sql.NullString `db:"defect_rate"`
	MeasurementPeriod sql.NullString `db:"measurement_period"`
	CreatedDate       time.Time      `db:"created_date"`
	UpdatedDate       time.Time      `db:"updated_date"`
}

func toDBMetric(bus metricsbus.Metric) metric {
	return metric{
		MetricID:          bus.MetricID,
		ProductID:         bus.ProductID,
		ReturnRate:        bus.ReturnRate.DBValue(),
		DefectRate:        bus.DefectRate.DBValue(),
		MeasurementPeriod: bus.MeasurementPeriod.DBValue(),
		CreatedDate:       bus.CreatedDate,
		UpdatedDate:       bus.UpdatedDate,
	}
}

func toBusMetric(db metric) (metricsbus.Metric, error) {

	returnRate, err := types.ParseRoundedFloat(db.ReturnRate.String)
	if err != nil {
		return metricsbus.Metric{}, fmt.Errorf("tobusmetric: %v", err)
	}

	defectRate, err := types.ParseRoundedFloat(db.DefectRate.String)
	if err != nil {
		return metricsbus.Metric{}, fmt.Errorf("tobusmetric: %v", err)
	}

	measurementPeriod, err := types.ParseInterval(db.MeasurementPeriod.String)
	if err != nil {
		return metricsbus.Metric{}, fmt.Errorf("tobusmetric: %v", err)
	}

	return metricsbus.Metric{
		MetricID:          db.MetricID,
		ProductID:         db.ProductID,
		ReturnRate:        returnRate,
		DefectRate:        defectRate,
		MeasurementPeriod: measurementPeriod,
		CreatedDate:       db.CreatedDate,
		UpdatedDate:       db.UpdatedDate,
	}, nil
}

func toBusMetrics(db []metric) ([]metricsbus.Metric, error) {
	bus := make([]metricsbus.Metric, len(db))

	for i, d := range db {
		m, err := toBusMetric(d)
		if err != nil {
			return nil, fmt.Errorf("tobusmetrics: %v", err)
		}
		bus[i] = m
	}

	return bus, nil
}
