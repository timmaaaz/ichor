package tagapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	tagBus *tagbus.Business
	auth   *auth.Auth
}

// NewApp constructs a tag app API for use.
func NewApp(tagBus *tagbus.Business) *App {
	return &App{
		tagBus: tagBus,
	}
}

// NewAppWithAuth constructs a tag app API for use with auth support.
func NewAppWithAuth(tagBus *tagbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:   ath,
		tagBus: tagBus,
	}
}

// Create adds a new tag to the system.
func (a *App) Create(ctx context.Context, app NewTag) (Tag, error) {
	tag, err := a.tagBus.Create(ctx, ToBusNewTag(app))
	if err != nil {
		if errors.Is(err, tagbus.ErrUniqueEntry) {
			return Tag{}, errs.New(errs.Aborted, tagbus.ErrUniqueEntry)
		}
		return Tag{}, errs.Newf(errs.Internal, "create: tag[%+v]: %s", tag, err)
	}

	return ToAppTag(tag), nil
}

// Update updates an existing tag.
func (a *App) Update(ctx context.Context, app UpdateTag, id uuid.UUID) (Tag, error) {
	ut := ToBusUpdateTag(app)

	at, err := a.tagBus.QueryByID(ctx, id)
	if err != nil {
		return Tag{}, errs.Newf(errs.NotFound, "update: tag[%s]: %s", id, err)
	}

	assetCondition, err := a.tagBus.Update(ctx, at, ut)
	if err != nil {
		if errors.Is(err, tagbus.ErrNotFound) {
			return Tag{}, errs.New(errs.NotFound, err)
		}
		return Tag{}, errs.Newf(errs.Internal, "update: tag[%+v]: %s", assetCondition, err)
	}

	return ToAppTag(assetCondition), nil
}

// Delete removes an existing tag.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := a.tagBus.QueryByID(ctx, id)
	if err != nil {
		return errs.Newf(errs.NotFound, "delete: tag[%s]: %s", id, err)
	}

	if err := a.tagBus.Delete(ctx, tag); err != nil {
		return errs.Newf(errs.Internal, "delete: tag[%+v]: %s", tag, err)
	}

	return nil
}

// Query returns a list of tags.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Tag], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Tag]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Tag]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Tag]{}, errs.NewFieldsError("orderby", err)
	}

	tags, err := a.tagBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Tag]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.tagBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Tag]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppTags(tags), total, page), nil
}

// QueryByID returns a single tag based on the id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Tag, error) {
	tag, err := a.tagBus.QueryByID(ctx, id)
	if err != nil {
		return Tag{}, errs.Newf(errs.NotFound, "query: tag[%s]: %s", id, err)
	}

	return ToAppTag(tag), nil
}
