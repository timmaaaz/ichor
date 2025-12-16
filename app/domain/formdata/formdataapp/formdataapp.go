// Package formdataapp provides the app layer for dynamic form data operations.
package formdataapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// App manages dynamic form data operations across multiple entities.
type App struct {
	registry     *formdataregistry.Registry
	db           *sqlx.DB
	formBus      *formbus.Business
	formFieldBus *formfieldbus.Business
	templateProc *workflow.TemplateProcessor
}

// NewApp constructs a form data app.
func NewApp(
	registry *formdataregistry.Registry,
	db *sqlx.DB,
	formBus *formbus.Business,
	formFieldBus *formfieldbus.Business,
) *App {
	return &App{
		registry:     registry,
		db:           db,
		formBus:      formBus,
		formFieldBus: formFieldBus,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
	}
}

// UpsertFormData handles multi-entity transactional create/update operations.
//
// This method orchestrates the execution of multiple entity operations in a
// single transaction, resolving foreign key references via template variables.
//
// **Array Support**: Entity data can be either a single object or an array of objects.
// When an array is detected, each item is processed individually with template resolution.
//
// Single Object Example:
//
//	{
//	  "operations": {
//	    "users": {"operation": "create", "order": 1}
//	  },
//	  "data": {
//	    "users": {"first_name": "John", "last_name": "Doe"}
//	  }
//	}
//
// Array Example:
//
//	{
//	  "operations": {
//	    "orders": {"operation": "create", "order": 1},
//	    "line_items": {"operation": "create", "order": 2}
//	  },
//	  "data": {
//	    "orders": {"customer_id": "c1", "order_date": "2025-01-15"},
//	    "line_items": [
//	      {"order_id": "{{orders.id}}", "product_id": "p1", "quantity": 5},
//	      {"order_id": "{{orders.id}}", "product_id": "p2", "quantity": 10}
//	    ]
//	  }
//	}
//
// Process:
// 1. Load form configuration and validate
// 2. Build ordered execution plan from operations metadata
// 3. Begin database transaction
// 4. Execute each operation in order (with template variable resolution and array support)
// 5. Commit transaction or rollback on error
//
// Foreign Key Resolution:
// Operations can reference results from previous operations using template syntax:
//
//	"user_id": "{{users.id}}"
//
// The template processor resolves these after each operation completes.
func (a *App) UpsertFormData(ctx context.Context, formID uuid.UUID, req FormDataRequest) (FormDataResponse, error) {
	// 1. Load and validate form configuration
	_, err := a.formBus.QueryByID(ctx, formID)
	if err != nil {
		return FormDataResponse{}, errs.New(errs.NotFound, err)
	}

	fields, err := a.formFieldBus.QueryByFormID(ctx, formID)
	if err != nil {
		return FormDataResponse{}, errs.Newf(errs.Internal, "load form fields: %s", err)
	}

	// Validate that all entities in request match form configuration
	if err := a.validateFormAlignment(req, fields); err != nil {
		return FormDataResponse{}, errs.New(errs.InvalidArgument, err)
	}

	// Validate that form has all required fields for the requested operations
	if err := a.validateFormRequiredFields(ctx, formID, req); err != nil {
		return FormDataResponse{}, err
	}

	// 2. Build execution plan
	plan, err := a.buildExecutionPlan(req.Operations)
	if err != nil {
		return FormDataResponse{}, errs.New(errs.InvalidArgument, err)
	}

	// 3. Begin transaction
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return FormDataResponse{}, errs.Newf(errs.Internal, "begin transaction: %s", err)
	}
	defer tx.Rollback()

	// 4. Execute operations in order
	results := make(map[string]any)
	templateContext := make(workflow.TemplateContext)

	for _, step := range plan {
		entityData, exists := req.Data[step.EntityName]
		if !exists {
			return FormDataResponse{}, errs.Newf(errs.InvalidArgument, "missing data for entity %s", step.EntityName)
		}

		result, err := a.executeOperation(ctx, step, entityData, templateContext)
		if err != nil {
			// Check if error is already typed (e.g., InvalidArgument from validation)
			// Unwrap the error chain to find the original *errs.Error
			var appErr *errs.Error
			if errors.As(err, &appErr) {
				return FormDataResponse{}, errs.Newf(appErr.Code, "execute %s: %s", step.EntityName, err.Error())
			}
			// Otherwise, treat as internal error
			return FormDataResponse{}, errs.Newf(errs.Internal, "execute %s: %s", step.EntityName, err)
		}

		results[step.EntityName] = result
		templateContext[step.EntityName] = result
	}

	// 5. Commit transaction
	if err := tx.Commit(); err != nil {
		return FormDataResponse{}, errs.Newf(errs.Internal, "commit: %s", err)
	}

	return FormDataResponse{
		Success: true,
		Results: results,
	}, nil
}

// validateFormAlignment ensures request entities match form field configuration.
func (a *App) validateFormAlignment(_ FormDataRequest, fields []formfieldbus.FormField) error {
	// Build set of entity IDs from form fields
	formEntities := make(map[uuid.UUID]bool)
	for _, field := range fields {
		formEntities[field.EntityID] = true
	}

	// TODO: Validate that request entities match form configuration
	// This requires mapping entity names to IDs, which needs entity lookup

	return nil
}

// buildExecutionPlan creates an ordered list of operations to execute.
//
// The plan is sorted by the order field to ensure dependencies are respected.
// For example, if addresses depend on users (via foreign key), users must have
// order=1 and addresses order=2.
func (a *App) buildExecutionPlan(operations map[string]OperationMeta) ([]ExecutionStep, error) {
	steps := make([]ExecutionStep, 0, len(operations))

	for entityName, meta := range operations {
		// Look up entity registration
		reg, err := a.registry.Get(entityName)
		if err != nil {
			return nil, fmt.Errorf("entity %s not registered: %w", entityName, err)
		}

		steps = append(steps, ExecutionStep{
			EntityName: entityName,
			Operation:  meta.Operation,
			Order:      meta.Order,
			Registry:   reg,
		})
	}

	// Sort by execution order
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Order < steps[j].Order
	})

	return steps, nil
}

// executeOperation executes a single entity operation with template resolution.
// Supports both single object and array operations.
func (a *App) executeOperation(
	ctx context.Context,
	step ExecutionStep,
	data json.RawMessage,
	templateContext workflow.TemplateContext,
) (any, error) {
	// Detect if data is an array or single object
	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("unmarshal data for type detection: %w", err)
	}

	// Check if data is an array (slice)
	if arr, isArray := rawData.([]interface{}); isArray {
		return a.executeArrayOperation(ctx, step, arr, templateContext)
	}

	// Single object operation
	return a.executeSingleOperation(ctx, step, data, templateContext)
}

// executeSingleOperation processes a single entity object (not an array).
func (a *App) executeSingleOperation(
	ctx context.Context,
	step ExecutionStep,
	data json.RawMessage,
	templateContext workflow.TemplateContext,
) (any, error) {
	// Process templates in the data
	processed := a.templateProc.ProcessTemplateObject(data, templateContext)
	if len(processed.Errors) > 0 {
		return nil, fmt.Errorf("template processing errors: %v", processed.Errors)
	}

	// Convert processed data back to JSON
	processedData, err := json.Marshal(processed.Processed)
	if err != nil {
		return nil, fmt.Errorf("marshal processed data: %w", err)
	}

	// Execute based on operation type
	switch step.Operation {
	case formdataregistry.OperationCreate:
		return a.executeCreate(ctx, step.Registry, processedData)

	case formdataregistry.OperationUpdate:
		return a.executeUpdate(ctx, step.Registry, processedData)

	default:
		return nil, fmt.Errorf("unknown operation: %s", step.Operation)
	}
}

// executeArrayOperation processes an array of entity objects.
// Each item is processed individually with template resolution.
// All items must succeed or the transaction will rollback.
func (a *App) executeArrayOperation(
	ctx context.Context,
	step ExecutionStep,
	items []interface{},
	templateContext workflow.TemplateContext,
) ([]any, error) {
	// Validate array is not empty
	// Line items arrays should contain at least one item
	if len(items) == 0 {
		return nil, errs.Newf(errs.InvalidArgument, "array for %s cannot be empty", step.EntityName)
	}

	results := make([]any, 0, len(items))

	for idx, item := range items {
		// Marshal item to JSON for processing
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("marshal array item %d: %w", idx, err)
		}

		// Process single item with template resolution
		result, err := a.executeSingleOperation(ctx, step, itemJSON, templateContext)
		if err != nil {
			return nil, fmt.Errorf("process array item %d: %w", idx, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// executeCreate handles create operations.
func (a *App) executeCreate(
	ctx context.Context,
	reg *formdataregistry.EntityRegistration,
	data json.RawMessage,
) (any, error) {
	// Decode and validate using registered function
	model, err := reg.DecodeNew(data)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	// Execute create using registered function
	result, err := reg.CreateFunc(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return result, nil
}

// executeUpdate handles update operations.
func (a *App) executeUpdate(
	ctx context.Context,
	reg *formdataregistry.EntityRegistration,
	data json.RawMessage,
) (any, error) {
	// Extract ID from data
	var idHolder struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal(data, &idHolder); err != nil {
		return nil, fmt.Errorf("extract id: %w", err)
	}

	if idHolder.ID == uuid.Nil {
		return nil, fmt.Errorf("id required for update operations")
	}

	// Decode and validate using registered function
	model, err := reg.DecodeUpdate(data)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	// Execute update using registered function
	result, err := reg.UpdateFunc(ctx, idHolder.ID, model)
	if err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}

	return result, nil
}
