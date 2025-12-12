package metricsbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("metric not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("metric entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, metric Metric) error
	Update(ctx context.Context, metric Metric) error
	Delete(ctx context.Context, metric Metric) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Metric, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, metricID uuid.UUID) (Metric, error)
}

// Business manages the set of APIs for brand access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a brand business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}, nil
}

// Create inserts a new metric into the database.
func (b *Business) Create(ctx context.Context, nm NewMetric) (Metric, error) {
	ctx, span := otel.AddSpan(ctx, "business.metricsbus.create")
	defer span.End()

	now := time.Now()

	metric := Metric{
		MetricID:          uuid.New(),
		ProductID:         nm.ProductID,
		ReturnRate:        nm.ReturnRate,
		DefectRate:        nm.DefectRate,
		MeasurementPeriod: nm.MeasurementPeriod,
		CreatedDate:       now,
		UpdatedDate:       now,
	}

	err := b.storer.Create(ctx, metric)
	if err != nil {
		return Metric{}, fmt.Errorf("create: %w", err)
	}

	return metric, nil
}

// Update modifies a metric in the system.
func (b *Business) Update(ctx context.Context, metric Metric, um UpdateMetric) (Metric, error) {
	ctx, span := otel.AddSpan(ctx, "business.metricsbus.update")
	defer span.End()

	if um.ProductID != nil {
		metric.ProductID = *um.ProductID
	}
	if um.ReturnRate != nil {
		metric.ReturnRate = *um.ReturnRate
	}
	if um.DefectRate != nil {
		metric.DefectRate = *um.DefectRate
	}
	if um.MeasurementPeriod != nil {
		metric.MeasurementPeriod = *um.MeasurementPeriod
	}

	metric.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, metric); err != nil {
		return Metric{}, fmt.Errorf("update: %w", err)
	}

	return metric, nil
}

// Delete removes a metric from the system.
func (b *Business) Delete(ctx context.Context, metric Metric) error {
	ctx, span := otel.AddSpan(ctx, "business.metricsbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, metric)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves metrics based on the given filter, order, and pagination parameters.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Metric, error) {
	ctx, span := otel.AddSpan(ctx, "business.metricsbus.query")
	defer span.End()

	metrics, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return metrics, nil
}

// Count returns the total number of metrics in the system.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.metricsbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a metric from the system by its ID.
func (b *Business) QueryByID(ctx context.Context, metricID uuid.UUID) (Metric, error) {
	ctx, span := otel.AddSpan(ctx, "business.metricsbus.querybyid")
	defer span.End()

	metric, err := b.storer.QueryByID(ctx, metricID)
	if err != nil {
		return Metric{}, fmt.Errorf("query by id: %w", err)
	}

	return metric, nil
}
