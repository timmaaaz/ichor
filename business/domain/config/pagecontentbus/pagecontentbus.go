package pagecontentbus

import (
	"context"
	"errors"
	"fmt"

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
	ErrNotFound           = errors.New("page content not found")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrMissingContentRef  = errors.New("missing required content reference")
	ErrOrphanTab          = errors.New("tab content must have parent container")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, content PageContent) error
	Update(ctx context.Context, content PageContent) error
	Delete(ctx context.Context, contentID uuid.UUID) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageReq page.Page) ([]PageContent, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, contentID uuid.UUID) (PageContent, error)
	QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) ([]PageContent, error)
	QueryWithChildren(ctx context.Context, pageConfigID uuid.UUID) ([]PageContent, error)
}

// Business manages the set of APIs for page content access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a page content business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
		del:    del,
	}
}

// NewWithTx constructs a new Business value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
		del:    b.del,
	}

	return &bus, nil
}

// Create adds a new page content block to the system.
func (b *Business) Create(ctx context.Context, nc NewPageContent) (PageContent, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.Create")
	defer span.End()

	// Validate content type
	if err := validateContentType(nc.ContentType); err != nil {
		return PageContent{}, err
	}

	// Validate content references match content type
	if err := validateContentReferences(nc.ContentType, nc.TableConfigID, nc.FormID, nc.ChartConfigID); err != nil {
		return PageContent{}, err
	}

	content := PageContent{
		ID:            uuid.New(),
		PageConfigID:  nc.PageConfigID,
		ContentType:   nc.ContentType,
		Label:         nc.Label,
		TableConfigID: nc.TableConfigID,
		FormID:        nc.FormID,
		ChartConfigID: nc.ChartConfigID,
		OrderIndex:    nc.OrderIndex,
		ParentID:      nc.ParentID,
		Layout:        nc.Layout,
		IsVisible:     nc.IsVisible,
		IsDefault:     nc.IsDefault,
	}

	if err := b.storer.Create(ctx, content); err != nil {
		return PageContent{}, fmt.Errorf("create: %w", err)
	}

	return content, nil
}

// Update modifies an existing page content block.
func (b *Business) Update(ctx context.Context, uc UpdatePageContent, contentID uuid.UUID) (PageContent, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.Update")
	defer span.End()

	// Fetch existing content
	content, err := b.storer.QueryByID(ctx, contentID)
	if err != nil {
		return PageContent{}, fmt.Errorf("query: %w", err)
	}

	// Apply updates
	if uc.Label != nil {
		content.Label = *uc.Label
	}
	if uc.OrderIndex != nil {
		content.OrderIndex = *uc.OrderIndex
	}
	if uc.Layout != nil {
		content.Layout = *uc.Layout
	}
	if uc.IsVisible != nil {
		content.IsVisible = *uc.IsVisible
	}
	if uc.IsDefault != nil {
		content.IsDefault = *uc.IsDefault
	}

	if err := b.storer.Update(ctx, content); err != nil {
		return PageContent{}, fmt.Errorf("update: %w", err)
	}

	return content, nil
}

// Delete removes a page content block from the system.
func (b *Business) Delete(ctx context.Context, contentID uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, contentID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of page content blocks based on filters.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageReq page.Page) ([]PageContent, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.Query")
	defer span.End()

	contents, err := b.storer.Query(ctx, filter, orderBy, pageReq)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return contents, nil
}

// Count returns the total number of page content blocks matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.Count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID finds a page content block by its ID.
func (b *Business) QueryByID(ctx context.Context, contentID uuid.UUID) (PageContent, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.QueryByID")
	defer span.End()

	content, err := b.storer.QueryByID(ctx, contentID)
	if err != nil {
		return PageContent{}, fmt.Errorf("query: %w", err)
	}

	return content, nil
}

// QueryByPageConfigID retrieves all content blocks for a specific page config.
func (b *Business) QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) ([]PageContent, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.QueryByPageConfigID")
	defer span.End()

	contents, err := b.storer.QueryByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return contents, nil
}

// QueryWithChildren retrieves content blocks and nests children under parents.
// This is especially useful for tabs, where tab items are children of a tabs container.
func (b *Business) QueryWithChildren(ctx context.Context, pageConfigID uuid.UUID) ([]PageContent, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagecontentbus.QueryWithChildren")
	defer span.End()

	contents, err := b.storer.QueryWithChildren(ctx, pageConfigID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return contents, nil
}

// =============================================================================
// Validation Functions
// =============================================================================

// validateContentType checks if the content type is valid
func validateContentType(contentType string) error {
	validTypes := map[string]bool{
		ContentTypeTable:     true,
		ContentTypeForm:      true,
		ContentTypeTabs:      true,
		ContentTypeContainer: true,
		ContentTypeText:      true,
		ContentTypeChart:     true,
	}

	if !validTypes[contentType] {
		return fmt.Errorf("%w: %s", ErrInvalidContentType, contentType)
	}

	return nil
}

// validateContentReferences ensures content type matches required references
func validateContentReferences(contentType string, tableConfigID, formID, chartConfigID uuid.UUID) error {
	switch contentType {
	case ContentTypeTable:
		if tableConfigID == uuid.Nil {
			return fmt.Errorf("%w: table content type requires tableConfigID", ErrMissingContentRef)
		}
	case ContentTypeForm:
		if formID == uuid.Nil {
			return fmt.Errorf("%w: form content type requires formID", ErrMissingContentRef)
		}
	case ContentTypeChart:
		if chartConfigID == uuid.Nil {
			return fmt.Errorf("%w: chart content type requires chartConfigID", ErrMissingContentRef)
		}
	}

	return nil
}
