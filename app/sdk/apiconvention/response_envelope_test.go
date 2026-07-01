package apiconvention

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// grandfathered lists app-layer response types that already violate the
// unpaged-list envelope convention and predate this guard.
//
// docs/layer-patterns.md prescribes exactly two collection shapes:
//   - paginated  -> query.Result[T]  ({items, total, page, rows_per_page})
//   - unpaged    -> bare-array wrapper (type Entities []Entity)
//
// A struct that exposes json:"items" WITHOUT json:"total" is query.Result[T]
// with its pagination fields amputated and matches neither shape. The entries
// below are known deviations awaiting the frontend-driven reshape (they feed
// dynamic/config-driven UI, so changing them is cross-cutting and must be
// sequenced from the FE). Do NOT add to this list: a NEW violation must be
// fixed to a documented shape, not grandfathered.
var grandfathered = map[string]bool{
	"Pages":           true, // app/domain/core/pageapp
	"UserPreferences": true, // app/domain/core/userpreferencesapp
	"TableConfigList": true, // app/domain/dataapp
}

// Test_ResponseEnvelope_ItemsRequiresTotal enforces that any app-layer response
// type (one that implements Encode) exposing a json:"items" field also exposes
// json:"total". This blocks new totals-less {items} envelopes, which match no
// documented convention (see docs/layer-patterns.md).
func Test_ResponseEnvelope_ItemsRequiresTotal(t *testing.T) {
	found := detectItemsWithoutTotal(t)

	var offenders []string
	for name, pos := range found {
		if grandfathered[name] {
			continue
		}
		offenders = append(offenders, name+" ("+pos+")")
	}
	sort.Strings(offenders)

	if len(offenders) > 0 {
		t.Fatalf("response type(s) expose json:\"items\" without json:\"total\".\n"+
			"Use query.Result[T] if the endpoint is paginated, or a bare-array wrapper "+
			"(type Entities []Entity) if it is not — see docs/layer-patterns.md.\nOffenders:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}

// Test_Grandfathered_StillViolate keeps the allowlist honest: every grandfathered
// type must still be a detected violation. If one no longer is (e.g. it was
// reshaped to a documented convention), this fails so the stale entry gets
// removed. It also proves the detector actually fires.
func Test_Grandfathered_StillViolate(t *testing.T) {
	found := detectItemsWithoutTotal(t)

	for name := range grandfathered {
		if _, ok := found[name]; !ok {
			t.Errorf("grandfathered type %q is no longer a json:\"items\"-without-\"total\" "+
				"violation — remove it from the grandfathered allowlist", name)
		}
	}
}

// detectItemsWithoutTotal scans app/domain for types that implement Encode and
// expose json:"items" but not json:"total". Returns name -> source position.
func detectItemsWithoutTotal(t *testing.T) map[string]string {
	t.Helper()

	domainDir := filepath.Join(repoRoot(t), "app", "domain")
	out := map[string]string{}

	err := filepath.WalkDir(domainDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		fset := token.NewFileSet()
		pkgs, perr := parser.ParseDir(fset, path, func(fi fs.FileInfo) bool {
			return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if perr != nil {
			// Skip a directory we can't parse rather than failing the whole sweep.
			return nil
		}

		for _, pkg := range pkgs {
			encoders := collectEncoders(pkg)
			for name, st := range collectStructs(pkg) {
				if !encoders[name] {
					continue
				}
				if structHasJSONKey(st, "items") && !structHasJSONKey(st, "total") {
					out[name] = fset.Position(st.Pos()).String()
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking app/domain: %v", err)
	}

	return out
}

// collectEncoders returns the set of type names in pkg that have a method
// `Encode() ([]byte, string, error)` — i.e. web response types.
func collectEncoders(pkg *ast.Package) map[string]bool {
	enc := map[string]bool{}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "Encode" || fn.Recv == nil {
				continue
			}
			if !isEncodeSig(fn.Type) {
				continue
			}
			if name := receiverName(fn.Recv); name != "" {
				enc[name] = true
			}
		}
	}
	return enc
}

// collectStructs maps struct type names to their definition in pkg.
func collectStructs(pkg *ast.Package) map[string]*ast.StructType {
	out := map[string]*ast.StructType{}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if st, ok := ts.Type.(*ast.StructType); ok {
					out[ts.Name.Name] = st
				}
			}
		}
	}
	return out
}

// isEncodeSig reports whether ft matches `() ([]byte, string, error)`.
func isEncodeSig(ft *ast.FuncType) bool {
	if ft.Params != nil && ft.Params.NumFields() != 0 {
		return false
	}
	return ft.Results != nil && ft.Results.NumFields() == 3
}

// receiverName extracts the base type name from a method receiver, unwrapping
// pointers (*T) and generics (T[U]).
func receiverName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.IndexExpr:
			expr = t.X
		case *ast.IndexListExpr:
			expr = t.X
		case *ast.Ident:
			return t.Name
		default:
			return ""
		}
	}
}

// structHasJSONKey reports whether st has a field whose json tag name is key.
func structHasJSONKey(st *ast.StructType, key string) bool {
	for _, field := range st.Fields.List {
		if field.Tag == nil {
			continue
		}
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		jsonTag := tag.Get("json")
		if jsonTag == "" {
			continue
		}
		if strings.Split(jsonTag, ",")[0] == key {
			return true
		}
	}
	return false
}

// repoRoot walks up from the working directory until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate go.mod above %s", dir)
		}
		dir = parent
	}
}
