// Package formdataapp provides the app layer for dynamic form data operations.
package formdataapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/calculations"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// maxArrayItems is the maximum number of items allowed in an array operation.
// This prevents DoS attacks via unbounded array submissions.
const maxArrayItems = 1000

// App manages dynamic form data operations across multiple entities.
type App struct {
	log          *logger.Logger
	registry     *formdataregistry.Registry
	db           *sqlx.DB
	formBus      *formbus.Business
	formFieldBus *formfieldbus.Business
	templateProc *workflow.TemplateProcessor
}

// NewApp constructs a form data app.
func NewApp(
	log *logger.Logger,
	registry *formdataregistry.Registry,
	db *sqlx.DB,
	formBus *formbus.Business,
	formFieldBus *formfieldbus.Business,
) *App {
	return &App{
		log:          log,
		registry:     registry,
		db:           db,
		formBus:      formBus,
		formFieldBus: formFieldBus,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
	}
}

// resolveFKByName resolves a human-readable FK value to a UUID.
//
// This method uses the registry's QueryByNameFunc to resolve names to UUIDs,
// following the Ardan Labs store → bus → app → api pattern. No raw SQL is used
// in the app layer.
//
// Returns the UUID unchanged if value is already a valid UUID (fast path).
func (a *App) resolveFKByName(ctx context.Context, dropdownConfig *formfieldbus.DropdownConfig, value string) (uuid.UUID, error) {
	// Fast path: value is already a UUID
	if id, err := uuid.Parse(value); err == nil {
		return id, nil
	}

	// Get entity registration from whitelist
	reg, err := a.registry.Get(dropdownConfig.Entity)
	if err != nil {
		return uuid.Nil, fmt.Errorf("entity %q not registered for FK resolution", dropdownConfig.Entity)
	}

	// Check if entity supports name-based lookup
	if reg.QueryByNameFunc == nil {
		return uuid.Nil, fmt.Errorf("entity %q does not support FK resolution by name", dropdownConfig.Entity)
	}

	// Use the registered business layer function to resolve the name
	id, err := reg.QueryByNameFunc(ctx, value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("value %q not found in %s: %w", value, dropdownConfig.Entity, err)
	}

	return id, nil
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

	// Set up template processor with builtins for $me and $now
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		// If no user ID in context, use nil UUID (anonymous)
		userID = uuid.Nil
	}
	a.templateProc.SetBuiltins(workflow.BuiltinContext{
		UserID:    userID.String(),
		Timestamp: time.Now().UTC(),
	})

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

		// Filter fields for this entity
		entityFields := a.filterFieldsByEntity(fields, step.EntityName)

		// Pass all fields so we can extract line item configs for array operations
		result, err := a.executeOperation(ctx, step, entityData, templateContext, entityFields, fields)
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

	// 4b. Post-processing: Recalculate order totals if applicable
	// This ensures backend is the single source of truth for financial calculations.
	// Frontend calculations are treated as display-only; backend always recalculates.
	if err := a.recalculateOrderTotalsIfNeeded(ctx, results); err != nil {
		// Map validation errors to InvalidArgument (400), others to Internal (500)
		if errors.Is(err, ErrParseInt) || errors.Is(err, ErrParseDecimal) ||
			errors.Is(err, ErrMissingField) || errors.Is(err, ErrInvalidType) ||
			errors.Is(err, calculations.ErrInvalidQuantity) ||
			errors.Is(err, calculations.ErrInvalidUnitPrice) ||
			errors.Is(err, calculations.ErrInvalidDiscount) ||
			errors.Is(err, calculations.ErrInvalidDiscountType) ||
			errors.Is(err, calculations.ErrInvalidTaxRate) ||
			errors.Is(err, calculations.ErrInvalidShippingCost) ||
			errors.Is(err, calculations.ErrOrderTotalExceeded) {
			return FormDataResponse{}, errs.New(errs.InvalidArgument, err)
		}
		return FormDataResponse{}, errs.Newf(errs.Internal, "recalculate order totals: %s", err)
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
// entityFields: fields filtered for this specific entity (used for single object default merging)
// allFields: all form fields (used to extract line item configs for array operations)
func (a *App) executeOperation(
	ctx context.Context,
	step ExecutionStep,
	data json.RawMessage,
	templateContext workflow.TemplateContext,
	entityFields []formfieldbus.FormField,
	allFields []formfieldbus.FormField,
) (any, error) {
	// Detect if data is an array or single object
	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("unmarshal data for type detection: %w", err)
	}

	// Check if data is an array (slice)
	if arr, isArray := rawData.([]interface{}); isArray {
		return a.executeArrayOperation(ctx, step, arr, templateContext, entityFields, allFields)
	}

	// Single object operation
	return a.executeSingleOperation(ctx, step, data, templateContext, entityFields)
}

// executeSingleOperation processes a single entity object (not an array).
func (a *App) executeSingleOperation(
	ctx context.Context,
	step ExecutionStep,
	data json.RawMessage,
	templateContext workflow.TemplateContext,
	entityFields []formfieldbus.FormField,
) (any, error) {
	// Merge field defaults before template processing
	// This injects values like {{$me}} and {{$now}} for audit fields
	mergedData, _, err := a.mergeFieldDefaults(ctx, data, entityFields, step.Operation)
	if err != nil {
		return nil, fmt.Errorf("merge field defaults: %w", err)
	}

	// Process templates in the data (including the injected {{$me}} and {{$now}})
	processed := a.templateProc.ProcessTemplateObject(mergedData, templateContext)
	if len(processed.Errors) > 0 {
		return nil, fmt.Errorf("template processing errors: %v", processed.Errors)
	}

	// Convert processed data to map for validation
	processedMap, ok := processed.Processed.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("processed data is not a map")
	}

	// Validate field constraints (defense in depth)
	for _, field := range entityFields {
		if value, exists := processedMap[field.Name]; exists {
			// Numeric constraints (min/max)
			if err := validateFieldConstraints(field, value); err != nil {
				return nil, err
			}
			// Date constraints (must_be_future, min_date, max_date)
			if err := validateDateConstraints(field, value, processedMap); err != nil {
				return nil, err
			}
		}
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
// allFields is used to extract LineItemField configs for default value injection.
func (a *App) executeArrayOperation(
	ctx context.Context,
	step ExecutionStep,
	items []interface{},
	templateContext workflow.TemplateContext,
	entityFields []formfieldbus.FormField,
	allFields []formfieldbus.FormField,
) ([]any, error) {
	// Validate array is not empty
	// Line items arrays should contain at least one item
	if len(items) == 0 {
		return nil, errs.Newf(errs.InvalidArgument, "array for %s cannot be empty", step.EntityName)
	}

	// Validate array size to prevent DoS attacks
	if len(items) > maxArrayItems {
		return nil, errs.Newf(errs.InvalidArgument, "array for %s exceeds maximum size of %d items", step.EntityName, maxArrayItems)
	}

	// Extract line item field configs from the parent lineitems field
	// These define defaults like Hidden and DefaultValue* for line item fields
	lineItemFields := a.extractLineItemFields(allFields, step.EntityName)

	results := make([]any, 0, len(items))

	for idx, item := range items {
		// Marshal item to JSON for processing
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("marshal array item %d: %w", idx, err)
		}

		// Apply line item field defaults before processing
		// This injects values like {{$me}} and {{$now}} for hidden audit fields
		if len(lineItemFields) > 0 {
			itemJSON, err = a.mergeLineItemFieldDefaults(ctx, itemJSON, lineItemFields, step.Operation)
			if err != nil {
				return nil, fmt.Errorf("merge line item defaults for item %d: %w", idx, err)
			}
		}

		// Process single item with template resolution
		result, err := a.executeSingleOperation(ctx, step, itemJSON, templateContext, entityFields)
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

// InjectionResult tracks which fields were auto-populated with default values.
type InjectionResult struct {
	EntityName     string            `json:"entity_name"`
	InjectedFields map[string]string `json:"injected_fields"` // field -> default value used
}

// mergeFieldDefaults merges default values from field configurations into the data.
// This injects values for fields that have default_value, default_value_create, or
// default_value_update configured, but only if the field is not already provided.
// For FK fields with dropdown config, names are resolved to UUIDs.
func (a *App) mergeFieldDefaults(
	ctx context.Context,
	data json.RawMessage,
	fieldConfigs []formfieldbus.FormField,
	operation formdataregistry.EntityOperation,
) (json.RawMessage, InjectionResult, error) {
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return data, InjectionResult{}, err
	}

	injected := InjectionResult{
		InjectedFields: make(map[string]string),
	}

	for _, field := range fieldConfigs {
		// Parse field config for default values
		var cfg formfieldbus.FieldDefaultConfig
		if err := json.Unmarshal(field.Config, &cfg); err != nil {
			// If config can't be parsed, skip this field
			continue
		}

		// Determine which default to use based on operation
		defaultVal := cfg.DefaultValue
		if operation == formdataregistry.OperationCreate && cfg.DefaultValueCreate != "" {
			defaultVal = cfg.DefaultValueCreate
		} else if operation == formdataregistry.OperationUpdate && cfg.DefaultValueUpdate != "" {
			defaultVal = cfg.DefaultValueUpdate
		}

		if defaultVal == "" {
			continue
		}

		// Only inject if field is not already provided in the data
		if _, exists := dataMap[field.Name]; !exists {
			// Check if field has dropdown config for FK resolution
			var dropdownCfg struct {
				Entity      string `json:"entity"`
				LabelColumn string `json:"label_column"`
				ValueColumn string `json:"value_column"`
			}
			if err := json.Unmarshal(field.Config, &dropdownCfg); err == nil && dropdownCfg.Entity != "" {
				ddConfig := &formfieldbus.DropdownConfig{
					Entity:      dropdownCfg.Entity,
					LabelColumn: dropdownCfg.LabelColumn,
					ValueColumn: dropdownCfg.ValueColumn,
				}
				resolvedID, err := a.resolveFKByName(ctx, ddConfig, defaultVal)
				if err != nil {
					a.log.Warn(ctx, "FK default resolution failed",
						"field", field.Name,
						"entity", dropdownCfg.Entity,
						"default", defaultVal,
						"error", err)
					continue
				}
				defaultVal = resolvedID.String()
			}

			dataMap[field.Name] = defaultVal
			injected.InjectedFields[field.Name] = defaultVal
		}
	}

	result, err := json.Marshal(dataMap)
	return result, injected, err
}

// filterFieldsByEntity filters form fields by entity name (schema.table format).
func (a *App) filterFieldsByEntity(fields []formfieldbus.FormField, entityName string) []formfieldbus.FormField {
	var filtered []formfieldbus.FormField
	for _, field := range fields {
		// Build entity name from schema and table
		fieldEntityName := fmt.Sprintf("%s.%s", field.EntitySchema, field.EntityTable)
		if fieldEntityName == entityName {
			filtered = append(filtered, field)
		}
	}
	return filtered
}

// extractLineItemFields finds LineItemField configs for a given entity from the lineitems field.
// It searches through all form fields looking for a "lineitems" field type that targets
// the specified entity name.
func (a *App) extractLineItemFields(fields []formfieldbus.FormField, entityName string) []formfieldbus.LineItemField {
	for _, field := range fields {
		if field.FieldType == "lineitems" {
			var config formfieldbus.LineItemsFieldConfig
			if err := json.Unmarshal(field.Config, &config); err == nil {
				if config.Entity == entityName {
					return config.Fields
				}
			}
		}
	}
	return nil
}

// mergeLineItemFieldDefaults injects default values for line item fields.
// This is similar to mergeFieldDefaults but operates on LineItemField configs
// instead of FormField configs. For FK fields with dropdown config, names are resolved to UUIDs.
func (a *App) mergeLineItemFieldDefaults(
	ctx context.Context,
	data json.RawMessage,
	lineItemFields []formfieldbus.LineItemField,
	operation formdataregistry.EntityOperation,
) (json.RawMessage, error) {
	if len(lineItemFields) == 0 {
		return data, nil
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return data, err
	}

	for _, field := range lineItemFields {
		// Determine which default to use based on operation
		defaultVal := field.DefaultValue
		if operation == formdataregistry.OperationCreate && field.DefaultValueCreate != "" {
			defaultVal = field.DefaultValueCreate
		} else if operation == formdataregistry.OperationUpdate && field.DefaultValueUpdate != "" {
			defaultVal = field.DefaultValueUpdate
		}

		if defaultVal == "" {
			continue
		}

		// Only inject if field is not already provided in the data
		if _, exists := dataMap[field.Name]; !exists {
			// Check if field has dropdown config for FK resolution
			if field.DropdownConfig != nil && field.DropdownConfig.Entity != "" {
				resolvedID, err := a.resolveFKByName(ctx, field.DropdownConfig, defaultVal)
				if err != nil {
					a.log.Warn(ctx, "FK default resolution failed for line item field",
						"field", field.Name,
						"entity", field.DropdownConfig.Entity,
						"default", defaultVal,
						"error", err)
					continue
				}
				defaultVal = resolvedID.String()
			}

			dataMap[field.Name] = defaultVal
		}
	}

	return json.Marshal(dataMap)
}

// ============================================================================
// Order Total Calculation Post-Processing
// ============================================================================

// recalculateOrderTotalsIfNeeded checks if both sales.orders and sales.order_line_items
// are in results, and if so, recalculates and updates the order totals.
//
// This ensures the backend is the single source of truth for order totals.
// Frontend-submitted totals are treated as display-only and always overridden
// with backend-calculated values.
//
// Design Decision: Why formdataapp instead of ordersbus?
// When ordersbus.Create() is called, line items haven't been created yet.
// This post-processing hook runs after all entities are created but before
// the transaction commits, giving us access to both the order and its line items.
func (a *App) recalculateOrderTotalsIfNeeded(ctx context.Context, results map[string]any) error {
	orderResult, hasOrder := results["sales.orders"]
	lineItemsResult, hasLineItems := results["sales.order_line_items"]

	if !hasOrder || !hasLineItems {
		return nil // Nothing to recalculate
	}

	// Type assertion with proper error handling (NOT silent failure)
	order, ok := orderResult.(map[string]any)
	if !ok {
		return fmt.Errorf("order result has unexpected type: %T", orderResult)
	}

	lineItems, ok := lineItemsResult.([]any)
	if !ok {
		return fmt.Errorf("line items result has unexpected type: %T (expected array)", lineItemsResult)
	}

	// Handle empty line items array
	if len(lineItems) == 0 {
		return nil // No line items to calculate from
	}

	// Convert line items to calculation input with error handling
	calcItems := make([]calculations.LineItemInput, 0, len(lineItems))
	for idx, item := range lineItems {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: line item %d has type %T", ErrInvalidType, idx, item)
		}

		// Validate required fields exist
		for _, field := range []string{"quantity", "unit_price"} {
			if _, exists := itemMap[field]; !exists {
				return fmt.Errorf("%w: line item %d missing %q", ErrMissingField, idx, field)
			}
		}

		// Parse with error handling
		quantity, err := parseIntFromAny(itemMap["quantity"])
		if err != nil {
			return fmt.Errorf("line item %d quantity: %w", idx, err)
		}
		unitPrice, err := parseDecimalFromAny(itemMap["unit_price"])
		if err != nil {
			return fmt.Errorf("line item %d unit_price: %w", idx, err)
		}
		discount, err := parseDecimalFromAny(itemMap["discount"])
		if err != nil {
			return fmt.Errorf("line item %d discount: %w", idx, err)
		}
		discountType, err := parseStringFromAny(itemMap["discount_type"])
		if err != nil {
			return fmt.Errorf("line item %d discount_type: %w", idx, err)
		}

		calcItems = append(calcItems, calculations.LineItemInput{
			Quantity:     quantity,
			UnitPrice:    unitPrice,
			Discount:     discount,
			DiscountType: discountType,
		})
	}

	// Get tax rate and shipping from order with error handling
	taxRate, err := parseDecimalFromAny(order["tax_rate"])
	if err != nil {
		return fmt.Errorf("order tax_rate: %w", err)
	}
	shippingCost, err := parseDecimalFromAny(order["shipping_cost"])
	if err != nil {
		return fmt.Errorf("order shipping_cost: %w", err)
	}

	// Calculate totals
	calculated, err := calculations.CalculateOrderTotals(calcItems, taxRate, shippingCost)
	if err != nil {
		return fmt.Errorf("calculate order totals: %w", err)
	}

	// Log if significantly different from submitted values (helps debug frontend bugs)
	submittedSubtotal, _ := parseDecimalFromAny(order["subtotal"]) // Ignore error - logging is best-effort
	diff := calculated.Subtotal.Sub(submittedSubtotal).Abs()
	if diff.GreaterThan(decimal.NewFromFloat(0.01)) {
		a.log.Info(ctx, "order totals recalculated",
			"order_id", order["id"],
			"submitted_subtotal", submittedSubtotal.String(),
			"calculated_subtotal", calculated.Subtotal.String(),
			"difference", diff.String(),
		)
	}

	// Update order with calculated values
	orderID, ok := order["id"].(string)
	if !ok {
		return fmt.Errorf("order id not found or not string: %T", order["id"])
	}

	return a.updateOrderTotals(ctx, orderID, calculated)
}

// updateOrderTotals updates the order record with calculated totals.
func (a *App) updateOrderTotals(ctx context.Context, orderID string, totals calculations.OrderTotals) error {
	reg, err := a.registry.Get("sales.orders")
	if err != nil {
		return fmt.Errorf("get orders registry: %w", err)
	}

	updateData := map[string]any{
		"id":           orderID,
		"subtotal":     totals.Subtotal.String(),
		"tax_amount":   totals.TaxAmount.String(),
		"total_amount": totals.Total.String(),
	}

	updateJSON, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("marshal update data: %w", err)
	}

	_, err = a.executeUpdate(ctx, reg, updateJSON)
	return err
}

// ============================================================================
// Field Validation (Defense in Depth)
// ============================================================================

// validateFieldConstraints checks field values against configured min/max constraints.
// Returns an error if validation fails.
func validateFieldConstraints(
	fieldConfig formfieldbus.FormField,
	value interface{},
) error {
	// Parse field config for validation rules
	var cfg struct {
		Min *int `json:"min"`
		Max *int `json:"max"`
	}
	if err := json.Unmarshal(fieldConfig.Config, &cfg); err != nil {
		return nil // If config can't be parsed, skip validation
	}

	if cfg.Min == nil && cfg.Max == nil {
		return nil // No numeric constraints configured
	}

	numVal, err := toFloat64(value)
	if err != nil {
		return nil // Non-numeric values handled elsewhere
	}

	if cfg.Min != nil && numVal < float64(*cfg.Min) {
		return errs.Newf(errs.InvalidArgument,
			"%s must be at least %d", fieldConfig.Name, *cfg.Min)
	}
	if cfg.Max != nil && numVal > float64(*cfg.Max) {
		return errs.Newf(errs.InvalidArgument,
			"%s must be at most %d", fieldConfig.Name, *cfg.Max)
	}
	return nil
}

// validateDateConstraints checks date field values against configured constraints.
// Supports: must_be_future, min_date ("today", "{{field}}", ISO date), max_date.
func validateDateConstraints(
	fieldConfig formfieldbus.FormField,
	value interface{},
	allValues map[string]interface{},
) error {
	// Parse field config for date validation rules
	var cfg struct {
		MustBeFuture bool   `json:"must_be_future"`
		MinDate      string `json:"min_date"`
		MaxDate      string `json:"max_date"`
	}
	if err := json.Unmarshal(fieldConfig.Config, &cfg); err != nil {
		return nil // If config can't be parsed, skip validation
	}

	if !cfg.MustBeFuture && cfg.MinDate == "" && cfg.MaxDate == "" {
		return nil // No date constraints configured
	}

	dateVal, err := toTime(value)
	if err != nil {
		return nil // Non-date values handled elsewhere
	}

	// Check must_be_future
	if cfg.MustBeFuture {
		today := time.Now().Truncate(24 * time.Hour)
		if dateVal.Before(today) {
			return errs.Newf(errs.InvalidArgument,
				"%s must be in the future", fieldConfig.Name)
		}
	}

	// Check min_date (can be "today", "{{field}}", or ISO date)
	if cfg.MinDate != "" {
		minDate, err := resolveDateConstraint(cfg.MinDate, allValues)
		if err == nil && dateVal.Before(minDate) {
			return errs.Newf(errs.InvalidArgument,
				"%s must be on or after %s", fieldConfig.Name, minDate.Format("2006-01-02"))
		}
	}

	// Check max_date
	if cfg.MaxDate != "" {
		maxDate, err := resolveDateConstraint(cfg.MaxDate, allValues)
		if err == nil && dateVal.After(maxDate) {
			return errs.Newf(errs.InvalidArgument,
				"%s must be on or before %s", fieldConfig.Name, maxDate.Format("2006-01-02"))
		}
	}

	return nil
}

// resolveDateConstraint resolves a date constraint string to a time.Time.
// Supports: "today", "{{field_name}}", ISO date strings.
func resolveDateConstraint(constraint string, allValues map[string]interface{}) (time.Time, error) {
	if constraint == "today" {
		return time.Now().Truncate(24 * time.Hour), nil
	}

	// Check for field reference: {{field_name}}
	if len(constraint) > 4 && constraint[:2] == "{{" && constraint[len(constraint)-2:] == "}}" {
		fieldName := constraint[2 : len(constraint)-2]
		if fieldValue, ok := allValues[fieldName]; ok {
			return toTime(fieldValue)
		}
		return time.Time{}, fmt.Errorf("field %s not found", fieldName)
	}

	// Try parsing as ISO date
	return time.Parse("2006-01-02", constraint)
}

// toFloat64 converts various numeric types to float64.
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return decimal.RequireFromString(v).InexactFloat64(), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// toTime converts various date types to time.Time.
func toTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try multiple formats
		for _, format := range []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05Z"} {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse %s as date", v)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", value)
	}
}
