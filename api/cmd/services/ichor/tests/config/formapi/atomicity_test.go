package formapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/formapp"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
)

// Test_Form_Import_Atomicity is the FF#2 §11 item-1 trip-wire for the non-atomic
// form-structure import (a pre-existing gap PR #191 flagged and explicitly excluded).
//
// formbus.ImportForms -> createFormWithFields calls b.Create(form) then loops
// b.formFieldBus.Create(field) with NO surrounding transaction, so each write is its own
// outbox.WriteAtomic begin+commit. A field that fails partway therefore leaves the form
// (and any earlier field) durably committed — an orphaned, half-built form definition that
// a workflow triggered by the form-create could observe.
//
// The import below is one form + two fields; the first field is fully valid, the second
// carries a well-formed but non-existent entity_id (FK -> workflow.entities). Both fields
// reuse a real entity's schema/table so they pass the storer's table-exists check; the
// payload therefore passes app validation (entity_id parses as a UUID) and fails only on
// the real Postgres FK (error 23503) AFTER the form and the first field are written.
//
//	RED  (pre-fix):  form + field #1 autocommit on the pool -> form COUNT == 1 -> fails.
//	GREEN (post-fix): the writes ride one app-level tx and roll back -> COUNT == 0.
func Test_Form_Import_Atomicity(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Form_Import_Atomicity")

	sd, err := insertSeedData(test.DB, test.Auth)
	require.NoError(t, err, "seeding")

	ctx := context.Background()

	// A real entity gives a valid (entity_id, schema, table) triple for the good field;
	// reusing its schema/table on the bad field passes the storer's pg_catalog table-exists
	// check so the ONLY defect on field #2 is the non-existent entity_id FK.
	entities, err := test.DB.BusDomain.Workflow.QueryEntities(ctx)
	require.NoError(t, err, "querying entities")
	require.NotEmpty(t, entities, "need at least one workflow entity for a valid field")
	ent := entities[0]

	formName := fmt.Sprintf("FORM-IMPORT-ATOMIC-%d", time.Now().UnixNano())

	pkg := formapp.ImportPackage{
		Mode: "skip", // form does not exist -> createFormWithFields runs
		Data: []formapp.FormPackage{
			{
				Form: formapp.Form{Name: formName},
				Fields: []formfieldapp.FormField{
					{
						EntityID:     ent.ID.String(),
						EntitySchema: ent.SchemaName,
						EntityTable:  ent.Name,
						Name:         "valid_field",
						Label:        "Valid Field",
						FieldType:    "text",
						FieldOrder:   1,
						Config:       json.RawMessage(`{}`),
					},
					{
						EntityID:     uuid.New().String(), // non-existent -> FK violation
						EntitySchema: ent.SchemaName,       // real table -> passes table-exists check
						EntityTable:  ent.Name,
						Name:         "bad_fk_field",
						Label:        "Bad FK Field",
						FieldType:    "text",
						FieldOrder:   2,
						Config:       json.RawMessage(`{}`),
					},
				},
			},
		},
	}

	body, err := json.Marshal(pkg)
	require.NoError(t, err, "marshal import package")

	// Snapshot the cascade outbox so we can assert the failed import emitted nothing durable.
	var cascadeBefore int
	require.NoError(t, test.DB.DB.GetContext(ctx, &cascadeBefore,
		"SELECT COUNT(*) FROM workflow.cascade_outbox"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/config/forms/import", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
	r.Header.Set("Content-Type", "application/json")
	test.ServeHTTP(w, r)

	require.GreaterOrEqual(t, w.Code, http.StatusBadRequest,
		"import must fail on the second field's FK violation; body=%s", w.Body.String())

	// No orphaned form: the form row must NOT be committed when a later field fails.
	var formCount int
	require.NoError(t, test.DB.DB.GetContext(ctx, &formCount,
		"SELECT COUNT(*) FROM config.forms WHERE name = $1", formName))
	require.Equal(t, 0, formCount,
		"atomicity violated: a config.forms row committed even though a later field write failed. "+
			"The import must wrap the form + all its fields in one transaction.")

	// No partial fields: no form_field may reference a form carrying our name.
	var fieldCount int
	require.NoError(t, test.DB.DB.GetContext(ctx, &fieldCount, `
		SELECT COUNT(*) FROM config.form_fields ff
		JOIN config.forms f ON ff.form_id = f.id
		WHERE f.name = $1`, formName))
	require.Equal(t, 0, fieldCount,
		"atomicity violated: a config.form_fields row committed even though the import failed.")

	// No partial cascade: every emit rode the same tx, so a rollback discards the outbox rows too.
	var cascadeAfter int
	require.NoError(t, test.DB.DB.GetContext(ctx, &cascadeAfter,
		"SELECT COUNT(*) FROM workflow.cascade_outbox"))
	require.Equal(t, cascadeBefore, cascadeAfter,
		"partial cascade: workflow.cascade_outbox grew even though the import failed and rolled back.")
}
