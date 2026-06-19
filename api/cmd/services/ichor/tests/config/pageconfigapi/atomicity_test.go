package pageconfigapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
)

// Test_PageConfig_Import_Atomicity is the FF#2 §11 item-1 trip-wire for the non-atomic
// page-config structure import (a pre-existing gap PR #191 flagged and explicitly excluded).
//
// pageconfigbus.ImportPageConfigs -> createPageConfigWithRelations writes the page config
// (b.Create) then its contents (b.pageContentBus.Create) then its actions
// (b.pageActionBus.CreateButton/...) with NO surrounding transaction — each its own
// outbox.WriteAtomic begin+commit. A child that fails partway therefore leaves the config
// (and any earlier content) durably committed: an orphaned, half-built page definition.
//
// The import below is one config + one valid "text" content + one button whose `variant`
// is a well-formed string but NOT a member of the config.button_variant enum. That payload
// passes app validation (only mode + name are checked) and the bus button-behavior check
// (variant is never inspected), then fails at the Postgres enum on INSERT into
// config.page_action_buttons — AFTER the page config and the content are written.
//
//	RED  (pre-fix):  config + content autocommit on the pool -> config COUNT == 1 -> fails.
//	GREEN (post-fix): the writes ride one app-level tx and roll back -> COUNT == 0.
func Test_PageConfig_Import_Atomicity(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PageConfig_Import_Atomicity")

	sd, err := insertSeedData(test.DB, test.Auth)
	require.NoError(t, err, "seeding")

	ctx := context.Background()

	marker := fmt.Sprintf("PAGECONFIG-IMPORT-ATOMIC-%d", time.Now().UnixNano())

	pkg := pageconfigapp.ImportPackage{
		Mode: "skip", // config does not exist -> createPageConfigWithRelations runs
		Data: []pageconfigapp.PageConfigPackage{
			{
				// IsDefault:true forces user_id NULL, satisfying the check_default_no_user constraint.
				PageConfig: pageconfigapp.PageConfig{Name: marker, IsDefault: true},
				Contents: []pageconfigapp.PageContentApp{
					{
						ContentType: "text", // no FK required -> commits cleanly before the bad button
						Label:       "Valid Text Block",
						OrderIndex:  1,
						Layout:      "{}", // valid JSONB; empty string would be invalid json
						IsVisible:   true,
					},
				},
				Actions: pageconfigapp.PageActionsApp{
					Buttons: []pageconfigapp.PageActionApp{
						{
							ActionType:  "button",
							ActionOrder: 1,
							IsActive:    true,
							Button: &pageconfigapp.ButtonActionApp{
								Label:      "Boom",
								TargetPath: "/somewhere",         // default "navigate" behavior requires a target
								Variant:    "not-a-real-variant", // invalid enum value -> config.button_variant rejects at INSERT (22P02)
								Alignment:  "left",
							},
						},
					},
					Dropdowns:  []pageconfigapp.PageActionApp{},
					Separators: []pageconfigapp.PageActionApp{},
				},
			},
		},
	}

	body, err := json.Marshal(pkg)
	require.NoError(t, err, "marshal import package")

	var cascadeBefore int
	require.NoError(t, test.DB.DB.GetContext(ctx, &cascadeBefore,
		"SELECT COUNT(*) FROM workflow.cascade_outbox"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/config/page-configs/import", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
	r.Header.Set("Content-Type", "application/json")
	test.ServeHTTP(w, r)

	require.GreaterOrEqual(t, w.Code, http.StatusBadRequest,
		"import must fail on the button's invalid variant enum; body=%s", w.Body.String())

	// No orphaned config: the page_configs row must NOT be committed when a later child fails.
	var configCount int
	require.NoError(t, test.DB.DB.GetContext(ctx, &configCount,
		"SELECT COUNT(*) FROM config.page_configs WHERE name = $1", marker))
	require.Equal(t, 0, configCount,
		"atomicity violated: a config.page_configs row committed even though a later child write failed. "+
			"The import must wrap the config + all its contents and actions in one transaction.")

	// No partial contents: no page_content may reference a config carrying our name.
	var contentCount int
	require.NoError(t, test.DB.DB.GetContext(ctx, &contentCount, `
		SELECT COUNT(*) FROM config.page_content pc
		JOIN config.page_configs pg ON pg.id = pc.page_config_id
		WHERE pg.name = $1`, marker))
	require.Equal(t, 0, contentCount,
		"atomicity violated: a config.page_content row committed even though the import failed.")

	// No partial actions: no page_action may reference a config carrying our name.
	var actionCount int
	require.NoError(t, test.DB.DB.GetContext(ctx, &actionCount, `
		SELECT COUNT(*) FROM config.page_actions pa
		JOIN config.page_configs pg ON pg.id = pa.page_config_id
		WHERE pg.name = $1`, marker))
	require.Equal(t, 0, actionCount,
		"atomicity violated: a config.page_actions row committed even though the import failed.")

	// No partial cascade: every emit rode the same tx, so a rollback discards the outbox rows too.
	var cascadeAfter int
	require.NoError(t, test.DB.DB.GetContext(ctx, &cascadeAfter,
		"SELECT COUNT(*) FROM workflow.cascade_outbox"))
	require.Equal(t, cascadeBefore, cascadeAfter,
		"partial cascade: workflow.cascade_outbox grew even though the import failed and rolled back.")
}
