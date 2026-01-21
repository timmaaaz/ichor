package formdataapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// TestResolveFKByName_UUIDPassthrough verifies that valid UUIDs are returned
// unchanged without any database lookup.
func TestResolveFKByName_UUIDPassthrough(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	// Create a registry with a test entity
	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	// Test with a valid UUID string
	testUUID := uuid.New()
	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "test.entity",
		LabelColumn: "name",
		ValueColumn: "id",
	}

	// When value is already a valid UUID, it should pass through unchanged
	result, err := app.resolveFKByName(ctx, dropdownConfig, testUUID.String())
	if err != nil {
		t.Fatalf("unexpected error for UUID passthrough: %v", err)
	}

	if result != testUUID {
		t.Errorf("UUID passthrough failed: got %v, want %v", result, testUUID)
	}
}

// TestResolveFKByName_UnknownEntity verifies that unregistered entities are
// rejected before any database query is attempted.
func TestResolveFKByName_UnknownEntity(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	// Create an empty registry (no entities registered)
	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "unregistered.entity",
		LabelColumn: "name",
		ValueColumn: "id",
	}

	// Non-UUID value with unregistered entity should fail
	_, err := app.resolveFKByName(ctx, dropdownConfig, "Pending")
	if err == nil {
		t.Error("expected error for unregistered entity, got nil")
	}

	// Error should mention the entity is not registered
	if err != nil && !containsSubstring(err.Error(), "not registered") {
		t.Errorf("error should mention entity not registered, got: %v", err)
	}
}

// TestResolveFKByName_EntityWithoutQueryByNameFunc verifies that entities without
// QueryByNameFunc return an appropriate error.
func TestResolveFKByName_EntityWithoutQueryByNameFunc(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	// Create registry with an entity that doesn't have QueryByNameFunc
	registry := formdataregistry.New()
	registry.Register(formdataregistry.EntityRegistration{
		Name: "test.entity",
		// No QueryByNameFunc set
	})

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "test.entity",
		LabelColumn: "name",
		ValueColumn: "id",
	}

	// Non-UUID value with entity that lacks QueryByNameFunc should fail
	_, err := app.resolveFKByName(ctx, dropdownConfig, "Pending")
	if err == nil {
		t.Error("expected error for entity without QueryByNameFunc, got nil")
	}

	// Error should mention the entity doesn't support name resolution
	if err != nil && !containsSubstring(err.Error(), "does not support FK resolution") {
		t.Errorf("error should mention entity doesn't support FK resolution, got: %v", err)
	}
}

// TestResolveFKByName_SuccessfulResolution verifies that when an entity has
// QueryByNameFunc, it's called and the result is returned.
func TestResolveFKByName_SuccessfulResolution(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	expectedID := uuid.New()

	// Create registry with an entity that has a mock QueryByNameFunc
	registry := formdataregistry.New()
	registry.Register(formdataregistry.EntityRegistration{
		Name: "test.statuses",
		QueryByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			if name == "Pending" {
				return expectedID, nil
			}
			return uuid.Nil, nil
		},
	})

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "test.statuses",
		LabelColumn: "name",
		ValueColumn: "id",
	}

	// Should successfully resolve "Pending" to the expected UUID
	result, err := app.resolveFKByName(ctx, dropdownConfig, "Pending")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expectedID {
		t.Errorf("expected resolved ID %v, got %v", expectedID, result)
	}
}

// TestResolveFKByName_UUIDPassthrough_SkipsRegistryCheck verifies that UUID values
// bypass the registry check entirely (performance optimization).
func TestResolveFKByName_UUIDPassthrough_SkipsRegistryCheck(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	// Create an empty registry (entity NOT registered)
	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	testUUID := uuid.New()
	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "unregistered.entity", // NOT in registry
		LabelColumn: "name",
		ValueColumn: "id",
	}

	// UUID should pass through even if entity is not registered
	// This is the fast path - no registry lookup needed for UUIDs
	result, err := app.resolveFKByName(ctx, dropdownConfig, testUUID.String())
	if err != nil {
		t.Fatalf("UUID passthrough should not check registry: %v", err)
	}

	if result != testUUID {
		t.Errorf("UUID passthrough failed: got %v, want %v", result, testUUID)
	}
}

// TestResolveFKByName_InvalidUUID verifies that invalid UUID strings are not
// mistakenly parsed as valid UUIDs.
func TestResolveFKByName_InvalidUUID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	// Create an empty registry
	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "unregistered.entity",
		LabelColumn: "name",
		ValueColumn: "id",
	}

	testCases := []string{
		"Pending",                         // Plain text
		"not-a-uuid",                      // Invalid format
		"12345678-1234-1234-1234-123456",  // Too short
		"12345678-1234-1234-1234-12345678901x", // Invalid character
		"",                                // Empty string
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// These should NOT be parsed as valid UUIDs
			// They should fail because the entity is not registered
			_, err := app.resolveFKByName(ctx, dropdownConfig, tc)
			if err == nil {
				t.Errorf("expected error for non-UUID value %q with unregistered entity", tc)
			}
		})
	}
}

// =============================================================================
// mergeFieldDefaults FK Resolution Integration Tests
// =============================================================================

// TestMergeFieldDefaults_FKResolutionAttempted verifies that mergeFieldDefaults
// attempts FK resolution when a field has dropdown config with entity.
// Since we don't have a real DB, this test verifies that:
// 1. Fields WITHOUT entity config get the default value as-is
// 2. Fields WITH entity config attempt resolution (and skip due to unregistered entity)
func TestMergeFieldDefaults_FKResolutionAttempted(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	// Field WITHOUT entity config - should inject default value directly
	fieldWithoutEntity := formfieldbus.FormField{
		Name:   "notes",
		Config: []byte(`{"default_value_create": "Default notes"}`),
	}

	// Field WITH entity config - should attempt FK resolution
	// Since entity is not registered, it will be skipped (logged as warning)
	fieldWithEntity := formfieldbus.FormField{
		Name:   "status_id",
		Config: []byte(`{"entity": "unregistered.statuses", "label_column": "name", "value_column": "id", "default_value_create": "Pending"}`),
	}

	fields := []formfieldbus.FormField{fieldWithoutEntity, fieldWithEntity}

	// Empty input data - both defaults should be considered
	inputData := []byte(`{}`)

	result, injected, err := app.mergeFieldDefaults(ctx, inputData, fields, formdataregistry.OperationCreate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Field WITHOUT entity should be injected
	if notes, ok := resultMap["notes"]; !ok || notes != "Default notes" {
		t.Errorf("expected notes to be 'Default notes', got %v", resultMap["notes"])
	}

	// Field WITH entity should NOT be injected (entity not registered)
	// The FK resolution fails and the field is skipped
	if _, exists := resultMap["status_id"]; exists {
		t.Errorf("expected status_id NOT to be injected when entity is not registered, but it was: %v", resultMap["status_id"])
	}

	// Verify only notes was injected
	if len(injected.InjectedFields) != 1 {
		t.Errorf("expected 1 injected field, got %d: %v", len(injected.InjectedFields), injected.InjectedFields)
	}
	if injected.InjectedFields["notes"] != "Default notes" {
		t.Errorf("expected notes in injected fields, got: %v", injected.InjectedFields)
	}

	// Verify warning was logged for FK resolution failure
	logOutput := buf.String()
	if !containsSubstring(logOutput, "FK default resolution failed") {
		t.Errorf("expected warning log for FK resolution failure, got: %s", logOutput)
	}
}

// TestMergeFieldDefaults_FKResolutionSuccess verifies that mergeFieldDefaults
// successfully resolves FK defaults when the entity has QueryByNameFunc.
func TestMergeFieldDefaults_FKResolutionSuccess(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	expectedStatusID := uuid.New()

	// Create registry with an entity that has QueryByNameFunc
	registry := formdataregistry.New()
	registry.Register(formdataregistry.EntityRegistration{
		Name: "sales.statuses",
		QueryByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			if name == "Pending" {
				return expectedStatusID, nil
			}
			return uuid.Nil, nil
		},
	})

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	// Field WITH entity config and QueryByNameFunc registered
	fieldWithEntity := formfieldbus.FormField{
		Name:   "status_id",
		Config: []byte(`{"entity": "sales.statuses", "label_column": "name", "value_column": "id", "default_value_create": "Pending"}`),
	}

	fields := []formfieldbus.FormField{fieldWithEntity}
	inputData := []byte(`{}`)

	result, injected, err := app.mergeFieldDefaults(ctx, inputData, fields, formdataregistry.OperationCreate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Field should be injected with the resolved UUID
	if statusID, ok := resultMap["status_id"]; !ok || statusID != expectedStatusID.String() {
		t.Errorf("expected status_id to be %s, got %v", expectedStatusID, resultMap["status_id"])
	}

	// Verify it was tracked as injected
	if injected.InjectedFields["status_id"] != expectedStatusID.String() {
		t.Errorf("expected status_id in injected fields with value %s, got: %v", expectedStatusID, injected.InjectedFields)
	}
}

// TestMergeFieldDefaults_UUIDDefaultPassthrough verifies that when the default value
// is already a UUID, it passes through without DB lookup (even if entity is not registered).
func TestMergeFieldDefaults_UUIDDefaultPassthrough(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	testUUID := uuid.New().String()

	// Field with entity config but UUID default value
	fieldWithUUIDDefault := formfieldbus.FormField{
		Name:   "status_id",
		Config: []byte(`{"entity": "unregistered.statuses", "label_column": "name", "value_column": "id", "default_value_create": "` + testUUID + `"}`),
	}

	fields := []formfieldbus.FormField{fieldWithUUIDDefault}
	inputData := []byte(`{}`)

	result, injected, err := app.mergeFieldDefaults(ctx, inputData, fields, formdataregistry.OperationCreate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// UUID default should be injected (passes through without DB lookup)
	if statusID, ok := resultMap["status_id"]; !ok || statusID != testUUID {
		t.Errorf("expected status_id to be %s, got %v", testUUID, resultMap["status_id"])
	}

	// Verify it was injected
	if injected.InjectedFields["status_id"] != testUUID {
		t.Errorf("expected status_id in injected fields with value %s, got: %v", testUUID, injected.InjectedFields)
	}

	// No warning should be logged (UUID passthrough doesn't check registry)
	logOutput := buf.String()
	if containsSubstring(logOutput, "FK default resolution failed") {
		t.Errorf("unexpected warning log for UUID passthrough: %s", logOutput)
	}
}

// TestMergeFieldDefaults_ExistingValueNotOverwritten verifies that user-provided
// values are never overwritten by defaults.
func TestMergeFieldDefaults_ExistingValueNotOverwritten(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	userProvidedUUID := uuid.New().String()

	fieldWithDefault := formfieldbus.FormField{
		Name:   "status_id",
		Config: []byte(`{"entity": "some.entity", "label_column": "name", "value_column": "id", "default_value_create": "Pending"}`),
	}

	fields := []formfieldbus.FormField{fieldWithDefault}

	// User already provided status_id
	inputData := []byte(`{"status_id": "` + userProvidedUUID + `"}`)

	result, injected, err := app.mergeFieldDefaults(ctx, inputData, fields, formdataregistry.OperationCreate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// User's value should be preserved
	if statusID, ok := resultMap["status_id"]; !ok || statusID != userProvidedUUID {
		t.Errorf("expected user's status_id %s to be preserved, got %v", userProvidedUUID, resultMap["status_id"])
	}

	// Nothing should be injected (user provided the value)
	if len(injected.InjectedFields) != 0 {
		t.Errorf("expected no injected fields when user provided value, got: %v", injected.InjectedFields)
	}
}

// TestMergeLineItemFieldDefaults_FKResolutionAttempted verifies that
// mergeLineItemFieldDefaults attempts FK resolution for line item fields.
func TestMergeLineItemFieldDefaults_FKResolutionAttempted(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	registry := formdataregistry.New()

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	// Line item field WITH dropdown config
	fieldWithDropdown := formfieldbus.LineItemField{
		Name:               "fulfillment_status_id",
		DefaultValueCreate: "Pending",
		DropdownConfig: &formfieldbus.DropdownConfig{
			Entity:      "unregistered.statuses",
			LabelColumn: "name",
			ValueColumn: "id",
		},
	}

	// Line item field WITHOUT dropdown config
	fieldWithoutDropdown := formfieldbus.LineItemField{
		Name:               "notes",
		DefaultValueCreate: "Line item note",
	}

	fields := []formfieldbus.LineItemField{fieldWithDropdown, fieldWithoutDropdown}
	inputData := []byte(`{"product_id": "some-uuid"}`)

	result, err := app.mergeLineItemFieldDefaults(ctx, inputData, fields, formdataregistry.OperationCreate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Field WITHOUT dropdown should be injected
	if notes, ok := resultMap["notes"]; !ok || notes != "Line item note" {
		t.Errorf("expected notes to be 'Line item note', got %v", resultMap["notes"])
	}

	// Field WITH dropdown should NOT be injected (entity not registered)
	if _, exists := resultMap["fulfillment_status_id"]; exists {
		t.Errorf("expected fulfillment_status_id NOT to be injected when entity not registered, but it was: %v", resultMap["fulfillment_status_id"])
	}

	// Verify warning was logged
	logOutput := buf.String()
	if !containsSubstring(logOutput, "FK default resolution failed for line item field") {
		t.Errorf("expected warning log for FK resolution failure, got: %s", logOutput)
	}
}

// TestResolveFKByName_CurrencyResolution verifies that currency FK resolution
// works correctly using QueryByCode.
func TestResolveFKByName_CurrencyResolution(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	expectedCurrencyID := uuid.New()

	// Create registry with currency entity that has QueryByNameFunc
	registry := formdataregistry.New()
	registry.Register(formdataregistry.EntityRegistration{
		Name: "core.currencies",
		QueryByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			if name == "USD" {
				return expectedCurrencyID, nil
			}
			return uuid.Nil, fmt.Errorf("currency with code %q not found", name)
		},
	})

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	dropdownConfig := &formfieldbus.DropdownConfig{
		Entity:      "core.currencies",
		LabelColumn: "name",
		ValueColumn: "id",
	}

	// Should successfully resolve "USD" to the expected UUID
	result, err := app.resolveFKByName(ctx, dropdownConfig, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expectedCurrencyID {
		t.Errorf("expected resolved ID %v, got %v", expectedCurrencyID, result)
	}
}

// TestMergeFieldDefaults_CurrencyFKResolutionSuccess verifies that mergeFieldDefaults
// successfully resolves currency FK defaults when the entity has QueryByNameFunc.
func TestMergeFieldDefaults_CurrencyFKResolutionSuccess(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })

	expectedCurrencyID := uuid.New()

	// Create registry with currency entity that has QueryByNameFunc
	registry := formdataregistry.New()
	registry.Register(formdataregistry.EntityRegistration{
		Name: "core.currencies",
		QueryByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			if name == "USD" {
				return expectedCurrencyID, nil
			}
			return uuid.Nil, fmt.Errorf("currency with code %q not found", name)
		},
	})

	app := &App{
		log:      log,
		registry: registry,
	}

	ctx := context.Background()

	// Field WITH currency entity config and QueryByNameFunc registered
	fieldWithCurrency := formfieldbus.FormField{
		Name:   "currency_id",
		Config: []byte(`{"entity": "core.currencies", "label_column": "name", "value_column": "id", "default_value_create": "USD"}`),
	}

	fields := []formfieldbus.FormField{fieldWithCurrency}
	inputData := []byte(`{}`)

	result, injected, err := app.mergeFieldDefaults(ctx, inputData, fields, formdataregistry.OperationCreate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse result
	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Field should be injected with the resolved UUID
	if currencyID, ok := resultMap["currency_id"]; !ok || currencyID != expectedCurrencyID.String() {
		t.Errorf("expected currency_id to be %s, got %v", expectedCurrencyID, resultMap["currency_id"])
	}

	// Verify it was tracked as injected
	if injected.InjectedFields["currency_id"] != expectedCurrencyID.String() {
		t.Errorf("expected currency_id in injected fields with value %s, got: %v", expectedCurrencyID, injected.InjectedFields)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
