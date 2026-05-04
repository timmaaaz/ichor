// Integration tests for settingsbus, exercising the production storer chain
// (settingscache wrapping settingsdb). These tests live in settingsbus rather
// than settingscache because they require a live database container; the
// cache logic is unit-tested separately under stores/settingscache.
//
// Each scenario seeds its own keys (under a dedicated "test." prefix) so the
// suite does not depend on whatever the SQL seed pipeline puts in
// config.settings. That keeps these tests independent of the seedFrontend
// path.
package settingsbus_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_Settings(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Settings")

	unitest.Run(t, queryAndPrefix(db.BusDomain), "queryAndPrefix")
	unitest.Run(t, count(db.BusDomain), "count")
	unitest.Run(t, queryByKey(db.BusDomain), "queryByKey")
	unitest.Run(t, createReadback(db.BusDomain), "createReadback")
	unitest.Run(t, updateReadback(db.BusDomain), "updateReadback")
	unitest.Run(t, deleteReadback(db.BusDomain), "deleteReadback")
}

// =============================================================================
// scenarios — each runs in its own scenario-prefix sandbox
// =============================================================================

func queryAndPrefix(busDomain dbtest.BusDomain) []unitest.Table {
	const ns = "test_query."
	mustSeed(busDomain, ns,
		"a", "b", "c",
	)
	mustSeed(busDomain, "test_other.", "x", "y")

	return []unitest.Table{
		{
			Name:    "prefix-filter-only-returns-matching-keys",
			ExpResp: []string{ns + "a", ns + "b", ns + "c"},
			ExcFunc: func(ctx context.Context) any {
				prefix := ns
				rows, err := busDomain.Settings.Query(ctx,
					settingsbus.QueryFilter{Prefix: &prefix},
					order.NewBy(settingsbus.OrderByKey, order.ASC),
					page.MustParse("1", "100"))
				if err != nil {
					return err
				}
				out := make([]string, 0, len(rows))
				for _, r := range rows {
					out = append(out, r.Key)
				}
				return out
			},
			CmpFunc: cmpStringSlice,
		},
		{
			Name:    "different-prefixes-stay-separated",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				p1 := "test_query."
				p2 := "test_other."
				q, err := busDomain.Settings.Query(ctx,
					settingsbus.QueryFilter{Prefix: &p1},
					order.NewBy(settingsbus.OrderByKey, order.ASC),
					page.MustParse("1", "100"))
				if err != nil {
					return err
				}
				o, err := busDomain.Settings.Query(ctx,
					settingsbus.QueryFilter{Prefix: &p2},
					order.NewBy(settingsbus.OrderByKey, order.ASC),
					page.MustParse("1", "100"))
				if err != nil {
					return err
				}
				for _, r := range q {
					if !strings.HasPrefix(r.Key, p1) {
						return false
					}
				}
				for _, r := range o {
					if !strings.HasPrefix(r.Key, p2) {
						return false
					}
				}
				return true
			},
			CmpFunc: cmpAny,
		},
	}
}

func count(busDomain dbtest.BusDomain) []unitest.Table {
	const ns = "test_count."
	mustSeed(busDomain, ns, "a", "b", "c", "d")

	return []unitest.Table{
		{
			Name:    "count-matches-query-length",
			ExpResp: 4,
			ExcFunc: func(ctx context.Context) any {
				prefix := ns
				n, err := busDomain.Settings.Count(ctx, settingsbus.QueryFilter{Prefix: &prefix})
				if err != nil {
					return err
				}
				return n
			},
			CmpFunc: cmpAny,
		},
	}
}

func queryByKey(busDomain dbtest.BusDomain) []unitest.Table {
	const k = "test_querybykey.solo"
	mustSeed(busDomain, "test_querybykey.", "solo")

	return []unitest.Table{
		{
			Name:    "round-trip-exact-key",
			ExpResp: k,
			ExcFunc: func(ctx context.Context) any {
				s, err := busDomain.Settings.QueryByKey(ctx, k)
				if err != nil {
					return err
				}
				return s.Key
			},
			CmpFunc: cmpAny,
		},
	}
}

func createReadback(busDomain dbtest.BusDomain) []unitest.Table {
	const k = "test_create.readback"
	return []unitest.Table{
		{
			Name:    "create-then-query-shows-row",
			ExpResp: k,
			ExcFunc: func(ctx context.Context) any {
				ns := settingsbus.NewSetting{
					Key:         k,
					Value:       json.RawMessage(`"hello"`),
					Description: "create_readback",
				}
				if _, err := busDomain.Settings.Create(ctx, ns); err != nil {
					return err
				}
				prefix := "test_create."
				rows, err := busDomain.Settings.Query(ctx,
					settingsbus.QueryFilter{Prefix: &prefix},
					order.NewBy(settingsbus.OrderByKey, order.ASC),
					page.MustParse("1", "100"))
				if err != nil {
					return err
				}
				for _, r := range rows {
					if r.Key == k {
						return r.Key
					}
				}
				return "missing key after Create"
			},
			CmpFunc: cmpAny,
		},
	}
}

func updateReadback(busDomain dbtest.BusDomain) []unitest.Table {
	const k = "test_update.readback"
	return []unitest.Table{
		{
			Name:    "update-invalidates-cache",
			ExpResp: `"after"`,
			ExcFunc: func(ctx context.Context) any {
				ns := settingsbus.NewSetting{Key: k, Value: json.RawMessage(`"before"`), Description: "x"}
				created, err := busDomain.Settings.Create(ctx, ns)
				if err != nil {
					return err
				}
				// Prime cache.
				if _, err := busDomain.Settings.QueryByKey(ctx, k); err != nil {
					return err
				}
				upd := settingsbus.UpdateSetting{Value: json.RawMessage(`"after"`)}
				if _, err := busDomain.Settings.Update(ctx, created, upd); err != nil {
					return err
				}
				got, err := busDomain.Settings.QueryByKey(ctx, k)
				if err != nil {
					return err
				}
				return string(got.Value)
			},
			CmpFunc: cmpAny,
		},
	}
}

func deleteReadback(busDomain dbtest.BusDomain) []unitest.Table {
	const k = "test_delete.readback"
	return []unitest.Table{
		{
			Name:    "delete-invalidates-cache",
			ExpResp: settingsbus.ErrNotFound,
			ExcFunc: func(ctx context.Context) any {
				ns := settingsbus.NewSetting{Key: k, Value: json.RawMessage(`"x"`), Description: "x"}
				created, err := busDomain.Settings.Create(ctx, ns)
				if err != nil {
					return err
				}
				if _, err := busDomain.Settings.QueryByKey(ctx, k); err != nil {
					return err
				}
				if err := busDomain.Settings.Delete(ctx, created); err != nil {
					return err
				}
				_, err = busDomain.Settings.QueryByKey(ctx, k)
				if err == nil {
					return "expected ErrNotFound after delete, got nil"
				}
				return err
			},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(error)
				if !ok {
					return fmt.Sprintf("expected error, got %#v", got)
				}
				expErr := exp.(error)
				if !errors.Is(gotErr, expErr) {
					return fmt.Sprintf("error chain missing %v: got %v", expErr, gotErr)
				}
				return ""
			},
		},
	}
}

// =============================================================================
// helpers
// =============================================================================

// mustSeed creates settings with the given prefix and suffixes. Panics on
// error to keep table-driven scenario setup terse.
func mustSeed(busDomain dbtest.BusDomain, prefix string, suffixes ...string) {
	ctx := context.Background()
	for _, sfx := range suffixes {
		ns := settingsbus.NewSetting{
			Key:         prefix + sfx,
			Value:       json.RawMessage(fmt.Sprintf(`"%s"`, sfx)),
			Description: "settingsbus_test seed",
		}
		if _, err := busDomain.Settings.Create(ctx, ns); err != nil {
			panic(fmt.Sprintf("seed %s: %v", ns.Key, err))
		}
	}
}

func cmpAny(got any, exp any) string {
	return cmp.Diff(got, exp)
}

func cmpStringSlice(got any, exp any) string {
	g, ok := got.([]string)
	if !ok {
		return fmt.Sprintf("expected []string, got %#v", got)
	}
	e := exp.([]string)
	return cmp.Diff(g, e)
}
