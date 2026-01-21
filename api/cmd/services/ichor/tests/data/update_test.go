package data_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func update200(sd apitest.SeedData) []apitest.Table {

	// Create a deep copy of SimpleConfig to avoid modifying the shared config
	updatedSimple := *SimpleConfig

	// Deep copy the VisualSettings.Columns map
	updatedSimple.VisualSettings.Columns = make(map[string]tablebuilder.ColumnConfig)
	for k, v := range SimpleConfig.VisualSettings.Columns {
		updatedSimple.VisualSettings.Columns[k] = v
	}

	// Deep copy the DataSource slice and its nested Columns slice
	updatedSimple.DataSource = make([]tablebuilder.DataSource, len(SimpleConfig.DataSource))
	copy(updatedSimple.DataSource, SimpleConfig.DataSource)
	updatedSimple.DataSource[0].Select.Columns = make([]tablebuilder.ColumnDefinition, len(SimpleConfig.DataSource[0].Select.Columns))
	copy(updatedSimple.DataSource[0].Select.Columns, SimpleConfig.DataSource[0].Select.Columns)

	// Add new columns to DataSource
	updatedSimple.DataSource[0].Select.Columns = append(updatedSimple.DataSource[0].Select.Columns,
		tablebuilder.ColumnDefinition{Name: "minimum_stock", TableColumn: "inventory_items.minimum_stock"},
		tablebuilder.ColumnDefinition{Name: "maximum_stock", TableColumn: "inventory_items.maximum_stock"},
	)

	// Add corresponding VisualSettings entries (required by validation)
	// Use TableColumn as the key since these columns don't have aliases
	updatedSimple.VisualSettings.Columns["inventory_items.minimum_stock"] = tablebuilder.ColumnConfig{
		Name:   "inventory_items.minimum_stock",
		Header: "Minimum Stock",
		Width:  120,
		Type:   "number",
	}
	updatedSimple.VisualSettings.Columns["inventory_items.maximum_stock"] = tablebuilder.ColumnConfig{
		Name:   "inventory_items.maximum_stock",
		Header: "Maximum Stock",
		Width:  120,
		Type:   "number",
	}

	jsonConfig, err := json.Marshal(updatedSimple)
	if err != nil {
		panic(err)
	}
	var raw *json.RawMessage
	tmp := json.RawMessage(jsonConfig)
	raw = &tmp

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/data/%s", sd.SimpleTableConfig.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: dataapp.UpdateTableConfig{
				Config: raw,
			},
			GotResp: &dataapp.TableConfig{},
			ExpResp: &dataapp.TableConfig{
				Name:        sd.SimpleTableConfig.Name,
				Description: sd.SimpleTableConfig.Description,
				Config:      json.RawMessage(jsonConfig),
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*dataapp.TableConfig)
				if !exists {
					return "could not convert got to *dataapp.TableConfig"
				}
				expResp, exists := exp.(*dataapp.TableConfig)
				if !exists {
					return "could not convert exp to *dataapp.TableConfig"
				}

				// Copy over the generated fields from got to exp
				expResp.ID = gotResp.ID
				expResp.CreatedBy = gotResp.CreatedBy
				expResp.UpdatedBy = gotResp.UpdatedBy
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				dbtest.NormalizeJSONFields(gotResp, &expResp)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}
