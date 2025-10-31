package formfieldbus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_FormField(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_FormField")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, queryByID(db.BusDomain, sd), "queryByID")
	unitest.Run(t, queryByFormID(db.BusDomain, sd), "queryByFormID")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	// Seed forms first since form fields depend on them
	forms, err := formbus.TestSeedForms(ctx, 5, busDomain.Form)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding forms : %w", err)
	}

	// Extract form IDs
	formIDs := make([]uuid.UUID, len(forms))
	for i, form := range forms {
		formIDs[i] = form.ID
	}

	// Get entities
	entities, err := busDomain.Workflow.QueryEntities(ctx)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("querying entities : %w", err)
	}

	// Seed form fields
	formFields, err := formfieldbus.TestSeedFormFields(ctx, 20, formIDs, busDomain.FormField, entities)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding form fields : %w", err)
	}

	return unitest.SeedData{
		Forms:      forms,
		FormFields: formFields,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []formfieldbus.FormField{
				sd.FormFields[0],
				sd.FormFields[1],
				sd.FormFields[2],
				sd.FormFields[3],
				sd.FormFields[4],
			},
			ExcFunc: func(ctx context.Context) any {
				formFields, err := busDomain.FormField.Query(ctx, formfieldbus.QueryFilter{}, order.NewBy(formfieldbus.OrderByName, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return formFields
			},
			CmpFunc: func(got any, exp any) string {
				dbtest.NormalizeJSONFields(got, exp)
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "queryByID",
			ExpResp: sd.FormFields[0],
			ExcFunc: func(ctx context.Context) any {
				formField, err := busDomain.FormField.QueryByID(ctx, sd.FormFields[0].ID)
				if err != nil {
					return err
				}
				return formField
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(formfieldbus.FormField)
				expResp := exp.(formfieldbus.FormField)
				dbtest.NormalizeJSONFields(&gotResp, &expResp)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func queryByFormID(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	// Find all form fields that belong to the first form
	var expectedFields []formfieldbus.FormField
	for _, ff := range sd.FormFields {
		if ff.FormID == sd.Forms[0].ID {
			expectedFields = append(expectedFields, ff)
		}
	}

	table := []unitest.Table{
		{
			Name:    "queryByFormID",
			ExpResp: expectedFields,
			ExcFunc: func(ctx context.Context) any {
				formFields, err := busDomain.FormField.QueryByFormID(ctx, sd.Forms[0].ID)
				if err != nil {
					return err
				}
				return formFields
			},
			CmpFunc: func(got any, exp any) string {
				dbtest.NormalizeJSONFields(got, exp)
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	config := map[string]any{
		"placeholder": "Test placeholder",
		"maxLength":   50,
	}
	configJSON, _ := json.Marshal(config)

	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: formfieldbus.FormField{
				FormID:     sd.Forms[0].ID,
				EntityID:   sd.FormFields[0].EntityID,
				Name:       "test_field",
				Label:      "Test Field",
				FieldType:  "text",
				FieldOrder: 999,
				Required:   true,
				Config:     json.RawMessage(configJSON),
			},
			ExcFunc: func(ctx context.Context) any {
				formField, err := busDomain.FormField.Create(ctx, formfieldbus.NewFormField{
					FormID:     sd.Forms[0].ID,
					EntityID:   sd.FormFields[0].EntityID,
					Name:       "test_field",
					Label:      "Test Field",
					FieldType:  "text",
					FieldOrder: 999,
					Required:   true,
					Config:     json.RawMessage(configJSON),
				})
				if err != nil {
					return err
				}
				return formField
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(formfieldbus.FormField)
				if !exists {
					return fmt.Sprintf("got is not a form field %v", got)
				}

				expResp := exp.(formfieldbus.FormField)
				expResp.ID = gotResp.ID

				dbtest.NormalizeJSONFields(&gotResp, &expResp)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	newConfig := map[string]any{
		"placeholder": "Updated placeholder",
		"maxLength":   200,
	}
	newConfigJSON, _ := json.Marshal(newConfig)
	newConfigRaw := json.RawMessage(newConfigJSON)

	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: formfieldbus.FormField{
				ID:         sd.FormFields[0].ID,
				EntityID:   sd.FormFields[0].EntityID,
				FormID:     sd.FormFields[0].FormID,
				Name:       sd.FormFields[0].Name,
				Label:      "Updated Label",
				FieldType:  sd.FormFields[0].FieldType,
				FieldOrder: 100,
				Required:   false,
				Config:     newConfigRaw,
			},
			ExcFunc: func(ctx context.Context) any {
				formField, err := busDomain.FormField.Update(ctx, sd.FormFields[0], formfieldbus.UpdateFormField{
					Label:      dbtest.StringPointer("Updated Label"),
					FieldOrder: dbtest.IntPointer(100),
					Required:   dbtest.BoolPointer(false),
					Config:     &newConfigRaw,
				})
				if err != nil {
					return err
				}
				return formField
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(formfieldbus.FormField)
				if !exists {
					return fmt.Sprintf("got is not a form field %v", got)
				}
				expResp := exp.(formfieldbus.FormField)
				dbtest.NormalizeJSONFields(&gotResp, &expResp)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.FormField.Delete(ctx, sd.FormFields[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
