package commands

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
)

// ValidateForms validates all seed form configurations from the registry.
// Returns nil if all forms are valid, otherwise returns validation errors.
// This command does not require a database connection.
func ValidateForms() error {
	// Dummy UUIDs for form field generators
	dummyFormID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dummyEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	var (
		hasErrors    bool
		validCount   int
		invalidCount int
		warnCount    int
	)

	fmt.Println("Validating seed form configurations...")
	fmt.Println()

	for _, entry := range seedmodels.FormRegistry {
		opts := formfieldbus.FormValidationOptions{
			SupportsUpdate: entry.SupportsUpdate,
			FormName:       entry.Name,
		}

		fields := entry.Generator(dummyFormID, dummyEntityID)
		result := formfieldbus.ValidateFormFields(fields, opts)

		if result.HasErrors() {
			hasErrors = true
			invalidCount++
			fmt.Printf("❌ %s:\n", entry.Name)
			for _, err := range result.Errors {
				fmt.Printf("   • %s: %s (%s)\n", err.Field, err.Message, err.Code)
			}
		} else {
			validCount++
			fmt.Printf("✓ %s\n", entry.Name)
		}

		// Show warnings regardless of error status
		for _, warn := range result.Warnings {
			warnCount++
			fmt.Printf("   ⚠ %s: %s\n", warn.Field, warn.Message)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d valid, %d invalid, %d warnings\n", validCount, invalidCount, warnCount)

	if hasErrors {
		return fmt.Errorf("validation failed: %d form(s) have errors", invalidCount)
	}

	fmt.Println("\nAll form configurations valid!")
	return nil
}
