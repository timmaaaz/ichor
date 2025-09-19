package metricsapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/products/metricsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	metricsapp *metricsapp.App
}

func newAPI(metricsApp *metricsapp.App) *api {
	return &api{
		metricsapp: metricsApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app metricsapp.NewMetric
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	metric, err := api.metricsapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return metric
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app metricsapp.UpdateMetric
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	metricID := web.Param(r, "metric_id")
	parsed, err := uuid.Parse(metricID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	metric, err := api.metricsapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return metric
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	metricID := web.Param(r, "metric_id")
	parsed, err := uuid.Parse(metricID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.metricsapp.Delete(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	metrics, err := api.metricsapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return metrics
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	metricID := web.Param(r, "metric_id")
	parsed, err := uuid.Parse(metricID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	metric, err := api.metricsapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return metric
}
