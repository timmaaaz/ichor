package pageconfigbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("page config not found")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, config PageConfig) error
	Update(ctx context.Context, config PageConfig) error
	Delete(ctx context.Context, configID uuid.UUID) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageReq page.Page) ([]PageConfig, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, configID uuid.UUID) (PageConfig, error)
	QueryByName(ctx context.Context, name string) (PageConfig, error)
	QueryByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (PageConfig, error)
	QueryAll(ctx context.Context) ([]PageConfig, error)

	// Batch validation methods to avoid N+1 queries
	ValidateTableConfigIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]bool, error)
	ValidateFormIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]bool, error)
}

// Business manages the set of APIs for page config access.
type Business struct {
	log            *logger.Logger
	storer         Storer
	del            *delegate.Delegate
	pageContentBus *pagecontentbus.Business
	pageActionBus  *pageactionbus.Business
	validator      *validator.Validate
}

// NewBusiness constructs a page config business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer, pageContentBus *pagecontentbus.Business, pageActionBus *pageactionbus.Business) *Business {
	// Initialize validator
	v := validator.New()

	// Register custom validation tags
	v.RegisterValidation("validContentType", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		validTypes := []string{"table", "form", "chart", "tabs", "container", "text"}
		return contains(validTypes, value)
	})

	return &Business{
		log:            log,
		storer:         storer,
		del:            del,
		pageContentBus: pageContentBus,
		pageActionBus:  pageActionBus,
		validator:      v,
	}
}

// NewWithTx constructs a new Business value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	pageContentBus, err := b.pageContentBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	pageActionBus, err := b.pageActionBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:            b.log,
		storer:         storer,
		del:            b.del,
		pageContentBus: pageContentBus,
		pageActionBus:  pageActionBus,
		validator:      b.validator,
	}

	return &bus, nil
}

// Create adds a new page configuration to the system.
func (b *Business) Create(ctx context.Context, nc NewPageConfig) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Create")
	defer span.End()

	// If is_default is true, ensure user_id is zero (will be NULL in database)
	userID := nc.UserID
	if nc.IsDefault {
		userID = uuid.UUID{}
	}

	config := PageConfig{
		ID:        uuid.New(),
		Name:      nc.Name,
		UserID:    userID,
		IsDefault: nc.IsDefault,
	}

	if err := b.storer.Create(ctx, config); err != nil {
		return PageConfig{}, fmt.Errorf("create: %w", err)
	}

	return config, nil
}

// Update modifies an existing page configuration.
func (b *Business) Update(ctx context.Context, uc UpdatePageConfig, configID uuid.UUID) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Update")
	defer span.End()

	// Fetch existing config
	config, err := b.storer.QueryByID(ctx, configID)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	// Apply updates
	if uc.Name != nil {
		config.Name = *uc.Name
	}
	if uc.UserID != nil {
		config.UserID = *uc.UserID
	}
	if uc.IsDefault != nil {
		config.IsDefault = *uc.IsDefault
	}

	// If is_default is true, ensure user_id is zero
	if config.IsDefault {
		config.UserID = uuid.UUID{}
	}

	if err := b.storer.Update(ctx, config); err != nil {
		return PageConfig{}, fmt.Errorf("update: %w", err)
	}

	return config, nil
}

// Delete removes a page configuration from the system.
func (b *Business) Delete(ctx context.Context, configID uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, configID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of page configurations based on filters.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageReq page.Page) ([]PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Query")
	defer span.End()

	configs, err := b.storer.Query(ctx, filter, orderBy, pageReq)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return configs, nil
}

// Count returns the total number of page configurations matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID finds a page configuration by its ID.
func (b *Business) QueryByID(ctx context.Context, configID uuid.UUID) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.QueryByID")
	defer span.End()

	config, err := b.storer.QueryByID(ctx, configID)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	return config, nil
}

// QueryByName retrieves the default page configuration by name.
// This returns the default page config that serves as a fallback for all users.
func (b *Business) QueryByName(ctx context.Context, name string) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.QueryByName")
	defer span.End()

	config, err := b.storer.QueryByName(ctx, name)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	return config, nil
}

// QueryByNameAndUserID retrieves a user-specific page configuration.
func (b *Business) QueryByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.QueryByNameAndUserID")
	defer span.End()

	config, err := b.storer.QueryByNameAndUserID(ctx, name, userID)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	return config, nil
}

// QueryAll retrieves all page configurations from the system.
func (b *Business) QueryAll(ctx context.Context) ([]PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.queryall")
	defer span.End()

	configs, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryall: %w", err)
	}

	return configs, nil
}

// =============================================================================
// Export/Import Methods

// ExportByIDs exports page configs with their content and actions by IDs.
func (b *Business) ExportByIDs(ctx context.Context, configIDs []uuid.UUID) ([]PageConfigWithRelations, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.exportbyids")
	defer span.End()

	var results []PageConfigWithRelations

	for _, configID := range configIDs {
		config, err := b.storer.QueryByID(ctx, configID)
		if err != nil {
			return nil, fmt.Errorf("query page config %s: %w", configID, err)
		}

		contents, err := b.pageContentBus.QueryByPageConfigID(ctx, configID)
		if err != nil {
			return nil, fmt.Errorf("query contents for page config %s: %w", configID, err)
		}

		actions, err := b.pageActionBus.QueryByPageConfigID(ctx, configID)
		if err != nil {
			return nil, fmt.Errorf("query actions for page config %s: %w", configID, err)
		}

		// Convert to export format
		exportContents := toExportPageContents(contents)
		exportActions := toExportPageActions(actions)

		results = append(results, PageConfigWithRelations{
			PageConfig: config,
			Contents:   exportContents,
			Actions:    exportActions,
		})
	}

	return results, nil
}

// ImportPageConfigs imports page configs with conflict resolution and nested content handling.
func (b *Business) ImportPageConfigs(ctx context.Context, packages []PageConfigWithRelations, mode string) (ImportStats, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.importpageconfigs")
	defer span.End()

	stats := ImportStats{}

	for _, pkg := range packages {
		// Check if config exists by name (considering both default and user-specific configs)
		existing, err := b.storer.QueryByName(ctx, pkg.PageConfig.Name)
		existsAlready := err == nil

		switch mode {
		case "skip":
			if existsAlready {
				stats.SkippedCount++
				continue
			}
			if err := b.createPageConfigWithRelations(ctx, pkg); err != nil {
				return stats, err
			}
			stats.ImportedCount++

		case "replace":
			if existsAlready {
				if err := b.Delete(ctx, existing.ID); err != nil {
					return stats, fmt.Errorf("delete existing: %w", err)
				}
				stats.UpdatedCount++
			}
			if err := b.createPageConfigWithRelations(ctx, pkg); err != nil {
				return stats, err
			}
			if !existsAlready {
				stats.ImportedCount++
			}

		case "merge":
			if existsAlready {
				if err := b.updatePageConfigWithRelations(ctx, existing.ID, pkg); err != nil {
					return stats, fmt.Errorf("update page config: %w", err)
				}
				stats.UpdatedCount++
			} else {
				if err := b.createPageConfigWithRelations(ctx, pkg); err != nil {
					return stats, err
				}
				stats.ImportedCount++
			}
		}
	}

	return stats, nil
}

func (b *Business) createPageConfigWithRelations(ctx context.Context, pkg PageConfigWithRelations) error {
	// Create page config
	newConfig := NewPageConfig{
		Name:      pkg.PageConfig.Name,
		UserID:    pkg.PageConfig.UserID,
		IsDefault: pkg.PageConfig.IsDefault,
	}

	config, err := b.Create(ctx, newConfig)
	if err != nil {
		return fmt.Errorf("create page config: %w", err)
	}

	// Create contents with ID remapping for parent/child relationships
	if err := b.createContentsWithRemapping(ctx, config.ID, pkg.Contents); err != nil {
		return err
	}

	// Create actions
	if err := b.createActions(ctx, config.ID, pkg.Actions); err != nil {
		return err
	}

	return nil
}

func (b *Business) createContentsWithRemapping(ctx context.Context, pageConfigID uuid.UUID, contents []PageContentExport) error {
	// Map old IDs to new IDs for parent/child relationships
	idMap := make(map[uuid.UUID]uuid.UUID)

	// First pass: Create all parent contents (where ParentID is zero/nil)
	for _, content := range contents {
		if content.ParentID == uuid.Nil {
			newContent := pagecontentbus.NewPageContent{
				PageConfigID:  pageConfigID,
				ContentType:   content.ContentType,
				Label:         content.Label,
				TableConfigID: content.TableConfigID,
				FormID:        content.FormID,
				OrderIndex:    content.OrderIndex,
				ParentID:      uuid.UUID{}, // Zero value for no parent
				Layout:        content.Layout,
				IsVisible:     content.IsVisible,
				IsDefault:     content.IsDefault,
			}
			created, err := b.pageContentBus.Create(ctx, newContent)
			if err != nil {
				return fmt.Errorf("create parent content: %w", err)
			}
			idMap[content.ID] = created.ID
		}
	}

	// Second pass: Create child contents with remapped ParentID
	for _, content := range contents {
		if content.ParentID != uuid.Nil {
			// Remap parent ID
			newParentID, ok := idMap[content.ParentID]
			if !ok {
				return fmt.Errorf("parent content %s not found in id map", content.ParentID)
			}

			newContent := pagecontentbus.NewPageContent{
				PageConfigID:  pageConfigID,
				ContentType:   content.ContentType,
				Label:         content.Label,
				TableConfigID: content.TableConfigID,
				FormID:        content.FormID,
				OrderIndex:    content.OrderIndex,
				ParentID:      newParentID,
				Layout:        content.Layout,
				IsVisible:     content.IsVisible,
				IsDefault:     content.IsDefault,
			}
			created, err := b.pageContentBus.Create(ctx, newContent)
			if err != nil {
				return fmt.Errorf("create child content: %w", err)
			}
			idMap[content.ID] = created.ID
		}
	}

	return nil
}

func (b *Business) createActions(ctx context.Context, pageConfigID uuid.UUID, actions PageActionsExport) error {
	// Create button actions
	for _, action := range actions.Buttons {
		if action.Button == nil {
			continue
		}
		newAction := pageactionbus.NewButtonAction{
			PageConfigID:       pageConfigID,
			ActionOrder:        action.ActionOrder,
			IsActive:           action.IsActive,
			Label:              action.Button.Label,
			Icon:               action.Button.Icon,
			TargetPath:         action.Button.TargetPath,
			Variant:            action.Button.Variant,
			Alignment:          action.Button.Alignment,
			ConfirmationPrompt: action.Button.ConfirmationPrompt,
		}
		if _, err := b.pageActionBus.CreateButton(ctx, newAction); err != nil {
			return fmt.Errorf("create button action: %w", err)
		}
	}

	// Create dropdown actions
	for _, action := range actions.Dropdowns {
		if action.Dropdown == nil {
			continue
		}
		// Convert dropdown items
		items := make([]pageactionbus.NewDropdownItem, len(action.Dropdown.Items))
		for i, item := range action.Dropdown.Items {
			items[i] = pageactionbus.NewDropdownItem{
				Label:      item.Label,
				TargetPath: item.TargetPath,
				ItemOrder:  item.ItemOrder,
			}
		}

		newAction := pageactionbus.NewDropdownAction{
			PageConfigID: pageConfigID,
			ActionOrder:  action.ActionOrder,
			IsActive:     action.IsActive,
			Label:        action.Dropdown.Label,
			Icon:         action.Dropdown.Icon,
			Items:        items,
		}
		if _, err := b.pageActionBus.CreateDropdown(ctx, newAction); err != nil {
			return fmt.Errorf("create dropdown action: %w", err)
		}
	}

	// Create separator actions
	for _, action := range actions.Separators {
		newAction := pageactionbus.NewSeparatorAction{
			PageConfigID: pageConfigID,
			ActionOrder:  action.ActionOrder,
			IsActive:     action.IsActive,
		}
		if _, err := b.pageActionBus.CreateSeparator(ctx, newAction); err != nil {
			return fmt.Errorf("create separator action: %w", err)
		}
	}

	return nil
}

func (b *Business) updatePageConfigWithRelations(ctx context.Context, configID uuid.UUID, pkg PageConfigWithRelations) error {
	// Update page config
	updateConfig := UpdatePageConfig{
		Name:      &pkg.PageConfig.Name,
		UserID:    &pkg.PageConfig.UserID,
		IsDefault: &pkg.PageConfig.IsDefault,
	}

	if _, err := b.Update(ctx, updateConfig, configID); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	// Delete existing contents
	existingContents, err := b.pageContentBus.QueryByPageConfigID(ctx, configID)
	if err != nil {
		return fmt.Errorf("query existing contents: %w", err)
	}

	for _, content := range existingContents {
		if err := b.pageContentBus.Delete(ctx, content.ID); err != nil {
			return fmt.Errorf("delete content %s: %w", content.ID, err)
		}
	}

	// Delete existing actions
	existingActions, err := b.pageActionBus.QueryByPageConfigID(ctx, configID)
	if err != nil {
		return fmt.Errorf("query existing actions: %w", err)
	}

	for _, action := range existingActions.Buttons {
		if err := b.pageActionBus.Delete(ctx, action); err != nil {
			return fmt.Errorf("delete button action %s: %w", action.ID, err)
		}
	}
	for _, action := range existingActions.Dropdowns {
		if err := b.pageActionBus.Delete(ctx, action); err != nil {
			return fmt.Errorf("delete dropdown action %s: %w", action.ID, err)
		}
	}
	for _, action := range existingActions.Separators {
		if err := b.pageActionBus.Delete(ctx, action); err != nil {
			return fmt.Errorf("delete separator action %s: %w", action.ID, err)
		}
	}

	// Recreate contents and actions
	if err := b.createContentsWithRemapping(ctx, configID, pkg.Contents); err != nil {
		return err
	}

	if err := b.createActions(ctx, configID, pkg.Actions); err != nil {
		return err
	}

	return nil
}

// Conversion helpers
func toExportPageContents(contents []pagecontentbus.PageContent) []PageContentExport {
	exports := make([]PageContentExport, len(contents))
	for i, c := range contents {
		exports[i] = PageContentExport{
			ID:            c.ID,
			PageConfigID:  c.PageConfigID,
			ContentType:   c.ContentType,
			Label:         c.Label,
			TableConfigID: c.TableConfigID,
			FormID:        c.FormID,
			OrderIndex:    c.OrderIndex,
			ParentID:      c.ParentID,
			Layout:        c.Layout,
			IsVisible:     c.IsVisible,
			IsDefault:     c.IsDefault,
		}
	}
	return exports
}

func toExportPageActions(actions pageactionbus.ActionsGroupedByType) PageActionsExport {
	export := PageActionsExport{
		Buttons:    make([]PageActionExport, len(actions.Buttons)),
		Dropdowns:  make([]PageActionExport, len(actions.Dropdowns)),
		Separators: make([]PageActionExport, len(actions.Separators)),
	}

	for i, action := range actions.Buttons {
		var buttonExport *ButtonActionExport
		if action.Button != nil {
			buttonExport = &ButtonActionExport{
				Label:              action.Button.Label,
				Icon:               action.Button.Icon,
				TargetPath:         action.Button.TargetPath,
				Variant:            action.Button.Variant,
				Alignment:          action.Button.Alignment,
				ConfirmationPrompt: action.Button.ConfirmationPrompt,
			}
		}
		export.Buttons[i] = PageActionExport{
			ID:           action.ID,
			PageConfigID: action.PageConfigID,
			ActionType:   string(action.ActionType),
			ActionOrder:  action.ActionOrder,
			IsActive:     action.IsActive,
			Button:       buttonExport,
		}
	}

	for i, action := range actions.Dropdowns {
		var dropdownExport *DropdownActionExport
		if action.Dropdown != nil {
			items := make([]DropdownItemExport, len(action.Dropdown.Items))
			for j, item := range action.Dropdown.Items {
				items[j] = DropdownItemExport{
					ID:         item.ID,
					Label:      item.Label,
					TargetPath: item.TargetPath,
					ItemOrder:  item.ItemOrder,
				}
			}
			dropdownExport = &DropdownActionExport{
				Label: action.Dropdown.Label,
				Icon:  action.Dropdown.Icon,
				Items: items,
			}
		}
		export.Dropdowns[i] = PageActionExport{
			ID:           action.ID,
			PageConfigID: action.PageConfigID,
			ActionType:   string(action.ActionType),
			ActionOrder:  action.ActionOrder,
			IsActive:     action.IsActive,
			Dropdown:     dropdownExport,
		}
	}

	for i, action := range actions.Separators {
		export.Separators[i] = PageActionExport{
			ID:           action.ID,
			PageConfigID: action.PageConfigID,
			ActionType:   string(action.ActionType),
			ActionOrder:  action.ActionOrder,
			IsActive:     action.IsActive,
		}
	}

	return export
}

// ValidateImportBlob validates a page config JSON blob before import.
func (b *Business) ValidateImportBlob(ctx context.Context, blob []byte) (ValidationResult, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.validateimportblob")
	defer span.End()

	b.log.Info(ctx, "validating page config import blob", "size", len(blob))

	var pkg PageConfigWithRelations
	if err := json.Unmarshal(blob, &pkg); err != nil {
		b.log.Error(ctx, "failed to unmarshal blob", "error", err)
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Field:   "root",
				Message: fmt.Sprintf("invalid JSON: %v", err),
				Code:    ErrCodeInvalidJSON,
			}},
		}, nil
	}

	// Validate structure
	errors := b.validatePageConfigStruct(ctx, pkg)

	result := ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}

	if !result.Valid {
		b.log.Warn(ctx, "validation failed", "error_count", len(errors))
	} else {
		b.log.Info(ctx, "validation successful")
	}

	return result, nil
}

func (b *Business) validatePageConfigStruct(ctx context.Context, pkg PageConfigWithRelations) []ValidationError {
	var errors []ValidationError

	// Required fields
	if pkg.PageConfig.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "page config name is required",
			Code:    ErrCodeRequiredField,
		})
	}

	// Collect all foreign key references for batch validation
	refs := b.collectReferences(pkg.Contents)

	// Validate references in batch (avoid N+1)
	refErrors := b.validateReferences(ctx, refs)
	errors = append(errors, refErrors...)

	// Validate content structure recursively
	for i, content := range pkg.Contents {
		contentErrors := b.validateContent(ctx, content, fmt.Sprintf("contents[%d]", i), 0)
		errors = append(errors, contentErrors...)
	}

	return errors
}

type validationRefs struct {
	tableConfigIDs map[uuid.UUID]bool
	formIDs        map[uuid.UUID]bool
}

func (b *Business) collectReferences(contents []PageContentExport) validationRefs {
	refs := validationRefs{
		tableConfigIDs: make(map[uuid.UUID]bool),
		formIDs:        make(map[uuid.UUID]bool),
	}

	var collect func([]PageContentExport)
	collect = func(items []PageContentExport) {
		for _, item := range items {
			if item.TableConfigID != uuid.Nil {
				refs.tableConfigIDs[item.TableConfigID] = true
			}
			if item.FormID != uuid.Nil {
				refs.formIDs[item.FormID] = true
			}
		}
	}

	collect(contents)
	return refs
}

func (b *Business) validateReferences(ctx context.Context, refs validationRefs) []ValidationError {
	var errors []ValidationError

	// Batch validate table config IDs
	if len(refs.tableConfigIDs) > 0 {
		ids := make([]uuid.UUID, 0, len(refs.tableConfigIDs))
		for id := range refs.tableConfigIDs {
			ids = append(ids, id)
		}

		validIDs, err := b.storer.ValidateTableConfigIDs(ctx, ids)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "contents",
				Message: "failed to validate table config references",
				Code:    ErrCodeInvalidReference,
			})
		} else {
			for _, id := range ids {
				if !validIDs[id] {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("tableConfigId:%s", id),
						Message: "table config does not exist",
						Code:    ErrCodeInvalidReference,
					})
				}
			}
		}
	}

	// Batch validate form IDs
	if len(refs.formIDs) > 0 {
		ids := make([]uuid.UUID, 0, len(refs.formIDs))
		for id := range refs.formIDs {
			ids = append(ids, id)
		}

		validIDs, err := b.storer.ValidateFormIDs(ctx, ids)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "contents",
				Message: "failed to validate form references",
				Code:    ErrCodeInvalidReference,
			})
		} else {
			for _, id := range ids {
				if !validIDs[id] {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("formId:%s", id),
						Message: "form does not exist",
						Code:    ErrCodeInvalidReference,
					})
				}
			}
		}
	}

	return errors
}

func (b *Business) validateContent(ctx context.Context, content PageContentExport, path string, depth int) []ValidationError {
	var errors []ValidationError

	// Check max nesting depth
	if depth > 10 {
		errors = append(errors, ValidationError{
			Field:   path,
			Message: "maximum nesting depth exceeded (10 levels)",
			Code:    ErrCodeMaxDepthExceeded,
		})
		return errors
	}

	// Content type validation
	if content.ContentType == "" {
		errors = append(errors, ValidationError{
			Field:   path + ".contentType",
			Message: "content type is required",
			Code:    ErrCodeRequiredField,
		})
		return errors
	}

	// Type-specific validation for all 6 content types
	switch content.ContentType {
	case "table":
		if content.TableConfigID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   path + ".tableConfigId",
				Message: "table config ID is required for content type 'table'",
				Code:    ErrCodeRequiredField,
			})
		}

	case "form":
		if content.FormID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   path + ".formId",
				Message: "form ID is required for content type 'form'",
				Code:    ErrCodeRequiredField,
			})
		}

	case "chart":
		// Charts use tableConfigId for data source
		if content.TableConfigID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   path + ".tableConfigId",
				Message: "table config ID is required for content type 'chart'",
				Code:    ErrCodeRequiredField,
			})
		}

	case "tabs", "container":
		// Tabs and containers should have children, but we won't enforce it strictly
		// as they might be added later

	case "text":
		// Text content has no special requirements

	default:
		errors = append(errors, ValidationError{
			Field:   path + ".contentType",
			Message: fmt.Sprintf("unknown content type: %s (valid types: table, form, chart, tabs, container, text)", content.ContentType),
			Code:    ErrCodeInvalidType,
		})
	}

	return errors
}

// ImportBlob imports a page config from JSON blob using existing ImportPageConfigs.
func (b *Business) ImportBlob(ctx context.Context, blob []byte, mode string) (ImportStats, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.importblob")
	defer span.End()

	b.log.Info(ctx, "importing page config blob", "mode", mode, "size", len(blob))

	var pkg PageConfigWithRelations
	if err := json.Unmarshal(blob, &pkg); err != nil {
		b.log.Error(ctx, "failed to unmarshal blob", "error", err)
		return ImportStats{}, fmt.Errorf("unmarshal: %w", err)
	}

	// Use existing ImportPageConfigs method which handles:
	// - Transactions
	// - ID remapping for nested content
	// - Conflict resolution (skip/replace/merge)
	// - Multi-table inserts (page_configs → page_content → page_actions)
	stats, err := b.ImportPageConfigs(ctx, []PageConfigWithRelations{pkg}, mode)
	if err != nil {
		b.log.Error(ctx, "import failed", "error", err)
		return ImportStats{}, err
	}

	b.log.Info(ctx, "import successful",
		"imported", stats.ImportedCount,
		"updated", stats.UpdatedCount,
		"skipped", stats.SkippedCount)

	return stats, nil
}
